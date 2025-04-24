#pragma once
#include <google/protobuf/message.h>

#include <chrono>
#include <functional>
#include <string>

namespace tableau {
enum class Format {
  kUnknown,
  kJSON,
  kText,
  kBin,
};

enum class LoadMode {
  kModeDefault,
  kModeOnlyMain,
  kModeOnlyPatch,
};

static const std::string kEmpty = "";

class Hub;

using Postprocessor = std::function<bool(const Hub& hub)>;
// ReadFunc reads the config file and returns its content.
using ReadFunc = std::function<bool(const std::string& filename, std::string& content)>;

struct LoadOptions {
  // postprocessor is called after loading all configurations.
  Postprocessor postprocessor;
  // read_func reads the config file and returns its content.
  ReadFunc read_func;
  // Whether to ignore unknown JSON fields during parsing.
  //
  // Refer https://protobuf.dev/reference/cpp/api-docs/google.protobuf.util.json_util/#JsonParseOptions.
  bool ignore_unknown_fields = false;
  // Paths maps each messager name to a corresponding config file path.
  // If specified, then the main messager will be parsed from the file
  // directly, other than the specified load dir.
  std::unordered_map<std::string, std::string> paths;
  // Patch paths maps each messager name to one or multiple corresponding patch file paths.
  // If specified, then main messager will be patched.
  std::unordered_map<std::string, std::vector<std::string>> patch_paths;
  // Patch dirs specifies the directory paths for config patching.
  std::vector<std::string> patch_dirs;
  // Mode specifies the loading mode for config patching.
  LoadMode mode = LoadMode::kModeDefault;
};

class Messager {
 public:
  struct Stats {
    std::chrono::microseconds duration;  // total load time consuming.
    // TODO: crc32 of config file to decide whether changed or not
    // std::string crc32;
    // int64_t last_modified_time = 0; // unix timestamp
  };

 public:
  virtual ~Messager() = default;
  static const std::string& Name() { return kEmpty; }
  const Stats& GetStats() { return stats_; }
  // Load fills message from file in the specified directory and format.
  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) = 0;
  // Message returns the inner message data.
  virtual const google::protobuf::Message* Message() const { return nullptr; }
  // callback after all messagers loaded.
  virtual bool ProcessAfterLoadAll(const Hub& hub) { return true; }

 protected:
  // callback after this messager loaded.
  virtual bool ProcessAfterLoad() { return true; };
  Stats stats_;
};

namespace util {
// Combine hash values
//
// References:
//  - https://stackoverflow.com/questions/2590677/how-do-i-combine-hash-values-in-c0x
//  - https://stackoverflow.com/questions/17016175/c-unordered-map-using-a-custom-class-type-as-the-key
inline void HashCombine(std::size_t& seed) {}

template <typename T, typename... O>
inline void HashCombine(std::size_t& seed, const T& v, O... others) {
  std::hash<T> hasher;
  seed ^= hasher(v) + 0x9e3779b9 + (seed << 6) + (seed >> 2);
  HashCombine(seed, others...);
}

template <typename T, typename... O>
inline std::size_t SugaredHashCombine(const T& v, O... others) {
  std::size_t seed = 0;  // start with a hash value 0
  HashCombine(seed, v, others...);
  return seed;
}

// Mkdir makes dir recursively.
int Mkdir(const std::string& path);
// GetDir returns all but the last element of path, typically the path's
// directory.
std::string GetDir(const std::string& path);
// GetExt returns the file name extension used by path.
// The extension is the suffix beginning at the final dot
// in the final element of path; it is empty if there is
// no dot.
std::string GetExt(const std::string& path);

class TimeProfiler {
 protected:
  std::chrono::time_point<std::chrono::steady_clock> last_;

 public:
  TimeProfiler() { Start(); }
  void Start() { last_ = std::chrono::steady_clock::now(); }
  // Calculate duration between the last time point and now,
  // and update last time point to now.
  std::chrono::microseconds Elapse() {
    auto now = std::chrono::steady_clock::now();
    auto duration = now - last_;  // This is of type std::chrono::duration
    last_ = now;
    return std::chrono::duration_cast<std::chrono::microseconds>(duration);
  }
};

}  // namespace util
}  // namespace tableau