package loader_test

import (
	"testing"

	"github.com/tableauio/loader/test/go-tableau-loader/protoconf"
)

// ---- ActivityConf: nested ordered map ----
// getOrderedMap exposes the full sorted tree; each node carries its value
// (.Second) and the next-level sub-map (.First).

func Test_ActivityConf_OrderedMap(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetActivityConf()
	if conf == nil {
		t.Fatal("ActivityConf is nil")
	}
	orderedMap := conf.GetOrderedMap()
	if orderedMap == nil || orderedMap.Size() == 0 {
		t.Fatal("ActivityConf ordered map is empty")
	}

	// 1st level: keys ascending; node.Second is the Activity value.
	prev := uint64(0)
	for iter := orderedMap.Iterator(); iter.Next(); {
		key := iter.Key()
		if key < prev {
			t.Errorf("ordered map keys not ascending: %d after %d", key, prev)
		}
		prev = key

		subOrderedMap := iter.Value().First
		for iter2 := subOrderedMap.Iterator(); iter2.Next(); {
			_ = iter2.Key()
			_ = iter2.Value().Second
		}
	}
}

// ---- ItemConf: 1st-level ordered map ----

func Test_ItemConf_OrderedMap(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetItemConf()
	if conf == nil {
		t.Fatal("ItemConf is nil")
	}
	orderedMap := conf.GetOrderedMap()
	if orderedMap == nil || orderedMap.Size() == 0 {
		t.Fatal("ItemConf ordered map is empty")
	}

	// keys ascending by id; prev starts at 0 so iteration 1 (key >= 0) holds.
	prev := uint32(0)
	orderedMap.Range(func(key uint32, _ *protoconf.ItemConf_Item) bool {
		if key < prev {
			t.Errorf("ItemConf ordered map keys not ascending: %d after %d", key, prev)
		}
		prev = key
		return true
	})
}
