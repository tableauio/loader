package loader_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tableauio/loader/test/go-tableau-loader/hub"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
	"github.com/tableauio/tableau/store"
	"google.golang.org/protobuf/proto"
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

// Test_Patch mirrors the patch tests in cpp-tableau-loader/src/main.cpp::TestPatch
// and csharp-tableau-loader/Program.cs::TestPatch to verify the Go patch logic.
func Test_Patch(t *testing.T) {
	const testdataDir = "../testdata"

	t.Run("patchconf", func(t *testing.T) {
		h := hub.NewMyHub()
		err := h.Load(testdataDir+"/conf/", format.JSON,
			load.IgnoreUnknownFields(),
			load.PatchDirs(testdataDir+"/patchconf/"),
		)
		if err != nil {
			t.Fatalf("failed to load with patchconf: %v", err)
		}

		mgr := h.GetRecursivePatchConf()
		if mgr == nil {
			t.Fatal("RecursivePatchConf is nil")
		}
		t.Logf("RecursivePatchConf: %v", mgr.Data())

		// Verify against the expected patch result.
		expected := &loader.RecursivePatchConf{}
		if err := expected.Load(testdataDir+"/patchresult/", format.JSON, nil); err != nil {
			t.Fatalf("failed to load patch result: %v", err)
		}
		t.Logf("Expected patch result: %v", expected.Data())
		if !proto.Equal(mgr.Data(), expected.Data()) {
			t.Fatalf("patch result not correct:\n got:      %v\n expected: %v", mgr.Data(), expected.Data())
		}

		t.Logf("PatchReplaceConf: %v", h.GetPatchReplaceConf().Data())
		t.Logf("PatchMergeConf: %v", h.GetPatchMergeConf().Data())
	})

	t.Run("patchconf2", func(t *testing.T) {
		h := hub.NewMyHub()
		err := h.Load(testdataDir+"/conf/", format.JSON,
			load.IgnoreUnknownFields(),
			load.PatchDirs(testdataDir+"/patchconf2/"),
		)
		if err != nil {
			t.Fatalf("failed to load with patchconf2: %v", err)
		}
		t.Logf("PatchMergeConf(patchconf2): %v", h.GetPatchMergeConf().Data())
	})

	t.Run("patchconf2-different-format", func(t *testing.T) {
		// patch_dirs uses .json for resolution, but messager_options overrides
		// PatchMergeConf with a .txtpb patch path.
		h := hub.NewMyHub()
		err := h.Load(testdataDir+"/conf/", format.JSON,
			load.IgnoreUnknownFields(),
			load.PatchDirs(testdataDir+"/patchconf2/"),
			load.WithMessagerOptions(map[string]*load.MessagerOptions{
				"PatchMergeConf": {
					PatchPaths: []string{testdataDir + "/patchconf2/PatchMergeConf.txtpb"},
				},
			}),
		)
		if err != nil {
			t.Fatalf("failed to load with patchconf2 (txtpb): %v", err)
		}
		t.Logf("PatchMergeConf(txtpb): %v", h.GetPatchMergeConf().Data())
	})

	t.Run("multiple-patch-files", func(t *testing.T) {
		h := hub.NewMyHub()
		err := h.Load(testdataDir+"/conf/", format.JSON,
			load.IgnoreUnknownFields(),
			load.WithMessagerOptions(map[string]*load.MessagerOptions{
				"PatchMergeConf": {
					PatchPaths: []string{
						testdataDir + "/patchconf/PatchMergeConf.json",
						testdataDir + "/patchconf2/PatchMergeConf.json",
					},
				},
			}),
		)
		if err != nil {
			t.Fatalf("failed to load with multiple patch files: %v", err)
		}
		got := h.GetPatchMergeConf().Data()
		t.Logf("PatchMergeConf(multi-patch): %v", got)

		// Sanity: 'item_map' should contain the merged entry from patchconf2 (id=999).
		if _, ok := got.GetItemMap()[999]; !ok {
			t.Fatalf("expected ItemMap to contain key 999 from patchconf2, got: %v", got.GetItemMap())
		}
		// 'replace_item_map' has PATCH_REPLACE field-level option, so the last
		// patch should fully replace any prior content (key 999 must remain).
		if _, ok := got.GetReplaceItemMap()[999]; !ok {
			t.Fatalf("expected ReplaceItemMap to contain key 999, got: %v", got.GetReplaceItemMap())
		}
	})

	t.Run("ModeOnlyMain", func(t *testing.T) {
		h := hub.NewMyHub()
		err := h.Load(testdataDir+"/conf/", format.JSON,
			load.IgnoreUnknownFields(),
			load.Mode(load.ModeOnlyMain),
			load.WithMessagerOptions(map[string]*load.MessagerOptions{
				"PatchMergeConf": {
					PatchPaths: []string{
						testdataDir + "/patchconf/PatchMergeConf.json",
						testdataDir + "/patchconf2/PatchMergeConf.json",
					},
				},
			}),
		)
		if err != nil {
			t.Fatalf("failed to load with ModeOnlyMain: %v", err)
		}

		got := h.GetPatchMergeConf().Data()
		t.Logf("PatchMergeConf(OnlyMain): %v", got)

		// Compare against the raw main file (no patches applied).
		mainMgr := &loader.PatchMergeConf{}
		if err := mainMgr.Load(testdataDir+"/conf/", format.JSON, &load.MessagerOptions{
			BaseOptions: load.BaseOptions{Mode: modeRef(load.ModeOnlyMain)},
		}); err != nil {
			t.Fatalf("failed to load main file: %v", err)
		}
		if !proto.Equal(got, mainMgr.Data()) {
			t.Fatalf("ModeOnlyMain should equal raw main file:\n got:      %v\n expected: %v", got, mainMgr.Data())
		}
	})

	t.Run("ModeOnlyPatch", func(t *testing.T) {
		h := hub.NewMyHub()
		err := h.Load(testdataDir+"/conf/", format.JSON,
			load.IgnoreUnknownFields(),
			load.Mode(load.ModeOnlyPatch),
			load.WithMessagerOptions(map[string]*load.MessagerOptions{
				"PatchMergeConf": {
					PatchPaths: []string{
						testdataDir + "/patchconf/PatchMergeConf.json",
						testdataDir + "/patchconf2/PatchMergeConf.json",
					},
				},
			}),
		)
		if err != nil {
			t.Fatalf("failed to load with ModeOnlyPatch: %v", err)
		}
		got := h.GetPatchMergeConf().Data()
		t.Logf("PatchMergeConf(OnlyPatch): %v", got)

		// Without main-file content, 'name' should be the value coming from the last
		// non-replace merge (still merged from patches), and replace_* fields should
		// equal the last patch only.
		if got.GetName() == "" {
			t.Fatalf("expected non-empty Name from patches, got: %q", got.GetName())
		}
	})
}

func modeRef(m load.LoadMode) *load.LoadMode { return &m }
