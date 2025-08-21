using pb = global::Google.Protobuf;
using pbr = global::Google.Protobuf.Reflection;
namespace Tableau
{
    public static class Load
    {
        public class Options
        {
            public bool IgnoreUnknownFields { get; set; }
            public Func<string, byte[]>? ReadFunc { get; set; }
        }

        public static bool LoadMessager(out pb::IMessage msg, pbr::MessageDescriptor desc, string dir, Format fmt, in Options? options = null)
        {
            string name = desc.Name;
            string path = Path.Combine(dir, name + Util.Format2Ext(fmt));
            try
            {
                var readFunc = options?.ReadFunc ?? File.ReadAllBytes;
                byte[] content = readFunc(path);
                return Unmarshal(content, out msg, desc, fmt, options);
            }
            catch (Exception)
            {
                msg = desc.Parser.ParseFrom(Array.Empty<byte>());
                return false;
            }
        }

        public static bool Unmarshal(byte[] content, out pb::IMessage msg, pbr::MessageDescriptor desc, Format fmt, in Options? options = null)
        {
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

        public abstract bool Load(string dir, Format fmt, in Load.Options? options = null);

        protected virtual bool ProcessAfterLoad() => true;

        public virtual bool ProcessAfterLoadAll(in Hub hub) => true;
    }
}
