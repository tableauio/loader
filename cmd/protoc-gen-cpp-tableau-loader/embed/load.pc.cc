#include "load.pc.h"

#if __cplusplus >= 201703L
#include <filesystem>
#endif

#include "logger.pc.h"
#include "util.pc.h"

namespace tableau {
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
      auto value_fd = fd->message_type()->map_value();
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

bool LoadMessageWithPatch(google::protobuf::Message& msg, const std::string& path, Format fmt, tableau::Patch patch,
                          const LoadOptions* options /* = nullptr*/) {
  if (options == nullptr) {
    return LoadMessageByPath(msg, path, fmt, nullptr);
  }
  if (options->mode == LoadMode::kModeOnlyMain) {
    // ignore patch files when LoadMode::kModeOnlyMain specified
    return LoadMessageByPath(msg, path, fmt, nullptr);
  }
  std::string name = util::GetProtoName(msg);
  std::vector<std::string> patch_paths;
  auto iter = options->patch_paths.find(name);
  if (iter != options->patch_paths.end()) {
    // patch path specified in PatchPaths, then use it instead of PatchDirs.
    patch_paths = iter->second;
  } else {
    std::string filename = name + util::Format2Ext(fmt);
    for (auto&& patch_dir : options->patch_dirs) {
#if __cplusplus >= 201703L
      patch_paths.emplace_back((std::filesystem::path(patch_dir) / filename).make_preferred().string());
#else
      patch_paths.emplace_back(patch_dir + kPathSeperator + filename);
#endif
    }
  }

  std::vector<std::string> existed_patch_paths;
  for (auto&& patch_path : patch_paths) {
    if (util::ExistsFile(patch_path)) {
      existed_patch_paths.emplace_back(patch_path);
    }
  }
  if (existed_patch_paths.empty()) {
    if (options->mode == LoadMode::kModeOnlyPatch) {
      // just returns empty message when LoadMode::kModeOnlyPatch specified but no valid patch file provided.
      return true;
    }
    // no valid patch path provided, then just load from the "main" file.
    return LoadMessageByPath(msg, path, fmt, options);
  }

  switch (patch) {
    case tableau::PATCH_REPLACE: {
      // just use the last "patch" file
      std::string& patch_path = existed_patch_paths.back();
      if (!LoadMessageByPath(msg, patch_path, util::Ext2Format(util::GetExt(patch_path)), options)) {
        return false;
      }
      break;
    }
    case tableau::PATCH_MERGE: {
      if (options->mode != LoadMode::kModeOnlyPatch) {
        // load msg from the "main" file
        if (!LoadMessageByPath(msg, path, fmt, options)) {
          return false;
        }
      }
      // Create a new instance of the same type of the original message
      google::protobuf::Message* patch_msg_ptr = msg.New();
      std::unique_ptr<google::protobuf::Message> _auto_release(msg.New());
      // load patch_msg from each "patch" file
      for (auto&& patch_path : existed_patch_paths) {
        if (!LoadMessageByPath(*patch_msg_ptr, patch_path, util::Ext2Format(util::GetExt(patch_path)), options)) {
          return false;
        }
        if (!PatchMessage(msg, *patch_msg_ptr)) {
          return false;
        }
      }
      break;
    }
    default: {
      SetErrMsg("unknown patch type: " + util::GetPatchName(patch));
      return false;
    }
  }
  ATOM_DEBUG("patched(%s) %s by %s: %s", util::GetPatchName(patch).c_str(), name.c_str(),
             ATOM_VECTOR_STR(existed_patch_paths).c_str(), msg.ShortDebugString().c_str());
  return true;
}

bool LoadMessageByPath(google::protobuf::Message& msg, const std::string& path, Format fmt,
                       const LoadOptions* options /* = nullptr*/) {
  std::string content;
  ReadFunc read_func = util::ReadFile;
  if (options != nullptr && options->read_func) {
    read_func = options->read_func;
  }
  bool ok = read_func(path, content);
  if (!ok) {
    return false;
  }
  switch (fmt) {
    case Format::kJSON: {
      return util::JSON2Message(content, msg, options);
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

bool LoadMessage(google::protobuf::Message& msg, const std::string& dir, Format fmt,
                 const LoadOptions* options /* = nullptr*/) {
  std::string name = util::GetProtoName(msg);
  std::string path;
  if (options) {
    auto iter = options->paths.find(name);
    if (iter != options->paths.end()) {
      // path specified in Paths, then use it instead of dir.
      path = iter->second;
      fmt = util::Ext2Format(util::GetExt(iter->second));
    }
  }
  if (path.empty()) {
    std::string filename = name + util::Format2Ext(fmt);
#if __cplusplus >= 201703L
    path = (std::filesystem::path(dir) / filename).make_preferred().string();
#else
    path = dir + kPathSeperator + filename;
#endif
  }

  const google::protobuf::Descriptor* descriptor = msg.GetDescriptor();
  if (!descriptor) {
    SetErrMsg("failed to get descriptor of message: " + name);
    return false;
  }
  // access the extension directly using the generated identifier
  const tableau::WorksheetOptions worksheet_options = descriptor->options().GetExtension(tableau::worksheet);
  if (worksheet_options.patch() != tableau::PATCH_NONE) {
    return LoadMessageWithPatch(msg, path, fmt, worksheet_options.patch(), options);
  }

  return LoadMessageByPath(msg, path, fmt, options);
}
}  // namespace tableau