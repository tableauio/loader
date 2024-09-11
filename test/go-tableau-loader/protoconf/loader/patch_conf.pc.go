// Code generated by protoc-gen-go-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-go-tableau-loader v0.4.0
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
	data protoconf.PatchReplaceConf
}

// Name returns the PatchReplaceConf's message name.
func (x *PatchReplaceConf) Name() string {
	if x != nil {
		return string((&x.data).ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the PatchReplaceConf's inner message data.
func (x *PatchReplaceConf) Data() *protoconf.PatchReplaceConf {
	if x != nil {
		return &x.data
	}
	return nil
}

// Load fills PatchReplaceConf's inner message from file in the specified directory and format.
func (x *PatchReplaceConf) Load(dir string, format format.Format, options ...load.Option) error {
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

// Store writes PatchReplaceConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *PatchReplaceConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *PatchReplaceConf) Messager() Messager {
	return x
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
	data protoconf.PatchMergeConf
}

// Name returns the PatchMergeConf's message name.
func (x *PatchMergeConf) Name() string {
	if x != nil {
		return string((&x.data).ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the PatchMergeConf's inner message data.
func (x *PatchMergeConf) Data() *protoconf.PatchMergeConf {
	if x != nil {
		return &x.data
	}
	return nil
}

// Load fills PatchMergeConf's inner message from file in the specified directory and format.
func (x *PatchMergeConf) Load(dir string, format format.Format, options ...load.Option) error {
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

// Store writes PatchMergeConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *PatchMergeConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *PatchMergeConf) Messager() Messager {
	return x
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

func init() {
	Register(func() Messager {
		return new(PatchReplaceConf)
	})
	Register(func() Messager {
		return new(PatchMergeConf)
	})
}
