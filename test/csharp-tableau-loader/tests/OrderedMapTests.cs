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

            uint prev = uint.MinValue;
            bool first = true;
            foreach (var key in orderedMap.Keys)
            {
                if (!first)
                {
                    Assert.True(key >= prev, $"ItemConf ordered map keys not ascending: {key} after {prev}");
                }
                first = false;
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
