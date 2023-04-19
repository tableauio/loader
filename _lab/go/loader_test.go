package main

import (
	"testing"

	"github.com/tableauio/loader/_lab/go/tableau"
	"github.com/tableauio/tableau/format"
)

func Test_Loader(t *testing.T) {
	err := GetHub().Load("../../testdata/", nil, format.JSON)
	if err != nil {
		panic(err)
	}

	conf := Get[*tableau.ActivityConf]()
	if conf == nil {
		panic("ActivityConf is nil")
	}
	chapter, err := conf.Get2(100001, 1)
	if err != nil {
		panic(err)
	}
	t.Logf("ActivityConf: %v", chapter)
}
