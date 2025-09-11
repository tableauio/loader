package hub

import (
	"sync"
	"time"

	"github.com/tableauio/loader/test/go-tableau-loader/customconf"
	tableau "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
)

type MyHub struct {
	*tableau.ContainerHub
}

var hubSingleton *MyHub
var once sync.Once

// GetHub return the singleton of MyHub
func GetHub() *MyHub {
	once.Do(func() {
		// new instance
		hubSingleton = &MyHub{
			ContainerHub: tableau.NewHub(
				tableau.WithMutableCheck(&tableau.MutableCheck{
					Interval: 1 * time.Second,
				}),
			),
		}
	})
	return hubSingleton
}

func (h *MyHub) GetCustomItemConf() *customconf.CustomItemConf {
	msger := h.GetMessager(customconf.CustomItemConfName)
	if msger != nil {
		if conf, ok := msger.(*customconf.CustomItemConf); ok {
			return conf
		}
	}
	return nil
}
