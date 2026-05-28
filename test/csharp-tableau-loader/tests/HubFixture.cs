using System;
using System.IO;
using Xunit;

namespace LoaderTests
{
    /// <summary>
    /// Common test paths. xUnit runs tests from the build output directory
    /// (e.g. bin/Debug/net8.0/), so we resolve testdata relative to the source
    /// tree by walking up until "testdata" is found.
    /// </summary>
    public static class TestPaths
    {
        public static string TestdataDir { get; } = ResolveTestdataDir();

        public static string ConfDir => Path.Combine(TestdataDir, "conf");
        public static string PatchConfDir => Path.Combine(TestdataDir, "patchconf");
        public static string PatchConf2Dir => Path.Combine(TestdataDir, "patchconf2");
        public static string PatchResultDir => Path.Combine(TestdataDir, "patchresult");
        public static string BinDir => Path.Combine(TestdataDir, "bin");

        private static string ResolveTestdataDir()
        {
            // Walk up from the assembly location, then from the current dir.
            foreach (var start in new[] { AppContext.BaseDirectory, Directory.GetCurrentDirectory() })
            {
                var dir = new DirectoryInfo(start);
                while (dir != null)
                {
                    var candidate = Path.Combine(dir.FullName, "testdata");
                    if (Directory.Exists(candidate))
                    {
                        return candidate;
                    }
                    // Also try one level above (e.g. when run from test/csharp-tableau-loader).
                    var sibling = Path.Combine(dir.FullName, "..", "testdata");
                    if (Directory.Exists(sibling))
                    {
                        return Path.GetFullPath(sibling);
                    }
                    dir = dir.Parent;
                }
            }
            throw new DirectoryNotFoundException("could not locate testdata directory");
        }
    }

    /// <summary>
    /// Provides a per-fixture pre-loaded Hub used by most tests. Loading is the
    /// expensive part; xUnit instantiates the fixture once per collection.
    ///
    /// Also serves as the single owner of <see cref="Tableau.Registry"/>
    /// initialization. The registry is a process-wide static collection that is
    /// not thread-safe, so all test classes that touch it must join
    /// <c>[Collection("HubCollection")]</c> to be serialized by xUnit.
    /// </summary>
    public class HubFixture
    {
        // Guards Registry init across all fixture constructions in the AppDomain.
        // xUnit may instantiate the fixture multiple times when test classes are
        // discovered in parallel; the lock plus the _registryInited flag make
        // Registry.Init() effectively idempotent.
        private static readonly object _registryLock = new object();
        private static bool _registryInited;

        public Tableau.Hub Hub { get; }

        public HubFixture()
        {
            lock (_registryLock)
            {
                if (!_registryInited)
                {
                    Tableau.Registry.Init();
                    Tableau.Registry.Register<Custom.CustomItemConf>();
                    _registryInited = true;
                }
            }

            var options = new Tableau.HubOptions
            {
                Filter = name => name != "TaskConf",
            };
            Hub = new Tableau.Hub(options);

            var loadOptions = new Tableau.Load.Options
            {
                IgnoreUnknownFields = true,
            };
            bool ok = Hub.Load(TestPaths.ConfDir, Tableau.Format.JSON, loadOptions);
            Assert.True(ok, $"hub.Load failed: {Tableau.Util.GetErrMsg()}");
        }
    }

    /// <summary>
    /// All test classes that read from <see cref="Tableau.Registry"/> or build a
    /// <see cref="Tableau.Hub"/> should belong to this collection so xUnit runs
    /// them serially and shares a single <see cref="HubFixture"/> instance.
    /// </summary>
    [CollectionDefinition("HubCollection")]
    public class HubCollection : ICollectionFixture<HubFixture> { }
}
