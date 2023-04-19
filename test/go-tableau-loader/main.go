package main

import (
	"fmt"
	"sync"

	tableau "github.com/tableauio/loader/test/protoconf/loader"
	"github.com/tableauio/loader/test/protoconf/loader/code"
	"github.com/tableauio/loader/test/protoconf/loader/xerrors"
	"github.com/tableauio/tableau/format"
	"github.com/tableauio/tableau/load"
)

type MyHub struct {
	*tableau.Hub
}

var hubSingleton *MyHub
var once sync.Once

// GetHub return the singleton of MyHub
func GetHub() *MyHub {
	once.Do(func() {
		// new instance
		hubSingleton = &MyHub{
			Hub: tableau.NewHub(),
		}
	})
	return hubSingleton
}

func main() {
	err := GetHub().Load("../testdata/", nil, format.JSON, load.IgnoreUnknownFields(true))
	if err != nil {
		panic(err)
	}

	conf := GetHub().GetActivityConf()
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

	if err := conf.Check(GetHub().Hub); err != nil {
		panic(err)
	}
}
