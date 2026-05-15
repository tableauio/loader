using System.IO;
using Xunit;

namespace LoaderTests
{
    public class BinTests
    {
        [Fact]
        public void HeroConf_LoadFromBin_Succeeds()
        {
            Tableau.Registry.Init();
            var heroConf = new Tableau.HeroConf();
            bool ok = heroConf.Load(TestPaths.BinDir, Tableau.Format.Bin);
            Assert.True(ok, $"failed to load HeroConf.binpb: {Tableau.Util.GetErrMsg()}");
            Assert.NotNull(heroConf.Data());
        }

        [Fact]
        public void HeroConf_LoadFromMissingDir_Fails()
        {
            Tableau.Registry.Init();
            var heroConf = new Tableau.HeroConf();
            string missingDir = Path.Combine(TestPaths.TestdataDir, "notexist");
            bool ok = heroConf.Load(missingDir, Tableau.Format.Bin);
            Assert.False(ok);
        }
    }
}
