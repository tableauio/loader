package main

import (
	"path/filepath"

	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"google.golang.org/protobuf/compiler/protogen"
)

// generateHub generates related hub files.
func generateHub(gen *protogen.Plugin) {
	filename := filepath.Join("Hub.cs")
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, nil, g, version)
	g.P(staticHubContent1)
	for _, messager := range messagers {
		g.P(helper.Indent(2), "public ", messager, "? ", messager, ";")
	}
	g.P(staticHubContent2)
	for _, messager := range messagers {
		g.P(helper.Indent(4), messager, " = Get<", messager, ">();")
	}
	g.P(staticHubContent3)
	for _, messager := range messagers {
		g.P()
		g.P(helper.Indent(2), "public ", messager, "? Get", messager, "() => _messagerContainer.", messager, ";")
	}
	g.P(staticHubContent4)
	for _, messager := range messagers {
		g.P("            Register<", messager, ">();")
	}
	g.P(staticHubContent5)
}

const staticHubContent1 = `using pb = global::Google.Protobuf;
namespace Tableau
{
    internal class MessagerContainer
    {
        public Dictionary<string, Messager> MessagerMap;
        public DateTime LastLoadedTime;`

var staticHubContent2 = `
        public MessagerContainer(in Dictionary<string, Messager>? messagerMap = null)
        {
            MessagerMap = messagerMap ?? [];
            LastLoadedTime = DateTime.Now;
            if (messagerMap != null)
            {`

var staticHubContent3 = `            }
        }

        public T? Get<T>() where T : Messager, IMessagerName => MessagerMap.TryGetValue(T.Name(), out var messager) ? (T)messager : null;
    }

    public class HubOptions
    {
        public Func<string, bool>? Filter { get; set; }
    }

    public class Hub(HubOptions? options = null)
    {
        private MessagerContainer _messagerContainer = new();
        private readonly HubOptions _options = options ?? new HubOptions();

        public bool Load(string dir, Format fmt, in Load.Options? options = null)
        {
            var messagerMap = NewMessagerMap();
            var opts = options ?? new Load.Options();
            foreach (var kvs in messagerMap)
            {
                string name = kvs.Key;
                if (!kvs.Value.Load(dir, fmt, opts.ParseMessagerOptionsByName(name)))
                {
                    return false;
                }
            }
            var tmpHub = new Hub();
            tmpHub.SetMessagerMap(messagerMap);
            foreach (var messager in messagerMap.Values)
            {
                if (!messager.ProcessAfterLoadAll(tmpHub))
                {
                    return false;
                }
            }
            SetMessagerMap(messagerMap);
            return true;
        }

        public ref Dictionary<string, Messager> GetMessagerMap() => ref _messagerContainer.MessagerMap;

        public void SetMessagerMap(in Dictionary<string, Messager> map) => _messagerContainer = new MessagerContainer(map);

        public T? Get<T>() where T : Messager, IMessagerName => _messagerContainer.Get<T>();`

const staticHubContent4 = `
        public DateTime GetLastLoadedTime() => _messagerContainer.LastLoadedTime;

        private Dictionary<string, Messager> NewMessagerMap()
        {
            var messagerMap = new Dictionary<string, Messager>();
            foreach (var kv in Registry.Registrar)
            {
                if (_options.Filter?.Invoke(kv.Key) ?? true)
                {
                    messagerMap[kv.Key] = kv.Value();
                }
            }
            return messagerMap;
        }
    }

    public class Registry
    {
        internal static readonly Dictionary<string, Func<Messager>> Registrar = [];

        public static void Register<T>() where T : Messager, IMessagerName, new() => Registrar[T.Name()] = () => new T();

        public static void Init()
        {`

const staticHubContent5 = `        }
    }
}`
