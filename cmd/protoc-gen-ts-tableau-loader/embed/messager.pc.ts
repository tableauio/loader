import { type Message } from "@bufbuild/protobuf";
import { Format } from "./util.pc.js";
import { type MessagerOptions } from "./load.pc.js";
// Type-only import to describe the hub passed to processAfterLoadAll. Importing
// only the type avoids a runtime cycle with the generated hub module.
import type { Hub } from "./hub.pc.js";

/**
 * Stats contains statistics info about loading.
 */
export interface Stats {
  /** Total load time consuming, in milliseconds. */
  durationMs: number;
}

/**
 * Messager is the base class for all generated configuration messagers.
 * It is designed for three goals:
 *   1. Easy use: simple yet powerful accessors.
 *   2. Elegant API: concise and clean functions.
 *   3. Extensibility: Map, OrderedMap, Index, OrderedIndex...
 */
export abstract class Messager {
  protected loadStats: Stats = { durationMs: 0 };

  /** getStats returns the loading stats info. */
  getStats(): Stats {
    return this.loadStats;
  }

  /** name returns the messager's message name. */
  abstract name(): string;

  /** load fills message from file in the specified directory and format. Throws on failure. */
  abstract load(dir: string, fmt: Format, options?: MessagerOptions): void;

  /** message returns the inner protobuf message data. */
  message(): Message | undefined {
    return undefined;
  }

  /** processAfterLoad is invoked after this messager is loaded. Throws on failure. */
  protected processAfterLoad(): void {}

  /** processAfterLoadAll is invoked after all messagers are loaded. Throws on failure. */
  processAfterLoadAll(_hub: Hub): void {}
}

/**
 * MessagerCtor is the constructor type for a concrete Messager.
 */
export type MessagerCtor = new () => Messager;
