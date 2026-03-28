package main

import (
	"context"
	"errors"
	"testing"

	"github.com/tableauio/loader/test/go-tableau-loader/hub"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
	"github.com/tableauio/tableau/store"
)

func prepareHub(t *testing.T) *hub.MyHub {
	t.Helper()
	h := hub.NewMyHub()
	err := h.Load("../testdata/conf/", format.JSON,
		load.IgnoreUnknownFields(),
		load.WithMessagerOptions(map[string]*load.MessagerOptions{
			"ItemConf": {
				Path: "../testdata/conf/ItemConf.json",
			},
		}),
	)
	if err != nil {
		t.Fatalf("failed to load hub: %v", err)
	}
	return h
}

func Test_Load(t *testing.T) {
	h := prepareHub(t)
	for name, msger := range h.GetMessagerMap() {
		t.Logf("%s: duration: %v", name, msger.GetStats().Duration)
	}
}

func Test_ActivityConf_NotFound(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetActivityConf()
	if conf == nil {
		t.Fatal("ActivityConf is nil")
	}
	_, err := conf.Get3(100001, 1, 999)
	if err == nil {
		t.Fatal("expected ErrNotFound, got nil")
	}
	if !errors.Is(err, loader.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func Test_ActivityConf_UpdateAndStore(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetActivityConf()
	if conf == nil {
		t.Fatal("ActivityConf is nil")
	}
	chapter, err := conf.Get3(100001, 1, 2)
	if err != nil {
		t.Fatalf("Get3 failed: %v", err)
	}
	chapter.SectionName = "updated section 2"
	err = h.Store(t.TempDir(), format.JSON,
		store.Pretty(true),
	)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
}

func Test_ActivityConf_OrderedMap(t *testing.T) {
	h := prepareHub(t)
	conf := h.GetActivityConf()
	if conf == nil {
		t.Fatal("ActivityConf is nil")
	}
	orderedMap := conf.GetOrderedMap()
	for iter := orderedMap.Iterator(); iter.Next(); {
		key := iter.Key()
		value := iter.Value().Second
		t.Logf("key: %v, value: %v", key, value)
		subOrderedMap := iter.Value().First
		for iter2 := subOrderedMap.Iterator(); iter2.Next(); {
			key2 := iter2.Key()
			value2 := iter2.Value().Second
			t.Logf("  key2: %v, value2: %v", key2, value2)
		}
	}
}

func Test_CustomItemConf(t *testing.T) {
	h := prepareHub(t)
	customConf := h.GetCustomItemConf()
	if customConf == nil {
		t.Fatal("CustomItemConf is nil")
	}
	t.Logf("specialItemName: %v", customConf.GetSpecialItemName())
}

func Test_HeroBaseConf(t *testing.T) {
	h := prepareHub(t)
	heroConf := h.GetHeroBaseConf()
	if heroConf == nil {
		t.Fatal("HeroBaseConf is nil")
	}
	t.Logf("HeroBaseConf: %v", heroConf.Data().GetHeroMap())
}

func Test_Context(t *testing.T) {
	h := prepareHub(t)

	// Save current messager container to ctx.
	ctx := h.NewContext(context.Background())

	// Load again with patch.
	err := h.Load("../testdata/conf/", format.JSON,
		load.IgnoreUnknownFields(),
		load.PatchDirs("../testdata/patchconf/"),
	)
	if err != nil {
		t.Fatalf("failed to load with patch: %v", err)
	}

	t.Logf("RecursivePatchConf: %v", h.GetRecursivePatchConf().Data())
	t.Logf("PatchReplaceConf: %v", h.GetPatchReplaceConf().Data())
	t.Logf("PatchReplaceConf(from ctx): %v", h.FromContext(ctx).GetPatchReplaceConf().Data())
	t.Logf("PatchReplaceConf(from background): %v", h.FromContext(context.Background()).GetPatchReplaceConf().Data())
}
