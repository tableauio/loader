package tableau

import (
	"github.com/tableauio/loader/_lab/go/protoconf"
	"github.com/tableauio/loader/_lab/go/tableau/code"
	"github.com/tableauio/loader/_lab/go/tableau/xerrors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

type ItemConf struct {
	data *protoconf.ItemConf
}

func (x *ItemConf) Name() string {
	return string(x.Data().ProtoReflect().Descriptor().Name())
}

func (x *ItemConf) Data() *protoconf.ItemConf {
	if x == nil {
		return nil
	}
	return x.data
}

// Messager is used to implement Checker interface.
func (x *ItemConf) Messager() Messager {
	return x
}

// Check is used to implement Checker interface.
func (x *ItemConf) Check(hub *Hub) error {
	return nil
}

func (x *ItemConf) Load(dir string, format format.Format, options ...load.Option) error {
	return load.Load(x.data, dir, format, options...)
}

func (x *ItemConf) Get1(key1 uint32) (*protoconf.ItemConf_Item, error) {
	d := x.Data().GetItemMap()
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
		return &ItemConf{data: &protoconf.ItemConf{}}
	})
}
