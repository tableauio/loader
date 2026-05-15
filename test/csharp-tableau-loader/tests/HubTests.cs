using Xunit;

namespace LoaderTests
{
    [Collection("HubCollection")]
    public class HubTests
    {
        private readonly Tableau.Hub _hub;

        public HubTests(HubFixture fixture)
        {
            _hub = fixture.Hub;
        }

        [Fact]
        public void TaskConf_FilteredOut_IsNull()
        {
            // HubFixture filters out TaskConf via HubOptions.Filter.
            var taskConf = _hub.Get<Tableau.TaskConf>();
            Assert.Null(taskConf);
        }

        [Fact]
        public void HeroConf_LoadedAndOrderedMapAccessible()
        {
            var heroConf = _hub.Get<Tableau.HeroConf>();
            Assert.NotNull(heroConf);
            Assert.NotNull(heroConf!.Data());

            var heroOrderedMap = heroConf.GetOrderedMap();
            Assert.NotNull(heroOrderedMap);
        }

        [Fact]
        public void GetItemConf_TypedAndGenericReturnSameInstance()
        {
            var itemConf1 = _hub.Get<Tableau.ItemConf>();
            var itemConf2 = _hub.GetItemConf();
            Assert.NotNull(itemConf1);
            Assert.Same(itemConf1, itemConf2);
        }

        [Fact]
        public void CustomItemConf_ProcessAfterLoadAll_ResolvesSpecialItem()
        {
            var customItemConf = _hub.Get<Custom.CustomItemConf>();
            Assert.NotNull(customItemConf);
            Assert.False(string.IsNullOrEmpty(customItemConf!.GetSpecialItemName()));
        }

        [Fact]
        public void ItemConf_FindItemInfoMap_NonEmpty()
        {
            var itemConf = _hub.GetItemConf();
            Assert.NotNull(itemConf);
            var itemInfoMap = itemConf!.FindItemInfoMap();
            Assert.NotNull(itemInfoMap);
            Assert.NotEmpty(itemInfoMap);
        }
    }
}
