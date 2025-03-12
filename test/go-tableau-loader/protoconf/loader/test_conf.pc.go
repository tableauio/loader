// Code generated by protoc-gen-go-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-go-tableau-loader v0.6.0
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
	proto "google.golang.org/protobuf/proto"
	time "time"
)

// OrderedMap types.
type ProtoconfSectionSectionRankMap_OrderedMap = treemap.TreeMap[uint32, int32]

type ProtoconfActivityConfActivityChapterSectionMap_OrderedMapValue = pair.Pair[*ProtoconfSectionSectionRankMap_OrderedMap, *protoconf.Section]
type ProtoconfActivityConfActivityChapterSectionMap_OrderedMap = treemap.TreeMap[uint32, *ProtoconfActivityConfActivityChapterSectionMap_OrderedMapValue]

type ProtoconfActivityConfActivityChapterMap_OrderedMapValue = pair.Pair[*ProtoconfActivityConfActivityChapterSectionMap_OrderedMap, *protoconf.ActivityConf_Activity_Chapter]
type ProtoconfActivityConfActivityChapterMap_OrderedMap = treemap.TreeMap[uint32, *ProtoconfActivityConfActivityChapterMap_OrderedMapValue]

type ProtoconfActivityConfActivityMap_OrderedMapValue = pair.Pair[*ProtoconfActivityConfActivityChapterMap_OrderedMap, *protoconf.ActivityConf_Activity]
type ProtoconfActivityConfActivityMap_OrderedMap = treemap.TreeMap[uint64, *ProtoconfActivityConfActivityMap_OrderedMapValue]

// Index types.
// Index: ChapterID
type ActivityConf_Index_ChapterMap = map[uint32][]*protoconf.ActivityConf_Activity_Chapter

// Index: ChapterName@NamedChapter
type ActivityConf_Index_NamedChapterMap = map[string][]*protoconf.ActivityConf_Activity_Chapter

// ActivityConf is a wrapper around protobuf message: protoconf.ActivityConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type ActivityConf struct {
	UnimplementedMessager
	data, originalData   *protoconf.ActivityConf
	orderedMap           *ProtoconfActivityConfActivityMap_OrderedMap
	indexChapterMap      ActivityConf_Index_ChapterMap
	indexNamedChapterMap ActivityConf_Index_NamedChapterMap
}

// Name returns the ActivityConf's message name.
func (x *ActivityConf) Name() string {
	if x != nil {
		return string(x.data.ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the ActivityConf's inner message data.
func (x *ActivityConf) Data() *protoconf.ActivityConf {
	if x != nil {
		return x.data
	}
	return nil
}

// Load fills ActivityConf's inner message from file in the specified directory and format.
func (x *ActivityConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	x.data = &protoconf.ActivityConf{}
	err := load.Load(x.data, dir, format, options...)
	if err != nil {
		return err
	}
	if x.backup {
		x.originalData = proto.Clone(x.data).(*protoconf.ActivityConf)
	}
	return x.processAfterLoad()
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

// Message returns the ActivityConf's inner message data.
func (x *ActivityConf) Message() proto.Message {
	return x.Data()
}

// originalMessage returns the ActivityConf's original inner message.
func (x *ActivityConf) originalMessage() proto.Message {
	return x.originalData
}

// mutable returns true if the ActivityConf's inner message is modified.
func (x *ActivityConf) mutable() bool {
	return !proto.Equal(x.originalData, x.data)
}

// processAfterLoad runs after this messager is loaded.
func (x *ActivityConf) processAfterLoad() error {
	// OrderedMap init.
	x.orderedMap = treemap.New[uint64, *ProtoconfActivityConfActivityMap_OrderedMapValue]()
	for k1, v1 := range x.Data().GetActivityMap() {
		map1 := x.orderedMap
		k1v := &ProtoconfActivityConfActivityMap_OrderedMapValue{
			First:  treemap.New[uint32, *ProtoconfActivityConfActivityChapterMap_OrderedMapValue](),
			Second: v1,
		}
		map1.Put(k1, k1v)
		for k2, v2 := range v1.GetChapterMap() {
			map2 := k1v.First
			k2v := &ProtoconfActivityConfActivityChapterMap_OrderedMapValue{
				First:  treemap.New[uint32, *ProtoconfActivityConfActivityChapterSectionMap_OrderedMapValue](),
				Second: v2,
			}
			map2.Put(k2, k2v)
			for k3, v3 := range v2.GetSectionMap() {
				map3 := k2v.First
				k3v := &ProtoconfActivityConfActivityChapterSectionMap_OrderedMapValue{
					First:  treemap.New[uint32, int32](),
					Second: v3,
				}
				map3.Put(k3, k3v)
				for k4, v4 := range v3.GetSectionRankMap() {
					map4 := k3v.First
					map4.Put(k4, v4)
				}
			}
		}
	}
	// Index init.
	// Index: ChapterID
	x.indexChapterMap = make(ActivityConf_Index_ChapterMap)
	for _, item1 := range x.data.GetActivityMap() {
		for _, item2 := range item1.GetChapterMap() {
			key := item2.GetChapterId()
			x.indexChapterMap[key] = append(x.indexChapterMap[key], item2)
		}
	}
	// Index: ChapterName@NamedChapter
	x.indexNamedChapterMap = make(ActivityConf_Index_NamedChapterMap)
	for _, item1 := range x.data.GetActivityMap() {
		for _, item2 := range item1.GetChapterMap() {
			key := item2.GetChapterName()
			x.indexNamedChapterMap[key] = append(x.indexNamedChapterMap[key], item2)
		}
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return
// NotFound error if the key is not found.
func (x *ActivityConf) Get1(activityId uint64) (*protoconf.ActivityConf_Activity, error) {
	d := x.Data().GetActivityMap()
	if val, ok := d[activityId]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "activityId(%v) not found", activityId)
	} else {
		return val, nil
	}
}

// Get2 finds value in the 2-level map. It will return
// NotFound error if the key is not found.
func (x *ActivityConf) Get2(activityId uint64, chapterId uint32) (*protoconf.ActivityConf_Activity_Chapter, error) {
	conf, err := x.Get1(activityId)
	if err != nil {
		return nil, err
	}
	d := conf.GetChapterMap()
	if val, ok := d[chapterId]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "chapterId(%v) not found", chapterId)
	} else {
		return val, nil
	}
}

// Get3 finds value in the 3-level map. It will return
// NotFound error if the key is not found.
func (x *ActivityConf) Get3(activityId uint64, chapterId uint32, sectionId uint32) (*protoconf.Section, error) {
	conf, err := x.Get2(activityId, chapterId)
	if err != nil {
		return nil, err
	}
	d := conf.GetSectionMap()
	if val, ok := d[sectionId]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "sectionId(%v) not found", sectionId)
	} else {
		return val, nil
	}
}

// Get4 finds value in the 4-level map. It will return
// NotFound error if the key is not found.
func (x *ActivityConf) Get4(activityId uint64, chapterId uint32, sectionId uint32, key4 uint32) (int32, error) {
	conf, err := x.Get3(activityId, chapterId, sectionId)
	if err != nil {
		return 0, err
	}
	d := conf.GetSectionRankMap()
	if val, ok := d[key4]; !ok {
		return 0, xerrors.Errorf(code.NotFound, "key4(%v) not found", key4)
	} else {
		return val, nil
	}
}

// GetOrderedMap returns the 1-level ordered map.
func (x *ActivityConf) GetOrderedMap() *ProtoconfActivityConfActivityMap_OrderedMap {
	return x.orderedMap
}

// GetOrderedMap1 finds value in the 1-level ordered map. It will return
// NotFound error if the key is not found.
func (x *ActivityConf) GetOrderedMap1(activityId uint64) (*ProtoconfActivityConfActivityChapterMap_OrderedMap, error) {
	conf := x.orderedMap
	if val, ok := conf.Get(activityId); !ok {
		return nil, xerrors.Errorf(code.NotFound, "activityId(%v) not found", activityId)
	} else {
		return val.First, nil
	}
}

// GetOrderedMap2 finds value in the 2-level ordered map. It will return
// NotFound error if the key is not found.
func (x *ActivityConf) GetOrderedMap2(activityId uint64, chapterId uint32) (*ProtoconfActivityConfActivityChapterSectionMap_OrderedMap, error) {
	conf, err := x.GetOrderedMap1(activityId)
	if err != nil {
		return nil, err
	}
	if val, ok := conf.Get(chapterId); !ok {
		return nil, xerrors.Errorf(code.NotFound, "chapterId(%v) not found", chapterId)
	} else {
		return val.First, nil
	}
}

// GetOrderedMap3 finds value in the 3-level ordered map. It will return
// NotFound error if the key is not found.
func (x *ActivityConf) GetOrderedMap3(activityId uint64, chapterId uint32, sectionId uint32) (*ProtoconfSectionSectionRankMap_OrderedMap, error) {
	conf, err := x.GetOrderedMap2(activityId, chapterId)
	if err != nil {
		return nil, err
	}
	if val, ok := conf.Get(sectionId); !ok {
		return nil, xerrors.Errorf(code.NotFound, "sectionId(%v) not found", sectionId)
	} else {
		return val.First, nil
	}
}

// Index: ChapterID

// FindChapterMap returns the index(ChapterID) to value(protoconf.ActivityConf_Activity_Chapter) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ActivityConf) FindChapterMap() ActivityConf_Index_ChapterMap {
	return x.indexChapterMap
}

// FindChapter returns a slice of all values of the given key.
func (x *ActivityConf) FindChapter(chapterId uint32) []*protoconf.ActivityConf_Activity_Chapter {
	return x.indexChapterMap[chapterId]
}

// FindFirstChapter returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ActivityConf) FindFirstChapter(chapterId uint32) *protoconf.ActivityConf_Activity_Chapter {
	val := x.indexChapterMap[chapterId]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: ChapterName@NamedChapter

// FindNamedChapterMap returns the index(ChapterName@NamedChapter) to value(protoconf.ActivityConf_Activity_Chapter) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ActivityConf) FindNamedChapterMap() ActivityConf_Index_NamedChapterMap {
	return x.indexNamedChapterMap
}

// FindNamedChapter returns a slice of all values of the given key.
func (x *ActivityConf) FindNamedChapter(chapterName string) []*protoconf.ActivityConf_Activity_Chapter {
	return x.indexNamedChapterMap[chapterName]
}

// FindFirstNamedChapter returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ActivityConf) FindFirstNamedChapter(chapterName string) *protoconf.ActivityConf_Activity_Chapter {
	val := x.indexNamedChapterMap[chapterName]
	if len(val) > 0 {
		return val[0]
	}
	return nil
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
	data, originalData *protoconf.ChapterConf
}

// Name returns the ChapterConf's message name.
func (x *ChapterConf) Name() string {
	if x != nil {
		return string(x.data.ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the ChapterConf's inner message data.
func (x *ChapterConf) Data() *protoconf.ChapterConf {
	if x != nil {
		return x.data
	}
	return nil
}

// Load fills ChapterConf's inner message from file in the specified directory and format.
func (x *ChapterConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	x.data = &protoconf.ChapterConf{}
	err := load.Load(x.data, dir, format, options...)
	if err != nil {
		return err
	}
	if x.backup {
		x.originalData = proto.Clone(x.data).(*protoconf.ChapterConf)
	}
	return x.processAfterLoad()
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

// Message returns the ChapterConf's inner message data.
func (x *ChapterConf) Message() proto.Message {
	return x.Data()
}

// originalMessage returns the ChapterConf's original inner message.
func (x *ChapterConf) originalMessage() proto.Message {
	return x.originalData
}

// mutable returns true if the ChapterConf's inner message is modified.
func (x *ChapterConf) mutable() bool {
	return !proto.Equal(x.originalData, x.data)
}

// Get1 finds value in the 1-level map. It will return
// NotFound error if the key is not found.
func (x *ChapterConf) Get1(id uint64) (*protoconf.ChapterConf_Chapter, error) {
	d := x.Data().GetChapterMap()
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
	data, originalData *protoconf.ThemeConf
}

// Name returns the ThemeConf's message name.
func (x *ThemeConf) Name() string {
	if x != nil {
		return string(x.data.ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the ThemeConf's inner message data.
func (x *ThemeConf) Data() *protoconf.ThemeConf {
	if x != nil {
		return x.data
	}
	return nil
}

// Load fills ThemeConf's inner message from file in the specified directory and format.
func (x *ThemeConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	x.data = &protoconf.ThemeConf{}
	err := load.Load(x.data, dir, format, options...)
	if err != nil {
		return err
	}
	if x.backup {
		x.originalData = proto.Clone(x.data).(*protoconf.ThemeConf)
	}
	return x.processAfterLoad()
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

// Message returns the ThemeConf's inner message data.
func (x *ThemeConf) Message() proto.Message {
	return x.Data()
}

// originalMessage returns the ThemeConf's original inner message.
func (x *ThemeConf) originalMessage() proto.Message {
	return x.originalData
}

// mutable returns true if the ThemeConf's inner message is modified.
func (x *ThemeConf) mutable() bool {
	return !proto.Equal(x.originalData, x.data)
}

// Get1 finds value in the 1-level map. It will return
// NotFound error if the key is not found.
func (x *ThemeConf) Get1(name string) (*protoconf.ThemeConf_Theme, error) {
	d := x.Data().GetThemeMap()
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
