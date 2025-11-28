package customconf

import (
	"fmt"

	"github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	tableau "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
)

const CustomItemConfName = "CustomItemConf"

type CustomItemConf struct {
	tableau.UnimplementedMessager
	specialItemConf *protoconf.ItemConf_Item
}

func (x *CustomItemConf) Name() string {
	return CustomItemConfName
}

func (x *CustomItemConf) ProcessAfterLoadAll(hub tableau.MessagerContainer) error {
	config, err := hub.GetItemConf().Get1(1)
	if err != nil {
		return err
	}
	x.specialItemConf = config
	fmt.Println("custom item conf processed")
	return nil
}

func (x *CustomItemConf) GetSpecialItemName() string {
	return x.specialItemConf.GetName()
}

func init() {
	tableau.Register(func() tableau.Messager {
		return new(CustomItemConf)
	})
}
