package loader_test

import (
	"context"
	"testing"

	"github.com/tableauio/loader/_lab/go127/protoconf/loader"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

func prepareHub(t *testing.T) *loader.Hub {
	t.Helper()
	h := loader.NewHub()
	err := h.Load("../../test/testdata/conf/", format.JSON,
		load.IgnoreUnknownFields(),
		load.WithMessagerOptions(map[string]*load.MessagerOptions{
			"ItemConf": {
				Path: "../../test/testdata/conf/ItemConf.json",
			},
		}),
	)
	if err != nil {
		t.Fatalf("failed to load hub: %v", err)
	}
	return h
}

func Test_Hub_Get_ItemConf(t *testing.T) {
	h := prepareHub(t)

	typed := h.GetItemConf()
	if typed == nil {
		t.Fatal("GetItemConf() returned nil")
	}

	got := h.Get[*loader.ItemConf]()
	if got == nil {
		t.Fatal("Get[*ItemConf]() returned nil")
	}
	if got != typed {
		t.Fatalf("Get[*ItemConf]() != GetItemConf(): %p vs %p", got, typed)
	}
}

func Test_MessagerContainer_Get_FromContext(t *testing.T) {
	h := prepareHub(t)
	ctx := h.NewContext(context.Background())
	mc := h.FromContext(ctx)

	typed := mc.GetItemConf()
	got := mc.Get[*loader.ItemConf]()
	if got == nil {
		t.Fatal("MessagerContainer.Get[*ItemConf]() returned nil")
	}
	if got != typed {
		t.Fatalf("container Get[*ItemConf]() != GetItemConf(): %p vs %p", got, typed)
	}
}
