using pb = global::Google.Protobuf;
using pbr = global::Google.Protobuf.Reflection;
namespace Tableau
{
    public static class Load
    {
        public delegate byte[] ReadFunc(string path);
        public delegate pb::IMessage? LoadFunc(pbr::MessageDescriptor desc, string dir, Format fmt, in MessagerOptions? options);

        public class BaseOptions
        {
            public bool? IgnoreUnknownFields { get; set; }
            public ReadFunc? ReadFunc { get; set; }
            public LoadFunc? LoadFunc { get; set; }
        }

        public class Options : BaseOptions
        {
            public Dictionary<string, MessagerOptions>? MessagerOptions { get; set; }
            public MessagerOptions ParseMessagerOptionsByName(string name)
            {
                var mopts = MessagerOptions?.TryGetValue(name, out var val) == true ? (MessagerOptions)val.Clone() : new MessagerOptions();
                mopts.IgnoreUnknownFields ??= IgnoreUnknownFields;
                mopts.ReadFunc ??= ReadFunc;
                mopts.LoadFunc ??= LoadFunc;
                return mopts;
            }
        }

        public class MessagerOptions : BaseOptions, ICloneable
        {
            public string? Path { get; set; }

            public object Clone()
            {
                return new MessagerOptions
                {
                    IgnoreUnknownFields = IgnoreUnknownFields,
                    ReadFunc = ReadFunc,
                    LoadFunc = LoadFunc,
                    Path = Path
                };
            }
        }

        public static pb::IMessage? LoadMessager(pbr::MessageDescriptor desc, string path, Format fmt, in MessagerOptions? options = null)
        {
            var readFunc = options?.ReadFunc ?? File.ReadAllBytes;
            byte[] content = readFunc(path);
            return Unmarshal(content, desc, fmt, options);
        }

        public static pb::IMessage? LoadMessagerInDir(pbr::MessageDescriptor desc, string dir, Format fmt, in MessagerOptions? options = null)
        {
            string name = desc.Name;
            string path = "";
            if (options?.Path != null)
            {
                path = options.Path;
                fmt = Util.GetFormat(path);
            }
            if (path == "")
            {
                string filename = name + Util.Format2Ext(fmt);
                path = Path.Combine(dir, filename);
            }
            var loadFunc = options?.LoadFunc ?? LoadMessager;
            return loadFunc(desc, path, fmt, options);
        }

        public static pb::IMessage? Unmarshal(byte[] content, pbr::MessageDescriptor desc, Format fmt, in MessagerOptions? options = null)
        {
            switch (fmt)
            {
                case Format.JSON:
                    var parser = new pb::JsonParser(
                        pb::JsonParser.Settings.Default.WithIgnoreUnknownFields(options?.IgnoreUnknownFields ?? false)
                    );
                    return parser.Parse(new StreamReader(new MemoryStream(content)), desc);
                case Format.Bin:
                    return desc.Parser.ParseFrom(content);
                default:
                    return null;
            }
        }
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

        public ref readonly Stats GetStats() => ref LoadStats;

        public abstract bool Load(string dir, Format fmt, in Load.MessagerOptions? options = null);

        public virtual pb::IMessage? Message() => null;

        protected virtual bool ProcessAfterLoad() => true;

        public virtual bool ProcessAfterLoadAll(in Hub hub) => true;
    }
}
