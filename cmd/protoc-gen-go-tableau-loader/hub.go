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
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
	"github.com/tableauio/tableau/store"
	"google.golang.org/protobuf/proto"
)

var marshalOpts = proto.MarshalOptions{Deterministic: true}

type Messager interface {
	Checker
	immutabilityChecker
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

type immutabilityChecker interface {
	modified() bool
}

type Options struct {
	// Filter can only filter in certain specific messagers based on the
	// condition that you provide.
	//
	// Default: nil.
	Filter FilterFunc

	// ImmutabilityCheck enables the immutability check of the loaded config,
	// and specifies its interval and error handler.
	//
	// Default: nil.
	ImmutabilityCheck *ImmutabilityCheck
}

// FilterFunc filter in messagers if returned value is true.
//
// NOTE: name is the protobuf message name, e.g.: "message ItemConf{...}".
type FilterFunc func(name string) bool

type ImmutabilityCheck struct {
	Interval     time.Duration
	ErrorHandler func(error)
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
//
// NOTE: only used in https://github.com/tableauio/loader.
func Filter(filter FilterFunc) Option {
	return func(opts *Options) {
		opts.Filter = filter
	}
}

// CheckImmutability specifies the interval and error handler of
// immutability check.
func CheckImmutability(interval time.Duration, handler func(error)) Option {
	return func(opts *Options) {
		if interval != 0 && handler != nil {
			opts.ImmutabilityCheck = &ImmutabilityCheck{
				Interval:     interval,
				ErrorHandler: handler,
			}
		}
	}
}

type Stats struct {
	Duration time.Duration // total load time consuming.
	md5      string
}

type UnimplementedMessager struct {
	Stats Stats
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

// calcMd5 calculates the md5 of the messager's inner message.
func (x *UnimplementedMessager) calcMd5() (string, error) {
	bytes, err := marshalOpts.Marshal(x.Message())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(bytes)), nil
}

// modified returns true if the messager's inner message is modified.
func (x *UnimplementedMessager) modified() bool {
	md5, err := x.calcMd5()
	if err != nil || md5 != x.Stats.md5 {
		return true
	}
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
	cancel         context.CancelFunc
}

func NewHub(options ...Option) *Hub {
	hub := &Hub{}
	hub.messagerMap.Store(&MessagerMap{})
	hub.lastLoadedTime.Store(&time.Time{})
	hub.opts = ParseOptions(options...)
	return hub
}

// NewMessagerMap creates a new MessagerMap.
func (h *Hub) NewMessagerMap(filter FilterFunc) MessagerMap {
	messagerMap := MessagerMap{}
	for name, gen := range getRegistrar().Generators {
		if filter == nil || filter(name) {
			messagerMap[name] = gen()
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
	if h.cancel != nil {
		h.cancel()
	}
	messagerMap := h.NewMessagerMap(h.opts.Filter)
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
	if h.opts.ImmutabilityCheck != nil {
		ctx, cancel := context.WithCancel(context.Background())
		h.cancel = cancel
		go h.checkImmutability(ctx, messagerMap)
	}
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

// checkImmutability checks if the messagers are modified or not.
func (h *Hub) checkImmutability(ctx context.Context, messagerMap MessagerMap) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(h.opts.ImmutabilityCheck.Interval)
			for name, msger := range messagerMap {
				if msger.modified() {
					h.opts.ImmutabilityCheck.ErrorHandler(errors.Errorf("msger modified: %v", name))
				}
			}
		}
	}
}

// GetLastLoadedTime returns the time when hub's messagerMap was last set.
func (h *Hub) GetLastLoadedTime() time.Time {
	return *h.lastLoadedTime.Load()
}

// Auto-generated getters below`
