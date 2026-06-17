import { readFileSync } from "node:fs";
import { existsSync } from "node:fs";
import { join } from "node:path";
import {
  create,
  fromBinary,
  fromJson,
  type DescMessage,
  type JsonValue,
  type Message,
  type MessageShape,
} from "@bufbuild/protobuf";
import * as tableaupb from "../tableau/protobuf/tableau_pb.js";
import {
  Format,
  format2Ext,
  getFormat,
  getSheetPatch,
  patchMessage,
} from "./util.pc.js";

export { Format } from "./util.pc.js";

/**
 * ReadFunc reads a config file and returns its content.
 */
export type ReadFunc = (path: string) => Uint8Array;

/**
 * LoadFunc loads a messager's content based on the given descriptor, path,
 * format, and options.
 */
export type LoadFunc = (
  desc: DescMessage,
  path: string,
  fmt: Format,
  options?: MessagerOptions,
) => Message;

/**
 * LoadMode controls patch loading behavior.
 */
export enum LoadMode {
  /** Load all related files (main + patches). Default. */
  ALL = "all",
  /** Only load the main file. */
  ONLY_MAIN = "only_main",
  /** Only load the patch files. */
  ONLY_PATCH = "only_patch",
}

/**
 * BaseOptions is the common options for both global-level and messager-level options.
 */
export interface BaseOptions {
  /** Whether to ignore unknown JSON fields during parsing. */
  ignoreUnknownFields?: boolean;
  /** Specify the directory paths for config patching. */
  patchDirs?: string[];
  /** Specify the loading mode for config patching. Default is LoadMode.ALL. */
  mode?: LoadMode;
  /** Custom read function to read a config file's content. Default is readFileSync. */
  readFunc?: ReadFunc;
  /** Custom load function to load a messager's content. Default is loadMessager. */
  loadFunc?: LoadFunc;
}

/**
 * MessagerOptions defines the options for loading a single messager.
 */
export interface MessagerOptions extends BaseOptions {
  /**
   * Path maps the messager to a corresponding config file path. If specified,
   * then the main messager will be parsed from the file directly, other than
   * the specified load dir.
   */
  path?: string;
  /**
   * PatchPaths maps the messager to one or multiple corresponding patch file
   * paths. If specified, then the main messager will be patched.
   */
  patchPaths?: string[];
}

/**
 * Options is the global-level options, which contains both global-level and
 * messager-level options.
 */
export interface Options extends BaseOptions {
  /**
   * messagerOptions maps each messager name to a MessagerOptions. If specified,
   * then the messager will be parsed with the given options directly.
   */
  messagerOptions?: { [name: string]: MessagerOptions };
}

/**
 * parseMessagerOptionsByName parses messager options with both global-level and
 * messager-level options taken into consideration.
 */
export function parseMessagerOptionsByName(options: Options, name: string): MessagerOptions {
  const mopts: MessagerOptions = options.messagerOptions?.[name]
    ? { ...options.messagerOptions[name] }
    : {};
  mopts.ignoreUnknownFields ??= options.ignoreUnknownFields;
  mopts.patchDirs ??= options.patchDirs;
  mopts.mode ??= options.mode;
  mopts.readFunc ??= options.readFunc;
  mopts.loadFunc ??= options.loadFunc;
  return mopts;
}

function defaultRead(path: string): Uint8Array {
  return new Uint8Array(readFileSync(path));
}

/**
 * loadMessager loads a protobuf message from the specified file path and format.
 */
export function loadMessager(
  desc: DescMessage,
  path: string,
  fmt: Format,
  options?: MessagerOptions,
): Message {
  const readFunc = options?.readFunc ?? defaultRead;
  let content: Uint8Array;
  try {
    content = readFunc(path);
  } catch (e) {
    throw new Error(`failed to read ${path}`, { cause: e });
  }
  return unmarshal(content, desc, fmt, options);
}

/**
 * loadMessagerInDir loads a protobuf message from the specified directory and
 * format. It resolves the file path based on the message descriptor name.
 */
export function loadMessagerInDir<Desc extends DescMessage>(
  desc: Desc,
  dir: string,
  fmt: Format,
  options?: MessagerOptions,
): MessageShape<Desc> {
  let path = "";
  if (options?.path) {
    path = options.path;
    fmt = getFormat(path);
  }
  if (path === "") {
    path = join(dir, desc.name + format2Ext(fmt));
  }
  const sheetPatch = getSheetPatch(desc);
  let msg: Message;
  if (sheetPatch !== tableaupb.Patch.NONE) {
    msg = loadMessagerWithPatch(desc, path, fmt, sheetPatch, options);
  } else {
    const loadFunc = options?.loadFunc ?? loadMessager;
    msg = loadFunc(desc, path, fmt, options);
  }
  return msg as MessageShape<Desc>;
}

/**
 * loadMessagerWithPatch loads a protobuf message with patch support.
 */
export function loadMessagerWithPatch(
  desc: DescMessage,
  path: string,
  fmt: Format,
  patch: tableaupb.Patch,
  options?: MessagerOptions,
): Message {
  const mode = options?.mode ?? LoadMode.ALL;
  const loadFunc = options?.loadFunc ?? loadMessager;
  if (mode === LoadMode.ONLY_MAIN) {
    // Ignore patch files when LoadMode.ONLY_MAIN specified.
    return loadFunc(desc, path, fmt, options);
  }

  let patchPaths: string[] = [];
  const explicitPaths = options?.patchPaths && options.patchPaths.length > 0;
  if (explicitPaths) {
    // PatchPaths takes precedence over PatchDirs.
    patchPaths = [...options!.patchPaths!];
  } else if (options?.patchDirs) {
    const filename = desc.name + format2Ext(fmt);
    for (const patchDir of options.patchDirs) {
      patchPaths.push(join(patchDir, filename));
    }
  }

  // Filter out non-existing patch files when relying on PatchDirs.
  let existedPatchPaths: string[];
  if (explicitPaths) {
    // Explicit paths are kept as-is; loadFunc surfaces errors if missing.
    existedPatchPaths = patchPaths;
  } else {
    existedPatchPaths = patchPaths.filter((p) => existsSync(p));
  }

  if (existedPatchPaths.length === 0) {
    if (mode === LoadMode.ONLY_PATCH) {
      // Return empty message when LoadMode.ONLY_PATCH specified but no valid
      // patch file provided.
      return create(desc);
    }
    // No valid patch path provided, then just load from the "main" file.
    return loadFunc(desc, path, fmt, options);
  }

  switch (patch) {
    case tableaupb.Patch.REPLACE: {
      // Just use the last "patch" file.
      const patchPath = existedPatchPaths[existedPatchPaths.length - 1]!;
      return loadFunc(desc, patchPath, getFormat(patchPath), options);
    }
    case tableaupb.Patch.MERGE: {
      let msg: Message;
      if (mode !== LoadMode.ONLY_PATCH) {
        // Load msg from the "main" file.
        msg = loadFunc(desc, path, fmt, options);
      } else {
        msg = create(desc);
      }
      for (const patchPath of existedPatchPaths) {
        const patchMsg = loadFunc(desc, patchPath, getFormat(patchPath), options);
        patchMessage(msg, patchMsg, desc);
      }
      return msg;
    }
    default:
      throw new Error(`unknown patch type: ${patch}`);
  }
}

/**
 * unmarshal parses the given byte content into a protobuf message based on the
 * specified format.
 */
export function unmarshal(
  content: Uint8Array,
  desc: DescMessage,
  fmt: Format,
  options?: MessagerOptions,
): Message {
  switch (fmt) {
    case Format.JSON:
      try {
        const json = JSON.parse(new TextDecoder().decode(content)) as JsonValue;
        return fromJson(desc, json, {
          ignoreUnknownFields: options?.ignoreUnknownFields ?? false,
        });
      } catch (e) {
        throw new Error(`failed to parse ${desc.name}.json`, { cause: e });
      }
    case Format.BIN:
      try {
        return fromBinary(desc, content);
      } catch (e) {
        throw new Error(`failed to parse ${desc.name}.binpb`, { cause: e });
      }
    default:
      throw new Error(`unknown format: ${fmt}`);
  }
}
