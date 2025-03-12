// Code generated by protoc-gen-go-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-go-tableau-loader v0.6.0
// - protoc                       v3.19.3
// source: patch_conf.proto

package loader

import (
	protoconf "github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	code "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/code"
	xerrors "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/xerrors"
	format "github.com/tableauio/tableau/format"
	load "github.com/tableauio/tableau/load"
	store "github.com/tableauio/tableau/store"
	proto "google.golang.org/protobuf/proto"
	time "time"
)

// PatchReplaceConf is a wrapper around protobuf message: protoconf.PatchReplaceConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type PatchReplaceConf struct {
	UnimplementedMessager
	data, originalData *protoconf.PatchReplaceConf
}

// Name returns the PatchReplaceConf's message name.
func (x *PatchReplaceConf) Name() string {
	if x != nil {
		return string(x.data.ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the PatchReplaceConf's inner message data.
func (x *PatchReplaceConf) Data() *protoconf.PatchReplaceConf {
	if x != nil {
		return x.data
	}
	return nil
}

// Load fills PatchReplaceConf's inner message from file in the specified directory and format.
func (x *PatchReplaceConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	x.data = &protoconf.PatchReplaceConf{}
	err := load.Load(x.data, dir, format, options...)
	if err != nil {
		return err
	}
	if x.backup {
		x.originalData = proto.Clone(x.data).(*protoconf.PatchReplaceConf)
	}
	return x.processAfterLoad()
}

// Store writes PatchReplaceConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *PatchReplaceConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *PatchReplaceConf) Messager() Messager {
	return x
}

// Message returns the PatchReplaceConf's inner message data.
func (x *PatchReplaceConf) Message() proto.Message {
	return x.Data()
}

// originalMessage returns the PatchReplaceConf's original inner message.
func (x *PatchReplaceConf) originalMessage() proto.Message {
	if x != nil {
		return x.originalData
	}
	return nil
}

// PatchMergeConf is a wrapper around protobuf message: protoconf.PatchMergeConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type PatchMergeConf struct {
	UnimplementedMessager
	data, originalData *protoconf.PatchMergeConf
}

// Name returns the PatchMergeConf's message name.
func (x *PatchMergeConf) Name() string {
	if x != nil {
		return string(x.data.ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the PatchMergeConf's inner message data.
func (x *PatchMergeConf) Data() *protoconf.PatchMergeConf {
	if x != nil {
		return x.data
	}
	return nil
}

// Load fills PatchMergeConf's inner message from file in the specified directory and format.
func (x *PatchMergeConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	x.data = &protoconf.PatchMergeConf{}
	err := load.Load(x.data, dir, format, options...)
	if err != nil {
		return err
	}
	if x.backup {
		x.originalData = proto.Clone(x.data).(*protoconf.PatchMergeConf)
	}
	return x.processAfterLoad()
}

// Store writes PatchMergeConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *PatchMergeConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *PatchMergeConf) Messager() Messager {
	return x
}

// Message returns the PatchMergeConf's inner message data.
func (x *PatchMergeConf) Message() proto.Message {
	return x.Data()
}

// originalMessage returns the PatchMergeConf's original inner message.
func (x *PatchMergeConf) originalMessage() proto.Message {
	if x != nil {
		return x.originalData
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return
// NotFound error if the key is not found.
func (x *PatchMergeConf) Get1(id uint32) (*protoconf.Item, error) {
	d := x.Data().GetItemMap()
	if val, ok := d[id]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "id(%v) not found", id)
	} else {
		return val, nil
	}
}

// RecursivePatchConf is a wrapper around protobuf message: protoconf.RecursivePatchConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type RecursivePatchConf struct {
	UnimplementedMessager
	data, originalData *protoconf.RecursivePatchConf
}

// Name returns the RecursivePatchConf's message name.
func (x *RecursivePatchConf) Name() string {
	if x != nil {
		return string(x.data.ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the RecursivePatchConf's inner message data.
func (x *RecursivePatchConf) Data() *protoconf.RecursivePatchConf {
	if x != nil {
		return x.data
	}
	return nil
}

// Load fills RecursivePatchConf's inner message from file in the specified directory and format.
func (x *RecursivePatchConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	x.data = &protoconf.RecursivePatchConf{}
	err := load.Load(x.data, dir, format, options...)
	if err != nil {
		return err
	}
	if x.backup {
		x.originalData = proto.Clone(x.data).(*protoconf.RecursivePatchConf)
	}
	return x.processAfterLoad()
}

// Store writes RecursivePatchConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *RecursivePatchConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *RecursivePatchConf) Messager() Messager {
	return x
}

// Message returns the RecursivePatchConf's inner message data.
func (x *RecursivePatchConf) Message() proto.Message {
	return x.Data()
}

// originalMessage returns the RecursivePatchConf's original inner message.
func (x *RecursivePatchConf) originalMessage() proto.Message {
	if x != nil {
		return x.originalData
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return
// NotFound error if the key is not found.
func (x *RecursivePatchConf) Get1(shopId uint32) (*protoconf.RecursivePatchConf_Shop, error) {
	d := x.Data().GetShopMap()
	if val, ok := d[shopId]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "shopId(%v) not found", shopId)
	} else {
		return val, nil
	}
}

// Get2 finds value in the 2-level map. It will return
// NotFound error if the key is not found.
func (x *RecursivePatchConf) Get2(shopId uint32, goodsId uint32) (*protoconf.RecursivePatchConf_Shop_Goods, error) {
	conf, err := x.Get1(shopId)
	if err != nil {
		return nil, err
	}
	d := conf.GetGoodsMap()
	if val, ok := d[goodsId]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "goodsId(%v) not found", goodsId)
	} else {
		return val, nil
	}
}

// Get3 finds value in the 3-level map. It will return
// NotFound error if the key is not found.
func (x *RecursivePatchConf) Get3(shopId uint32, goodsId uint32, type_ uint32) (*protoconf.RecursivePatchConf_Shop_Goods_Currency, error) {
	conf, err := x.Get2(shopId, goodsId)
	if err != nil {
		return nil, err
	}
	d := conf.GetCurrencyMap()
	if val, ok := d[type_]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "type_(%v) not found", type_)
	} else {
		return val, nil
	}
}

// Get4 finds value in the 4-level map. It will return
// NotFound error if the key is not found.
func (x *RecursivePatchConf) Get4(shopId uint32, goodsId uint32, type_ uint32, key4 int32) (int32, error) {
	conf, err := x.Get3(shopId, goodsId, type_)
	if err != nil {
		return 0, err
	}
	d := conf.GetValueList()
	if val, ok := d[key4]; !ok {
		return 0, xerrors.Errorf(code.NotFound, "key4(%v) not found", key4)
	} else {
		return val, nil
	}
}

func init() {
	Register(func() Messager {
		return new(PatchReplaceConf)
	})
	Register(func() Messager {
		return new(PatchMergeConf)
	})
	Register(func() Messager {
		return new(RecursivePatchConf)
	})
}
