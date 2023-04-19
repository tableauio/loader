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
	Name() string
	Load(dir string, fmt format.Format, options ...load.Option) error
}

type Checker interface {
	Messager() Messager
	Check(hub *Hub) error
}

type MessagerMap = map[string]Messager
type MessagerGenerator = func() Messager
type Registrar struct {
	generators map[string]MessagerGenerator
}

func (r *Registrar) register(name string, gen MessagerGenerator) {
	r.generators[name] = gen
}

var registrarSingleton *Registrar
var once sync.Once

func getRegistrar() *Registrar {
	once.Do(func() {
		registrarSingleton = &Registrar{
			generators: map[string]MessagerGenerator{},
		}
	})
	return registrarSingleton
}

func register(name string, gen MessagerGenerator) {
	getRegistrar().register(name, gen)
}

type Filter interface {
	Filter(name string) bool
}

// Hub is the holder for managing configurations.
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
	for name, gen := range getRegistrar().generators {
		if filter == nil || filter.Filter(name) {
			messagerMap[name] = gen()
		}
	}
	return messagerMap
}

func (h *Hub) SetMessagerMap(messagerMap MessagerMap) {
	h.messagerMap = messagerMap
}

func (h *Hub) Load(dir string, filter Filter, format format.Format, options ...load.Option) error {
	messagerMap := h.NewMessagerMap(filter)
	for name, msger := range messagerMap {
		if err := msger.Load(dir, format, options...); err != nil {
			return errors.WithMessagef(err, "failed to load: %v", name)
		}
		fmt.Println("Loaded: " + msger.Name())
	}
	h.SetMessagerMap(messagerMap)
	return nil
}

func Get[T Messager](h *Hub) T {
	var t T
	if h == nil {
		return t
	}
	msgers := h.messagerMap
	msger := msgers[t.Name()]
	if msger != nil {
		if t, ok := msger.(T); ok {
			return t
		}
	}
	return t
}

`
