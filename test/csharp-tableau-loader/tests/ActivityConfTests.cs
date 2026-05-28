using Xunit;

namespace LoaderTests
{
    [Collection("HubCollection")]
    public class ActivityConfTests
    {
        private readonly Tableau.Hub _hub;

        public ActivityConfTests(HubFixture fixture)
        {
            _hub = fixture.Hub;
        }

        [Fact]
        public void Get3_Found_ReturnsSection()
        {
            var conf = _hub.GetActivityConf();
            Assert.NotNull(conf);
            var section = conf!.Get3(100001, 1, 2);
            Assert.NotNull(section);
            Assert.Equal(2u, section!.SectionId);
        }

        [Fact]
        public void Get3_NotFound_ReturnsNull()
        {
            var conf = _hub.GetActivityConf();
            Assert.NotNull(conf);
            var notFound = conf!.Get3(100001, 1, 999);
            Assert.Null(notFound);
        }

        [Fact]
        public void GetOrderedMap_Traverses_NonEmpty()
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
    }
}
