package main

import (
	"path/filepath"

	"google.golang.org/protobuf/compiler/protogen"
)

// generateHub generates related hub files.
func generateHub(gen *protogen.Plugin) {
	filename := filepath.Join("hub." + pcExt + ".go")
	g := gen.NewGeneratedFile(filename, "")
	generateCommonHeader(gen, g)
	g.P()
	g.P("package ", *pkg)
	g.P()
	g.P(staticHubContent)
	g.P()

	for _, messager := range messagers {
		g.P("func (h *Hub) Get", messager, "() *", messager, " {")
		g.P(`msger := h.GetMessager("`, messager, `")`)
		g.P("if msger != nil {")
		g.P("if conf, ok := msger.(*", messager, "); ok {")
		g.P("return conf")
		g.P("}")
		g.P("}")
		g.P("return nil")
		g.P("}")
		g.P()
	}
}

const staticHubContent = `import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
	"github.com/tableauio/tableau/store"
	"google.golang.org/protobuf/proto"
)

type Messager interface {
	Checker
	immutableChecker
	// Name returns the unique message name.
	Name() string
	// GetStats returns stats info.
	GetStats() *Stats
	// Load fills message from file in the specified directory and format.
	Load(dir string, fmt format.Format, options ...load.Option) error
	// Store writes message to file in the specified directory and format.
	Store(dir string, fmt format.Format, options ...store.Option) error
	// processAfterLoad is invoked after this messager loaded.
	processAfterLoad() error
	// ProcessAfterLoadAll is invoked after all messagers loaded.
	ProcessAfterLoadAll(hub *Hub) error
	// Message returns the inner message data.
	Message() proto.Message
}

type Checker interface {
	Messager() Messager
	Check(hub *Hub) error
	CheckCompatibility(hub, newHub *Hub) error
}

type immutableChecker interface {
	enableBackup()
	originalMessage() proto.Message
	mutable() bool
}

type Options struct {
	// Filter can only filter in certain specific messagers based on the
	// condition that you provide.
	//
	// Default: nil.
	Filter FilterFunc

	// ImmutableCheck enables the immutable check of the loaded config,
	// and specifies its interval and mutable handler.
	//
	// Default: nil.
	ImmutableCheck *ImmutableCheck
}

// FilterFunc filter in messagers if returned value is true.
//
// NOTE: name is the protobuf message name, e.g.: "message ItemConf{...}".
type FilterFunc func(name string) bool

type ImmutableCheck struct {
	// Interval is the gap duration between two checks.
	// Default: 60s.
	Interval time.Duration
	// MutableHandler is called when encouters modifications, with messager's name,
	// original message and current message.
	MutableHandler func(name string, original, current proto.Message)
}

// Option is the functional option type.
type Option func(*Options)

// newDefault returns a default Options.
func newDefault() *Options {
	return &Options{}
}

// ParseOptions parses functional options and merge them to default Options.
func ParseOptions(setters ...Option) *Options {
	// Default Options
	opts := newDefault()
	for _, setter := range setters {
		setter(opts)
	}
	return opts
}

// Filter can only filter in certain specific messagers based on the
// condition that you provide.
func Filter(filter FilterFunc) Option {
	return func(opts *Options) {
		opts.Filter = filter
	}
}

// WithImmutableCheck enables the immutable check with given params.
func WithImmutableCheck(check *ImmutableCheck) Option {
	return func(opts *Options) {
		opts.ImmutableCheck = check
	}
}

type Stats struct {
	Duration time.Duration // total load time consuming.
}

type UnimplementedMessager struct {
	Stats Stats
	backup bool
}

func (x *UnimplementedMessager) Name() string {
	return ""
}

func (x *UnimplementedMessager) GetStats() *Stats {
	return &x.Stats
}

func (x *UnimplementedMessager) Load(dir string, format format.Format, options ...load.Option) error {
	return nil
}

func (x *UnimplementedMessager) Store(dir string, format format.Format, options ...store.Option) error {
	return nil
}

func (x *UnimplementedMessager) processAfterLoad() error {
	return nil
}

func (x *UnimplementedMessager) ProcessAfterLoadAll(hub *Hub) error {
	return nil
}

func (x *UnimplementedMessager) Message() proto.Message {
	return nil
}

func (x *UnimplementedMessager) Messager() Messager {
	return nil
}

func (x *UnimplementedMessager) Check(hub *Hub) error {
	return nil
}

func (x *UnimplementedMessager) CheckCompatibility(hub, newHub *Hub) error {
	return nil
}

func (x *UnimplementedMessager) enableBackup() {
	x.backup = true
}

func (x *UnimplementedMessager) originalMessage() proto.Message {
	return nil
}

func (x *UnimplementedMessager) mutable() bool {
	return false
}

type MessagerMap = map[string]Messager
type MessagerGenerator = func() Messager
type Registrar struct {
	Generators map[string]MessagerGenerator
}

func NewRegistrar() *Registrar {
	return &Registrar{
		Generators: map[string]MessagerGenerator{},
	}
}

func (r *Registrar) Register(gen MessagerGenerator) {
	if _, ok := r.Generators[gen().Name()]; ok {
		panic("register duplicate messager: " + gen().Name())
	}
	r.Generators[gen().Name()] = gen
}

var registrarSingleton *Registrar
var once sync.Once

func getRegistrar() *Registrar {
	once.Do(func() {
		registrarSingleton = NewRegistrar()
	})
	return registrarSingleton
}

func Register(gen MessagerGenerator) {
	getRegistrar().Register(gen)
}

func BoolToInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

// Hub is the messager manager.
type Hub struct {
	messagerMap    atomic.Pointer[MessagerMap]
	lastLoadedTime atomic.Pointer[time.Time]
	opts           *Options
}

func NewHub(options ...Option) *Hub {
	hub := &Hub{}
	hub.messagerMap.Store(&MessagerMap{})
	hub.lastLoadedTime.Store(&time.Time{})
	hub.opts = ParseOptions(options...)
	go hub.immutableCheck()
	return hub
}

// NewMessagerMap creates a new MessagerMap.
func (h *Hub) NewMessagerMap() MessagerMap {
	messagerMap := MessagerMap{}
	for name, gen := range getRegistrar().Generators {
		if h.opts.Filter == nil || h.opts.Filter(name) {
			messager := gen()
			if h.opts.ImmutableCheck != nil {
				messager.enableBackup()
			}
			messagerMap[name] = messager
		}
	}
	return messagerMap
}

// GetMessagerMap returns hub's inner field messagerMap.
func (h *Hub) GetMessagerMap() MessagerMap {
	return *h.messagerMap.Load()
}

// SetMessagerMap sets hub's inner field messagerMap.
func (h *Hub) SetMessagerMap(messagerMap MessagerMap) {
	h.messagerMap.Store(&messagerMap)
	now := time.Now()
	h.lastLoadedTime.Store(&now)
}

// GetMessager finds and returns the specified Messenger in hub.
func (h *Hub) GetMessager(name string) Messager {
	return h.GetMessagerMap()[name]
}

// Load fills messages from files in the specified directory and format.
func (h *Hub) Load(dir string, format format.Format, options ...load.Option) error {
	messagerMap := h.NewMessagerMap()
	for name, msger := range messagerMap {
		if err := msger.Load(dir, format, options...); err != nil {
			return errors.WithMessagef(err, "failed to load: %v", name)
		}
	}
	// create a temporary hub with messager container for post process
	tmpHub := &Hub{}
	tmpHub.SetMessagerMap(messagerMap)
	for name, msger := range messagerMap {
		if err := msger.ProcessAfterLoadAll(tmpHub); err != nil {
			return errors.WithMessagef(err, "failed to process messager %s after load all", name)
		}
	}
	h.SetMessagerMap(messagerMap)
	return nil
}

// Store stores protobuf messages to files in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (h *Hub) Store(dir string, format format.Format, options ...store.Option) error {
	opts := store.ParseOptions(options...)
	for name, msger := range h.GetMessagerMap() {
		if opts.Filter == nil || opts.Filter(name) {
			if err := msger.Store(dir, format, options...); err != nil {
				return errors.WithMessagef(err, "failed to store: %v", name)
			}
		}
	}
	return nil
}

// immutableCheck checks if the messagers are mutable or not.
func (h *Hub) immutableCheck() {
	if h.opts.ImmutableCheck == nil {
		return
	}
	interval := h.opts.ImmutableCheck.Interval
	if interval == 0 {
		interval = time.Minute
	}
	handler := h.opts.ImmutableCheck.MutableHandler
	if handler == nil {
		handler = h.defaultMutableHandler
	}
	for {
		time.Sleep(interval)
		messagerMap := h.GetMessagerMap()
		for name, msger := range messagerMap {
			time.Sleep(time.Second)
			if msger.mutable() {
				handler(name, msger.originalMessage(), msger.Message())
			}
		}
	}
}

func (h *Hub) defaultMutableHandler(name string, original, current proto.Message) {
	originalText, _ := store.MarshalToText(original, true)
	currentText, _ := store.MarshalToText(current, true)
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(originalText)),
		B:        difflib.SplitLines(string(currentText)),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	fmt.Fprintf(os.Stderr,
		"==== %s DIFF BEGIN ====\n%s==== %s DIFF END ====\n",
		name, text, name)
}

// GetLastLoadedTime returns the time when hub's messagerMap was last set.
func (h *Hub) GetLastLoadedTime() time.Time {
	return *h.lastLoadedTime.Load()
}

// Auto-generated getters below`
