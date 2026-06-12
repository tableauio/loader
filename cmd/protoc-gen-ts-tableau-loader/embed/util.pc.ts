import { clone, getExtension, type DescField, type DescMessage, type Message } from "@bufbuild/protobuf";
import * as tableaupb from "../tableau/protobuf/tableau_pb.js";

/**
 * Format specifies the format of the configuration file.
 */
export enum Format {
  UNKNOWN = "unknown",
  JSON = "json",
  BIN = "bin",
}

const UNKNOWN_EXT = ".unknown";
const JSON_EXT = ".json";
const BIN_EXT = ".binpb";

/**
 * getFormat returns the Format determined by the file extension of the given path.
 */
export function getFormat(path: string): Format {
  if (path.endsWith(JSON_EXT)) return Format.JSON;
  if (path.endsWith(BIN_EXT)) return Format.BIN;
  return Format.UNKNOWN;
}

/**
 * format2Ext returns the file extension corresponding to the given format.
 */
export function format2Ext(fmt: Format): string {
  switch (fmt) {
    case Format.JSON:
      return JSON_EXT;
    case Format.BIN:
      return BIN_EXT;
    default:
      return UNKNOWN_EXT;
  }
}

// AnyMessage is a structural view over a protobuf-es message used by the
// reflection-based patch routines below.
type AnyMessage = Record<string, unknown>;

/**
 * getSheetPatch returns the sheet-level patch type for a message descriptor.
 */
export function getSheetPatch(desc: DescMessage): tableaupb.Patch {
  const opts = desc.proto.options;
  if (!opts) return tableaupb.Patch.NONE;
  const ws = getExtension(opts, tableaupb.worksheet);
  return ws?.patch ?? tableaupb.Patch.NONE;
}

function getFieldPatch(fd: DescField): tableaupb.Patch {
  const opts = fd.proto.options;
  if (!opts) return tableaupb.Patch.NONE;
  const fieldOpts = getExtension(opts, tableaupb.field);
  return fieldOpts?.prop?.patch ?? tableaupb.Patch.NONE;
}

function isFieldPopulated(msg: AnyMessage, fd: DescField): boolean {
  const v = msg[fd.localName];
  if (v === undefined || v === null) return false;
  switch (fd.fieldKind) {
    case "list":
      return Array.isArray(v) && v.length > 0;
    case "map":
      return typeof v === "object" && Object.keys(v as object).length > 0;
    case "message":
      return true;
    default:
      if (typeof v === "bigint") return v !== 0n;
      if (typeof v === "number") return v !== 0;
      if (typeof v === "string") return v.length > 0;
      if (typeof v === "boolean") return v;
      if (v instanceof Uint8Array) return v.length > 0;
      return true;
  }
}

function clearField(msg: AnyMessage, fd: DescField): void {
  switch (fd.fieldKind) {
    case "list":
      msg[fd.localName] = [];
      break;
    case "map":
      msg[fd.localName] = {};
      break;
    default:
      delete msg[fd.localName];
      break;
  }
}

/**
 * patchMessage patches src into dst, which must be messages with the same descriptor.
 *
 * Default mechanism:
 *   - scalar: populated scalar fields in src are copied to dst.
 *   - message: populated singular messages in src are merged into dst recursively,
 *     or replace dst message if PATCH_REPLACE is specified for the field.
 *   - list: src elements are appended to dst, or replace dst list on PATCH_REPLACE.
 *   - map: src entries are merged into dst, or replace dst map on PATCH_REPLACE.
 */
export function patchMessage(dst: Message, src: Message, desc: DescMessage): void {
  patchMessageInternal(dst as unknown as AnyMessage, src as unknown as AnyMessage, desc);
}

function patchMessageInternal(dst: AnyMessage, src: AnyMessage, desc: DescMessage): void {
  for (const fd of desc.fields) {
    if (!isFieldPopulated(src, fd)) continue;
    if (getFieldPatch(fd) === tableaupb.Patch.REPLACE) {
      clearField(dst, fd);
    }
    switch (fd.fieldKind) {
      case "map":
        patchMap(dst, src, fd);
        break;
      case "list":
        patchList(dst, src, fd);
        break;
      case "message": {
        const name = fd.localName;
        const srcChild = src[name] as AnyMessage;
        const dstChild = dst[name] as AnyMessage | undefined;
        if (dstChild === undefined || dstChild === null) {
          dst[name] = clone(fd.message, srcChild as never);
        } else {
          patchMessageInternal(dstChild, srcChild, fd.message);
        }
        break;
      }
      default:
        dst[fd.localName] = src[fd.localName];
        break;
    }
  }
}

function patchList(dst: AnyMessage, src: AnyMessage, fd: DescField): void {
  const srcList = src[fd.localName] as unknown[];
  let dstList = dst[fd.localName] as unknown[] | undefined;
  if (!Array.isArray(dstList)) {
    dstList = [];
    dst[fd.localName] = dstList;
  }
  const isMessage = fd.fieldKind === "list" && fd.listKind === "message";
  for (const item of srcList) {
    if (isMessage && fd.message) {
      dstList.push(clone(fd.message, item as never));
    } else {
      dstList.push(item);
    }
  }
}

function patchMap(dst: AnyMessage, src: AnyMessage, fd: DescField): void {
  const srcMap = src[fd.localName] as Record<string, unknown>;
  let dstMap = dst[fd.localName] as Record<string, unknown> | undefined;
  if (typeof dstMap !== "object" || dstMap === null) {
    dstMap = {};
    dst[fd.localName] = dstMap;
  }
  const isMessageValue = fd.fieldKind === "map" && fd.mapKind === "message";
  for (const [k, v] of Object.entries(srcMap)) {
    if (isMessageValue && fd.message) {
      const existing = dstMap[k] as AnyMessage | undefined;
      if (existing !== undefined && existing !== null) {
        // NOTE: this MERGES into the existing value (differs from a simple replace).
        patchMessageInternal(existing, v as AnyMessage, fd.message);
      } else {
        dstMap[k] = clone(fd.message, v as never);
      }
    } else {
      dstMap[k] = v;
    }
  }
}

// ---------------------------------------------------------------------------
// Index container runtime
//
// Helpers shared by generated loaders to build index, ordered index and
// ordered map containers.
// ---------------------------------------------------------------------------

/**
 * makeIndexKey serializes a tuple of multi-column index key parts into a unique
 * string so it can be used as a Map key with by-value equality (JS Maps compare
 * object keys by reference, which would break composite keys).
 *
 * Each part is tagged with its runtime type to avoid collisions between values
 * that stringify to the same text (e.g. the number 1 and the string "1"), and
 * parts are joined with a NUL separator that cannot appear in normal strings.
 *
 * Module-private: it is an implementation detail of TupleKeyMap (the only
 * caller). Generated loaders never serialize keys themselves; they go through
 * TupleKeyMap, so this is intentionally not exported.
 */
function makeIndexKey(parts: readonly unknown[]): string {
  return parts.map((p) => `${typeof p}:${String(p)}`).join("\u0000");
}

/**
 * compareValues compares two ordered scalar key parts of the same logical type
 * (number, bigint, string or boolean), returning a negative/zero/positive
 * number suitable for Array.prototype.sort.
 */
export function compareValues(a: unknown, b: unknown): number {
  if (typeof a === "bigint" || typeof b === "bigint") {
    const x = BigInt(a as never);
    const y = BigInt(b as never);
    return x < y ? -1 : x > y ? 1 : 0;
  }
  if (typeof a === "string" || typeof b === "string") {
    const x = String(a);
    const y = String(b);
    return x < y ? -1 : x > y ? 1 : 0;
  }
  if (typeof a === "boolean" || typeof b === "boolean") {
    return (a ? 1 : 0) - (b ? 1 : 0);
  }
  return (a as number) - (b as number);
}

/**
 * compareTuples compares two equal-length tuples of ordered scalar parts
 * lexicographically.
 */
export function compareTuples(a: readonly unknown[], b: readonly unknown[]): number {
  const n = Math.min(a.length, b.length);
  for (let i = 0; i < n; i++) {
    const c = compareValues(a[i], b[i]);
    if (c !== 0) return c;
  }
  return a.length - b.length;
}

/**
 * sortMapByKey sorts a Map in place by key using the given comparator and
 * returns the same Map instance. JS Maps preserve insertion order, so the
 * entries are re-inserted in sorted order (an "ordered map"). Sorting in place
 * (rather than returning a new Map) keeps the Map reference stable across
 * reloads, so views cached over it (e.g. TupleKeyMap) stay valid.
 */
export function sortMapByKey<K, V>(map: Map<K, V>, cmp: (a: K, b: K) => number): Map<K, V> {
  const entries = [...map.entries()].sort((x, y) => cmp(x[0], y[0]));
  map.clear();
  for (const [k, v] of entries) map.set(k, v);
  return map;
}

/**
 * OrderedMapValue is the value stored at a non-leaf level of a nested ordered
 * map. It pairs the next-level ordered (sub-)map with the message value at the
 * current level, mirroring the Pair<subMap, value> node the Go / C++ / C#
 * loaders expose (Go .First/.Second, C# .Item1/.Item2, C++ .first/.second), so
 * callers can both descend into the sub-map (.first) and read the level's own
 * value (.second). Leaf levels store the plain value directly, without this
 * wrapper.
 *
 * Type parameters:
 *  - M: the next-level ordered map type (a Map keyed in sorted order).
 *  - E: the message value at the current level.
 */
export interface OrderedMapValue<M, E> {
  /** first is the next-level ordered (sub-)map, sorted by key. */
  first: M;
  /** second is the value at the current level. */
  second: E;
}

/**
 * TupleKeyMap is the multi-column index / ordered-index container. A composite
 * key (e.g. [param, extType]) cannot be used directly as a JS Map key (Maps
 * compare object keys by reference), so it owns an internal Map keyed by the
 * opaque serialized string of the key columns (see makeIndexKey) plus a side
 * table mapping that string back to the original key tuple. Lookups and
 * iteration accept / return by-value tuple keys, mirroring the structured
 * composite keys the Go / C++ / C# loaders expose for multi-column indexes
 * (where the map key is a comparable key struct).
 *
 * Unlike a plain read-only view, it owns its storage and exposes a small build
 * API (clear / getOrSet / sortKeys) used by the generated processAfterLoad, so
 * a messager holds this single container per multi-column index (or per
 * multi-key leveled container) instead of a raw Map, a separate serialized-key
 * -> tuple side table, and a cached view. It is mutated in place across
 * reloads, so its reference stays stable and callers holding the container
 * returned by a finder stay valid.
 *
 * Type parameters:
 *  - K: the readonly tuple of key columns (e.g. readonly [id: number,
 *    name: string]); the generated loaders pass a named alias (e.g.
 *    ItemConf.Index_AwardItemKey) so lookups / iteration are fully typed. K is
 *    purely compile-time: the runtime always serializes via makeIndexKey, so
 *    the key type carries no runtime cost.
 *  - E: the value type stored under each key tuple. This is a plain tuple-keyed
 *    map (one key -> one value), so the value is whatever the caller stores:
 *    a value list (E = V[]) for a leaf multi-column index, or an inner index
 *    container for a multi-key leveled container.
 */
export class TupleKeyMap<K extends readonly unknown[], E> {
  private readonly map = new Map<string, E>();
  private readonly tuples = new Map<string, readonly unknown[]>();

  /** size returns the number of key buckets. */
  get size(): number {
    return this.map.size;
  }

  /** clear empties the container (called at the start of each (re)build). */
  clear(): void {
    this.map.clear();
    this.tuples.clear();
  }

  /**
   * getOrSet returns the value stored under the given key columns, creating and
   * inserting it via make() on first access (and recording the original key
   * tuple then, for sortKeys and by-value key iteration). makeIndexKey is
   * computed exactly once. The build side is generator-driven, so keyParts is
   * loosely typed; the query side (get/has/keys/entries) stays strongly typed.
   */
  getOrSet(keyParts: readonly unknown[], make: () => E): E {
    const k = makeIndexKey(keyParts);
    let v = this.map.get(k);
    if (v === undefined) {
      v = make();
      this.map.set(k, v);
      this.tuples.set(k, keyParts);
    }
    return v;
  }

  /**
   * sortKeys reorders the buckets by comparing their key tuples
   * lexicographically (used by ordered indexes). JS Maps preserve insertion
   * order, so entries are re-inserted in sorted order. Sorting in place keeps
   * the container reference stable across reloads.
   */
  sortKeys(): void {
    const entries = [...this.map.entries()].sort((x, y) =>
      compareTuples(this.tuples.get(x[0]) ?? [], this.tuples.get(y[0]) ?? []),
    );
    this.map.clear();
    for (const [k, v] of entries) this.map.set(k, v);
  }

  /** get returns the value stored under the given key columns, or undefined. */
  get(key: K): E | undefined {
    return this.map.get(makeIndexKey(key));
  }

  /** has reports whether the given key columns are present. */
  has(key: K): boolean {
    return this.map.has(makeIndexKey(key));
  }

  /** keys iterates the key-column tuples in container (sorted, if ordered) order. */
  *keys(): IterableIterator<K> {
    for (const k of this.map.keys()) {
      yield (this.tuples.get(k) ?? []) as unknown as K;
    }
  }

  /** values iterates the values in container order. */
  values(): IterableIterator<E> {
    return this.map.values();
  }

  /** entries iterates [keyColumns, value] pairs in container order. */
  *entries(): IterableIterator<[K, E]> {
    for (const [k, v] of this.map) {
      yield [(this.tuples.get(k) ?? []) as unknown as K, v];
    }
  }

  [Symbol.iterator](): IterableIterator<[K, E]> {
    return this.entries();
  }
}
