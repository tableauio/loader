package hub

import (
	"testing"

	tableau "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

func prepareHubForTest() *tableau.Hub {
	hub := tableau.NewHub()
	err := hub.Load("../../testdata/conf/", format.JSON,
		load.IgnoreUnknownFields(),
		load.Paths(map[string]string{
			"ItemConf": "../../testdata/conf/ItemConf.json",
		}))
	if err != nil {
		panic(err)
	}
	return hub
}

func Benchmark_GetMessager(b *testing.B) {
	// var once sync.Once
	hub := prepareHubForTest()
	for i := 0; i < b.N; i++ {
		_ = hub.GetMessager("ItemConf").(*tableau.ItemConf)
		// once.Do(func() { fmt.Println(msger.Data()) })
	}
}

func Benchmark_GetItemConf(b *testing.B) {
	// var once sync.Once
	hub := prepareHubForTest()
	for i := 0; i < b.N; i++ {
		_ = hub.GetItemConf()
		// once.Do(func() { fmt.Println(msger.Data()) })
	}
}
