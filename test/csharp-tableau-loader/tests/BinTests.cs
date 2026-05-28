using System.IO;
using Xunit;

namespace LoaderTests
{
    [Collection("HubCollection")]
    public class BinTests
    {
        // Depending on HubFixture guarantees Tableau.Registry.Init() has run
        // exactly once before any test in this class executes, and the
        // collection serialization prevents concurrent registry mutation.
        public BinTests(HubFixture _) { }

        [Fact]
        public void HeroConf_LoadFromBin_Succeeds()
        {
            var heroConf = new Tableau.HeroConf();
            bool ok = heroConf.Load(TestPaths.BinDir, Tableau.Format.Bin);
            Assert.True(ok, $"failed to load HeroConf.binpb: {Tableau.Util.GetErrMsg()}");
            Assert.NotNull(heroConf.Data());
        }

        [Fact]
        public void HeroConf_LoadFromMissingDir_Fails()
        {
            var heroConf = new Tableau.HeroConf();
            string missingDir = Path.Combine(TestPaths.TestdataDir, "notexist");
            bool ok = heroConf.Load(missingDir, Tableau.Format.Bin);
            Assert.False(ok);
        }
    }
}
