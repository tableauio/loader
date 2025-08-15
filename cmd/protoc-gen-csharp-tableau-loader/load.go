package main

import (
	"path/filepath"

	"github.com/tableauio/loader/cmd/protoc-gen-csharp-tableau-loader/helper"
	"google.golang.org/protobuf/compiler/protogen"
)

// generateLoad generates related load files.
func generateLoad(gen *protogen.Plugin) {
	filename := filepath.Join("Load.cs")
	g := gen.NewGeneratedFile(filename, "")
	helper.GenerateFileHeader(gen, nil, g, version)
	g.P(staticLoadContent1)
}

const staticLoadContent1 = `using pb = global::Google.Protobuf;
using pbr = global::Google.Protobuf.Reflection;
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
        public Func<string, byte[]> ReadFunc { get; set; } = File.ReadAllBytes;
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

        protected Stats LoadStats = new();

        public ref Stats GetStats() => ref LoadStats;

        public abstract bool Load(string dir, Format fmt, in LoadOptions? options = null);

        protected virtual bool ProcessAfterLoad() => true;

        public virtual bool ProcessAfterLoadAll(in Hub hub) => true;
    }

    public static class MessageParser
    {
        public static bool LoadMessageByPath(out pb::IMessage msg, pbr::MessageDescriptor desc, string dir, Format fmt, in LoadOptions? options = null)
        {
            string name = desc.Name;
            string path = Path.Combine(dir, name + Format2Ext(fmt));
            try
            {
                var readFunc = options?.ReadFunc ?? File.ReadAllBytes;
                byte[] content = readFunc(path);

                switch (fmt)
                {
                    case Format.JSON:
                        var parser = options is null
                            ? pb::JsonParser.Default
                            : new pb::JsonParser(pb::JsonParser.Settings.Default.WithIgnoreUnknownFields(options.IgnoreUnknownFields));
                        msg = parser.Parse(new StreamReader(new MemoryStream(content)), desc);
                        return true;
                    case Format.Bin:
                        msg = desc.Parser.ParseFrom(content);
                        return true;
                    default:
                        msg = desc.Parser.ParseFrom(Array.Empty<byte>());
                        return false;
                }
            }
            catch (Exception)
            {
                msg = desc.Parser.ParseFrom(Array.Empty<byte>());
                return false;
            }
        }

        public static string Format2Ext(Format fmt)
        {
            return fmt switch
            {
                Format.JSON => ".json",
                Format.Bin => ".bin",
                _ => ".unknown",
            };
        }
    }
}`
