package main

import (
	"fmt"
	"time"

	"github.com/tableauio/loader/test/go-tableau-loader/hub"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/code"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/xerrors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
	"github.com/tableauio/tableau/store"
)

func main() {
	err := hub.GetHub().Load("../testdata/conf/", format.JSON,
		load.IgnoreUnknownFields(),
		load.Paths(map[string]string{
			"ItemConf": "../testdata/conf/ItemConf.json",
		}))
	if err != nil {
		panic(err)
	}

	for name, msger := range hub.GetHub().GetMessagerMap() {
		fmt.Printf("%s: duration: %v\n", name, msger.GetStats().Duration)
	}

	conf := hub.GetHub().GetActivityConf()
	if conf == nil {
		panic("ActivityConf is nil")
	}

	// error: not found
	if _, err := conf.Get3(100001, 1, 999); err != nil {
		if xerrors.Is(err, code.NotFound) {
			fmt.Println("error: not found:", err)
		}
	}

	// update and store
	chapter, err := conf.Get3(100001, 1, 2)
	if err != nil {
		panic(err)
	}
	chapter.SectionName = "updated section 2"
	err = hub.GetHub().Store("_out/", format.JSON,
		store.Pretty(true),
	)
	if err != nil {
		panic(err)
	}

	// OrderedMap
	orderedMap := conf.GetOrderedMap()
	for iter := orderedMap.Iterator(); iter.Next(); {
		key := iter.Key()
		value := iter.Value().Second
		fmt.Println("key:", key)
		fmt.Println("value:", value)
		fmt.Println()
		subOrderedMap := iter.Value().First
		for iter2 := subOrderedMap.Iterator(); iter2.Next(); {
			key2 := iter2.Key()
			value2 := iter2.Value().Second
			fmt.Println("key2:", key2)
			fmt.Println("value2:", value2)
			fmt.Println()
		}
	}
	fmt.Printf("specialItemName: %v\n", hub.GetHub().GetCustomItemConf().GetSpecialItemName())
	fmt.Printf("HeroBaseConf: %v\n", hub.GetHub().GetHeroBaseConf().Data().GetHeroMap())

	// patchconf
	err = hub.GetHub().Load("../testdata/conf/", format.JSON,
		load.IgnoreUnknownFields(),
		load.PatchDirs("../testdata/patchconf/"),
	)
	if err != nil {
		panic(err)
	}
	// print recursive patch conf
	fmt.Printf("RecursivePatchConf: %v\n", hub.GetHub().GetRecursivePatchConf().Data())

	// test immutable check
	delete(hub.GetHub().GetActivityConf().Data().ActivityMap, 100001)
	hub.GetHub().GetActivityConf().Data().ThemeName = "theme2"
	time.Sleep(time.Minute)
}
