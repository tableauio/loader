package tableau

import (
	"github.com/tableauio/loader/_lab/go/protoconf"
	"github.com/tableauio/loader/_lab/go/tableau/code"
	"github.com/tableauio/loader/_lab/go/tableau/xerrors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

type ItemConf struct {
	data protoconf.ItemConf
}

func (x *ItemConf) Name() string {
	return string((&x.data).ProtoReflect().Descriptor().Name())
}

func (x *ItemConf) Data() *protoconf.ItemConf {
	return &x.data
}

// Messager is used to implement Checker interface.
func (x *ItemConf) Messager() Messager {
	return x
}

// Check is used to implement Checker interface.
func (x *ItemConf) Check() error {
	return nil
}

func (x *ItemConf) Load(dir string, format format.Format) error {
	return load.Load(&x.data, dir, format)
}

func (x *ItemConf) InternalCheck(hub *Hub) error {
	return nil
}

func (x *ItemConf) Get1(key1 uint32) (*protoconf.ItemConf_Item, error) {
	d := x.data.ItemMap
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ItemMap is nil")
	}
	if val, ok := d[key1]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "key1(%v)not found", key1)
	} else {
		return val, nil
	}
}

func init() {
	register("ItemConf", func() Messager {
		return &ItemConf{}
	})
}
