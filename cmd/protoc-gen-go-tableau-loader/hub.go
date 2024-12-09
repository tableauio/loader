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
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
	"github.com/tableauio/tableau/store"
	"google.golang.org/protobuf/proto"
)

type Messager interface {
	Checker
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

type Stats struct {
	Duration time.Duration // total load time consuming.
	// TODO: crc32 of config file to decide whether changed or not
	// CRC32 string
	// LastModifiedTime time.Time
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
}

func NewHub() *Hub {
	hub := &Hub{}
	hub.messagerMap.Store(&MessagerMap{})
	hub.lastLoadedTime.Store(&time.Time{})
	return hub
}

// NewMessagerMap creates a new MessagerMap.
func (h *Hub) NewMessagerMap(filter load.FilterFunc) MessagerMap {
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
	opts := load.ParseOptions(options...)
	messagerMap := h.NewMessagerMap(opts.Filter)
	for name, msger := range messagerMap {
		if err := msger.Load(dir, format, options...); err != nil {
			return errors.WithMessagef(err, "failed to load: %v", name)
		}
		fmt.Println("Loaded: " + msger.Name())
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
			fmt.Println("Stored: " + msger.Name())
		}
	}
	return nil
}

// GetLastLoadedTime returns the time when hub's messagerMap was last set.
func (h *Hub) GetLastLoadedTime() time.Time {
	return *h.lastLoadedTime.Load()
}

// Auto-generated getters below`
