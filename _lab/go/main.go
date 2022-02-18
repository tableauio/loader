package main

import (
	"fmt"
	"sync"

	"github.com/tableauio/loader/_lab/go/tableau"
	"github.com/tableauio/tableau/options"
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
	err := GetHub().Load("../../test/testdata/", nil, options.JSON)
	if err != nil {
		panic(err)
	}

	conf := GetHub().GetActivityConf()
	if conf == nil {
		panic("ActivityConf is nil")
	}
	chapter, err := conf.Get3(100001, 1, 2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("ActivityConf: %v\n", chapter)
}
