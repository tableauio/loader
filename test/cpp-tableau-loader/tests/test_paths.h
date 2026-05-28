#pragma once
#include <filesystem>
#include <string>

namespace test {

// TestPaths resolves the testdata directory relative to the test binary location,
// walking up the source tree until "testdata" is found. This keeps tests
// runnable both via direct invocation and via ctest.
class TestPaths {
 public:
  static const std::filesystem::path& Testdata() {
    static const std::filesystem::path kTestdata = Resolve();
    return kTestdata;
  }

  static std::filesystem::path Conf() { return Testdata() / "conf"; }
  static std::filesystem::path PatchConf() { return Testdata() / "patchconf"; }
  static std::filesystem::path PatchConf2() { return Testdata() / "patchconf2"; }
  static std::filesystem::path PatchResult() { return Testdata() / "patchresult"; }

 private:
  static std::filesystem::path Resolve() {
    namespace fs = std::filesystem;
    // Walk up from the current path until we find a "testdata" sibling directory.
    fs::path dir = fs::current_path();
    for (int i = 0; i < 8; ++i) {
      auto candidate = dir / "testdata";
      if (fs::exists(candidate) && fs::is_directory(candidate)) {
        return candidate;
      }
      auto sibling = dir / ".." / "testdata";
      if (fs::exists(sibling) && fs::is_directory(sibling)) {
        return fs::canonical(sibling);
      }
      if (!dir.has_parent_path() || dir.parent_path() == dir) {
        break;
      }
      dir = dir.parent_path();
    }
    throw std::runtime_error("could not locate testdata directory");
  }
};

}  // namespace test
