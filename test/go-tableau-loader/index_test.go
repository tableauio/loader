package main

import (
	"errors"
	"testing"

	"github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
)

// fruitType constants matching FruitConf.json / Fruit6Conf.json
var (
	fruitTypeApple  = int32(protoconf.FruitType_FRUIT_TYPE_APPLE)
	fruitTypeOrange = int32(protoconf.FruitType_FRUIT_TYPE_ORANGE)
	fruitTypeBanana = int32(protoconf.FruitType_FRUIT_TYPE_BANANA)
)

// ---- FruitConf ----

func Test_FruitConf_Get1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// found
	fruit, err := conf.Get1(fruitTypeApple)
	if err != nil {
		t.Fatalf("Get1(%d) unexpected error: %v", fruitTypeApple, err)
	}
	if fruit == nil {
		t.Fatal("Get1: returned nil fruit")
	}

	// not found
	_, err = conf.Get1(999)
	if err == nil {
		t.Fatal("Get1(999): expected ErrNotFound, got nil")
	}
	if !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get1(999): expected ErrNotFound, got: %v", err)
	}
}

func Test_FruitConf_Get2(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruitConf()

	// found: APPLE -> item 1001
	item, err := conf.Get2(fruitTypeApple, 1001)
	if err != nil {
		t.Fatalf("Get2(%d, 1001) unexpected error: %v", fruitTypeApple, err)
	}
	if item.GetId() != 1001 {
		t.Errorf("Get2: expected id=1001, got %d", item.GetId())
	}
	if item.GetPrice() != 10 {
		t.Errorf("Get2: expected price=10, got %d", item.GetPrice())
	}

	// not found: wrong item id
	_, err = conf.Get2(fruitTypeApple, 9999)
	if err == nil {
		t.Fatal("Get2(apple, 9999): expected ErrNotFound, got nil")
	}
	if !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get2(apple, 9999): expected ErrNotFound, got: %v", err)
	}

	// not found: wrong fruitType
	_, err = conf.Get2(999, 1001)
	if !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get2(999, 1001): expected ErrNotFound, got: %v", err)
	}
}

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

	// ORANGE(2) -> price=15 -> items [2001, 2002], verify IDs are in ascending order
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

// ---- Fruit6Conf ----

func Test_Fruit6Conf_Get1(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetFruit6Conf()

	// found
	fruit, err := conf.Get1(fruitTypeApple)
	if err != nil {
		t.Fatalf("Get1(%d) unexpected error: %v", fruitTypeApple, err)
	}
	if fruit == nil {
		t.Fatal("Get1: returned nil fruit")
	}

	// not found
	_, err = conf.Get1(999)
	if !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("Get1(999): expected ErrNotFound, got: %v", err)
	}
}

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
		t.Fatalf("FindItem1(orange, 15): expected 1 item, got %d", len(items))
	}
	if items[0].GetId() != 2000 {
		t.Errorf("FindItem1(orange, 15): expected id=2001, got %d", items[0].GetId())
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

	// ORANGE(2) -> price=15 -> items [2001, 2002], verify IDs are in ascending order
	items = conf.FindOrderedFruit1(fruitTypeOrange, 15)
	if len(items) != 3 {
		t.Fatalf("FindOrderedFruit1(orange, 15): expected 2 items, got %d", len(items))
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
