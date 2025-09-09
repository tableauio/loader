using pb = global::Google.Protobuf;
namespace Tableau
{
    internal class MessagerContainer(in Dictionary<string, Messager>? messagerMap = null)
    {
        public Dictionary<string, Messager> MessagerMap = messagerMap ?? [];
        public DateTime LastLoadedTime = DateTime.Now;
{{ range . }}        public {{ . }}? {{ . }} = InternalGet<{{ . }}>(messagerMap);
{{ end }}
        public T? Get<T>() where T : Messager, IMessagerName => InternalGet<T>(MessagerMap);

        private static T? InternalGet<T>(in Dictionary<string, Messager>? messagerMap) where T : Messager, IMessagerName =>
           messagerMap?.TryGetValue(T.Name(), out var messager) == true ? (T)messager : null;
    }

    internal class Atomic<T> where T : class
    {
        private T? _value;

        public T? Value
        {
            get => Interlocked.CompareExchange(ref _value, null, null);
            set => Interlocked.Exchange(ref _value, value);
        }
    }

    public class HubOptions
    {
        public Func<string, bool>? Filter { get; set; }
    }

    public class Hub(HubOptions? options = null)
    {
        private readonly Atomic<MessagerContainer> _messagerContainer = new();
        private readonly HubOptions? _options = options;

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

        public IReadOnlyDictionary<string, Messager>? GetMessagerMap() => _messagerContainer.Value?.MessagerMap;

        public void SetMessagerMap(in Dictionary<string, Messager> map) => _messagerContainer.Value = new MessagerContainer(map);

        public T? Get<T>() where T : Messager, IMessagerName => _messagerContainer.Value?.Get<T>();
{{ range . }}
        public {{ . }}? Get{{ . }}() => _messagerContainer.Value?.{{ . }};
{{ end }}
        public DateTime? GetLastLoadedTime() => _messagerContainer.Value?.LastLoadedTime;

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

    public class Registry
    {
        internal static readonly Dictionary<string, Func<Messager>> Registrar = [];

        public static void Register<T>() where T : Messager, IMessagerName, new() => Registrar[T.Name()] = () => new T();

        public static void Init()
        {
{{ range . }}            Register<{{ . }}>();
{{ end }}        }
    }
}
