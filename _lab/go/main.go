package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/tableauio/loader/_lab/go/tableau"
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

func Get[T tableau.Messager]() T {
	return tableau.Get[T](GetHub().Hub)
}

func main() {
	err := GetHub().Load("../../test/testdata/", nil, format.JSON, load.IgnoreUnknownFields(true))
	if err != nil {
		panic(err)
	}

	conf := Get[*tableau.ActivityConf]()
	if conf == nil {
		panic("ActivityConf is nil")
	}
	chapter, err := conf.Get3(100001, 1, 2)
	if err != nil {
		fmt.Println(err)
	}
	if err := conf.Check(GetHub().Hub); err != nil {
		panic(err)
	}
	fmt.Printf("ActivityConf: %v\n", chapter)

	debug()
	time.AfterFunc(time.Second*5, func() {
		err := GetHub().Load("../../test/testdata/", nil, format.JSON, load.IgnoreUnknownFields(true))
		if err != nil {
			panic(err)
		}
		debug()
	})
	time.Sleep(time.Second * 30)
}

func debug() {
	itemConf := Get[*tableau.ItemConf]() 
	if itemConf == nil {
		panic("ItemConf is nil")
	}
	fmt.Printf("ItemConf: %v\n", itemConf.Data())
}
