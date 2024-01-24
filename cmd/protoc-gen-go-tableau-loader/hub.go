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

	"github.com/pkg/errors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

type Messager interface {
	Checker
	// Name returns the unique message name.
	Name() string
	// Load fills inner message data from the specified direcotry and format.
	Load(dir string, fmt format.Format, options ...load.Option) error
	// ProcessAfterLoadAll is invoked after all messagers loaded.
	ProcessAfterLoadAll(hub *Hub) error
}

type Checker interface {
	Messager() Messager
	Check(hub *Hub) error
	CheckCompatibility(hub, newHub *Hub) error
}

type UnimplementedMessager struct {
}

func (x *UnimplementedMessager) Name() string {
	return ""
}

func (x *UnimplementedMessager) Load(dir string, format format.Format, options ...load.Option) error {
	return nil
}

func (x *UnimplementedMessager) ProcessAfterLoadAll(hub *Hub) error {
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
	if _, ok:= r.Generators[gen().Name()]; ok{
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

type Filter interface {
	Filter(name string) bool
}

func BoolToInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

// Hub is the messager manager.
type Hub struct {
	messagerMap MessagerMap
}

func NewHub() *Hub {
	return &Hub{
		messagerMap: MessagerMap{},
	}
}

func (h *Hub) NewMessagerMap(filter Filter) MessagerMap {
	messagerMap := MessagerMap{}
	for name, gen := range getRegistrar().Generators {
		if filter == nil || filter.Filter(name) {
			messagerMap[name] = gen()
		}
	}
	return messagerMap
}

func (h *Hub) SetMessagerMap(messagerMap MessagerMap) {
	h.messagerMap = messagerMap
}

func (h *Hub) GetMessager(name string) Messager {
	return h.messagerMap[name]
}

func (h *Hub) Load(dir string, filter Filter, format format.Format, options ...load.Option) error {
	messagerMap := h.NewMessagerMap(filter)
	for name, msger := range messagerMap {
		if err := msger.Load(dir, format, options...); err != nil {
			return errors.WithMessagef(err, "failed to load: %v", name)
		}
		fmt.Println("Loaded: " + msger.Name())
	}
	// create a temporary hub with messager container for post process
  	tmpHub := &Hub{messagerMap: messagerMap};
	for name, msger := range messagerMap {
		if err := msger.ProcessAfterLoadAll(tmpHub); err != nil {
			return errors.WithMessagef(err, "failed to process messager %s after load all", name)
		}
	}
	h.SetMessagerMap(messagerMap)
	return nil
}

// Auto-generated getters below`
