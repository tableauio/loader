using Xunit;

namespace LoaderTests
{
    /// <summary>
    /// OrderedMap block: GetOrderedMap traversal. Mirrors the same scenarios in:
    ///   - Go:  test/go-tableau-loader/ordered_map_test.go
    ///   - C++: test/cpp-tableau-loader/tests/ordered_map_test.cpp
    ///   - TS:  test/ts-tableau-loader/tests/ordered_map.test.ts
    /// </summary>
    [Collection("HubCollection")]
    public class OrderedMapTests
    {
        private readonly Tableau.Hub _hub;

        public OrderedMapTests(HubFixture fixture)
        {
            _hub = fixture.Hub;
        }

        // ---- ActivityConf: nested ordered map ----

        [Fact]
        public void ActivityConf_GetOrderedMap_Traverses_NonEmpty()
        {
            var conf = _hub.GetActivityConf();
            Assert.NotNull(conf);
            var orderedMap = conf!.GetOrderedMap();
            Assert.NotNull(orderedMap);
            Assert.NotEmpty(orderedMap);

            int activities = 0;
            foreach (var activityPair in orderedMap)
            {
                activities++;
                var chapterOrderedMap = activityPair.Value.Item1;
                Assert.NotNull(chapterOrderedMap);
                // 2nd-level sub-map is reachable and non-empty (every activity
                // has chapters).
                Assert.NotEmpty(chapterOrderedMap);
            }
            Assert.True(activities > 0, "expected at least one activity");
        }

        // ---- ItemConf: 1st-level ordered map ----

        [Fact]
        public void ItemConf_GetOrderedMap_Ascending()
        {
            var item = _hub.GetItemConf();
            Assert.NotNull(item);
            var orderedMap = item!.GetOrderedMap();
            Assert.NotEmpty(orderedMap);

            // keys ascending by id; prev starts at 0 so iteration 1 (key >= 0) holds.
            uint prev = uint.MinValue;
            foreach (var key in orderedMap.Keys)
            {
                Assert.True(key >= prev, $"ItemConf ordered map keys not ascending: {key} after {prev}");
                prev = key;
            }
        }

        // ---- HeroConf: ordered map accessible ----

        [Fact]
        public void HeroConf_GetOrderedMap_Accessible()
        {
            var heroConf = _hub.Get<Tableau.HeroConf>();
            Assert.NotNull(heroConf);
            Assert.NotNull(heroConf!.Data());

            var heroOrderedMap = heroConf.GetOrderedMap();
            Assert.NotNull(heroOrderedMap);
        }
    }
}
