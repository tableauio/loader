package main

import (
	"github.com/tableauio/loader/test/go-tableau-loader/hub"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

func main() {
	err := hub.GetHub().Load("../testdata/conf/", format.JSON,
		load.IgnoreUnknownFields(),
		load.WithMessagerOptions(map[string]*load.MessagerOptions{
			"ItemConf": {
				Path: "../testdata/conf/ItemConf.json",
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	// // test mutable check
	// delete(hub.GetHub().GetActivityConf().Data().ActivityMap, 100001)
	// hub.GetHub().GetActivityConf().Data().ThemeName = "theme2"
	// time.Sleep(time.Minute)
}
