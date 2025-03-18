// Code generated by protoc-gen-go-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-go-tableau-loader v0.6.0
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
	store "github.com/tableauio/tableau/store"
	proto "google.golang.org/protobuf/proto"
	time "time"
)

// OrderedMap types.
type ProtoconfItemConfItemMap_OrderedMap = treemap.TreeMap[uint32, *protoconf.ItemConf_Item]

// Index types.
// Index: Type
type ItemConf_Index_ItemMap = map[protoconf.FruitType][]*protoconf.ItemConf_Item

// Index: Param@ItemInfo
type ItemConf_Index_ItemInfoMap = map[int32][]*protoconf.ItemConf_Item

// Index: Default@ItemDefaultInfo
type ItemConf_Index_ItemDefaultInfoMap = map[string][]*protoconf.ItemConf_Item

// Index: ExtType@ItemExtInfo
type ItemConf_Index_ItemExtInfoMap = map[protoconf.FruitType][]*protoconf.ItemConf_Item

// Index: (ID,Name)@AwardItem
type ItemConf_Index_AwardItemKey struct {
	Id   uint32
	Name string
}
type ItemConf_Index_AwardItemMap = map[ItemConf_Index_AwardItemKey][]*protoconf.ItemConf_Item

// Index: (ID,Type,Param,ExtType)@SpecialItem
type ItemConf_Index_SpecialItemKey struct {
	Id      uint32
	Type    protoconf.FruitType
	Param   int32
	ExtType protoconf.FruitType
}
type ItemConf_Index_SpecialItemMap = map[ItemConf_Index_SpecialItemKey][]*protoconf.ItemConf_Item

// Index: PathDir@ItemPathDir
type ItemConf_Index_ItemPathDirMap = map[string][]*protoconf.ItemConf_Item

// Index: PathName@ItemPathName
type ItemConf_Index_ItemPathNameMap = map[string][]*protoconf.ItemConf_Item

// Index: PathFriendID@ItemPathFriendID
type ItemConf_Index_ItemPathFriendIDMap = map[uint32][]*protoconf.ItemConf_Item

// Index: UseEffectType@UseEffectType
type ItemConf_Index_UseEffectTypeMap = map[protoconf.UseEffect_Type][]*protoconf.ItemConf_Item

// ItemConf is a wrapper around protobuf message: protoconf.ItemConf.
//
// It is designed for three goals:
//
//  1. Easy use: simple yet powerful accessers.
//  2. Elegant API: concise and clean functions.
//  3. Extensibility: Map, OrdererdMap, Index...
type ItemConf struct {
	UnimplementedMessager
	data, originalData       *protoconf.ItemConf
	orderedMap               *ProtoconfItemConfItemMap_OrderedMap
	indexItemMap             ItemConf_Index_ItemMap
	indexItemInfoMap         ItemConf_Index_ItemInfoMap
	indexItemDefaultInfoMap  ItemConf_Index_ItemDefaultInfoMap
	indexItemExtInfoMap      ItemConf_Index_ItemExtInfoMap
	indexAwardItemMap        ItemConf_Index_AwardItemMap
	indexSpecialItemMap      ItemConf_Index_SpecialItemMap
	indexItemPathDirMap      ItemConf_Index_ItemPathDirMap
	indexItemPathNameMap     ItemConf_Index_ItemPathNameMap
	indexItemPathFriendIdMap ItemConf_Index_ItemPathFriendIDMap
	indexUseEffectTypeMap    ItemConf_Index_UseEffectTypeMap
}

// Name returns the ItemConf's message name.
func (x *ItemConf) Name() string {
	if x != nil {
		return string(x.data.ProtoReflect().Descriptor().Name())
	}
	return ""
}

// Data returns the ItemConf's inner message data.
func (x *ItemConf) Data() *protoconf.ItemConf {
	if x != nil {
		return x.data
	}
	return nil
}

// Load fills ItemConf's inner message from file in the specified directory and format.
func (x *ItemConf) Load(dir string, format format.Format, options ...load.Option) error {
	start := time.Now()
	defer func() {
		x.Stats.Duration = time.Since(start)
	}()
	x.data = &protoconf.ItemConf{}
	err := load.Load(x.data, dir, format, options...)
	if err != nil {
		return err
	}
	if x.backup {
		x.originalData = proto.Clone(x.data).(*protoconf.ItemConf)
	}
	return x.processAfterLoad()
}

// Store writes ItemConf's inner message to file in the specified directory and format.
// Available formats: JSON, Bin, and Text.
func (x *ItemConf) Store(dir string, format format.Format, options ...store.Option) error {
	return store.Store(x.Data(), dir, format, options...)
}

// Message returns the ItemConf's inner message data.
func (x *ItemConf) Message() proto.Message {
	return x.Data()
}

// originalMessage returns the ItemConf's original inner message.
func (x *ItemConf) originalMessage() proto.Message {
	if x != nil {
		return x.originalData
	}
	return nil
}

// processAfterLoad runs after this messager is loaded.
func (x *ItemConf) processAfterLoad() error {
	// OrderedMap init.
	x.orderedMap = treemap.New[uint32, *protoconf.ItemConf_Item]()
	for k1, v1 := range x.Data().GetItemMap() {
		map1 := x.orderedMap
		map1.Put(k1, v1)
	}
	// Index init.
	// Index: Type
	x.indexItemMap = make(ItemConf_Index_ItemMap)
	for _, item1 := range x.data.GetItemMap() {
		key := item1.GetType()
		x.indexItemMap[key] = append(x.indexItemMap[key], item1)
	}
	// Index: Param@ItemInfo
	x.indexItemInfoMap = make(ItemConf_Index_ItemInfoMap)
	for _, item1 := range x.data.GetItemMap() {
		for _, item2 := range item1.GetParamList() {
			key := item2
			x.indexItemInfoMap[key] = append(x.indexItemInfoMap[key], item1)
		}
	}
	// Index: Default@ItemDefaultInfo
	x.indexItemDefaultInfoMap = make(ItemConf_Index_ItemDefaultInfoMap)
	for _, item1 := range x.data.GetItemMap() {
		key := item1.GetDefault()
		x.indexItemDefaultInfoMap[key] = append(x.indexItemDefaultInfoMap[key], item1)
	}
	// Index: ExtType@ItemExtInfo
	x.indexItemExtInfoMap = make(ItemConf_Index_ItemExtInfoMap)
	for _, item1 := range x.data.GetItemMap() {
		for _, item2 := range item1.GetExtTypeList() {
			key := item2
			x.indexItemExtInfoMap[key] = append(x.indexItemExtInfoMap[key], item1)
		}
	}
	// Index: (ID,Name)@AwardItem
	x.indexAwardItemMap = make(ItemConf_Index_AwardItemMap)
	for _, item1 := range x.data.GetItemMap() {
		key := ItemConf_Index_AwardItemKey{item1.GetId(), item1.GetName()}
		x.indexAwardItemMap[key] = append(x.indexAwardItemMap[key], item1)
	}
	// Index: (ID,Type,Param,ExtType)@SpecialItem
	x.indexSpecialItemMap = make(ItemConf_Index_SpecialItemMap)
	for _, item1 := range x.data.GetItemMap() {
		for _, indexItem2 := range item1.GetParamList() {
			for _, indexItem3 := range item1.GetExtTypeList() {
				key := ItemConf_Index_SpecialItemKey{item1.GetId(), item1.GetType(), indexItem2, indexItem3}
				x.indexSpecialItemMap[key] = append(x.indexSpecialItemMap[key], item1)
			}
		}
	}
	// Index: PathDir@ItemPathDir
	x.indexItemPathDirMap = make(ItemConf_Index_ItemPathDirMap)
	for _, item1 := range x.data.GetItemMap() {
		key := item1.GetPath().GetDir()
		x.indexItemPathDirMap[key] = append(x.indexItemPathDirMap[key], item1)
	}
	// Index: PathName@ItemPathName
	x.indexItemPathNameMap = make(ItemConf_Index_ItemPathNameMap)
	for _, item1 := range x.data.GetItemMap() {
		for _, item2 := range item1.GetPath().GetNameList() {
			key := item2
			x.indexItemPathNameMap[key] = append(x.indexItemPathNameMap[key], item1)
		}
	}
	// Index: PathFriendID@ItemPathFriendID
	x.indexItemPathFriendIdMap = make(ItemConf_Index_ItemPathFriendIDMap)
	for _, item1 := range x.data.GetItemMap() {
		key := item1.GetPath().GetFriend().GetId()
		x.indexItemPathFriendIdMap[key] = append(x.indexItemPathFriendIdMap[key], item1)
	}
	// Index: UseEffectType@UseEffectType
	x.indexUseEffectTypeMap = make(ItemConf_Index_UseEffectTypeMap)
	for _, item1 := range x.data.GetItemMap() {
		key := item1.GetUseEffect().GetType()
		x.indexUseEffectTypeMap[key] = append(x.indexUseEffectTypeMap[key], item1)
	}
	return nil
}

// Get1 finds value in the 1-level map. It will return
// NotFound error if the key is not found.
func (x *ItemConf) Get1(id uint32) (*protoconf.ItemConf_Item, error) {
	d := x.Data().GetItemMap()
	if val, ok := d[id]; !ok {
		return nil, xerrors.Errorf(code.NotFound, "id(%v) not found", id)
	} else {
		return val, nil
	}
}

// GetOrderedMap returns the 1-level ordered map.
func (x *ItemConf) GetOrderedMap() *ProtoconfItemConfItemMap_OrderedMap {
	return x.orderedMap
}

// Index: Type

// FindItemMap returns the index(Type) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindItemMap() ItemConf_Index_ItemMap {
	return x.indexItemMap
}

// FindItem returns a slice of all values of the given key.
func (x *ItemConf) FindItem(type_ protoconf.FruitType) []*protoconf.ItemConf_Item {
	return x.indexItemMap[type_]
}

// FindFirstItem returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstItem(type_ protoconf.FruitType) *protoconf.ItemConf_Item {
	val := x.indexItemMap[type_]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: Param@ItemInfo

// FindItemInfoMap returns the index(Param@ItemInfo) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindItemInfoMap() ItemConf_Index_ItemInfoMap {
	return x.indexItemInfoMap
}

// FindItemInfo returns a slice of all values of the given key.
func (x *ItemConf) FindItemInfo(param int32) []*protoconf.ItemConf_Item {
	return x.indexItemInfoMap[param]
}

// FindFirstItemInfo returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstItemInfo(param int32) *protoconf.ItemConf_Item {
	val := x.indexItemInfoMap[param]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: Default@ItemDefaultInfo

// FindItemDefaultInfoMap returns the index(Default@ItemDefaultInfo) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindItemDefaultInfoMap() ItemConf_Index_ItemDefaultInfoMap {
	return x.indexItemDefaultInfoMap
}

// FindItemDefaultInfo returns a slice of all values of the given key.
func (x *ItemConf) FindItemDefaultInfo(default_ string) []*protoconf.ItemConf_Item {
	return x.indexItemDefaultInfoMap[default_]
}

// FindFirstItemDefaultInfo returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstItemDefaultInfo(default_ string) *protoconf.ItemConf_Item {
	val := x.indexItemDefaultInfoMap[default_]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: ExtType@ItemExtInfo

// FindItemExtInfoMap returns the index(ExtType@ItemExtInfo) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindItemExtInfoMap() ItemConf_Index_ItemExtInfoMap {
	return x.indexItemExtInfoMap
}

// FindItemExtInfo returns a slice of all values of the given key.
func (x *ItemConf) FindItemExtInfo(extType protoconf.FruitType) []*protoconf.ItemConf_Item {
	return x.indexItemExtInfoMap[extType]
}

// FindFirstItemExtInfo returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstItemExtInfo(extType protoconf.FruitType) *protoconf.ItemConf_Item {
	val := x.indexItemExtInfoMap[extType]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: (ID,Name)@AwardItem

// FindAwardItemMap returns the index((ID,Name)@AwardItem) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindAwardItemMap() ItemConf_Index_AwardItemMap {
	return x.indexAwardItemMap
}

// FindAwardItem returns a slice of all values of the given key.
func (x *ItemConf) FindAwardItem(key ItemConf_Index_AwardItemKey) []*protoconf.ItemConf_Item {
	return x.indexAwardItemMap[key]
}

// FindFirstAwardItem returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstAwardItem(key ItemConf_Index_AwardItemKey) *protoconf.ItemConf_Item {
	val := x.indexAwardItemMap[key]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: (ID,Type,Param,ExtType)@SpecialItem

// FindSpecialItemMap returns the index((ID,Type,Param,ExtType)@SpecialItem) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindSpecialItemMap() ItemConf_Index_SpecialItemMap {
	return x.indexSpecialItemMap
}

// FindSpecialItem returns a slice of all values of the given key.
func (x *ItemConf) FindSpecialItem(key ItemConf_Index_SpecialItemKey) []*protoconf.ItemConf_Item {
	return x.indexSpecialItemMap[key]
}

// FindFirstSpecialItem returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstSpecialItem(key ItemConf_Index_SpecialItemKey) *protoconf.ItemConf_Item {
	val := x.indexSpecialItemMap[key]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: PathDir@ItemPathDir

// FindItemPathDirMap returns the index(PathDir@ItemPathDir) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindItemPathDirMap() ItemConf_Index_ItemPathDirMap {
	return x.indexItemPathDirMap
}

// FindItemPathDir returns a slice of all values of the given key.
func (x *ItemConf) FindItemPathDir(dir string) []*protoconf.ItemConf_Item {
	return x.indexItemPathDirMap[dir]
}

// FindFirstItemPathDir returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstItemPathDir(dir string) *protoconf.ItemConf_Item {
	val := x.indexItemPathDirMap[dir]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: PathName@ItemPathName

// FindItemPathNameMap returns the index(PathName@ItemPathName) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindItemPathNameMap() ItemConf_Index_ItemPathNameMap {
	return x.indexItemPathNameMap
}

// FindItemPathName returns a slice of all values of the given key.
func (x *ItemConf) FindItemPathName(name string) []*protoconf.ItemConf_Item {
	return x.indexItemPathNameMap[name]
}

// FindFirstItemPathName returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstItemPathName(name string) *protoconf.ItemConf_Item {
	val := x.indexItemPathNameMap[name]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: PathFriendID@ItemPathFriendID

// FindItemPathFriendIDMap returns the index(PathFriendID@ItemPathFriendID) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindItemPathFriendIDMap() ItemConf_Index_ItemPathFriendIDMap {
	return x.indexItemPathFriendIdMap
}

// FindItemPathFriendID returns a slice of all values of the given key.
func (x *ItemConf) FindItemPathFriendID(id uint32) []*protoconf.ItemConf_Item {
	return x.indexItemPathFriendIdMap[id]
}

// FindFirstItemPathFriendID returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstItemPathFriendID(id uint32) *protoconf.ItemConf_Item {
	val := x.indexItemPathFriendIdMap[id]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

// Index: UseEffectType@UseEffectType

// FindUseEffectTypeMap returns the index(UseEffectType@UseEffectType) to value(protoconf.ItemConf_Item) map.
// One key may correspond to multiple values, which are contained by a slice.
func (x *ItemConf) FindUseEffectTypeMap() ItemConf_Index_UseEffectTypeMap {
	return x.indexUseEffectTypeMap
}

// FindUseEffectType returns a slice of all values of the given key.
func (x *ItemConf) FindUseEffectType(type_ protoconf.UseEffect_Type) []*protoconf.ItemConf_Item {
	return x.indexUseEffectTypeMap[type_]
}

// FindFirstUseEffectType returns the first value of the given key,
// or nil if the key correspond to no value.
func (x *ItemConf) FindFirstUseEffectType(type_ protoconf.UseEffect_Type) *protoconf.ItemConf_Item {
	val := x.indexUseEffectTypeMap[type_]
	if len(val) > 0 {
		return val[0]
	}
	return nil
}

func init() {
	Register(func() Messager {
		return new(ItemConf)
	})
}
