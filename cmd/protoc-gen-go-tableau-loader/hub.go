package main

import (
	"path/filepath"

	"google.golang.org/protobuf/compiler/protogen"
)

// generateHub generates related hub files.
func generateHub(gen *protogen.Plugin) {
	filename := filepath.Join("hub."+pcExt+".go")
	g := gen.NewGeneratedFile(filename, "")
	generateCommonHeader(gen, g)
	g.P()
	g.P("package ", *pkg)
	g.P()
	g.P(staticHubContent)
	g.P()

	for _, messager := range messagers {
		g.P("func (h *Hub) Get", messager, "() *", messager, " {")
		g.P(`msger := h.configMap["`, messager, `"]`)
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
)

type Messager interface {
	Name() string
	Load(dir string, fmt format.Format) error
}

type ConfigMap = map[string]Messager
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
	configMap ConfigMap
}

func NewHub() *Hub {
	return &Hub{
		configMap: map[string]Messager{},
	}
}

func (h *Hub) newConfigMap(filter Filter) ConfigMap {
	configMap := map[string]Messager{}
	for name, gen := range getRegistrar().generators {
		if filter == nil || filter.Filter(name) {
			configMap[name] = gen()
		}
	}
	return configMap
}

func (h *Hub) SetConfigMap(configMap ConfigMap) {
	h.configMap = configMap
}

func (h *Hub) Load(dir string, filter Filter, format format.Format) error {
	configMap := h.newConfigMap(filter)
	for name, msger := range configMap {
		if err := msger.Load(dir, format); err != nil {
			return errors.WithMessagef(err, "failed to load: %v", name)
		}
		fmt.Println("Loaded: " + msger.Name())
	}
	h.SetConfigMap(configMap)
	return nil
}`
