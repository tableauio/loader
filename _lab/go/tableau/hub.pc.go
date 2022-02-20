package tableau

import (
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
		fmt.Println("Loaded successfully: " + msger.Name())
	}
	h.SetConfigMap(configMap)
	return nil
}

// auto-generated
func (h *Hub) GetActivityConf() *ActivityConf {
	msger := h.configMap["ActivityConf"]
	if msger != nil {
		if conf, ok := msger.(*ActivityConf); ok {
			return conf
		}
	}
	return nil
}
