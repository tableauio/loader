package tableau

import (
	"github.com/tableauio/loader/_lab/go/protoconf"
	"github.com/tableauio/loader/_lab/go/tableau/code"
	"github.com/tableauio/loader/_lab/go/tableau/xerrors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

type ActivityConf struct {
	data protoconf.ActivityConf
}

func (x *ActivityConf) Name() string {
	return string((&x.data).ProtoReflect().Descriptor().Name())
}

func (x *ActivityConf) Data() *protoconf.ActivityConf {
	return &x.data
}

// Messager is used to implement Checker interface.
func (x *ActivityConf) Messager() Messager {
	return x
}

// Check is used to implement Checker interface.
func (x *ActivityConf) Check() error {
	return nil
}

func (x *ActivityConf) Load(dir string, fmt format.Format) error {
	return load.Load(&x.data, dir, fmt)
}

func (x *ActivityConf) InternalCheck(hub *Hub) error {
	return nil
}

func (x *ActivityConf) Get1(key1 uint64) (*protoconf.ActivityConf_Activity, error) {
	d := x.data.ActivityMap
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ActivityMap is nil")
	}
	if val, ok := d[key1]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "key1(%v) not found", key1)
	} else {
		return val, nil
	}
}

func (x *ActivityConf) Get2(key1 uint64, key2 uint32) (*protoconf.ActivityConf_Activity_Chapter, error) {
	conf, err := x.Get1(key1)
	if err != nil {
		return nil, err
	}

	d := conf.ChapterMap
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ChapterMap is nil")
	}
	if val, ok := d[key2]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "key2(%v) not found", key2)
	} else {
		return val, nil
	}
}

func (x *ActivityConf) Get3(key1 uint64, key2 uint32, key3 uint32) (*protoconf.Section, error) {
	conf, err := x.Get2(key1, key2)
	if err != nil {
		return nil, err
	}

	d := conf.SectionMap
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ChapterMap is nil")
	}
	if val, ok := d[key3]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "key3(%v) not found", key3)
	} else {
		return val, nil
	}
}

func (x *ActivityConf) Get4(key1 uint64, key2 uint32, key3 uint32, key4 uint32) (int32, error) {
	conf, err := x.Get3(key1, key2, key3)
	if err != nil {
		return 0, err
	}

	d := conf.SectionRankMap
	if d == nil {
		return 0, xerrors.Errorf(code.Nil, "SectionRankMap is nil")
	}
	if val, ok := d[key4]; !ok {
		return 0, xerrors.Errorf(code.NotFound, "key4(%v)not found", key1)
	} else {
		return val, nil
	}
}

func init() {
	register("ActivityConf", func() Messager {
		return &ActivityConf{}
	})
}
