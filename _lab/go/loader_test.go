package main

import (
	"testing"

	"github.com/tableauio/tableau/format"
)

func Test_Loader(t *testing.T) {
	err := GetHub().Load("../../testdata/", nil, format.JSON)
	if err != nil {
		panic(err)
	}

	conf := GetHub().GetActivityConf()
	if conf == nil {
		panic("ActivityConf is nil")
	}
	chapter, err := conf.Get2(100001, 1)
	if err != nil {
		panic(err)
	}
	t.Logf("ActivityConf: %v", chapter)
}
