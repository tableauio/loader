package tableau

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tableauio/tableau/options"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var registrar Registrar

func init() {
	registrar = make(Registrar, 1024)
}

type Messager interface {
	Name() string
	Load(dir string, fmt options.Format) error
}

type ConfigMap = map[string]Messager
type MessagerGenerator = func() Messager
type Registrar = map[string]MessagerGenerator

func register(name string, gen MessagerGenerator) {
	registrar[name] = gen
}

type Filter interface {
	Filter(name string) bool
}

func load(msg proto.Message, dir string, format options.Format) error {
	ext, err := options.Format2Ext(format)
	if err != nil {
		return fmt.Errorf("invalid format: %v", format)
	}
	md := msg.ProtoReflect().Descriptor()
	msgName := string(md.Name())
	path := filepath.Join(dir, msgName+ext)

	if content, err := os.ReadFile(path); err != nil {
		return fmt.Errorf("failed to read file %v: %v", path, err)
	} else {
		if err := protojson.Unmarshal(content, msg); err != nil {
			return fmt.Errorf("failed to parse message %v: %v", msgName, err)
		}
	}
	return nil
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
	for name, gen := range registrar {
		if filter == nil || filter.Filter(name) {
			configMap[name] = gen()
		}
	}
	return configMap
}

func (h *Hub) Load(dir string, filter Filter, format options.Format) error {
	configMap := h.newConfigMap(filter)
	for name, msger := range configMap {
		if err := msger.Load(dir, format); err != nil {
			return fmt.Errorf("failed to load %v: %v", name, err)
		}
		fmt.Println("Loaded successfully: " + msger.Name())
	}
	h.configMap = configMap
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
