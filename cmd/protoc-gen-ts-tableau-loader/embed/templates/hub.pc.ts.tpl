import { Messager, type MessagerCtor } from "./messager.pc.js";
import { Format } from "./util.pc.js";
import { type Options, parseMessagerOptionsByName } from "./load.pc.js";
{{- range .Imports }}
import { {{ range $i, $n := .Names }}{{ if $i }}, {{ end }}{{ $n }}{{ end }} } from "./{{ .Module }}.js";
{{- end }}

/**
 * HubOptions is the options for Hub.
 */
export interface HubOptions {
  /**
   * filter can only filter in certain specific messagers based on the
   * condition that you provide.
   */
  filter?: (name: string) => boolean;
}

/**
 * Registry manages the registration of all messager generators.
 */
export class Registry {
  static readonly registrar: Map<string, MessagerCtor> = new Map();

  /** register registers a messager constructor. */
  static register(ctor: MessagerCtor): void {
    Registry.registrar.set(new ctor().name(), ctor);
  }

  /** init registers all generated messagers. */
  static init(): void {
{{- range .Messagers }}
    Registry.register({{ . }});
{{- end }}
  }
}

Registry.init();

/**
 * Hub is the messager manager. It manages loading, accessing, and storing
 * all configuration messagers.
 */
export class Hub {
  private messagerMap: Map<string, Messager> = new Map();
  private lastLoadedTime: Date | undefined;
  private readonly options?: HubOptions;

  constructor(options?: HubOptions) {
    this.options = options;
  }

  /**
   * load fills messages from files in the specified directory and format.
   * Throws an Error (with the underlying cause chained) on failure.
   */
  load(dir: string, fmt: Format, options?: Options): void {
    const messagerMap = this.newMessagerMap();
    const opts = options ?? {};
    for (const [name, messager] of messagerMap) {
      try {
        messager.load(dir, fmt, parseMessagerOptionsByName(opts, name));
      } catch (e) {
        throw new Error(`load ${name} failed`, { cause: e });
      }
    }
    const tmpHub = new Hub();
    tmpHub.setMessagerMap(messagerMap);
    for (const [name, messager] of messagerMap) {
      try {
        messager.processAfterLoadAll(tmpHub);
      } catch (e) {
        throw new Error(`hub call processAfterLoadAll failed, messager: ${name}`, { cause: e });
      }
    }
    this.setMessagerMap(messagerMap);
  }

  /** getMessagerMap returns the current messager map. */
  getMessagerMap(): Map<string, Messager> {
    return this.messagerMap;
  }

  /** setMessagerMap sets the messager map. */
  setMessagerMap(map: Map<string, Messager>): void {
    this.messagerMap = map;
    this.lastLoadedTime = new Date();
  }

  /** getMessager returns the messager by name. */
  getMessager(name: string): Messager | undefined {
    return this.messagerMap.get(name);
  }

  /** getLastLoadedTime returns the time when hub's messager map was last set. */
  getLastLoadedTime(): Date | undefined {
    return this.lastLoadedTime;
  }

  private newMessagerMap(): Map<string, Messager> {
    const messagerMap = new Map<string, Messager>();
    for (const [name, ctor] of Registry.registrar) {
      if (!this.options?.filter || this.options.filter(name)) {
        messagerMap.set(name, new ctor());
      }
    }
    return messagerMap;
  }
{{ range .Messagers }}
  /** get{{ . }} returns the {{ . }} messager. */
  get{{ . }}(): {{ . }} | undefined {
    return this.messagerMap.get("{{ . }}") as {{ . }} | undefined;
  }
{{ end }}}
