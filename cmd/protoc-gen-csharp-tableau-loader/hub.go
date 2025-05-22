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
		g.P("            Register<Tableau.", messager, ">();")
	}
	g.P(staticHubContent2)
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

        public ref Stats GetStats() { return ref LoadStats; }

        public abstract bool Load(string dir, Format fmt, LoadOptions? options = null);

        protected virtual bool ProcessAfterLoad() { return true; }

        public virtual bool ProcessAfterLoadAll(in Hub hub) { return true; }

        internal bool LoadMessageByPath<T>(out T msg, string dir, Format fmt, LoadOptions? options = null) where T : Google.Protobuf.IMessage<T>, new()
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
                            var parser = Google.Protobuf.JsonParser.Default;
                            if (options != null)
                            {
                                parser = new Google.Protobuf.JsonParser(Google.Protobuf.JsonParser.Settings.Default.WithIgnoreUnknownFields(true));
                            }
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
        public Dictionary<string, Messager> MessagerMap = new Dictionary<string, Messager>();
        public DateTime LastLoadedTime;
    }

    public class HubOptions
    {
        public Func<string, bool>? Filter { get; set; }
    }

    public class Hub
    {
        private MessagerContainer MessagerContainer = new MessagerContainer();
        private readonly HubOptions Options;

        public Hub(HubOptions? options = null)
        {
            Options = options ?? new HubOptions();
        }

        public bool Load(string dir, Format fmt, LoadOptions? options = null)
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
            tmpHub.SetMessagerMap(MessagerContainer.MessagerMap);
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

        public ref Dictionary<string, Messager> GetMessagerMap()
        {
            return ref MessagerContainer.MessagerMap;
        }

        public void SetMessagerMap(in Dictionary<string, Messager> map)
        {
            MessagerContainer.MessagerMap = map;
            MessagerContainer.LastLoadedTime = DateTime.Now;
        }

        public T? Get<T>() where T : Messager, IMessagerName, new()
        {
            string name = T.Name();
            if (MessagerContainer.MessagerMap.TryGetValue(name, out var messager))
            {
                return (T)messager;
            }
            return default;
        }

        private Messager GetMessager(string name)
        {
            return GetMessagerMap()[name];
        }

        public DateTime GetLastLoadedTime()
        {
            return MessagerContainer.LastLoadedTime;
        }

        private Dictionary<string, Messager> NewMessagerMap()
        {
            var messagerMap = new Dictionary<string, Messager>();
            foreach (var kv in Registry.Registrar)
            {
                if (Options.Filter == null || Options.Filter(kv.Key))
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

        public static void Register<T>() where T : Messager, IMessagerName, new()
        {
            string name = T.Name();
            Registrar[name] = () => new T();
        }

        public static void Init()
        {`

const staticHubContent2 = `        }
    }
}`
