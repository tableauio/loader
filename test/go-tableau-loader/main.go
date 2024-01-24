package main

import (
	"fmt"

	"github.com/tableauio/loader/test/go-tableau-loader/hub"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/code"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader/xerrors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

func main() {
	err := hub.GetHub().Load("../testdata/", nil, format.JSON,
		load.IgnoreUnknownFields(),
		load.Paths(map[string]string{
			"ItemConf": "../testdata/ItemConf.json",
		}))
	if err != nil {
		panic(err)
	}

	conf := hub.GetHub().GetActivityConf()
	if conf == nil {
		panic("ActivityConf is nil")
	}
	// chapter, err := conf.Get3(100001, 1, 9)
	chapter, err := conf.Get3(100001, 1, 2)
	if err != nil {
		if xerrors.Is(err, code.NotFound) {
			panic(err)
		}
	}
	fmt.Printf("ActivityConf: %v\n", chapter)
	fmt.Println()

	ordMap := conf.GetOrderedMap()
	for iter := ordMap.Iterator(); iter.Next(); {
		key := iter.Key()
		conf := iter.Value().Second
		fmt.Printf("Key1: %v\n", key)
		fmt.Printf("Conf1: %v\n", conf)
		fmt.Println()
		nextMap := iter.Value().First
		for iter2 := nextMap.Iterator(); iter2.Next(); {
			key2 := iter2.Key()
			conf2 := iter2.Value().Second
			fmt.Printf("Key2: %v\n", key2)
			fmt.Printf("Conf2: %v\n", conf2)
			fmt.Println()
		}
	}

	if err := conf.Check(hub.GetHub().Hub); err != nil {
		panic(err)
	}
	fmt.Printf("specialItemName: %v\n", hub.GetHub().GetCustomItemConf().GetSpecialItemName())
}
