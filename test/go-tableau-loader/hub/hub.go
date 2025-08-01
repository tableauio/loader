package hub

import (
	"context"
	"sync"
	"time"

	"github.com/tableauio/loader/test/go-tableau-loader/customconf"
	tableau "github.com/tableauio/loader/test/go-tableau-loader/protoconf/loader"
)

type MyHub struct {
	*tableau.Hub
}

var hubSingleton *MyHub
var once sync.Once

type messagerContainerKey struct{}

// GetHub return the singleton of MyHub
func GetHub() *MyHub {
	once.Do(func() {
		// new instance
		hubSingleton = &MyHub{
			Hub: tableau.NewHub(
				tableau.WithMutableCheck(&tableau.MutableCheck{
					Interval: 1 * time.Second,
				}),
				tableau.MessagerContainerProvider(func(ctx context.Context, hub *tableau.Hub) *tableau.MessagerContainer {
					if container, ok := ctx.Value(messagerContainerKey{}).(*tableau.MessagerContainer); ok {
						return container
					}
					return hub.GetMessagerContainer()
				}),
			),
		}
	})
	return hubSingleton
}

// NewContext returns a new Context that carries MessagerContainer.
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, messagerContainerKey{}, hubSingleton.GetMessagerContainer())
}

func (h *MyHub) GetCustomItemConf(ctx context.Context) *customconf.CustomItemConf {
	msger := h.GetMessager(ctx, customconf.CustomItemConfName)
	if msger != nil {
		if conf, ok := msger.(*customconf.CustomItemConf); ok {
			return conf
		}
	}
	return nil
}
