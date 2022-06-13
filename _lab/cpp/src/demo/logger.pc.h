#pragma once
#include <sys/stat.h>
#include <sys/types.h>

#include <fstream>
#include <iostream>
#include <mutex>

namespace tableau {

enum LogLevel : int {
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
  int Init(const std::string& path);
  // Log with guarantee of thread-safety.
  void Log(const SourceLocation& loc, LogLevel level, const char* format, ...);

 private:
  std::ofstream ofs_;
  std::ostream* os_;
  std::mutex mutex_;
};

const char* NowStr();
Logger* DefaultLogger();
void SetDefaultLogger(Logger* logger);
}  // namespace tableau

#define ATOM_LOGGER_CALL(logger, level, ...) \
  (logger)->Log(tableau::SourceLocation{__FILE__, __LINE__, static_cast<const char*>(__FUNCTION__)}, level, __VA_ARGS__)

#define ATOM_TRACE(...) ATOM_LOGGER_CALL(tableau::DefaultLogger(), tableau::kTrace, __VA_ARGS__)
#define ATOM_DEBUG(...) ATOM_LOGGER_CALL(tableau::DefaultLogger(), tableau::kDebug, __VA_ARGS__)
#define ATOM_INFO(...) ATOM_LOGGER_CALL(tableau::DefaultLogger(), tableau::kInfo, __VA_ARGS__)
#define ATOM_WARN(...) ATOM_LOGGER_CALL(tableau::DefaultLogger(), tableau::kWarn, __VA_ARGS__)
#define ATOM_ERROR(...) ATOM_LOGGER_CALL(tableau::DefaultLogger(), tableau::kError, __VA_ARGS__)
#define ATOM_FATAL(...) ATOM_LOGGER_CALL(tableau::DefaultLogger(), tableau::kFatal, __VA_ARGS__)
