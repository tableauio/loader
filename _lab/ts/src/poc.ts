/**
 * PoC — validate that protobuf-es (@bufbuild/protobuf) can faithfully consume
 * the protojson that tableau (github.com/tableauio/tableau) emits, so it can
 * serve as the codegen base for `protoc-gen-ts-tableau-loader`.
 *
 * Run:  npm run generate && npm run poc
 *
 * What this exercises (the things that actually decide the selection):
 *   1. Canonical proto3 JSON ("protojson") parsing via fromJson(Schema, ...).
 *   2. Mixed field-name casing in one file: ItemConf.json has `itemMap`,
 *      `extTypeList`, `nameList` (camelCase) AND `param_list` (original
 *      snake_case). proto3 JSON requires BOTH to be accepted — protobuf-es
 *      does; protobufjs's fromObject does not (it would silently drop one).
 *   3. 64-bit ints as JSON strings: ThemeConf.Theme.value is uint64, encoded
 *      as "1" in JSON, must parse into a bigint.
 *   4. Well-known types: ItemConf.Item.expiry is google.protobuf.Timestamp,
 *      encoded as an RFC-3339 string.
 *   5. Binary round-trip (toBinary -> fromBinary) — the `.binpb` load path.
 *   6. Re-emitting canonical protojson via toJson (note snake `param_list`
 *      comes back as camel `paramList`).
 */
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import {
  fromJson,
  toJson,
  toBinary,
  fromBinary,
  type JsonValue,
} from "@bufbuild/protobuf";

import { ItemConfSchema } from "./protoconf/item_conf_pb.js";
import { ThemeConfSchema } from "./protoconf/test_conf_pb.js";

const HERE = dirname(fileURLToPath(import.meta.url));
const CONF = join(HERE, "../../../test/testdata/conf");

let failures = 0;
function check(label: string, cond: unknown): void {
  const ok = Boolean(cond);
  console.log(`  ${ok ? "PASS" : "FAIL"}  ${label}`);
  if (!ok) failures++;
}

// JSON.stringify helper that survives bigint (64-bit) fields.
const bigintSafe = (_k: string, v: unknown) =>
  typeof v === "bigint" ? v.toString() : v;

function readJson(file: string): JsonValue {
  return JSON.parse(readFileSync(join(CONF, file), "utf8")) as JsonValue;
}

// --- 1) ItemConf: the rich case (maps / enums / WKT / mixed casing) --------
console.log("\n[ItemConf] protojson -> message");
const itemConf = fromJson(ItemConfSchema, readJson("ItemConf.json"));
const apple = itemConf.itemMap["1"];

check("map<uint32,Item> parsed; key '1' present", apple !== undefined);
check("camelCase 'name' -> apple", apple?.name === "apple");
check(
  "snake_case 'param_list' accepted -> [1,2,3]",
  JSON.stringify(apple?.paramList) === "[1,2,3]",
);
check(
  "camelCase 'extTypeList' accepted -> 2 entries",
  apple?.extTypeList.length === 2,
);
check(
  "nested 'path.nameList' (camel) -> ['icon.png']",
  apple?.path?.nameList[0] === "icon.png",
);
check("enum 'type' parsed by name -> non-zero", (apple?.type ?? 0) !== 0);
check(
  "WKT Timestamp 'expiry' parsed -> seconds > 0",
  (apple?.expiry?.seconds ?? 0n) > 0n,
);

// --- 2) ThemeConf: uint64-as-string -----------------------------------------
console.log("\n[ThemeConf] uint64 encoded as JSON string");
const themeConf = fromJson(ThemeConfSchema, readJson("ThemeConf.json"));
const theme1 = themeConf.themeMap["theme1"];
check("map<string,Theme> key 'theme1' present", theme1 !== undefined);
check('uint64 "1" -> bigint 1n', theme1?.value === 1n);

// --- 3) binary round-trip (the .binpb load path) ----------------------------
console.log("\n[ItemConf] binary round-trip");
const bin = toBinary(ItemConfSchema, itemConf);
const back = fromBinary(ItemConfSchema, bin);
check("toBinary produced bytes", bin.length > 0);
check(
  "fromBinary preserves apple.name",
  back.itemMap["1"]?.name === "apple",
);
check(
  "fromBinary preserves paramList",
  JSON.stringify(back.itemMap["1"]?.paramList) === "[1,2,3]",
);

// --- 4) re-emit canonical protojson -----------------------------------------
console.log("\n[ItemConf] message -> canonical protojson (toJson)");
const reJson = toJson(ItemConfSchema, itemConf) as Record<string, any>;
const reApple = reJson.itemMap["1"];
check(
  "re-emitted with camelCase 'paramList' (was snake in input)",
  Array.isArray(reApple.paramList),
);
check(
  "re-emitted Timestamp as RFC-3339 string",
  typeof reApple.expiry === "string" && reApple.expiry.includes("T"),
);
console.log(
  "  sample:",
  JSON.stringify(reApple, bigintSafe).slice(0, 120) + " ...",
);

// --- verdict ----------------------------------------------------------------
console.log(
  `\n${failures === 0 ? "ALL CHECKS PASSED" : `${failures} CHECK(S) FAILED`}`,
);
process.exit(failures === 0 ? 0 : 1);
