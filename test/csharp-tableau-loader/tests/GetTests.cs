using Xunit;

namespace LoaderTests
{
    /// <summary>
    /// Get block: point-lookup getters (map getters). Mirrors the same scenarios in:
    ///   - Go:  test/go-tableau-loader/get_test.go
    ///   - C++: test/cpp-tableau-loader/tests/get_test.cpp
    ///   - TS:  test/ts-tableau-loader/tests/get.test.ts
    /// </summary>
    [Collection("HubCollection")]
    public class GetTests
    {
        private readonly Tableau.Hub _hub;

        public GetTests(HubFixture fixture)
        {
            _hub = fixture.Hub;
        }

        [Fact]
        public void GetItemConf_TypedAndGenericReturnSameInstance()
        {
            var itemConf1 = _hub.Get<Tableau.ItemConf>();
            var itemConf2 = _hub.GetItemConf();
            Assert.NotNull(itemConf1);
            Assert.Same(itemConf1, itemConf2);
        }

        // ---- ItemConf: 1st-level int-keyed map getter ----

        [Fact]
        public void ItemConf_Get1()
        {
            var item = _hub.GetItemConf();
            Assert.NotNull(item);
            Assert.Equal("apple", item!.Get1(1)!.Name);
            Assert.Equal("coin1", item.Get1(0)!.Name);
            Assert.Null(item.Get1(12345));
        }

        // ---- ActivityConf: nested map getters with a 64-bit outer key ----

        [Fact]
        public void ActivityConf_Get()
        {
            var conf = _hub.GetActivityConf();
            Assert.NotNull(conf);

            Assert.Equal("活动1", conf!.Get1(100001)!.ActivityName);
            Assert.Equal("签到活动章1", conf.Get2(100001, 1)!.ChapterName);

            var section = conf.Get3(100001, 1, 2);
            Assert.NotNull(section);
            Assert.Equal(2u, section!.SectionId);

            // sectionRankMap["2001"] === 2
            Assert.Equal(2, conf.Get4(100001, 1, 2, 2001));
        }

        [Fact]
        public void ActivityConf_Get3_NotFound_ReturnsNull()
        {
            var conf = _hub.GetActivityConf();
            Assert.NotNull(conf);
            Assert.Null(conf!.Get3(100001, 1, 999));
        }

        // ---- ThemeConf: string-keyed map getter ----

        [Fact]
        public void ThemeConf_Get1()
        {
            var theme = _hub.GetThemeConf();
            Assert.NotNull(theme);
            Assert.NotEmpty(theme!.Data().ThemeMap);
        }

        // ---- FruitConf: leveled point getters ----

        [Fact]
        public void FruitConf_Get1_Get2()
        {
            var fruit = _hub.GetFruitConf();
            Assert.NotNull(fruit);

            Assert.NotNull(fruit!.Get1(FruitTypes.Apple));
            Assert.Null(fruit.Get1(999));

            // APPLE -> item 1001, price 10
            var item = fruit.Get2(FruitTypes.Apple, 1001);
            Assert.NotNull(item);
            Assert.Equal(1001, item!.Id);
            Assert.Equal(10, item.Price);
            Assert.Null(fruit.Get2(FruitTypes.Apple, 9999));
            Assert.Null(fruit.Get2(999, 1001));
        }
    }
}
