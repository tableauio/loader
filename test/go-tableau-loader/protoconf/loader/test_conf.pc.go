// Code generated by protoc-gen-go-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-go-tableau-loader v0.2.6
// - protoc                       v3.19.3
// source: test_conf.proto

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
type ActivityConf_int32_OrderedMap = treemap.TreeMap[uint32, int32]

type Uint32_Section_OrderedMapValue = pair.Pair[*ActivityConf_int32_OrderedMap, *protoconf.Section]
type Uint32_Section_OrderedMap = treemap.TreeMap[uint32, Uint32_Section_OrderedMapValue]

type ActivityConf_Activity_Chapter_OrderedMapValue = pair.Pair[*Uint32_Section_OrderedMap, *protoconf.ActivityConf_Activity_Chapter]
type ActivityConf_Activity_Chapter_OrderedMap = treemap.TreeMap[uint32, ActivityConf_Activity_Chapter_OrderedMapValue]

type ActivityConf_Activity_OrderedMapValue = pair.Pair[*ActivityConf_Activity_Chapter_OrderedMap, *protoconf.ActivityConf_Activity]
type ActivityConf_Activity_OrderedMap = treemap.TreeMap[uint64, ActivityConf_Activity_OrderedMapValue]

// ActivityConf is a wrapper around protobuf message: protoconf.ActivityConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type ActivityConf struct {
	UnimplementedMessager
	data       protoconf.ActivityConf
	orderedMap *ActivityConf_Activity_OrderedMap
}

// Name returns the ActivityConf's message name.
func (x *ActivityConf) Name() string {
	if x != nil {
		return string((&x.data).ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the ActivityConf's inner message data.
func (x *ActivityConf) Data() *protoconf.ActivityConf {
	if x != nil {
		return &x.data
	}
	return nil
}

// Load fills ActivityConf's inner message from file in the specified directory and format.
func (x *ActivityConf) Load(dir string, format format.Format, options ...load.Option) error {
	err := load.Load(x.Data(), dir, format, options...)
	if err != nil {
		return err
	}
	return x.AfterLoad()
}

// Store writes ActivityConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *ActivityConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *ActivityConf) Messager() Messager {
	return x
}

// AfterLoad runs after this messager is loaded.
func (x *ActivityConf) AfterLoad() error {
	// OrderedMap init.
	x.orderedMap = treemap.New[uint64, ActivityConf_Activity_OrderedMapValue]()
	for k1, v1 := range x.Data().GetActivityMap() {
		map1 := x.orderedMap
		map1.Put(k1, ActivityConf_Activity_OrderedMapValue{
			First:  treemap.New[uint32, ActivityConf_Activity_Chapter_OrderedMapValue](),
			Second: v1,
		})
		k1v, _ := map1.Get(k1)
		for k2, v2 := range v1.GetChapterMap() {
			map2 := k1v.First
			map2.Put(k2, ActivityConf_Activity_Chapter_OrderedMapValue{
				First:  treemap.New[uint32, Uint32_Section_OrderedMapValue](),
				Second: v2,
			})
			k2v, _ := map2.Get(k2)
			for k3, v3 := range v2.GetSectionMap() {
				map3 := k2v.First
				map3.Put(k3, Uint32_Section_OrderedMapValue{
					First:  treemap.New[uint32, int32](),
					Second: v3,
				})
				k3v, _ := map3.Get(k3)
				for k4, v4 := range v3.GetSectionRankMap() {
					map4 := k3v.First
					map4.Put(k4, v4)
				}
			}
		}
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ActivityConf) Get1(activityID uint64) (*protoconf.ActivityConf_Activity, error) {
	d := x.Data().GetActivityMap()
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ActivityMap is nil")
	}
	if val, ok := d[activityID]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "activityID(%v) not found", activityID)
	} else {
		return val, nil
	}
}

// Get2 finds value in the 2-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ActivityConf) Get2(activityID uint64, chapterID uint32) (*protoconf.ActivityConf_Activity_Chapter, error) {
	conf, err := x.Get1(activityID)
	if err != nil {
		return nil, err
	}

	d := conf.GetChapterMap()
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ChapterMap is nil")
	}
	if val, ok := d[chapterID]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "chapterID(%v) not found", chapterID)
	} else {
		return val, nil
	}
}

// Get3 finds value in the 3-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ActivityConf) Get3(activityID uint64, chapterID uint32, sectionID uint32) (*protoconf.Section, error) {
	conf, err := x.Get2(activityID, chapterID)
	if err != nil {
		return nil, err
	}

	d := conf.GetSectionMap()
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "SectionMap is nil")
	}
	if val, ok := d[sectionID]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "sectionID(%v) not found", sectionID)
	} else {
		return val, nil
	}
}

// Get4 finds value in the 4-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ActivityConf) Get4(activityID uint64, chapterID uint32, sectionID uint32, key4 uint32) (int32, error) {
	conf, err := x.Get3(activityID, chapterID, sectionID)
	if err != nil {
		return 0, err
	}

	d := conf.GetSectionRankMap()
	if d == nil {
		return 0, xerrors.Errorf(code.Nil, "SectionRankMap is nil")
	}
	if val, ok := d[key4]; !ok {
		return 0, xerrors.Errorf(code.NotFound, "key4(%v) not found", key4)
	} else {
		return val, nil
	}
}

// GetOrderedMap returns the 1-level ordered map.
func (x *ActivityConf) GetOrderedMap() *ActivityConf_Activity_OrderedMap {
	return x.orderedMap
}

// GetOrderedMap1 finds value in the 1-level ordered map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ActivityConf) GetOrderedMap1(activityID uint64) (*ActivityConf_Activity_Chapter_OrderedMap, error) {
	conf := x.orderedMap
	if val, ok := conf.Get(activityID); !ok {
		return nil, xerrors.Errorf(code.NotFound, "activityID(%v) not found", activityID)
	} else {
		return val.First, nil
	}
}

// GetOrderedMap2 finds value in the 2-level ordered map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ActivityConf) GetOrderedMap2(activityID uint64, chapterID uint32) (*Uint32_Section_OrderedMap, error) {
	conf, err := x.GetOrderedMap1(activityID)
	if err != nil {
		return nil, err
	}
	if val, ok := conf.Get(chapterID); !ok {
		return nil, xerrors.Errorf(code.NotFound, "chapterID(%v) not found", chapterID)
	} else {
		return val.First, nil
	}
}

// GetOrderedMap3 finds value in the 3-level ordered map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ActivityConf) GetOrderedMap3(activityID uint64, chapterID uint32, sectionID uint32) (*ActivityConf_int32_OrderedMap, error) {
	conf, err := x.GetOrderedMap2(activityID, chapterID)
	if err != nil {
		return nil, err
	}
	if val, ok := conf.Get(sectionID); !ok {
		return nil, xerrors.Errorf(code.NotFound, "sectionID(%v) not found", sectionID)
	} else {
		return val.First, nil
	}
}

// ChapterConf is a wrapper around protobuf message: protoconf.ChapterConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type ChapterConf struct {
	UnimplementedMessager
	data protoconf.ChapterConf
}

// Name returns the ChapterConf's message name.
func (x *ChapterConf) Name() string {
	if x != nil {
		return string((&x.data).ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the ChapterConf's inner message data.
func (x *ChapterConf) Data() *protoconf.ChapterConf {
	if x != nil {
		return &x.data
	}
	return nil
}

// Load fills ChapterConf's inner message from file in the specified directory and format.
func (x *ChapterConf) Load(dir string, format format.Format, options ...load.Option) error {
	err := load.Load(x.Data(), dir, format, options...)
	if err != nil {
		return err
	}
	return x.AfterLoad()
}

// Store writes ChapterConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *ChapterConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *ChapterConf) Messager() Messager {
	return x
}

// AfterLoad runs after this messager is loaded.
func (x *ChapterConf) AfterLoad() error {
	return nil
}

// Get1 finds value in the 1-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ChapterConf) Get1(id uint64) (*protoconf.ChapterConf_Chapter, error) {
	d := x.Data().GetChapterMap()
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ChapterMap is nil")
	}
	if val, ok := d[id]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "id(%v) not found", id)
	} else {
		return val, nil
	}
}

// ThemeConf is a wrapper around protobuf message: protoconf.ThemeConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type ThemeConf struct {
	UnimplementedMessager
	data protoconf.ThemeConf
}

// Name returns the ThemeConf's message name.
func (x *ThemeConf) Name() string {
	if x != nil {
		return string((&x.data).ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the ThemeConf's inner message data.
func (x *ThemeConf) Data() *protoconf.ThemeConf {
	if x != nil {
		return &x.data
	}
	return nil
}

// Load fills ThemeConf's inner message from file in the specified directory and format.
func (x *ThemeConf) Load(dir string, format format.Format, options ...load.Option) error {
	err := load.Load(x.Data(), dir, format, options...)
	if err != nil {
		return err
	}
	return x.AfterLoad()
}

// Store writes ThemeConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *ThemeConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Messager is used to implement Checker interface.
func (x *ThemeConf) Messager() Messager {
	return x
}

// AfterLoad runs after this messager is loaded.
func (x *ThemeConf) AfterLoad() error {
	return nil
}

// Get1 finds value in the 1-level map. It will return nil if
// the deepest key is not found, otherwise return an error.
func (x *ThemeConf) Get1(name string) (*protoconf.ThemeConf_Theme, error) {
	d := x.Data().GetThemeMap()
	if d == nil {
		return nil, xerrors.Errorf(code.Nil, "ThemeMap is nil")
	}
	if val, ok := d[name]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "name(%v) not found", name)
	} else {
		return val, nil
	}
}

func init() {
	Register(func() Messager {
		return new(ActivityConf)
	})
	Register(func() Messager {
		return new(ChapterConf)
	})
	Register(func() Messager {
		return new(ThemeConf)
	})
}