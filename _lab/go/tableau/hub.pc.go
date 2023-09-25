package tableau

import (
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
	CheckCompatibility(hub, newHub *Hub) error
}

type MessagerMap = map[string]Messager
type MessagerGenerator = func() Messager
type Registrar struct {
	generators map[string]MessagerGenerator
}

func NewRegistrar() *Registrar {
	return &Registrar{
		generators: map[string]MessagerGenerator{},
	}
}

func (r *Registrar) Register(name string, gen MessagerGenerator) {
	r.generators[name] = gen
}

var registrarSingleton *Registrar
var once sync.Once

func getRegistrar() *Registrar {
	once.Do(func() {
		registrarSingleton = NewRegistrar()
	})
	return registrarSingleton
}

func register(name string, gen MessagerGenerator) {
	getRegistrar().Register(name, gen)
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

// Auto-generated getters below

func (h *Hub) GetActivityConf() *ActivityConf {
	msger := h.messagerMap["ActivityConf"]
	if msger != nil {
		if conf, ok := msger.(*ActivityConf); ok {
			return conf
		}
	}
	return nil
}

func (h *Hub) GetItemConf() *ItemConf {
	msger := h.messagerMap["ItemConf"]
	if msger != nil {
		if conf, ok := msger.(*ItemConf); ok {
			return conf
		}
	}
	return nil
}
