// Shared harness for the TS tableau loader tests.
//
// The unit tests are split into four blocks, one file each, mirroring the
// Go / C# / C++ test layout:
//   - load.test.ts        (Load, filter, Bin, Patch)
//   - get.test.ts         (point-lookup getters)
//   - ordered_map.test.ts (getOrderedMap traversal)
//   - index.test.ts       (find* index finders)
//
// `tests/smoke.ts` is the entry point that runs all four blocks.
import { strict as assert } from "node:assert";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

import { Hub } from "../tableau/hub.pc.js";
import { Format } from "../tableau/load.pc.js";
import * as protoconf from "../tableau/barrel/protoconf.pc.js";

const here = dirname(fileURLToPath(import.meta.url));
export const testdata = join(here, "..", "..", "testdata");
export const confDir = join(testdata, "conf");
export const binDir = join(testdata, "bin");
export const patchConf = join(testdata, "patchconf");
export const patchConf2 = join(testdata, "patchconf2");
export const patchResult = join(testdata, "patchresult");

// fruitType map keys (== FruitType enum values) used by FruitConf.json.
export const fruitTypeApple = protoconf.FruitType.APPLE;
export const fruitTypeOrange = protoconf.FruitType.ORANGE;
export const fruitTypeBanana = protoconf.FruitType.BANANA;

export { assert, join };

let passed = 0;
export function check(name: string, fn: () => void): void {
  fn();
  passed++;
  console.log(`  ok - ${name}`);
}
export function passedCount(): number {
  return passed;
}

// Shared hub, loaded lazily once and reused across the Get / OrderedMap / Index
// blocks. Mirrors Go's prepareHub helper, C#'s HubFixture and C++'s HubFixture.
let sharedHub: Hub | undefined;
export function prepareHub(): Hub {
  if (!sharedHub) {
    sharedHub = new Hub();
    sharedHub.load(confDir, Format.JSON, { ignoreUnknownFields: true });
  }
  return sharedHub;
}
