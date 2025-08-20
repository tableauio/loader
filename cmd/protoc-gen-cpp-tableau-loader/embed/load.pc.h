#pragma once
#include <google/protobuf/message.h>

#include <chrono>
#include <filesystem>
#include <functional>
#include <memory>
#include <optional>
#include <string>
#include <unordered_map>
#include <vector>

#include "tableau/protobuf/tableau.pb.h"
#include "util.pc.h"

namespace tableau {
namespace load {
enum class LoadMode {
  kAll,        // Load all related files
  kOnlyMain,   // Only load the main file
  kOnlyPatch,  // Only load the patch files
};

struct MessagerOptions;

bool LoadMessager(google::protobuf::Message& msg, const std::filesystem::path& path, Format fmt = Format::kJSON,
                  std::shared_ptr<const MessagerOptions> options = nullptr);
bool LoadMessagerInDir(google::protobuf::Message& msg, const std::filesystem::path& dir, Format fmt = Format::kJSON,
                       std::shared_ptr<const MessagerOptions> options = nullptr);
bool LoadMessagerWithPatch(google::protobuf::Message& msg, const std::filesystem::path& path, Format fmt,
                           tableau::Patch patch, std::shared_ptr<const MessagerOptions> options = nullptr);
bool Unmarshal(const std::string& content, google::protobuf::Message& msg, Format fmt,
               std::shared_ptr<const MessagerOptions> options = nullptr);

// ReadFunc reads the config file and returns its content.
using ReadFunc = std::function<bool(const std::filesystem::path& filename, std::string& content)>;
// LoadFunc defines a func which can load message's content based on the given
// path, format, and options.
using LoadFunc = std::function<bool(google::protobuf::Message& msg, const std::filesystem::path& path, Format fmt,
                                    std::shared_ptr<const MessagerOptions> options)>;

// BaseOptions is the common struct for both global-level and messager-level
// options.
struct BaseOptions {
  // Whether to ignore unknown JSON fields during parsing.
  //
  // Refer https://protobuf.dev/reference/cpp/api-docs/google.protobuf.util.json_util/#JsonParseOptions.
  std::optional<bool> ignore_unknown_fields;
  // Specify the directory paths for config patching.
  std::vector<std::filesystem::path> patch_dirs;
  // Specify the loading mode for config patching.
  // Default is LoadMode::kModeAll.
  std::optional<LoadMode> mode;
  // You can specify custom read function to read a config file's content.
  // Default is util::ReadFile.
  ReadFunc read_func;
  // You can specify custom load function to load a messager's content.
  // Default is LoadMessager.
  LoadFunc load_func;

 public:
  inline bool GetIgnoreUnknownFields() const { return ignore_unknown_fields.value_or(false); }
  inline LoadMode GetMode() const { return mode.value_or(LoadMode::kAll); }
  inline ReadFunc GetReadFunc() const { return read_func ? read_func : util::ReadFile; }
  inline LoadFunc GetLoadFunc() const { return load_func ? load_func : LoadMessager; }
};

// Options is the options struct, which contains both global-level and
// messager-level options.
struct Options : public BaseOptions {
  // messager_options maps each messager name to a MessageOptions.
  // If specified, then the messager will be parsed with the given options
  // directly.
  std::unordered_map<std::string, std::shared_ptr<const MessagerOptions>> messager_options;

 public:
  // ParseMessagerOptions parses messager options with both global-level and
  // messager-level options taken into consideration.
  std::shared_ptr<const MessagerOptions> ParseMessagerOptionsByName(const std::string& name) const;
};

// MessagerOptions defines the options for loading a messager.
struct MessagerOptions : public BaseOptions {
  // Path maps each messager name to a corresponding config file path.
  // If specified, then the main messager will be parsed from the file
  // directly, other than the specified load dir.
  std::filesystem::path path;
  // Patch paths maps each messager name to one or multiple corresponding patch file paths.
  // If specified, then main messager will be patched.
  std::vector<std::filesystem::path> patch_paths;
};
}  // namespace load

class Hub;

class Messager {
 public:
  struct Stats {
    std::chrono::microseconds duration;  // total load time consuming.
  };

 public:
  virtual ~Messager() = default;
  static const std::string& Name() = delete;
  const Stats& GetStats() { return stats_; }
  // Load fills message from file in the specified directory and format.
  virtual bool Load(const std::filesystem::path& dir, Format fmt,
                    std::shared_ptr<const load::MessagerOptions> options = nullptr) = 0;
  // Message returns the inner message data.
  virtual const google::protobuf::Message* Message() const { return nullptr; }
  // callback after all messagers loaded.
  virtual bool ProcessAfterLoadAll(const Hub&) { return true; }

 protected:
  // callback after this messager loaded.
  virtual bool ProcessAfterLoad() { return true; };
  Stats stats_;
};
}  // namespace tableau