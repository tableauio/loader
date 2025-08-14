#include "util.pc.h"

#include <google/protobuf/text_format.h>
#include <google/protobuf/util/json_util.h>

#include <filesystem>
#include <fstream>
#include <sstream>
#include <string>

#include "load.pc.h"
#include "logger.pc.h"

namespace tableau {
static thread_local std::string g_err_msg;
const std::string& GetErrMsg() { return g_err_msg; }
void SetErrMsg(const std::string& msg) { g_err_msg = msg; }

const std::string kUnknownExt = ".unknown";
const std::string kJSONExt = ".json";
const std::string kTextExt = ".txt";
const std::string kBinExt = ".bin";

namespace util {
int Mkdir(const std::string& path) {
  std::error_code ec;
  if (!std::filesystem::create_directories(path, ec)) {
    if (ec) {
      std::cerr << "system error: " << ec.message() << std::endl;
      return -1;
    }
  }
  return 0;
}

std::string GetDir(const std::string& path) { return std::filesystem::path(path).parent_path().string(); }

bool ExistsFile(const std::string& filename) { return std::filesystem::exists(filename); }

bool ReadFile(const std::string& filename, std::string& content) {
  std::ifstream file(filename);
  if (!file.is_open()) {
    SetErrMsg("failed to open " + filename + ": " + strerror(errno));
    return false;
  }
  std::stringstream ss;
  ss << file.rdbuf();
  content = ss.str();
  return true;
}

std::string GetExt(const std::string& path) {
  std::size_t pos = path.find_last_of(".");
  if (pos != std::string::npos) {
    return path.substr(pos);
  }
  return kEmpty;
}

Format Ext2Format(const std::string& ext) {
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

bool JSON2Message(const std::string& json, google::protobuf::Message& msg,
                  std::shared_ptr<const MessagerOptions> options /* = nullptr */) {
  google::protobuf::util::Status status;
  if (options != nullptr) {
    google::protobuf::util::JsonParseOptions parse_options;
    parse_options.ignore_unknown_fields = options->ignore_unknown_fields.value_or(false);
    status = google::protobuf::util::JsonStringToMessage(json, &msg, parse_options);
  } else {
    status = google::protobuf::util::JsonStringToMessage(json, &msg);
  }
  if (!status.ok()) {
    SetErrMsg("failed to parse " + GetProtoName(msg) + kJSONExt + ": " + status.ToString());
    return false;
  }
  return true;
}

bool Text2Message(const std::string& text, google::protobuf::Message& msg) {
  if (!google::protobuf::TextFormat::ParseFromString(text, &msg)) {
    SetErrMsg("failed to parse " + GetProtoName(msg) + kTextExt);
    return false;
  }
  return true;
}
bool Bin2Message(const std::string& bin, google::protobuf::Message& msg) {
  if (!msg.ParseFromString(bin)) {
    SetErrMsg("failed to parse " + GetProtoName(msg) + kBinExt);
    return false;
  }
  return true;
}

const std::string& GetProtoName(const google::protobuf::Message& msg) {
  const auto* md = msg.GetDescriptor();
  return md != nullptr ? md->name() : kEmpty;
}

std::string GetPatchName(tableau::Patch patch) {
  auto* descriptor = tableau::Patch_descriptor();
  if (descriptor) {
    auto* value = descriptor->FindValueByNumber(patch);
    if (value) {
      return value->name();
    }
  }
  return std::to_string(static_cast<int>(patch));
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