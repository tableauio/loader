package main

import (
	"fmt"
	"sync"

	tableau "github.com/tableauio/loader/test/protoconf/loader"
	"github.com/tableauio/loader/test/protoconf/loader/code"
	"github.com/tableauio/loader/test/protoconf/loader/xerrors"
	"github.com/tableauio/tableau/format"
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
	err := GetHub().Load("../testdata/", nil, format.JSON)
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
	if err := conf.Check(GetHub().Hub); err != nil {
		panic(err)
	}
}
