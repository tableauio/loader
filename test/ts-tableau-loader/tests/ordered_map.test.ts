// OrderedMap block: getOrderedMap traversal. Mirrors the same scenarios in:
//   - Go:  test/go-tableau-loader/ordered_map_test.go
//   - C#:  test/csharp-tableau-loader/tests/OrderedMapTests.cs
//   - C++: test/cpp-tableau-loader/tests/ordered_map_test.cpp
import { assert, check, prepareHub } from "./harness.js";

export function run(): void {
  // ---- ActivityConf: nested ordered map ----
  // getOrderedMap() exposes the full sorted tree; each node carries its value
  // (.second) and the next-level sub-map (.first), and the leveled
  // getOrderedMap1/2/3 getters descend it.
  check("ActivityConf nested ordered map", () => {
    const activity = prepareHub().getActivityConf();
    assert.ok(activity);

    // 1st level: bigint keys, ascending; node.second is the Activity value.
    const top = activity!.getOrderedMap();
    assert.ok(top.size > 0);
    let prev = -1n;
    for (const key of top.keys()) {
      assert.ok(key >= prev, `ordered map keys not ascending: ${key} after ${prev}`);
      prev = key;
    }
    assert.equal(top.get(100001n)?.second.activityName, "活动1");

    // descending via .first equals the scoped getter (2nd level: chapter map).
    const chapters = activity!.getOrderedMap1(100001n);
    assert.equal(chapters, top.get(100001n)?.first);
    assert.equal(chapters?.get(1)?.second.chapterName, "签到活动章1");

    // 3rd level: section map, scoped to (activityId, chapterId).
    const sections = activity!.getOrderedMap2(100001n, 1);
    assert.equal(sections, chapters?.get(1)?.first);

    // 4th level (leaf): scalar map sectionRankMap["2001"] === 2, ascending keys.
    const ranks = activity!.getOrderedMap3(100001n, 1, 2);
    assert.equal(ranks, sections?.get(2)?.first);
    assert.equal(ranks?.get(2001), 2);
    let prevRank = -Infinity;
    for (const key of ranks!.keys()) {
      assert.ok(key >= prevRank, `leaf ordered map keys not ascending: ${key} after ${prevRank}`);
      prevRank = key;
    }

    // missing keys return undefined at every level.
    assert.equal(activity!.getOrderedMap1(999n), undefined);
    assert.equal(activity!.getOrderedMap3(100001n, 1, 999), undefined);
  });

  // ---- ItemConf: 1st-level ordered map ----
  check("ItemConf ordered map ascending", () => {
    const item = prepareHub().getItemConf();
    assert.ok(item);
    const orderedMap = item!.getOrderedMap();
    assert.ok(orderedMap.size > 0);

    // keys ascending by id
    let prev = -Infinity;
    for (const key of orderedMap.keys()) {
      assert.ok(key >= prev, `ItemConf ordered map keys not ascending: ${key} after ${prev}`);
      prev = key;
    }
  });
}
