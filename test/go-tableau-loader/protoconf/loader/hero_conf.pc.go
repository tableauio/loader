// Code generated by protoc-gen-go-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-go-tableau-loader v0.2.6
// - protoc                       v3.19.3
// source: hero_conf.proto

package loader

import (
	pair "github.com/tableauio/loader/pkg/pair"
	treemap "github.com/tableauio/loader/pkg/treemap"
	protoconf "github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	code "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/code"
	xerrors "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/xerrors"
	format "github.com/tableauio/tableau/format"
	load "github.com/tableauio/tableau/load"
	store "github.com/tableauio/tableau/store"
)

// OrderedMap types.
type HeroConf_Hero_Attr_OrderedMap = treemap.TreeMap[string, *protoconf.HeroConf_Hero_Attr]

type HeroConf_Hero_OrderedMapValue = pair.Pair[*HeroConf_Hero_Attr_OrderedMap, *protoconf.HeroConf_Hero]
type HeroConf_Hero_OrderedMap = treemap.TreeMap[string, HeroConf_Hero_OrderedMapValue]

// HeroConf is a wrapper around protobuf message: protoconf.HeroConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type HeroConf struct {
	UnimplementedMessager
	data       protoconf.HeroConf
	orderedMap *HeroConf_Hero_OrderedMap
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
	err := load.Load(x.Data(), dir, format, options...)
	if err != nil {
		return err
	}
	return x.AfterLoad()
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

// AfterLoad runs after this messager is loaded.
func (x *HeroConf) AfterLoad() error {
	// OrderedMap init.
	x.orderedMap = treemap.New[string, HeroConf_Hero_OrderedMapValue]()
	for k1, v1 := range x.Data().GetHeroMap() {
		map1 := x.orderedMap
		map1.Put(k1, HeroConf_Hero_OrderedMapValue{
			First:  treemap.New[string, *protoconf.HeroConf_Hero_Attr](),
			Second: v1,
		})
		k1v, _ := map1.Get(k1)
		for k2, v2 := range v1.GetAttrMap() {
			map2 := k1v.First
			map2.Put(k2, v2)
		}
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *HeroConf) Get1(name string) (*protoconf.HeroConf_Hero, error) {
	d := x.Data().GetHeroMap()
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "HeroMap is nil")
	}
	if val, ok := d[name]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "name(%v) not found", name)
	} else {
		return val, nil
	}
}

// Get2 finds value in the 2-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *HeroConf) Get2(name string, title string) (*protoconf.HeroConf_Hero_Attr, error) {
	conf, err := x.Get1(name)
	if err != nil {
		return nil, err
	}

	d := conf.GetAttrMap()
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "AttrMap is nil")
	}
	if val, ok := d[title]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "title(%v) not found", title)
	} else {
		return val, nil
	}
}

// GetOrderedMap returns the 1-level ordered map.
func (x *HeroConf) GetOrderedMap() *HeroConf_Hero_OrderedMap {
	return x.orderedMap
}

// GetOrderedMap1 finds value in the 1-level ordered map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *HeroConf) GetOrderedMap1(name string) (*HeroConf_Hero_Attr_OrderedMap, error) {
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
}