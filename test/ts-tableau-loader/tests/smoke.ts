import { strict as assert } from "node:assert";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

import { equals } from "@bufbuild/protobuf";

import { Hub } from "../tableau/hub.pc.js";
import { Format, LoadMode } from "../tableau/load.pc.js";
import {
  PatchMergeConf,
  PatchReplaceConf,
  RecursivePatchConf,
} from "../tableau/patch_conf.pc.js";
import { HeroConf } from "../tableau/hero_conf.pc.js";
import * as protoconf from "../tableau/barrel/protoconf.pc.js";

const here = dirname(fileURLToPath(import.meta.url));
const testdata = join(here, "..", "..", "testdata");
const confDir = join(testdata, "conf");
const binDir = join(testdata, "bin");
const patchConf = join(testdata, "patchconf");
const patchConf2 = join(testdata, "patchconf2");
const patchResult = join(testdata, "patchresult");

let passed = 0;
function check(name: string, fn: () => void): void {
  fn();
  passed++;
  console.log(`  ok - ${name}`);
}

// 1. Hub loads all messagers from a JSON directory.
const hub = new Hub();
check("hub.load(conf, JSON)", () => {
  hub.load(confDir, Format.JSON, { ignoreUnknownFields: true });
});

// 2. ItemConf: first-level int-keyed map getter.
check("ItemConf map getter", () => {
  const item = hub.getItemConf();
  assert.ok(item);
  assert.equal(item!.name(), "ItemConf");
  assert.equal(item!.get1(1)?.name, "apple");
  assert.equal(item!.get1(0)?.name, "coin1");
  assert.equal(item!.get1(12345), undefined);
});

// 3. ActivityConf: 4-level nested map getter with a 64-bit (bigint) outer key.
check("ActivityConf nested map getters", () => {
  const activity = hub.getActivityConf();
  assert.ok(activity);
  assert.equal(activity!.get1(100001n)?.activityName, "活动1");
  assert.equal(activity!.get2(100001n, 1)?.chapterName, "签到活动章1");
  // sectionRankMap["2001"] === 2
  assert.equal(activity!.get4(100001n, 1, 2, 2001), 2);
  assert.equal(activity!.get3(100001n, 1, 99), undefined);
});

// 4. ThemeConf: string-keyed map getter.
check("ThemeConf string-keyed map", () => {
  const theme = hub.getThemeConf();
  assert.ok(theme);
  assert.ok(Object.keys(theme!.data().themeMap).length > 0);
});

// 5. Hub typed accessor for an unloaded-by-filter messager returns undefined.
check("Hub filter", () => {
  const filtered = new Hub({ filter: (n) => n === "ItemConf" });
  filtered.load(confDir, Format.JSON, { ignoreUnknownFields: true });
  assert.ok(filtered.getItemConf());
  assert.equal(filtered.getActivityConf(), undefined);
});

// 6. Binary format load via a single messager.
check("HeroConf BIN load", () => {
  const hero = new HeroConf();
  hero.load(binDir, Format.BIN);
  assert.equal(hero.name(), "HeroConf");
});

// 7. Patch MERGE: main + two patch files merged in order.
check("PatchMergeConf MERGE", () => {
  const pm = new PatchMergeConf();
  pm.load(confDir, Format.JSON, {
    ignoreUnknownFields: true,
    patchPaths: [
      join(patchConf, "PatchMergeConf.json"),
      join(patchConf2, "PatchMergeConf.json"),
    ],
  });
  // Scalar overridden by the first patch.
  assert.equal(pm.data().name, "orange");
  // itemMap key 999 contributed by the second patch (merge, not replace).
  assert.ok(pm.get1(999));
  assert.ok(pm.get1(1));
});

// 8. LoadMode.ONLY_MAIN ignores the patch files.
check("PatchMergeConf ONLY_MAIN", () => {
  const pm = new PatchMergeConf();
  pm.load(confDir, Format.JSON, {
    ignoreUnknownFields: true,
    mode: LoadMode.ONLY_MAIN,
    patchPaths: [
      join(patchConf, "PatchMergeConf.json"),
      join(patchConf2, "PatchMergeConf.json"),
    ],
  });
  assert.equal(pm.data().name, "apple");
  assert.equal(pm.get1(999), undefined);
});

// 9. PATCH_MERGE golden: conf + patchconf merged must equal the canonical
//    patchresult, checked as a whole (mirrors Go/C#'s proto.Equal comparison).
check("RecursivePatchConf MERGE golden", () => {
  const got = new RecursivePatchConf();
  got.load(confDir, Format.JSON, {
    ignoreUnknownFields: true,
    patchDirs: [patchConf],
  });

  // The golden file is itself a RecursivePatchConf sheet (PATCH_MERGE), so load
  // it with no patch sources to get the main file content verbatim.
  const expected = new RecursivePatchConf();
  expected.load(patchResult, Format.JSON, { ignoreUnknownFields: true });

  assert.ok(
    equals(protoconf.RecursivePatchConfSchema, got.message(), expected.message()),
    "patched RecursivePatchConf does not match the golden patchresult",
  );
});

// 10. PATCH_REPLACE: the sheet-level option replaces the whole message with the
//     last patch file, rather than merging field-by-field.
check("PatchReplaceConf PATCH_REPLACE", () => {
  const pr = new PatchReplaceConf();
  pr.load(confDir, Format.JSON, {
    ignoreUnknownFields: true,
    patchDirs: [patchConf],
  });
  // Main file has name "apple" / priceList [10,100]; the patch fully replaces
  // it with name "orange" / priceList [20,200] (replace, not append).
  assert.equal(pr.data().name, "orange");
  assert.deepEqual(pr.data().priceList.map(Number), [20, 200]);
});

// 11. LoadMode.ONLY_PATCH starts from an empty message and applies patches only.
check("PatchMergeConf ONLY_PATCH", () => {
  const pm = new PatchMergeConf();
  pm.load(confDir, Format.JSON, {
    ignoreUnknownFields: true,
    mode: LoadMode.ONLY_PATCH,
    patchPaths: [
      join(patchConf, "PatchMergeConf.json"),
      join(patchConf2, "PatchMergeConf.json"),
    ],
  });
  // Name comes from the first patch (second patch carries no name field), and
  // both itemMap entries come purely from the patches, not the main file.
  assert.equal(pm.data().name, "orange");
  assert.ok(pm.get1(1));
  assert.ok(pm.get1(999));
});

// 12. BIN load from a non-existent directory throws (mirrors C#'s failure case).
check("HeroConf BIN load failure", () => {
  const hero = new HeroConf();
  assert.throws(() => hero.load(join(testdata, "no-such-bin-dir"), Format.BIN));
});

console.log(`\nAll ${passed} smoke checks passed.`);
