#include "logger.pc.h"

#include <stdarg.h>
#include <stdio.h>
#include <string.h>
#include <sys/syscall.h>
#include <sys/time.h>
#include <unistd.h>

#include <thread>
#include <unordered_map>

#include "hub.pc.h"

#define gettid() syscall(SYS_gettid)

namespace tableau {
namespace log {

// clang-format off
static const std::unordered_map<int, std::string> kLevelMap = {
    {kTrace, "TRACE"},
    {kDebug, "DEBUG"},
    {kInfo, "INFO"},
    {kWarn, "WARN"},
    {kError, "ERROR"},
    {kFatal, "FATAL"}
};
// clang-format on

static Logger* g_default_logger;

Logger* DefaultLogger() {
  if (g_default_logger == nullptr) {
    g_default_logger = new Logger();
  }
  return g_default_logger;
}

void SetDefaultLogger(Logger* logger) {
  g_default_logger = logger;
}

int Logger::Init(const std::string& path, Level level) {
  std::string dir = path.substr(0, path.find_last_of('/'));
  // prepare the specified directory
  int status = util::Mkdir(dir);
  if (status == -1) {
    return status;
  }
  ofs_.open(path, std::ofstream::out | std::ofstream::app);
  os_ = &ofs_;  // use file stream as output stream
  level_ = level;
  return 0;
}

void Logger::Log(const SourceLocation& loc, Level level, const char* format, ...) {
  if (level < level_) {
    return;
  }
  // scoped auto-release lock.
  std::unique_lock<std::mutex> lock(mutex_);
  static thread_local char content[1024] = {0};
  va_list ap;
  va_start(ap, format);
  vsnprintf(content, sizeof(content), format, ap);
  va_end(ap);

  LevelInfo level_info{level, kLevelMap.at(static_cast<int>(level))};
  writer_(os_, loc, level_info, content);
}

void DefaultWrite(std::ostream* os, const SourceLocation& loc, const LevelInfo& lvl, const std::string& content) {
  // clang-format off
  *os << NowStr() << "|"
    // << std::this_thread::get_id() << "|"
    << gettid() << "|"
    << lvl.name << "|" 
    << loc.filename << ":" << loc.line << "|" 
    << loc.funcname << "|" 
    << content
    << std::endl << std::flush;
  // clang-format on
}

const char* NowStr() {
  static char fmt[64], buf[64];
  struct timeval tv;
  struct tm* tm;

  gettimeofday(&tv, NULL);
  if ((tm = localtime(&tv.tv_sec)) != NULL) {
    // strftime(fmt, sizeof fmt, "%Y-%m-%d %H:%M:%S.%%06u %z", tm);
    strftime(fmt, sizeof fmt, "%Y-%m-%d %H:%M:%S.%%06u", tm);
    snprintf(buf, sizeof buf, fmt, tv.tv_usec);
  }
  return buf;
}

}  // namespace log
}  // namespace tableau
