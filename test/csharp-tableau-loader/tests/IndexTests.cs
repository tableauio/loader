using Xunit;

namespace LoaderTests
{
    /// <summary>
    /// Index block: index finders (leveled, ordered and multi-column). Mirrors:
    ///   - Go:  test/go-tableau-loader/index_test.go
    ///   - C++: test/cpp-tableau-loader/tests/index_test.cpp
    ///   - TS:  test/ts-tableau-loader/tests/index.test.ts (cases 13-16)
    /// </summary>
    [Collection("HubCollection")]
    public class IndexTests
    {
        private readonly Tableau.Hub _hub;

        public IndexTests(HubFixture fixture)
        {
            _hub = fixture.Hub;
        }

        // ---- FruitConf: leveled index (Price<ID>) finders ----

        [Fact]
        public void FruitConf_IndexFinders()
        {
            var fruit = _hub.GetFruitConf();
            Assert.NotNull(fruit);

            // global index: price -> items
            var items = fruit!.FindItem(10);
            Assert.NotNull(items);
            Assert.Single(items!);
            Assert.Equal(1001, items![0].Id);
            Assert.Equal(1002, fruit.FindFirstItem(20)!.Id);
            Assert.Null(fruit.FindItem(999));
            // 6 items, all unique prices -> 6 entries
            Assert.Equal(6, fruit.FindItemMap().Count);

            // 1st-level scoped index: (fruitType) -> price -> items
            Assert.Equal(2001, fruit.FindItem1(FruitTypes.Orange, 15)![0].Id);
            Assert.Equal(3001, fruit.FindFirstItem1(FruitTypes.Banana, 8)!.Id);
            Assert.Equal(2, fruit.FindItemMap1(FruitTypes.Apple)!.Count);
            Assert.Null(fruit.FindItemMap1(999));
            Assert.Null(fruit.FindItem1(999, 15));
        }

        // ---- FruitConf: ordered index (Price<ID>@OrderedFruit) finders ----

        [Fact]
        public void FruitConf_OrderedIndexFinders()
        {
            var fruit = _hub.GetFruitConf();
            Assert.NotNull(fruit);

            Assert.Equal(1001, fruit!.FindOrderedFruit(10)![0].Id);
            Assert.Equal(2002, fruit.FindFirstOrderedFruit(25)!.Id);
            Assert.Null(fruit.FindOrderedFruit(999));

            // ordered map is keyed in ascending price order (SortedDictionary).
            var m = fruit.FindOrderedFruitMap();
            Assert.Equal(6, m.Count);
            int prev = int.MinValue;
            foreach (var key in m.Keys)
            {
                Assert.True(key >= prev, $"ordered map keys not ascending: {key} after {prev}");
                prev = key;
            }

            // 1st-level scoped ordered index
            Assert.Equal(2002, fruit.FindOrderedFruit1(FruitTypes.Orange, 25)![0].Id);
            Assert.Equal(3002, fruit.FindFirstOrderedFruit1(FruitTypes.Banana, 12)!.Id);
            Assert.Null(fruit.FindOrderedFruit1(999, 25));
        }

        // ---- ItemConf: single-column index ----

        [Fact]
        public void ItemConf_FindItemInfoMap_NonEmpty()
        {
            var itemConf = _hub.GetItemConf();
            Assert.NotNull(itemConf);
            var itemInfoMap = itemConf!.FindItemInfoMap();
            Assert.NotNull(itemInfoMap);
            Assert.NotEmpty(itemInfoMap);
        }

        // ---- ItemConf: multi-column (composite-key) indexes ----

        [Fact]
        public void ItemConf_ParamExtType_OrderedMultiColumnIndex()
        {
            var item = _hub.GetItemConf();
            Assert.NotNull(item);

            // apple has param_list [1,2,3] x extTypeList [APPLE, ORANGE] -> 6 buckets,
            // and it is the only item carrying both columns, so the map has 6 entries.
            var m = item!.FindParamExtTypeMap();
            Assert.Equal(6, m.Count);

            // SortedDictionary iterates in ascending key order; each key round-trips
            // through the point-lookup finder.
            int prevParam = int.MinValue;
            foreach (var kv in m)
            {
                Assert.True(kv.Key.Param >= prevParam, "ParamExtType keys not ascending by param");
                prevParam = kv.Key.Param;
                Assert.Equal(kv.Value, item.FindParamExtType(kv.Key.Param, kv.Key.ExtType));
            }

            // a concrete known bucket: (param=1, extType=APPLE) -> apple
            var apple = item.FindParamExtType(1, Protoconf.FruitType.Apple);
            Assert.NotNull(apple);
            Assert.Single(apple!);
            Assert.Equal("apple", apple![0].Name);
            Assert.Null(item.FindParamExtType(999, Protoconf.FruitType.Apple));
        }

        [Fact]
        public void ItemConf_AwardItem_MultiColumnIndex()
        {
            var item = _hub.GetItemConf();
            Assert.NotNull(item);

            var m = item!.FindAwardItemMap();
            foreach (var kv in m)
            {
                Assert.Equal(kv.Value, item.FindAwardItem(kv.Key.Id, kv.Key.Name));
            }

            // apple is keyed by (id=1, name="apple")
            var apple = item.FindAwardItem(1, "apple");
            Assert.NotNull(apple);
            Assert.Single(apple!);
            Assert.Equal(1u, apple![0].Id);
            // FindFirst* variant returns the first match for the same key.
            Assert.NotNull(item.FindFirstAwardItem(1, "apple"));
        }
    }
}
