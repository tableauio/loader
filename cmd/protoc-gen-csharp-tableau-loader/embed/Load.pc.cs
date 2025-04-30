using pb = global::Google.Protobuf;
using pbr = global::Google.Protobuf.Reflection;
namespace Tableau
{
    /// <summary>
    /// Load provides functions for loading and parsing protobuf messages from files.
    /// </summary>
    public static class Load
    {
        /// <summary>
        /// ReadFunc reads the config file and returns its content.
        /// </summary>
        public delegate byte[] ReadFunc(string path);
        /// <summary>
        /// LoadFunc defines a func which can load message's content based on the
        /// given descriptor, path, format, and options.
        /// </summary>
        public delegate pb::IMessage? LoadFunc(pbr::MessageDescriptor desc, string dir, Format fmt, in MessagerOptions? options);

        /// <summary>
        /// BaseOptions is the common options for both global-level and messager-level options.
        /// </summary>
        public class BaseOptions
        {
            /// <summary>
            /// Whether to ignore unknown JSON fields during parsing.
            /// </summary>
            public bool? IgnoreUnknownFields { get; set; }
            /// <summary>
            /// You can specify custom read function to read a config file's content.
            /// Default is File.ReadAllBytes.
            /// </summary>
            public ReadFunc? ReadFunc { get; set; }
            /// <summary>
            /// You can specify custom load function to load a messager's content.
            /// Default is LoadMessager.
            /// </summary>
            public LoadFunc? LoadFunc { get; set; }
        }

        /// <summary>
        /// Options is the global-level options, which contains both global-level and
        /// messager-level options.
        /// </summary>
        public class Options : BaseOptions
        {
            /// <summary>
            /// MessagerOptions maps each messager name to a MessagerOptions.
            /// If specified, then the messager will be parsed with the given options directly.
            /// </summary>
            public Dictionary<string, MessagerOptions>? MessagerOptions { get; set; }
            /// <summary>
            /// ParseMessagerOptionsByName parses messager options with both global-level and
            /// messager-level options taken into consideration.
            /// </summary>
            public MessagerOptions ParseMessagerOptionsByName(string name)
            {
                var mopts = MessagerOptions?.TryGetValue(name, out var val) == true ? (MessagerOptions)val.Clone() : new MessagerOptions();
                mopts.IgnoreUnknownFields ??= IgnoreUnknownFields;
                mopts.ReadFunc ??= ReadFunc;
                mopts.LoadFunc ??= LoadFunc;
                return mopts;
            }
        }

        /// <summary>
        /// MessagerOptions defines the options for loading a messager.
        /// </summary>
        public class MessagerOptions : BaseOptions, ICloneable
        {
            /// <summary>
            /// Path maps each messager name to a corresponding config file path.
            /// If specified, then the main messager will be parsed from the file
            /// directly, other than the specified load dir.
            /// </summary>
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

        /// <summary>
        /// LoadMessager loads a protobuf message from the specified file path and format.
        /// </summary>
        public static pb::IMessage? LoadMessager(pbr::MessageDescriptor desc, string path, Format fmt, in MessagerOptions? options = null)
        {
            var readFunc = options?.ReadFunc ?? File.ReadAllBytes;
            byte[] content;
            try
            {
                content = readFunc(path);
            }
            catch (Exception ex)
            {
                Util.SetErrMsg($"failed to read {path}: {ex.Message}");
                throw;
            }
            return Unmarshal(content, desc, fmt, options);
        }

        /// <summary>
        /// LoadMessagerInDir loads a protobuf message from the specified directory and format.
        /// It resolves the file path based on the message descriptor name.
        /// </summary>
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

        /// <summary>
        /// Unmarshal parses the given byte content into a protobuf message based on the specified format.
        /// </summary>
        public static pb::IMessage? Unmarshal(byte[] content, pbr::MessageDescriptor desc, Format fmt, in MessagerOptions? options = null)
        {
            switch (fmt)
            {
                case Format.JSON:
                    try
                    {
                        var parser = new pb::JsonParser(
                            pb::JsonParser.Settings.Default.WithIgnoreUnknownFields(options?.IgnoreUnknownFields ?? false)
                        );
                        return parser.Parse(new StreamReader(new MemoryStream(content)), desc);
                    }
                    catch (Exception ex)
                    {
                        Util.SetErrMsg($"failed to parse {desc.Name}.json: {ex.Message}");
                        throw;
                    }
                case Format.Bin:
                    try
                    {
                        return desc.Parser.ParseFrom(content);
                    }
                    catch (Exception ex)
                    {
                        Util.SetErrMsg($"failed to parse {desc.Name}.binpb: {ex.Message}");
                        throw;
                    }
                default:
                    Util.SetErrMsg($"unknown format: {fmt}");
                    return null;
            }
        }
    }

    /// <summary>
    /// IMessagerName is an interface that provides the static Name() method for messagers.
    /// </summary>
    public interface IMessagerName
    {
        static abstract string Name();
    }

    /// <summary>
    /// Messager is the base class for all generated configuration messagers.
    /// It is designed for three goals:
    ///   1. Easy use: simple yet powerful accessors.
    ///   2. Elegant API: concise and clean functions.
    ///   3. Extensibility: Map, OrderedMap, Index, OrderedIndex...
    /// </summary>
    public abstract class Messager
    {
        /// <summary>
        /// Stats contains statistics info about loading.
        /// </summary>
        public class Stats
        {
            /// <summary>Total load time consuming.</summary>
            public TimeSpan Duration;
        }

        protected Stats LoadStats = new();

        /// <summary>
        /// GetStats returns the loading stats info.
        /// </summary>
        public ref readonly Stats GetStats() => ref LoadStats;

        /// <summary>
        /// Load fills message from file in the specified directory and format.
        /// </summary>
        public abstract bool Load(string dir, Format fmt, in Load.MessagerOptions? options = null);

        /// <summary>
        /// Message returns the inner protobuf message data.
        /// </summary>
        public virtual pb::IMessage? Message() => null;

        /// <summary>
        /// ProcessAfterLoad is invoked after this messager loaded.
        /// </summary>
        protected virtual bool ProcessAfterLoad() => true;

        /// <summary>
        /// ProcessAfterLoadAll is invoked after all messagers loaded.
        /// </summary>
        public virtual bool ProcessAfterLoadAll(in Hub hub) => true;
    }
}