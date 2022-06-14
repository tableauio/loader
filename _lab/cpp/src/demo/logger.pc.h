#pragma once
#include <sys/stat.h>
#include <sys/types.h>

#include <fstream>
#include <iostream>
#include <mutex>

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

struct SourceLocation {
  SourceLocation() = default;
  SourceLocation(const char* filename_in, int line_in, const char* funcname_in)
      : filename{filename_in}, line{line_in}, funcname{funcname_in} {}

  bool empty() const { return line == 0; }
  const char* filename{nullptr};
  int line{0};
  const char* funcname{nullptr};
};

// A simple multi-threaded logger.
class Logger {
 public:
  Logger() {
    // default: write to stdout
    os_ = &std::cout;
  }
  // Init the logger with the specified path.
  // NOTE: no guarantee of thread-safety.
  int Init(const std::string& path, Level level);
  // Log with guarantee of thread-safety.
  void Log(const SourceLocation& loc, Level level, const char* format, ...);

 private:
  Level level_ = kTrace;
  std::ofstream ofs_;
  std::ostream* os_ = nullptr;
  std::mutex mutex_;
};

const char* NowStr();
Logger* DefaultLogger();
void SetDefaultLogger(Logger* logger);

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
