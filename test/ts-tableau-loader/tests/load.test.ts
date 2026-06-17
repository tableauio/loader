// Load block: hub loading, filter, binary-format loading and patch loading.
// Mirrors the same scenarios in:
//   - Go:  test/go-tableau-loader/load_test.go
//   - C#:  test/csharp-tableau-loader/tests/LoadTests.cs
//   - C++: test/cpp-tableau-loader/tests/load_test.cpp
import { equals } from "@bufbuild/protobuf";

import { Hub, Registry } from "../tableau/hub.pc.js";
import { Format, LoadMode } from "../tableau/load.pc.js";
import {
  PatchMergeConf,
  PatchReplaceConf,
  RecursivePatchConf,
} from "../tableau/patch_conf.pc.js";
import { HeroConf } from "../tableau/hero_conf.pc.js";
import * as protoconf from "../tableau/barrel/protoconf.pc.js";

import { CustomItemConf, CustomItemConfName } from "../custom/custom_item_conf.js";

import {
  assert,
  check,
  confDir,
  binDir,
  patchConf,
  patchConf2,
  patchResult,
  testdata,
  join,
  prepareHub,
} from "./harness.js";

// Register the hand-written CustomItemConf before any hub is built, so the
// shared hub (and any hub loaded afterwards) includes it. Mirrors Go's init()
// Register, C#'s HubFixture registration and C++'s InitCustomMessager.
Registry.register(CustomItemConf);

export function run(): void {
  // ---- Load ----

  // Hub loads all messagers from a JSON directory.
  check("hub.load(conf, JSON)", () => {
    const hub = prepareHub();
    assert.ok(hub.getItemConf());
    assert.ok(hub.getActivityConf());
  });

  // ---- CustomConf ----

  // CustomItemConf is a hand-written messager: it has no file of its own and
  // resolves a "special" item from ItemConf via processAfterLoadAll. Mirrors
  // Go's Test_CustomItemConf and C#/C++'s
  // CustomItemConf_ProcessAfterLoadAll_ResolvesSpecialItem.
  check("CustomItemConf processAfterLoadAll", () => {
    const hub = prepareHub();
    const custom = hub.getMessager(CustomItemConfName) as CustomItemConf | undefined;
    assert.ok(custom);
    // ItemConf item 1 is "apple", so the resolved special item name is non-empty.
    assert.notEqual(custom!.getSpecialItemName(), "");
    assert.equal(custom!.getSpecialItemName(), "apple");
  });

  // Hub typed accessor for an unloaded-by-filter messager returns undefined.
  check("Hub filter", () => {
    const filtered = new Hub({ filter: (n) => n === "ItemConf" });
    filtered.load(confDir, Format.JSON, { ignoreUnknownFields: true });
    assert.ok(filtered.getItemConf());
    assert.equal(filtered.getActivityConf(), undefined);
  });

  // ---- Bin ----

  // Binary format load via a single messager.
  check("HeroConf BIN load", () => {
    const hero = new HeroConf();
    hero.load(binDir, Format.BIN);
    assert.equal(hero.name(), "HeroConf");
  });

  // BIN load from a non-existent directory throws (mirrors C#'s failure case).
  check("HeroConf BIN load failure", () => {
    const hero = new HeroConf();
    assert.throws(() => hero.load(join(testdata, "no-such-bin-dir"), Format.BIN));
  });

  // ---- Patch ----

  // Patch MERGE: main + two patch files merged in order.
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
    // itemMap merges field-by-field, so the main file's key 2 survives.
    assert.ok(pm.get1(2));
    // replaceItemMap has the field-level PATCH_REPLACE option: the last patch
    // that carries it fully replaces the field. patchconf2 provides {1, 999},
    // so key 999 is present while the main file's key 2 is dropped (replace,
    // not merge). Mirrors Go/C#/C++'s ReplaceItemMap[999] assertion.
    assert.ok(pm.data().replaceItemMap[999]);
    assert.equal(pm.data().replaceItemMap[2], undefined);
  });

  // LoadMode.ONLY_MAIN ignores the patch files.
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

  // PATCH_MERGE golden: conf + patchconf merged must equal the canonical
  // patchresult, checked as a whole (mirrors Go/C#'s proto.Equal comparison).
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

  // PATCH_REPLACE: the sheet-level option replaces the whole message with the
  // last patch file, rather than merging field-by-field.
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

  // LoadMode.ONLY_PATCH starts from an empty message and applies patches only.
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
}
