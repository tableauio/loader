package tableau

import (
	"errors"
	"fmt"

	"github.com/tableauio/loader/_lab/go/protoconf"
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

func (x *ActivityConf) Load(dir string, fmt format.Format) error {
	return load.Load(&x.data, dir, fmt)
}

func (x *ActivityConf) Get1(key1 uint64) (*protoconf.ActivityConf_Activity, error) {
	d := x.data.ActivityMap
	if d == nil {
		return nil, errors.New("ActivityMap is nil")
	}
	if val, ok := d[key1]; !ok {
		return nil, fmt.Errorf("key1(%v)not found", key1)
	} else {
		return val, nil
	}
}

func (x *ActivityConf) Get2(key1 uint64, key2 uint32) (*protoconf.ActivityConf_Activity_Chapter, error) {
	conf, err := x.Get1(key1)
	if err != nil {
		return nil, fmt.Errorf("Get1 failed: %v", err)
	}

	d := conf.ChapterMap
	if d == nil {
		return nil, errors.New("ChapterMap is nil")
	}
	if val, ok := d[key2]; !ok {
		return nil, fmt.Errorf("key2(%v)not found", key1)
	} else {
		return val, nil
	}
}

func (x *ActivityConf) Get3(key1 uint64, key2 uint32, key3 uint32) (*protoconf.Section, error) {
	conf, err := x.Get2(key1, key2)
	if err != nil {
		return nil, fmt.Errorf("Get1 failed: %v", err)
	}

	d := conf.SectionMap
	if d == nil {
		return nil, errors.New("ChapterMap is nil")
	}
	if val, ok := d[key3]; !ok {
		return nil, fmt.Errorf("key2(%v)not found", key1)
	} else {
		return val, nil
	}
}

func (x *ActivityConf) Get4(key1 uint64, key2 uint32, key3 uint32, key4 uint32) (*protoconf.Item, error) {
	conf, err := x.Get3(key1, key2, key3)
	if err != nil {
		return nil, fmt.Errorf("Get3 failed: %v", err)
	}

	d := conf.SectionItemMap
	if d == nil {
		return nil, errors.New("SectionItemMap is nil")
	}
	if val, ok := d[key4]; !ok {
		return nil, fmt.Errorf("key4(%v)not found", key1)
	} else {
		return val, nil
	}
}

func init() {
	register("ActivityConf", func() Messager {
		return &ActivityConf{}
	})
}
