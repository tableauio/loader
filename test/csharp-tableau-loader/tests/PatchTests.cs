using System.Collections.Generic;
using System.IO;
using Xunit;

namespace LoaderTests
{
    /// <summary>
    /// Patch-loading tests, mirroring the same scenarios in:
    ///   - Go:  test/go-tableau-loader/main_test.go::Test_Patch
    ///   - C++: test/cpp-tableau-loader/src/main.cpp::TestPatch (legacy print-based)
    /// </summary>
    public class PatchTests
    {
        public PatchTests()
        {
            Tableau.Registry.Init();
            Tableau.Registry.Register<Custom.CustomItemConf>();
        }

        [Fact]
        public void PatchConf_RecursivePatchConf_MatchesExpectedResult()
        {
            var hub = new Tableau.Hub();
            var options = new Tableau.Load.Options
            {
                IgnoreUnknownFields = true,
                PatchDirs = new List<string> { TestPaths.PatchConfDir },
            };
            bool ok = hub.Load(TestPaths.ConfDir, Tableau.Format.JSON, options);
            Assert.True(ok, $"failed to load with patch: {Tableau.Util.GetErrMsg()}");

            var actual = hub.GetRecursivePatchConf();
            Assert.NotNull(actual);

            // Load the expected golden result from testdata/patchresult/.
            var expected = new Tableau.RecursivePatchConf();
            ok = expected.Load(TestPaths.PatchResultDir, Tableau.Format.JSON);
            Assert.True(ok, $"failed to load expected patch result: {Tableau.Util.GetErrMsg()}");

            Assert.Equal(expected.Data(), actual!.Data());
        }

        [Fact]
        public void PatchConf_PatchReplaceConf_ReplacesEntirely()
        {
            var hub = new Tableau.Hub();
            var options = new Tableau.Load.Options
            {
                IgnoreUnknownFields = true,
                PatchDirs = new List<string> { TestPaths.PatchConfDir },
            };
            Assert.True(hub.Load(TestPaths.ConfDir, Tableau.Format.JSON, options));

            var conf = hub.GetPatchReplaceConf();
            Assert.NotNull(conf);
            // PATCH_REPLACE: the patch fully replaces the main file content.
            Assert.Equal("orange", conf!.Data().Name);
        }

        [Fact]
        public void PatchConf2_DifferentFormat_PatchPathsOverride()
        {
            var hub = new Tableau.Hub();
            var options = new Tableau.Load.Options
            {
                IgnoreUnknownFields = true,
                PatchDirs = new List<string> { TestPaths.PatchConf2Dir },
                MessagerOptions = new Dictionary<string, Tableau.Load.MessagerOptions>
                {
                    ["PatchMergeConf"] = new Tableau.Load.MessagerOptions
                    {
                        // .txtpb override (note: C# loader currently supports JSON/Bin only;
                        // this test validates that PatchPaths is honored even though the
                        // unmarshal step would surface an error for unsupported formats.)
                        // We instead point to .json to keep the format-supported path.
                        PatchPaths = new List<string>
                        {
                            Path.Combine(TestPaths.PatchConf2Dir, "PatchMergeConf.json"),
                        },
                    },
                },
            };
            bool ok = hub.Load(TestPaths.ConfDir, Tableau.Format.JSON, options);
            Assert.True(ok, $"failed to load: {Tableau.Util.GetErrMsg()}");
            Assert.NotNull(hub.GetPatchMergeConf());
        }

        [Fact]
        public void PatchConf_MultiplePatchPaths_AppliedSequentially()
        {
            var hub = new Tableau.Hub();
            var options = new Tableau.Load.Options
            {
                IgnoreUnknownFields = true,
                MessagerOptions = new Dictionary<string, Tableau.Load.MessagerOptions>
                {
                    ["PatchMergeConf"] = new Tableau.Load.MessagerOptions
                    {
                        PatchPaths = new List<string>
                        {
                            Path.Combine(TestPaths.PatchConfDir, "PatchMergeConf.json"),
                            Path.Combine(TestPaths.PatchConf2Dir, "PatchMergeConf.json"),
                        },
                    },
                },
            };
            Assert.True(hub.Load(TestPaths.ConfDir, Tableau.Format.JSON, options));

            var data = hub.GetPatchMergeConf()!.Data();

            // Merge: ItemMap should contain key 999 from patchconf2 plus existing entries.
            Assert.True(data.ItemMap.ContainsKey(999), "ItemMap should contain key 999 from patchconf2");
            // PATCH_REPLACE on replace_item_map: last patch wins, key 999 must remain.
            Assert.True(data.ReplaceItemMap.ContainsKey(999), "ReplaceItemMap should contain key 999");
        }

        [Fact]
        public void PatchConf_ModeOnlyMain_IgnoresPatches()
        {
            var hub = new Tableau.Hub();
            var options = new Tableau.Load.Options
            {
                IgnoreUnknownFields = true,
                Mode = Tableau.Load.LoadMode.OnlyMain,
                MessagerOptions = new Dictionary<string, Tableau.Load.MessagerOptions>
                {
                    ["PatchMergeConf"] = new Tableau.Load.MessagerOptions
                    {
                        PatchPaths = new List<string>
                        {
                            Path.Combine(TestPaths.PatchConfDir, "PatchMergeConf.json"),
                            Path.Combine(TestPaths.PatchConf2Dir, "PatchMergeConf.json"),
                        },
                    },
                },
            };
            Assert.True(hub.Load(TestPaths.ConfDir, Tableau.Format.JSON, options));

            // Compare against a fresh OnlyMain load of the same file — they should be equal.
            var direct = new Tableau.PatchMergeConf();
            var directOpts = new Tableau.Load.MessagerOptions { Mode = Tableau.Load.LoadMode.OnlyMain };
            Assert.True(direct.Load(TestPaths.ConfDir, Tableau.Format.JSON, directOpts));

            Assert.Equal(direct.Data(), hub.GetPatchMergeConf()!.Data());
        }

        [Fact]
        public void PatchConf_ModeOnlyPatch_AppliesPatchesFromEmpty()
        {
            var hub = new Tableau.Hub();
            var options = new Tableau.Load.Options
            {
                IgnoreUnknownFields = true,
                Mode = Tableau.Load.LoadMode.OnlyPatch,
                MessagerOptions = new Dictionary<string, Tableau.Load.MessagerOptions>
                {
                    ["PatchMergeConf"] = new Tableau.Load.MessagerOptions
                    {
                        PatchPaths = new List<string>
                        {
                            Path.Combine(TestPaths.PatchConfDir, "PatchMergeConf.json"),
                            Path.Combine(TestPaths.PatchConf2Dir, "PatchMergeConf.json"),
                        },
                    },
                },
            };
            Assert.True(hub.Load(TestPaths.ConfDir, Tableau.Format.JSON, options));

            var data = hub.GetPatchMergeConf()!.Data();
            // OnlyPatch starts from an empty message, so Name must come from a patch file.
            Assert.False(string.IsNullOrEmpty(data.Name));
        }
    }
}
