// Code generated by protoc-gen-go-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-go-tableau-loader v0.2.6
// - protoc                       v3.19.3
// source: item_conf.proto

package loader

import (
	treemap "github.com/tableauio/loader/pkg/treemap"
	protoconf "github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	code "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/code"
	xerrors "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/xerrors"
	format "github.com/tableauio/tableau/format"
	load "github.com/tableauio/tableau/load"
)

// OrderedMap types.
type ItemConf_Item_OrderedMap = treemap.TreeMap[uint32, *protoconf.ItemConf_Item]

// ItemConf is a wrapper around protobuf message: protoconf.ItemConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type ItemConf struct {
	UnimplementedMessager
	data       protoconf.ItemConf
	orderedMap *ItemConf_Item_OrderedMap
}

// Name returns the ItemConf's message name.
func (x *ItemConf) Name() string {
	if x != nil {
		return string((&x.data).ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the ItemConf's inner message data.
func (x *ItemConf) Data() *protoconf.ItemConf {
	if x != nil {
		return &x.data
	}
	return nil
}

// Load fills ItemConf's inner message data from the specified direcotry and format.
func (x *ItemConf) Load(dir string, format format.Format, options ...load.Option) error {
	err := load.Load(x.Data(), dir, format, options...)
	if err != nil {
		return err
	}
	return x.AfterLoad()
}

// AfterLoad runs after this messager is loaded.
func (x *ItemConf) AfterLoad() error {
	// OrderedMap init.
	x.orderedMap = treemap.New[uint32, *protoconf.ItemConf_Item]()
	for k1, v1 := range x.Data().GetItemMap() {
		map1 := x.orderedMap
		map1.Put(k1, v1)
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ItemConf) Get1(id uint32) (*protoconf.ItemConf_Item, error) {
	d := x.Data().GetItemMap()
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ItemMap is nil")
	}
	if val, ok := d[id]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "id(%v)not found", id)
	} else {
		return val, nil
	}
}

// GetOrderedMap returns the 1-level ordered map.
func (x *ItemConf) GetOrderedMap() *ItemConf_Item_OrderedMap {
	return x.orderedMap
}

func init() {
	Register(func() Messager {
		return new(ItemConf)
	})
}
