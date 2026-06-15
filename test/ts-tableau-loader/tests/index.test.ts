// Index block: index finders (leveled, ordered and multi-column). Mirrors:
//   - Go:  test/go-tableau-loader/index_test.go
//   - C#:  test/csharp-tableau-loader/tests/IndexTests.cs
//   - C++: test/cpp-tableau-loader/tests/index_test.cpp
import * as protoconf from "../tableau/barrel/protoconf.pc.js";

import {
  assert,
  check,
  fruitTypeApple,
  fruitTypeOrange,
  fruitTypeBanana,
  prepareHub,
} from "./harness.js";

export function run(): void {
  // ---- FruitConf: leveled index (Price<ID>) finders ----
  check("FruitConf index finders", () => {
    const fruit = prepareHub().getFruitConf();
    assert.ok(fruit);

    // global index: price -> items
    assert.deepEqual(fruit!.findItem(10)?.map((i) => i.id), [1001]);
    assert.equal(fruit!.findFirstItem(20)?.id, 1002);
    assert.equal(fruit!.findItem(999), undefined);
    // 6 items, all unique prices -> 6 entries
    assert.equal(fruit!.findItemMap().size, 6);

    // 1st-level scoped index: (fruitType) -> price -> items
    assert.deepEqual(fruit!.findItem1(fruitTypeOrange, 15)?.map((i) => i.id), [2001]);
    assert.equal(fruit!.findFirstItem1(fruitTypeBanana, 8)?.id, 3001);
    assert.equal(fruit!.findItemMap1(fruitTypeApple)?.size, 2);
    assert.equal(fruit!.findItemMap1(999), undefined);
    assert.equal(fruit!.findItem1(999, 15), undefined);
  });

  // ---- FruitConf: ordered index (Price<ID>@OrderedFruit) finders + ascending map ----
  check("FruitConf ordered index finders", () => {
    const fruit = prepareHub().getFruitConf();
    assert.ok(fruit);

    assert.deepEqual(fruit!.findOrderedFruit(10)?.map((i) => i.id), [1001]);
    assert.equal(fruit!.findFirstOrderedFruit(25)?.id, 2002);
    assert.equal(fruit!.findOrderedFruit(999), undefined);

    // ordered map is keyed in ascending price order
    const m = fruit!.findOrderedFruitMap();
    assert.equal(m.size, 6);
    let prev = -Infinity;
    for (const key of m.keys()) {
      assert.ok(key >= prev, `ordered map keys not ascending: ${key} after ${prev}`);
      prev = key;
    }

    // 1st-level scoped ordered index
    assert.deepEqual(fruit!.findOrderedFruit1(fruitTypeOrange, 25)?.map((i) => i.id), [2002]);
    assert.equal(fruit!.findFirstOrderedFruit1(fruitTypeBanana, 12)?.id, 3002);
    assert.equal(fruit!.findOrderedFruit1(999, 25), undefined);
  });

  // ---- ItemConf: single-column index ----
  check("ItemConf findItemInfoMap non-empty", () => {
    const item = prepareHub().getItemConf();
    assert.ok(item);
    assert.ok(item!.findItemInfoMap().size > 0);
  });

  // ---- ItemConf: multi-column ordered index (ParamExtType) ----
  // Returns a TupleKeyMap whose keys are readable [param, extType] tuples (not
  // opaque serialized strings), are sorted ascending, and round-trip through get().
  check("ItemConf multi-column ordered index (TupleKeyMap)", () => {
    const item = prepareHub().getItemConf();
    assert.ok(item);

    // apple has param_list [1,2,3] x extTypeList [APPLE, ORANGE] -> 6 buckets.
    const m = item!.findParamExtTypeMap();
    assert.equal(m.size, 6);
    assert.equal([...m.keys()].length, 6);

    let prevParam = -Infinity;
    for (const [key, values] of m) {
      // keys are decoded 2-tuples, not opaque serialized strings
      assert.equal(key.length, 2);
      assert.ok((key[0] as number) >= prevParam, "ParamExtType keys not ascending by param");
      prevParam = key[0] as number;
      // get(tuple) round-trips to the very same value list reference
      assert.equal(m.get(key), values);
      // and agrees with the point-lookup finder
      assert.equal(
        item!.findParamExtType(key[0] as number, key[1] as protoconf.FruitType),
        values,
      );
    }

    // a concrete known bucket: (param=1, extType=APPLE) -> apple
    assert.deepEqual(m.get([1, fruitTypeApple])?.map((i) => i.name), ["apple"]);
    assert.equal(m.get([999, fruitTypeApple]), undefined);
  });

  // ---- ItemConf: non-ordered multi-column index (AwardItem) ----
  // Also exposed as a TupleKeyMap with readable [id, name] tuple keys.
  check("ItemConf multi-column index (TupleKeyMap)", () => {
    const item = prepareHub().getItemConf();
    assert.ok(item);

    const m = item!.findAwardItemMap();
    for (const [key, values] of m) {
      assert.equal(key.length, 2);
      assert.equal(m.get(key), values);
      assert.equal(
        item!.findAwardItem(key[0] as number, key[1] as string),
        values,
      );
    }
    // apple is keyed by (id=1, name="apple")
    assert.deepEqual(m.get([1, "apple"])?.map((i) => i.id), [1]);
    // findFirst* variant returns the first match for the same key.
    assert.equal(item!.findFirstAwardItem(1, "apple")?.id, 1);
  });
}
