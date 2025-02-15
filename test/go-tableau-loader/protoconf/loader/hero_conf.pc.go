// Code generated by protoc-gen-go-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-go-tableau-loader v0.6.0
// - protoc                       v3.19.3
// source: hero_conf.proto

package loader

import (
	pair "github.com/tableauio/loader/pkg/pair"
	treemap "github.com/tableauio/loader/pkg/treemap"
	protoconf "github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	base "github.com/tableauio/loader/test/go-tableau-loader/protoconf/base"
	code "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/code"
	xerrors "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/xerrors"
	format "github.com/tableauio/tableau/format"
	load "github.com/tableauio/tableau/load"
	store "github.com/tableauio/tableau/store"
	proto "google.golang.org/protobuf/proto"
	time "time"
)

// Index types.
// Index: Title
type HeroConf_Index_AttrMap = map[string][]*protoconf.HeroConf_Hero_Attr

// HeroConf is a wrapper around protobuf message: protoconf.HeroConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type HeroConf struct {
	UnimplementedMessager
	data         protoconf.HeroConf
	indexAttrMap HeroConf_Index_AttrMap
}

// Name returns the HeroConf's message name.
func (x *HeroConf) Name() string {
	if x != nil {
		return string((&x.data).ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the HeroConf's inner message data.
func (x *HeroConf) Data() *protoconf.HeroConf {
	if x != nil {
		return &x.data
	}
	return nil
}

// Load fills HeroConf's inner message from file in the specified directory and format.
func (x *HeroConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	err := load.Load(x.Data(), dir, format, options...)
	if err != nil {
		return err
	}
	return x.processAfterLoad()
}

// Store writes HeroConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *HeroConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *HeroConf) Messager() Messager {
	return x
}

// Message returns the HeroConf's inner message data.
func (x *HeroConf) Message() proto.Message {
	if x != nil {
		return &x.data
	}
	return nil
}

// processAfterLoad runs after this messager is loaded.
func (x *HeroConf) processAfterLoad() error {
	// Index init.
	// Index: Title
	x.indexAttrMap = make(HeroConf_Index_AttrMap)
	for _, item1 := range x.data.GetHeroMap() {
		for _, item2 := range item1.GetAttrMap() {
			key := item2.GetTitle()
			x.indexAttrMap[key] = append(x.indexAttrMap[key], item2)
		}
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return
// NotFound error if the key is not found.
func (x *HeroConf) Get1(name string) (*protoconf.HeroConf_Hero, error) {
	d := x.Data().GetHeroMap()
	if val, ok := d[name]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "name(%v) not found", name)
	} else {
		return val, nil
	}
}

// Get2 finds value in the 2-level map. It will return
// NotFound error if the key is not found.
func (x *HeroConf) Get2(name string, title string) (*protoconf.HeroConf_Hero_Attr, error) {
	conf, err := x.Get1(name)
	if err != nil {
		return nil, err
	}
	d := conf.GetAttrMap()
	if val, ok := d[title]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "title(%v) not found", title)
	} else {
		return val, nil
	}
}

// Index: Title

// FindAttrMap returns the index(Title) to value(protoconf.HeroConf_Hero_Attr) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *HeroConf) FindAttrMap() HeroConf_Index_AttrMap {
	return x.indexAttrMap
}

// FindAttr returns a slice of all values of the given key.
func (x *HeroConf) FindAttr(title string) []*protoconf.HeroConf_Hero_Attr {
	return x.indexAttrMap[title]
}

// FindFirstAttr returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *HeroConf) FindFirstAttr(title string) *protoconf.HeroConf_Hero_Attr {
	val := x.indexAttrMap[title]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// OrderedMap types.
type BaseHeroItemMap_OrderedMap = treemap.TreeMap[string, *base.Item]

type ProtoconfHeroBaseConfHeroMap_OrderedMapValue = pair.Pair[*BaseHeroItemMap_OrderedMap, *base.Hero]
type ProtoconfHeroBaseConfHeroMap_OrderedMap = treemap.TreeMap[string, *ProtoconfHeroBaseConfHeroMap_OrderedMapValue]

// HeroBaseConf is a wrapper around protobuf message: protoconf.HeroBaseConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type HeroBaseConf struct {
	UnimplementedMessager
	data       protoconf.HeroBaseConf
	orderedMap *ProtoconfHeroBaseConfHeroMap_OrderedMap
}

// Name returns the HeroBaseConf's message name.
func (x *HeroBaseConf) Name() string {
	if x != nil {
		return string((&x.data).ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the HeroBaseConf's inner message data.
func (x *HeroBaseConf) Data() *protoconf.HeroBaseConf {
	if x != nil {
		return &x.data
	}
	return nil
}

// Load fills HeroBaseConf's inner message from file in the specified directory and format.
func (x *HeroBaseConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	err := load.Load(x.Data(), dir, format, options...)
	if err != nil {
		return err
	}
	return x.processAfterLoad()
}

// Store writes HeroBaseConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *HeroBaseConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *HeroBaseConf) Messager() Messager {
	return x
}

// Message returns the HeroBaseConf's inner message data.
func (x *HeroBaseConf) Message() proto.Message {
	if x != nil {
		return &x.data
	}
	return nil
}

// processAfterLoad runs after this messager is loaded.
func (x *HeroBaseConf) processAfterLoad() error {
	// OrderedMap init.
	x.orderedMap = treemap.New[string, *ProtoconfHeroBaseConfHeroMap_OrderedMapValue]()
	for k1, v1 := range x.Data().GetHeroMap() {
		map1 := x.orderedMap
		k1v := &ProtoconfHeroBaseConfHeroMap_OrderedMapValue{
			First:  treemap.New[string, *base.Item](),
			Second: v1,
		}
		map1.Put(k1, k1v)
		for k2, v2 := range v1.GetItemMap() {
			map2 := k1v.First
			map2.Put(k2, v2)
		}
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return
// NotFound error if the key is not found.
func (x *HeroBaseConf) Get1(name string) (*base.Hero, error) {
	d := x.Data().GetHeroMap()
	if val, ok := d[name]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "name(%v) not found", name)
	} else {
		return val, nil
	}
}

// Get2 finds value in the 2-level map. It will return
// NotFound error if the key is not found.
func (x *HeroBaseConf) Get2(name string, id string) (*base.Item, error) {
	conf, err := x.Get1(name)
	if err != nil {
		return nil, err
	}
	d := conf.GetItemMap()
	if val, ok := d[id]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "id(%v) not found", id)
	} else {
		return val, nil
	}
}

// GetOrderedMap returns the 1-level ordered map.
func (x *HeroBaseConf) GetOrderedMap() *ProtoconfHeroBaseConfHeroMap_OrderedMap {
	return x.orderedMap
}

// GetOrderedMap1 finds value in the 1-level ordered map. It will return
// NotFound error if the key is not found.
func (x *HeroBaseConf) GetOrderedMap1(name string) (*BaseHeroItemMap_OrderedMap, error) {
	conf := x.orderedMap
	if val, ok := conf.Get(name); !ok {
		return nil, xerrors.Errorf(code.NotFound, "name(%v) not found", name)
	} else {
		return val.First, nil
	}
}

func init() {
	Register(func() Messager {
		return new(HeroConf)
	})
	Register(func() Messager {
		return new(HeroBaseConf)
	})
}
