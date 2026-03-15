using System;
using System.Collections.Generic;
using System.Threading;
using pb = global::Google.Protobuf;
namespace Tableau
{
    /// <summary>
    /// MessagerContainer holds all messager instances and provides fast access.
    /// </summary>
    internal class MessagerContainer
    {
        public Dictionary<string, Messager> MessagerMap;
        public DateTime LastLoadedTime;
{{ range . }}        public {{ . }}? {{ . }};
{{ end }}
        public MessagerContainer(Dictionary<string, Messager>? messagerMap = null)
        {
            MessagerMap = messagerMap ?? new Dictionary<string, Messager>();
            LastLoadedTime = DateTime.Now;
{{ range . }}            {{ . }} = InternalGet<{{ . }}>(messagerMap);
{{ end }}        }

        /// <summary>
        /// Get returns the messager of type T from the container.
        /// </summary>
        public T? Get<T>() where T : Messager, IMessagerName, new() => InternalGet<T>(MessagerMap);

        private static T? InternalGet<T>(Dictionary<string, Messager>? messagerMap) where T : Messager, IMessagerName, new()
        {
            if (messagerMap == null) return null;
            string name = new T().Name();
            return messagerMap.TryGetValue(name, out var messager) ? (T)messager : null;
        }
    }

    /// <summary>
    /// Atomic provides a thread-safe wrapper for reference types.
    /// </summary>
    internal class Atomic<T> where T : class
    {
        private T? _value;

        public T? Value
        {
            get => Interlocked.CompareExchange(ref _value, null, null);
            set => Interlocked.Exchange(ref _value, value);
        }
    }

    /// <summary>
    /// HubOptions is the options for Hub.
    /// </summary>
    public class HubOptions
    {
        /// <summary>
        /// Filter can only filter in certain specific messagers based on the
        /// condition that you provide.
        /// </summary>
        public Func<string, bool>? Filter { get; set; }
    }

    /// <summary>
    /// Hub is the messager manager. It manages loading, accessing, and storing
    /// all configuration messagers.
    /// </summary>
    public class Hub
    {
        private readonly Atomic<MessagerContainer> _messagerContainer = new Atomic<MessagerContainer>();
        private readonly HubOptions? _options;

        public Hub(HubOptions? options = null)
        {
            _options = options;
        }

        /// <summary>
        /// Load fills messages from files in the specified directory and format.
        /// </summary>
        public bool Load(string dir, Format fmt, in Load.Options? options = null)
        {
            var messagerMap = NewMessagerMap();
            var opts = options ?? new Load.Options();
            foreach (var kvs in messagerMap)
            {
                string name = kvs.Key;
                if (!kvs.Value.Load(dir, fmt, opts.ParseMessagerOptionsByName(name)))
                {
                    Console.Error.WriteLine($"load {name} failed: {Util.GetErrMsg()}");
                    return false;
                }
            }
            var tmpHub = new Hub();
            tmpHub.SetMessagerMap(messagerMap);
            foreach (var messager in messagerMap)
            {
                if (!messager.Value.ProcessAfterLoadAll(tmpHub))
                {
                    Console.Error.WriteLine($"hub call ProcessAfterLoadAll failed, messager: {messager.Key}");
                    return false;
                }
            }
            SetMessagerMap(messagerMap);
            return true;
        }

        /// <summary>
        /// GetMessagerMap returns the current messager map.
        /// </summary>
        public IReadOnlyDictionary<string, Messager>? GetMessagerMap() => _messagerContainer.Value?.MessagerMap;

        /// <summary>
        /// SetMessagerMap sets the messager map with thread-safe guarantee.
        /// </summary>
        public void SetMessagerMap(in Dictionary<string, Messager> map) => _messagerContainer.Value = new MessagerContainer(map);

        /// <summary>
        /// Get returns the messager of type T from the hub.
        /// </summary>
        public T? Get<T>() where T : Messager, IMessagerName, new() => _messagerContainer.Value?.Get<T>();
{{ range . }}
        public {{ . }}? Get{{ . }}() => _messagerContainer.Value?.{{ . }};
{{ end }}
        /// <summary>
        /// GetLastLoadedTime returns the time when hub's messager container was last set.
        /// </summary>
        public DateTime? GetLastLoadedTime() => _messagerContainer.Value?.LastLoadedTime;

        /// <summary>
        /// NewMessagerMap creates a new MessagerMap based on the registered messagers.
        /// </summary>
        private Dictionary<string, Messager> NewMessagerMap()
        {
            var messagerMap = new Dictionary<string, Messager>();
            foreach (var kv in Registry.Registrar)
            {
                if (_options?.Filter?.Invoke(kv.Key) ?? true)
                {
                    messagerMap[kv.Key] = kv.Value();
                }
            }
            return messagerMap;
        }
    }

    /// <summary>
    /// Registry manages the registration of all messager generators.
    /// </summary>
    public class Registry
    {
        internal static readonly Dictionary<string, Func<Messager>> Registrar = new Dictionary<string, Func<Messager>>();

        /// <summary>
        /// Register registers a messager generator for type T.
        /// </summary>
        public static void Register<T>() where T : Messager, IMessagerName, new() => Registrar[new T().Name()] = () => new T();

        /// <summary>
        /// Init registers all generated messagers.
        /// </summary>
        public static void Init()
        {
{{ range . }}            Register<{{ . }}>();
{{ end }}        }
    }
}
