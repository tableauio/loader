#pragma once
#include <filesystem>
#include <fstream>
#include <functional>
#include <iostream>
#include <mutex>
#include <sstream>
#include <string>

namespace tableau {
namespace log {
enum Level : int {
  kTrace = 0,
  kDebug = 1,
  kInfo = 2,
  kWarn = 3,
  kError = 4,
  kFatal = 5,
};

struct LevelInfo {
  Level level;
  const std::string& name;
};

struct SourceLocation {
  SourceLocation() = default;
  SourceLocation(const char* filename_in, int line_in, const char* funcname_in)
      : filename{filename_in}, line{line_in}, funcname{funcname_in} {}

  bool empty() const { return line == 0; }
  const char* filename{nullptr};
  int line{0};
  const char* funcname{nullptr};
};

using Writer = std::function<void(std::ostream* os, const SourceLocation& loc, const LevelInfo& level,
                                  const std::string& content)>;
// Default write: write the log to output stream.
void DefaultWrite(std::ostream* os, const SourceLocation& loc, const LevelInfo& level, const std::string& content);

// A simple thread-safe logger.
class Logger {
 public:
  Logger() {
    // default: write to stdout
    os_ = &std::cout;
    writer_ = DefaultWrite;
  }
  ~Logger() { ofs_.close(); }
  // Init the logger with the specified path.
  // NOTE: no guarantee of thread-safety.
  int Init(const std::filesystem::path& path, Level level);
  // Set the writer for writing log.
  void SetWriter(Writer writer) { writer_ = writer; }
  // Log with guarantee of thread-safety.
  void Log(const SourceLocation& loc, Level level, const char* format, ...);

 private:
  Level level_ = kTrace;
  std::ofstream ofs_;
  std::ostream* os_ = nullptr;
  std::mutex mutex_;
  Writer writer_;
};

const char* NowStr();
Logger* DefaultLogger();
void SetDefaultLogger(Logger* logger);

template <typename T>
std::string DebugString(const T& a,
                        std::function<void(std::stringstream& ss, typename T::const_iterator it)> print_func) {
  std::stringstream ss;
  ss << "[";
  for (auto it = a.begin(); it != a.end(); ++it) {
    print_func(ss, it);
    auto tmp_it = it;
    if (++tmp_it != a.end()) {
      ss << ", ";
    }
  }
  ss << "]";
  return ss.str();
}

}  // namespace log
}  // namespace tableau

#define ATOM_LOGGER_CALL(logger, level, ...)                                                                     \
  (logger)->Log(tableau::log::SourceLocation{__FILE__, __LINE__, static_cast<const char*>(__FUNCTION__)}, level, \
                __VA_ARGS__)

#define ATOM_TRACE(...) ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), tableau::log::kTrace, __VA_ARGS__)
#define ATOM_DEBUG(...) ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), tableau::log::kDebug, __VA_ARGS__)
#define ATOM_INFO(...) ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), tableau::log::kInfo, __VA_ARGS__)
#define ATOM_WARN(...) ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), tableau::log::kWarn, __VA_ARGS__)
#define ATOM_ERROR(...) ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), tableau::log::kError, __VA_ARGS__)
#define ATOM_FATAL(...) ATOM_LOGGER_CALL(tableau::log::DefaultLogger(), tableau::log::kFatal, __VA_ARGS__)

#define ATOM_VECTOR_STR(v) \
  tableau::log::DebugString(v, [](std::stringstream& ss, decltype(v)::const_iterator it) { ss << *it; })
#define ATOM_MAP_STR(m)                                                                    \
  tableau::log::DebugString(m, [](std::stringstream& ss, decltype(m)::const_iterator it) { \
    ss << "{" << it->first << ": " << it->second << "}";                                   \
  })