#include "load.pc.h"

#include <filesystem>

#include "logger.pc.h"
#include "util.pc.h"

namespace tableau {

// Forward declaration of the PatchMessage function
bool PatchMessage(google::protobuf::Message& dst, const google::protobuf::Message& src);

std::shared_ptr<const MessagerOptions> ParseMessagerOptions(std::shared_ptr<const LoadOptions> opts,
                                                            const std::string& name) {
  std::shared_ptr<MessagerOptions> mopts = std::make_shared<MessagerOptions>();
  if (!opts) {
    return mopts;
  }
  if (auto iter = opts->messager_options.find(name); iter != opts->messager_options.end() && iter->second) {
    mopts = std::make_shared<MessagerOptions>(*iter->second);
  }
  if (!mopts->ignore_unknown_fields.has_value()) {
    mopts->ignore_unknown_fields = opts->ignore_unknown_fields;
  }
  if (mopts->patch_dirs.empty()) {
    mopts->patch_dirs = opts->patch_dirs;
  }
  if (!mopts->mode.has_value()) {
    mopts->mode = opts->mode;
  }
  if (mopts->read_func == nullptr) {
    mopts->read_func = opts->read_func;
  }
  if (mopts->load_func == nullptr) {
    mopts->load_func = opts->load_func;
  }
  return mopts;
}

bool LoadMessagerWithPatch(google::protobuf::Message& msg, const std::filesystem::path& path, Format fmt,
                           tableau::Patch patch, std::shared_ptr<const MessagerOptions> options /* = nullptr*/) {
  auto mode = options->GetMode();
  auto load_func = options->GetLoadFunc();
  if (mode == LoadMode::kOnlyMain) {
    // ignore patch files when LoadMode::kModeOnlyMain specified
    return load_func(msg, path, fmt, nullptr);
  }
  const std::string& name = msg.GetDescriptor()->name();
  std::vector<std::filesystem::path> patch_paths;
  if (!options->patch_paths.empty()) {
    // patch path specified in PatchPaths, then use it instead of PatchDirs.
    patch_paths = options->patch_paths;
  } else {
    std::filesystem::path filename = name + util::Format2Ext(fmt);
    for (auto&& patch_dir : options->patch_dirs) {
      patch_paths.emplace_back(patch_dir / filename);
    }
  }

  std::vector<std::filesystem::path> existed_patch_paths;
  for (auto&& patch_path : patch_paths) {
    if (std::filesystem::exists(patch_path)) {
      existed_patch_paths.emplace_back(patch_path);
    }
  }
  if (existed_patch_paths.empty()) {
    if (mode == LoadMode::kOnlyPatch) {
      // just returns empty message when LoadMode::kModeOnlyPatch specified but no valid patch file provided.
      return true;
    }
    // no valid patch path provided, then just load from the "main" file.
    return load_func(msg, path, fmt, options);
  }

  switch (patch) {
    case tableau::PATCH_REPLACE: {
      // just use the last "patch" file
      std::filesystem::path& patch_path = existed_patch_paths.back();
      if (!load_func(msg, patch_path, util::GetFormat(patch_path), options)) {
        return false;
      }
      break;
    }
    case tableau::PATCH_MERGE: {
      if (mode != LoadMode::kOnlyPatch) {
        // load msg from the "main" file
        if (!load_func(msg, path, fmt, options)) {
          return false;
        }
      }
      // Create a new instance of the same type of the original message
      google::protobuf::Message* patch_msg_ptr = msg.New();
      std::unique_ptr<google::protobuf::Message> _auto_release(msg.New());
      // load patch_msg from each "patch" file
      for (auto&& patch_path : existed_patch_paths) {
        if (!load_func(*patch_msg_ptr, patch_path, util::GetFormat(patch_path), options)) {
          return false;
        }
        if (!PatchMessage(msg, *patch_msg_ptr)) {
          return false;
        }
      }
      break;
    }
    default: {
      SetErrMsg("unknown patch type: " + tableau::Patch_Name(patch));
      return false;
    }
  }
  ATOM_DEBUG("patched(%s) %s by %s: %s", tableau::Patch_Name(patch).c_str(), name.c_str(),
             ATOM_VECTOR_STR(existed_patch_paths).c_str(), msg.ShortDebugString().c_str());
  return true;
}

bool LoadMessager(google::protobuf::Message& msg, const std::filesystem::path& path, Format fmt,
                  std::shared_ptr<const MessagerOptions> options /* = nullptr*/) {
  options = options ? options : std::make_shared<MessagerOptions>();
  std::string content;
  bool ok = options->GetReadFunc()(path, content);
  if (!ok) {
    return false;
  }
  switch (fmt) {
    case Format::kJSON: {
      return util::JSON2Message(content, msg, options->GetIgnoreUnknownFields());
    }
    case Format::kText: {
      return util::Text2Message(content, msg);
    }
    case Format::kBin: {
      return util::Bin2Message(content, msg);
    }
    default: {
      SetErrMsg("unknown format: " + std::to_string(static_cast<int>(fmt)));
      return false;
    }
  }
}

bool LoadMessagerInDir(google::protobuf::Message& msg, const std::filesystem::path& dir, Format fmt,
                       std::shared_ptr<const MessagerOptions> options /* = nullptr*/) {
  options = options ? options : std::make_shared<MessagerOptions>();
  const std::string& name = msg.GetDescriptor()->name();
  std::filesystem::path path;
  if (!options->path.empty()) {
    // path specified in Paths, then use it instead of dir.
    path = options->path;
    fmt = util::GetFormat(path);
  }
  if (path.empty()) {
    std::filesystem::path filename = name + util::Format2Ext(fmt);
    path = dir / filename;
  }

  const google::protobuf::Descriptor* descriptor = msg.GetDescriptor();
  // access the extension directly using the generated identifier
  const tableau::WorksheetOptions& worksheet_options = descriptor->options().GetExtension(tableau::worksheet);
  if (worksheet_options.patch() != tableau::PATCH_NONE) {
    return LoadMessagerWithPatch(msg, path, fmt, worksheet_options.patch(), options);
  }
  return options->GetLoadFunc()(msg, path, fmt, options);
}

#ifdef _WIN32
#undef GetMessage
#endif

// PatchMessage patches src into dst, which must be a message with the same descriptor.
//
// # Default PatchMessage mechanism
//   - scalar: Populated scalar fields in src are copied to dst.
//   - message: Populated singular messages in src are merged into dst by
//     recursively calling [xproto.PatchMessage], or replace dst message if
//     "PATCH_REPLACE" is specified for this field.
//   - list: The elements of every list field in src are appended to the
//     corresponded list fields in dst, or replace dst list if "PATCH_REPLACE"
//     is specified for this field.
//   - map: The entries of every map field in src are MERGED (different from
//     the behavior of proto.Merge) into the corresponding map field in dst,
//     or replace dst map if "PATCH_REPLACE" is specified for this field.
//   - unknown: The unknown fields of src are appended to the unknown
//     fields of dst (TODO: untested).
//
// # References:
//  - https://protobuf.dev/reference/cpp/api-docs/google.protobuf.message/#Reflection
//  - https://protobuf.dev/reference/cpp/api-docs/google.protobuf.descriptor/#Descriptor
//  - https://protobuf.dev/reference/cpp/api-docs/google.protobuf.descriptor/#FieldDescriptor
//  - https://protobuf.dev/reference/cpp/api-docs/google.protobuf.message/#Message.MergeFrom.details
bool PatchMessage(google::protobuf::Message& dst, const google::protobuf::Message& src) {
  const google::protobuf::Descriptor* dst_descriptor = dst.GetDescriptor();
  const google::protobuf::Descriptor* src_descriptor = src.GetDescriptor();
  // Ensure both messages are of the same type
  if (dst_descriptor != src_descriptor) {
    SetErrMsg("dst and src are not messages with the same descriptor");
    ATOM_ERROR("dst %s and src %s are not messages with the same descriptor", dst_descriptor->name().c_str(),
               src_descriptor->name().c_str());
    return false;
  }

  // Get the reflection and descriptor for the messages
  const google::protobuf::Reflection* dst_reflection = dst.GetReflection();
  const google::protobuf::Reflection* src_reflection = src.GetReflection();

  // List all populated fields
  std::vector<const google::protobuf::FieldDescriptor*> fields;
  src_reflection->ListFields(src, &fields);

  // Iterates over every populated field.
  for (auto fd : fields) {
    const tableau::FieldOptions& opts = fd->options().GetExtension(tableau::field);
    tableau::Patch patch = opts.prop().patch();
    if (patch == tableau::PATCH_REPLACE) {
      dst_reflection->ClearField(&dst, fd);
    }
    if (fd->is_map()) {
      // Reference:
      // https://github.com/protocolbuffers/protobuf/blob/95ef4134d3f65237b7adfb66e5e7aa10fcfa1fa3/src/google/protobuf/map_field.cc#L500
      auto key_fd = fd->message_type()->map_key();
      int src_count = src_reflection->FieldSize(src, fd);
      int dst_count = dst_reflection->FieldSize(dst, fd);
      switch (key_fd->cpp_type()) {
#define HANDLE_TYPE(CPPTYPE, METHOD, TYPENAME)                                       \
  case google::protobuf::FieldDescriptor::CPPTYPE_##CPPTYPE: {                       \
    std::unordered_map<TYPENAME, int> dst_key_index_map;                             \
    for (int i = 0; i < dst_count; i++) {                                            \
      auto&& entry = dst_reflection->GetRepeatedMessage(dst, fd, i);                 \
      TYPENAME key = entry.GetReflection()->Get##METHOD(entry, key_fd);              \
      dst_key_index_map[key] = i;                                                    \
    }                                                                                \
    for (int j = 0; j < src_count; j++) {                                            \
      auto&& src_entry = src_reflection->GetRepeatedMessage(src, fd, j);             \
      TYPENAME key = src_entry.GetReflection()->Get##METHOD(src_entry, key_fd);      \
      auto it = dst_key_index_map.find(key);                                         \
      if (it != dst_key_index_map.end()) {                                           \
        int index = it->second;                                                      \
        auto&& dst_entry = *dst_reflection->MutableRepeatedMessage(&dst, fd, index); \
        PatchMessage(dst_entry, src_entry);                                          \
      } else {                                                                       \
        PatchMessage(*dst_reflection->AddMessage(&dst, fd), src_entry);              \
      }                                                                              \
    }                                                                                \
    break;                                                                           \
  }

        HANDLE_TYPE(INT32, Int32, int32_t);
        HANDLE_TYPE(INT64, Int64, int64_t);
        HANDLE_TYPE(UINT32, UInt32, uint32_t);
        HANDLE_TYPE(UINT64, UInt64, uint64_t);
        HANDLE_TYPE(BOOL, Bool, bool);
        HANDLE_TYPE(STRING, String, std::string);
        default: {
          // other types are impossible to be protobuf map key
          ATOM_FATAL("invalid map key type: %d", key_fd->cpp_type());
          break;
        }
#undef HANDLE_TYPE
      }
    } else if (fd->is_repeated()) {
      // Reference:
      // https://github.com/protocolbuffers/protobuf/blob/95ef4134d3f65237b7adfb66e5e7aa10fcfa1fa3/src/google/protobuf/reflection_ops.cc#L68
      int count = src_reflection->FieldSize(src, fd);
      for (int j = 0; j < count; j++) {
        switch (fd->cpp_type()) {
#define HANDLE_TYPE(CPPTYPE, METHOD)                                                        \
  case google::protobuf::FieldDescriptor::CPPTYPE_##CPPTYPE: {                              \
    dst_reflection->Add##METHOD(&dst, fd, src_reflection->GetRepeated##METHOD(src, fd, j)); \
    break;                                                                                  \
  }

          HANDLE_TYPE(INT32, Int32);
          HANDLE_TYPE(INT64, Int64);
          HANDLE_TYPE(UINT32, UInt32);
          HANDLE_TYPE(UINT64, UInt64);
          HANDLE_TYPE(FLOAT, Float);
          HANDLE_TYPE(DOUBLE, Double);
          HANDLE_TYPE(BOOL, Bool);
          HANDLE_TYPE(STRING, String);
          HANDLE_TYPE(ENUM, Enum);
#undef HANDLE_TYPE

          case google::protobuf::FieldDescriptor::CPPTYPE_MESSAGE: {
            const google::protobuf::Message& src_child = src_reflection->GetRepeatedMessage(src, fd, j);
            PatchMessage(*dst_reflection->AddMessage(&dst, fd), src_child);
            break;
          }
        }
      }
    } else {
      switch (fd->cpp_type()) {
#define HANDLE_TYPE(CPPTYPE, METHOD)                                             \
  case google::protobuf::FieldDescriptor::CPPTYPE_##CPPTYPE:                     \
    dst_reflection->Set##METHOD(&dst, fd, src_reflection->Get##METHOD(src, fd)); \
    break;

        HANDLE_TYPE(INT32, Int32);
        HANDLE_TYPE(INT64, Int64);
        HANDLE_TYPE(UINT32, UInt32);
        HANDLE_TYPE(UINT64, UInt64);
        HANDLE_TYPE(FLOAT, Float);
        HANDLE_TYPE(DOUBLE, Double);
        HANDLE_TYPE(BOOL, Bool);
        HANDLE_TYPE(STRING, String);
        HANDLE_TYPE(ENUM, Enum);
#undef HANDLE_TYPE

        case google::protobuf::FieldDescriptor::CPPTYPE_MESSAGE:
          const google::protobuf::Message& src_child = src_reflection->GetMessage(src, fd);
          PatchMessage(*dst_reflection->MutableMessage(&dst, fd), src_child);
          break;
      }
    }
  }

  dst_reflection->MutableUnknownFields(&dst)->MergeFrom(src_reflection->GetUnknownFields(src));
  return true;
}

bool BaseOptions::GetIgnoreUnknownFields() const { return ignore_unknown_fields.value_or(false); }
LoadMode BaseOptions::GetMode() const { return mode.value_or(LoadMode::kAll); }
ReadFunc BaseOptions::GetReadFunc() const { return read_func ? read_func : util::ReadFile; }
LoadFunc BaseOptions::GetLoadFunc() const { return load_func ? load_func : LoadMessager; }
}  // namespace tableau