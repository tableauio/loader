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
	helper.GenerateCommonHeader(gen, g, version)
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
		g.P(helper.Indent(2), "public ", messager, "? Get", messager, "() => MessagerContainer.", messager, ";")
	}
	g.P(staticHubContent4)
	for _, messager := range messagers {
		g.P("            Register<Tableau.", messager, ">();")
	}
	g.P(staticHubContent5)
}

const staticHubContent1 = `using System;
using System.Collections.Generic;
using Google.Protobuf;
using System.IO;

namespace Tableau
{
    public enum Format
    {
        Unknown,
        JSON,
        Bin
    }

    public class LoadOptions
    {
        public bool IgnoreUnknownFields { get; set; } = false;
    }

    public interface IMessagerName
    {
        static abstract string Name();
    }

    public abstract class Messager
    {
        public class Stats
        {
            public TimeSpan Duration;
        }

        protected Stats LoadStats = new Stats();

        public ref Stats GetStats() => ref LoadStats;

        public abstract bool Load(string dir, Format fmt, in LoadOptions? options = null);

        protected virtual bool ProcessAfterLoad() => true;

        public virtual bool ProcessAfterLoadAll(in Hub hub) => true;

        internal bool LoadMessageByPath<T>(out T msg, string dir, Format fmt, in LoadOptions? options = null) where T : IMessage<T>, new()
        {
            msg = new T();
            string name = msg.Descriptor.Name;
            string path = Path.Combine(dir, name + Format2Ext(fmt));
            try
            {
                switch (fmt)
                {
                    case Format.JSON:
                        {
                            string content = File.ReadAllText(path);
                            var parser = options is null ? JsonParser.Default : new JsonParser(JsonParser.Settings.Default.WithIgnoreUnknownFields(options.IgnoreUnknownFields));
                            msg = parser.Parse<T>(content);
                            break;
                        }
                    case Format.Bin:
                        {
                            byte[] content = File.ReadAllBytes(path);
                            var parser = new MessageParser<T>(() => new T());
                            msg = parser.ParseFrom(content);
                            break;
                        }
                    default:
                        return false;
                }
            }
            catch (Exception)
            {
                return false;
            }
            return true;
        }

        internal string Format2Ext(Format fmt)
        {
            switch (fmt)
            {
                case Format.JSON: return ".json";
                case Format.Bin: return ".bin";
                default: return ".unknown";
            }
        }
    }

    internal class MessagerContainer
    {
        public Dictionary<string, Messager> MessagerMap;
        public DateTime LastLoadedTime;`

var staticHubContent2 = `
        public MessagerContainer(in Dictionary<string, Messager>? messagerMap = null)
        {
            MessagerMap = messagerMap ?? new Dictionary<string, Messager>();
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

    public class Hub
    {
        private MessagerContainer MessagerContainer = new MessagerContainer();
        private readonly HubOptions Options;

        public Hub(HubOptions? options = null) => Options = options ?? new HubOptions();

        public bool Load(string dir, Format fmt, in LoadOptions? options = null)
        {
            var messagerMap = NewMessagerMap();
            foreach (var messager in messagerMap.Values)
            {
                if (!messager.Load(dir, fmt, options))
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

        public ref Dictionary<string, Messager> GetMessagerMap() => ref MessagerContainer.MessagerMap;

        public void SetMessagerMap(in Dictionary<string, Messager> map) => MessagerContainer = new MessagerContainer(map);

        public T? Get<T>() where T : Messager, IMessagerName => MessagerContainer.Get<T>();`

const staticHubContent4 = `
        public DateTime GetLastLoadedTime() => MessagerContainer.LastLoadedTime;

        private Dictionary<string, Messager> NewMessagerMap()
        {
            var messagerMap = new Dictionary<string, Messager>();
            foreach (var kv in Registry.Registrar)
            {
                if (Options.Filter?.Invoke(kv.Key) ?? true)
                {
                    messagerMap[kv.Key] = kv.Value();
                }
            }
            return messagerMap;
        }
    }

    public class Registry
    {
        internal static readonly Dictionary<string, Func<Messager>> Registrar = new Dictionary<string, Func<Messager>>();

        public static void Register<T>() where T : Messager, IMessagerName, new() => Registrar[T.Name()] = () => new T();

        public static void Init()
        {`

const staticHubContent5 = `        }
    }
}`
