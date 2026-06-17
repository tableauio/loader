// Get block: point-lookup getters (map getters). Mirrors the same scenarios in:
//   - Go:  test/go-tableau-loader/get_test.go
//   - C#:  test/csharp-tableau-loader/tests/GetTests.cs
//   - C++: test/cpp-tableau-loader/tests/get_test.cpp
import { assert, check, fruitTypeApple, prepareHub } from "./harness.js";

export function run(): void {
  // ---- ItemConf: first-level int-keyed map getter ----
  check("ItemConf map getter", () => {
    const item = prepareHub().getItemConf();
    assert.ok(item);
    assert.equal(item!.name(), "ItemConf");
    assert.equal(item!.get1(1)?.name, "apple");
    assert.equal(item!.get1(0)?.name, "coin1");
    assert.equal(item!.get1(12345), undefined);
  });

  // ---- ActivityConf: nested map getters with a 64-bit (bigint) outer key ----
  check("ActivityConf nested map getters", () => {
    const activity = prepareHub().getActivityConf();
    assert.ok(activity);
    assert.equal(activity!.get1(100001n)?.activityName, "活动1");
    assert.equal(activity!.get2(100001n, 1)?.chapterName, "签到活动章1");
    assert.equal(activity!.get3(100001n, 1, 2)?.sectionId, 2);
    // sectionRankMap["2001"] === 2
    assert.equal(activity!.get4(100001n, 1, 2, 2001), 2);
    // not found
    assert.equal(activity!.get3(100001n, 1, 99), undefined);
  });

  // ---- ThemeConf: string-keyed map getter ----
  check("ThemeConf string-keyed map", () => {
    const theme = prepareHub().getThemeConf();
    assert.ok(theme);
    assert.ok(Object.keys(theme!.data().themeMap).length > 0);
  });

  // ---- FruitConf: leveled point getters ----
  check("FruitConf leveled getters", () => {
    const fruit = prepareHub().getFruitConf();
    assert.ok(fruit);

    // get1: found / not found
    assert.ok(fruit!.get1(fruitTypeApple));
    assert.equal(fruit!.get1(999), undefined);

    // get2: APPLE -> item 1001, price 10
    const item = fruit!.get2(fruitTypeApple, 1001);
    assert.ok(item);
    assert.equal(item!.id, 1001);
    assert.equal(item!.price, 10);
    assert.equal(fruit!.get2(fruitTypeApple, 9999), undefined);
    assert.equal(fruit!.get2(999, 1001), undefined);
  });
}
