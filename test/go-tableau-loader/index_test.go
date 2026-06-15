package loader_test

import (
	"testing"

	"github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
)

// fruitType constants matching FruitConf.json / Fruit6Conf.json. Shared by the
// get and index test blocks.
var (
	fruitTypeApple  = int32(protoconf.FruitType_FRUIT_TYPE_APPLE)
	fruitTypeOrange = int32(protoconf.FruitType_FRUIT_TYPE_ORANGE)
	fruitTypeBanana = int32(protoconf.FruitType_FRUIT_TYPE_BANANA)
)

// ---- FruitConf: leveled index (Price<ID>) finders ----

func Test_FruitConf_FindItem(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// price=10 belongs to APPLE item 1001
	items := conf.FindItem(10)
	if len(items) != 1 {
		t.Fatalf("FindItem(10): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 1001 {
		t.Errorf("FindItem(10): expected id=1001, got %d", items[0].GetId())
	}

	// price not present
	items = conf.FindItem(999)
	if len(items) != 0 {
		t.Errorf("FindItem(999): expected 0 items, got %d", len(items))
	}
}

func Test_FruitConf_FindFirstItem(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// price=20 belongs to APPLE item 1002
	item := conf.FindFirstItem(20)
	if item == nil {
		t.Fatal("FindFirstItem(20): expected non-nil, got nil")
	}
	if item.GetId() != 1002 {
		t.Errorf("FindFirstItem(20): expected id=1002, got %d", item.GetId())
	}

	// price not present
	item = conf.FindFirstItem(999)
	if item != nil {
		t.Errorf("FindFirstItem(999): expected nil, got %v", item)
	}
}

func Test_FruitConf_FindItemMap(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	m := conf.FindItemMap()
	if m == nil {
		t.Fatal("FindItemMap: returned nil")
	}
	// 6 items total, each with a unique price → 6 entries
	if len(m) != 6 {
		t.Errorf("FindItemMap: expected 6 entries, got %d", len(m))
	}
}

func Test_FruitConf_FindItem1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// ORANGE(2) -> price=15 -> item 2001
	items := conf.FindItem1(fruitTypeOrange, 15)
	if len(items) != 1 {
		t.Fatalf("FindItem1(orange, 15): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 2001 {
		t.Errorf("FindItem1(orange, 15): expected id=2001, got %d", items[0].GetId())
	}

	// wrong fruitType → nil slice
	items = conf.FindItem1(999, 15)
	if len(items) != 0 {
		t.Errorf("FindItem1(999, 15): expected 0 items, got %d", len(items))
	}

	// correct fruitType, wrong price → nil slice
	items = conf.FindItem1(fruitTypeOrange, 999)
	if len(items) != 0 {
		t.Errorf("FindItem1(orange, 999): expected 0 items, got %d", len(items))
	}
}

func Test_FruitConf_FindFirstItem1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// BANANA(3) -> price=8 -> item 3001
	item := conf.FindFirstItem1(fruitTypeBanana, 8)
	if item == nil {
		t.Fatal("FindFirstItem1(banana, 8): expected non-nil, got nil")
	}
	if item.GetId() != 3001 {
		t.Errorf("FindFirstItem1(banana, 8): expected id=3001, got %d", item.GetId())
	}

	// not found
	item = conf.FindFirstItem1(fruitTypeBanana, 999)
	if item != nil {
		t.Errorf("FindFirstItem1(banana, 999): expected nil, got %v", item)
	}
}

func Test_FruitConf_FindItemMap1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// APPLE(1) has 2 items → 2 price entries
	m := conf.FindItemMap1(fruitTypeApple)
	if len(m) != 2 {
		t.Errorf("FindItemMap1(apple): expected 2 entries, got %d", len(m))
	}

	// non-existent fruitType → nil map
	m = conf.FindItemMap1(999)
	if m != nil {
		t.Errorf("FindItemMap1(999): expected nil, got %v", m)
	}
}

// ---- FruitConf: ordered index (Price<ID>@OrderedFruit) finders ----

func Test_FruitConf_FindOrderedFruit(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// price=10 → APPLE item 1001
	items := conf.FindOrderedFruit(10)
	if len(items) != 1 {
		t.Fatalf("FindOrderedFruit(10): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 1001 {
		t.Errorf("FindOrderedFruit(10): expected id=1001, got %d", items[0].GetId())
	}

	// price not present
	items = conf.FindOrderedFruit(999)
	if len(items) != 0 {
		t.Errorf("FindOrderedFruit(999): expected 0 items, got %d", len(items))
	}
}

func Test_FruitConf_FindFirstOrderedFruit(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	item := conf.FindFirstOrderedFruit(25)
	if item == nil {
		t.Fatal("FindFirstOrderedFruit(25): expected non-nil, got nil")
	}
	if item.GetId() != 2002 {
		t.Errorf("FindFirstOrderedFruit(25): expected id=2002, got %d", item.GetId())
	}

	item = conf.FindFirstOrderedFruit(999)
	if item != nil {
		t.Errorf("FindFirstOrderedFruit(999): expected nil, got %v", item)
	}
}

func Test_FruitConf_FindOrderedFruitMap(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	m := conf.FindOrderedFruitMap()
	if m == nil {
		t.Fatal("FindOrderedFruitMap: returned nil")
	}
	if m.Size() != 6 {
		t.Errorf("FindOrderedFruitMap: expected size=6, got %d", m.Size())
	}
	// verify ascending order of keys
	prev := int32(-1)
	m.Range(func(key int32, _ []*protoconf.FruitConf_Fruit_Item) bool {
		if key < prev {
			t.Errorf("FindOrderedFruitMap: keys not in ascending order: %d after %d", key, prev)
		}
		prev = key
		return true
	})
}

func Test_FruitConf_FindOrderedFruit1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// ORANGE(2) -> price=25 -> item 2002
	items := conf.FindOrderedFruit1(fruitTypeOrange, 25)
	if len(items) != 1 {
		t.Fatalf("FindOrderedFruit1(orange, 25): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 2002 {
		t.Errorf("FindOrderedFruit1(orange, 25): expected id=2002, got %d", items[0].GetId())
	}

	// ORANGE(2) -> price=15 -> items, verify IDs are in ascending order
	items = conf.FindOrderedFruit1(fruitTypeOrange, 15)
	if len(items) != 1 {
		t.Fatalf("FindOrderedFruit1(orange, 15): expected 1 item, got %d", len(items))
	}
	for i := 1; i < len(items); i++ {
		if items[i].GetId() < items[i-1].GetId() {
			t.Errorf("FindOrderedFruit1(orange, 15): items not in ascending id order at index %d: id=%d after id=%d",
				i, items[i].GetId(), items[i-1].GetId())
		}
	}

	// wrong fruitType
	items = conf.FindOrderedFruit1(999, 25)
	if len(items) != 0 {
		t.Errorf("FindOrderedFruit1(999, 25): expected 0 items, got %d", len(items))
	}
}

func Test_FruitConf_FindFirstOrderedFruit1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// BANANA(3) -> price=12 -> item 3002
	item := conf.FindFirstOrderedFruit1(fruitTypeBanana, 12)
	if item == nil {
		t.Fatal("FindFirstOrderedFruit1(banana, 12): expected non-nil, got nil")
	}
	if item.GetId() != 3002 {
		t.Errorf("FindFirstOrderedFruit1(banana, 12): expected id=3002, got %d", item.GetId())
	}

	// not found
	item = conf.FindFirstOrderedFruit1(fruitTypeBanana, 999)
	if item != nil {
		t.Errorf("FindFirstOrderedFruit1(banana, 999): expected nil, got %v", item)
	}
}

// ---- Fruit6Conf (Go-specific extra) ----

func Test_Fruit6Conf_FindItem(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	// price=10 → APPLE item 1001
	items := conf.FindItem(10)
	if len(items) != 1 {
		t.Fatalf("FindItem(10): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 1001 {
		t.Errorf("FindItem(10): expected id=1001, got %d", items[0].GetId())
	}

	// price not present
	items = conf.FindItem(999)
	if len(items) != 0 {
		t.Errorf("FindItem(999): expected 0 items, got %d", len(items))
	}
}

func Test_Fruit6Conf_FindFirstItem(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	item := conf.FindFirstItem(20)
	if item == nil {
		t.Fatal("FindFirstItem(20): expected non-nil, got nil")
	}
	if item.GetId() != 1002 {
		t.Errorf("FindFirstItem(20): expected id=1002, got %d", item.GetId())
	}

	item = conf.FindFirstItem(999)
	if item != nil {
		t.Errorf("FindFirstItem(999): expected nil, got %v", item)
	}
}

func Test_Fruit6Conf_FindItemMap(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	m := conf.FindItemMap()
	if m == nil {
		t.Fatal("FindItemMap: returned nil")
	}
	if len(m) != 6 {
		t.Errorf("FindItemMap: expected 6 entries, got %d", len(m))
	}
}

func Test_Fruit6Conf_FindItem1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	// ORANGE(2) -> price=15 -> item 2001
	items := conf.FindItem1(fruitTypeOrange, 15)
	if len(items) != 3 {
		t.Fatalf("FindItem1(orange, 15): expected 3 items, got %d", len(items))
	}
	if items[0].GetId() != 2000 {
		t.Errorf("FindItem1(orange, 15): expected id=2000, got %d", items[0].GetId())
	}

	// wrong fruitType
	items = conf.FindItem1(999, 15)
	if len(items) != 0 {
		t.Errorf("FindItem1(999, 15): expected 0 items, got %d", len(items))
	}

	// correct fruitType, wrong price
	items = conf.FindItem1(fruitTypeOrange, 999)
	if len(items) != 0 {
		t.Errorf("FindItem1(orange, 999): expected 0 items, got %d", len(items))
	}
}

func Test_Fruit6Conf_FindFirstItem1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	// BANANA(3) -> price=8 -> item 3001
	item := conf.FindFirstItem1(fruitTypeBanana, 8)
	if item == nil {
		t.Fatal("FindFirstItem1(banana, 8): expected non-nil, got nil")
	}
	if item.GetId() != 3001 {
		t.Errorf("FindFirstItem1(banana, 8): expected id=3001, got %d", item.GetId())
	}

	item = conf.FindFirstItem1(fruitTypeBanana, 999)
	if item != nil {
		t.Errorf("FindFirstItem1(banana, 999): expected nil, got %v", item)
	}
}

func Test_Fruit6Conf_FindItemMap1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	m := conf.FindItemMap1(fruitTypeApple)
	if len(m) != 2 {
		t.Errorf("FindItemMap1(apple): expected 2 entries, got %d", len(m))
	}

	m = conf.FindItemMap1(999)
	if m != nil {
		t.Errorf("FindItemMap1(999): expected nil, got %v", m)
	}
}

func Test_Fruit6Conf_FindOrderedFruit(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	items := conf.FindOrderedFruit(10)
	if len(items) != 1 {
		t.Fatalf("FindOrderedFruit(10): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 1001 {
		t.Errorf("FindOrderedFruit(10): expected id=1001, got %d", items[0].GetId())
	}

	items = conf.FindOrderedFruit(999)
	if len(items) != 0 {
		t.Errorf("FindOrderedFruit(999): expected 0 items, got %d", len(items))
	}
}

func Test_Fruit6Conf_FindFirstOrderedFruit(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	item := conf.FindFirstOrderedFruit(25)
	if item == nil {
		t.Fatal("FindFirstOrderedFruit(25): expected non-nil, got nil")
	}
	if item.GetId() != 2002 {
		t.Errorf("FindFirstOrderedFruit(25): expected id=2002, got %d", item.GetId())
	}

	item = conf.FindFirstOrderedFruit(999)
	if item != nil {
		t.Errorf("FindFirstOrderedFruit(999): expected nil, got %v", item)
	}
}

func Test_Fruit6Conf_FindOrderedFruitMap(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	m := conf.FindOrderedFruitMap()
	if m == nil {
		t.Fatal("FindOrderedFruitMap: returned nil")
	}
	if m.Size() != 6 {
		t.Errorf("FindOrderedFruitMap: expected size=6, got %d", m.Size())
	}
	// verify ascending order of keys
	prev := int32(-1)
	m.Range(func(key int32, _ []*protoconf.Fruit6Conf_Fruit_Item) bool {
		if key < prev {
			t.Errorf("FindOrderedFruitMap: keys not in ascending order: %d after %d", key, prev)
		}
		prev = key
		return true
	})
}

func Test_Fruit6Conf_FindOrderedFruit1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	// ORANGE(2) -> price=25 -> item 2002
	items := conf.FindOrderedFruit1(fruitTypeOrange, 25)
	if len(items) != 1 {
		t.Fatalf("FindOrderedFruit1(orange, 25): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 2002 {
		t.Errorf("FindOrderedFruit1(orange, 25): expected id=2002, got %d", items[0].GetId())
	}

	// ORANGE(2) -> price=15 -> items, verify IDs are in ascending order
	items = conf.FindOrderedFruit1(fruitTypeOrange, 15)
	if len(items) != 3 {
		t.Fatalf("FindOrderedFruit1(orange, 15): expected 3 items, got %d", len(items))
	}
	for i := 1; i < len(items); i++ {
		if items[i].GetId() < items[i-1].GetId() {
			t.Errorf("FindOrderedFruit1(orange, 15): items not in ascending id order at index %d: id=%d after id=%d",
				i, items[i].GetId(), items[i-1].GetId())
		}
	}

	items = conf.FindOrderedFruit1(999, 25)
	if len(items) != 0 {
		t.Errorf("FindOrderedFruit1(999, 25): expected 0 items, got %d", len(items))
	}
}

func Test_Fruit6Conf_FindFirstOrderedFruit1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	// BANANA(3) -> price=12 -> item 3002
	item := conf.FindFirstOrderedFruit1(fruitTypeBanana, 12)
	if item == nil {
		t.Fatal("FindFirstOrderedFruit1(banana, 12): expected non-nil, got nil")
	}
	if item.GetId() != 3002 {
		t.Errorf("FindFirstOrderedFruit1(banana, 12): expected id=3002, got %d", item.GetId())
	}

	item = conf.FindFirstOrderedFruit1(fruitTypeBanana, 999)
	if item != nil {
		t.Errorf("FindFirstOrderedFruit1(banana, 999): expected nil, got %v", item)
	}
}

// ---- ItemConf: single-column index ----

func Test_ItemConf_FindItemInfoMap(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetItemConf()
	if conf == nil {
		t.Fatal("ItemConf is nil")
	}
	if len(conf.FindItemInfoMap()) == 0 {
		t.Fatal("FindItemInfoMap: expected non-empty map")
	}
}

// ---- ItemConf: multi-column (composite-key) indexes ----
// Mirrors smoke.ts cases 15 (ParamExtType ordered index) and 16 (AwardItem index).

// Test_ItemConf_FindParamExtType covers the ordered multi-column index
// (Param,ExtType)<ID>@ParamExtType: a sorted map whose composite keys round-trip
// through the point-lookup finder.
func Test_ItemConf_FindParamExtType(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetItemConf()
	if conf == nil {
		t.Fatal("ItemConf is nil")
	}

	// apple (id=1) has param_list [1,2,3] x extTypeList [APPLE, ORANGE] -> 6 buckets,
	// and it is the only item carrying both columns, so the map has exactly 6 entries.
	m := conf.FindParamExtTypeMap()
	if m == nil {
		t.Fatal("FindParamExtTypeMap: returned nil")
	}
	if m.Size() != 6 {
		t.Errorf("FindParamExtTypeMap: expected size=6, got %d", m.Size())
	}

	// Keys are sorted ascending by (param, extType); each key round-trips through
	// the point-lookup finder and agrees with the map's stored value list.
	prevParam := int32(-1)
	m.Range(func(key loader.ItemConf_OrderedIndex_ParamExtTypeKey, values []*protoconf.ItemConf_Item) bool {
		if key.Param < prevParam {
			t.Errorf("FindParamExtTypeMap: keys not ascending by param: %d after %d", key.Param, prevParam)
		}
		prevParam = key.Param
		if got := conf.FindParamExtType(key.Param, key.ExtType); len(got) != len(values) {
			t.Errorf("FindParamExtType(%d, %v): point-lookup (%d) disagrees with map iteration (%d)",
				key.Param, key.ExtType, len(got), len(values))
		}
		return true
	})

	// A concrete known bucket: (param=1, extType=APPLE) -> apple.
	items := conf.FindParamExtType(1, protoconf.FruitType_FRUIT_TYPE_APPLE)
	if len(items) != 1 {
		t.Fatalf("FindParamExtType(1, APPLE): expected 1 item, got %d", len(items))
	}
	if items[0].GetName() != "apple" {
		t.Errorf("FindParamExtType(1, APPLE): expected name=apple, got %s", items[0].GetName())
	}

	// Missing key -> empty slice.
	if items := conf.FindParamExtType(999, protoconf.FruitType_FRUIT_TYPE_APPLE); len(items) != 0 {
		t.Errorf("FindParamExtType(999, APPLE): expected 0 items, got %d", len(items))
	}
}

// Test_ItemConf_FindAwardItem covers the non-ordered multi-column index
// (ID,Name)<Type,UseEffectType>@AwardItem.
func Test_ItemConf_FindAwardItem(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetItemConf()
	if conf == nil {
		t.Fatal("ItemConf is nil")
	}

	m := conf.FindAwardItemMap()
	if m == nil {
		t.Fatal("FindAwardItemMap: returned nil")
	}

	// Every map entry round-trips through the point-lookup finder.
	for key, values := range m {
		if got := conf.FindAwardItem(key.Id, key.Name); len(got) != len(values) {
			t.Errorf("FindAwardItem(%d, %q): point-lookup (%d) disagrees with map iteration (%d)",
				key.Id, key.Name, len(got), len(values))
		}
	}

	// apple is keyed by (id=1, name="apple").
	items := conf.FindAwardItem(1, "apple")
	if len(items) != 1 {
		t.Fatalf("FindAwardItem(1, apple): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 1 {
		t.Errorf("FindAwardItem(1, apple): expected id=1, got %d", items[0].GetId())
	}
}
