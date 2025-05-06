#pragma once
#include <google/protobuf/message.h>

#include <functional>
#include <string>

#include "util.pc.h"

namespace tableau {
enum class LoadMode {
  kModeDefault,
  kModeOnlyMain,
  kModeOnlyPatch,
};

// ReadFunc reads the config file and returns its content.
using ReadFunc = std::function<bool(const std::string& filename, std::string& content)>;

struct LoadOptions {
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

bool LoadMessageByPath(google::protobuf::Message& msg, const std::string& path, Format fmt = Format::kJSON,
                       const LoadOptions* options = nullptr);
bool LoadMessage(google::protobuf::Message& msg, const std::string& dir, Format fmt = Format::kJSON,
                 const LoadOptions* options = nullptr);
}  // namespace tableau