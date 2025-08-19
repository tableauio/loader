#include "util.pc.h"

#include <fstream>

#include "logger.pc.h"
#include "tableau/protobuf/tableau.pb.h"

namespace tableau {
static thread_local std::string g_err_msg;
const std::string& GetErrMsg() { return g_err_msg; }
void SetErrMsg(const std::string& msg) { g_err_msg = msg; }

const std::string kUnknownExt = ".unknown";
const std::string kJSONExt = ".json";
const std::string kTextExt = ".txt";
const std::string kBinExt = ".bin";

namespace util {
bool ReadFile(const std::filesystem::path& filename, std::string& content) {
  std::ifstream file(filename);
  if (!file.is_open()) {
    SetErrMsg("failed to open " + filename.string() + ": " + strerror(errno));
    return false;
  }
  content.assign(std::istreambuf_iterator<char>(file), {});
  return true;
}

Format GetFormat(const std::filesystem::path& path) {
  auto ext = path.extension();
  if (ext == kJSONExt) {
    return Format::kJSON;
  } else if (ext == kTextExt) {
    return Format::kText;
  } else if (ext == kBinExt) {
    return Format::kBin;
  } else {
    return Format::kUnknown;
  }
}

const std::string& Format2Ext(Format fmt) {
  switch (fmt) {
    case Format::kJSON:
      return kJSONExt;
    case Format::kText:
      return kTextExt;
    case Format::kBin:
      return kBinExt;
    default:
      return kUnknownExt;
  }
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

// refer: https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/stubs/logging.h
void ProtobufLogHandler(google::protobuf::LogLevel level, const char* filename, int line, const std::string& msg) {
  static const std::unordered_map<int, log::Level> kLevelMap = {{google::protobuf::LOGLEVEL_INFO, log::kInfo},
                                                                {google::protobuf::LOGLEVEL_WARNING, log::kWarn},
                                                                {google::protobuf::LOGLEVEL_ERROR, log::kError},
                                                                {google::protobuf::LOGLEVEL_FATAL, log::kFatal}};
  log::Level lvl = log::kWarn;  // default
  auto iter = kLevelMap.find(level);
  if (iter != kLevelMap.end()) {
    lvl = iter->second;
  }
  ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), lvl, "[libprotobuf %s:%d] %s", filename, line, msg.c_str());
}
}  // namespace util
}  // namespace tableau