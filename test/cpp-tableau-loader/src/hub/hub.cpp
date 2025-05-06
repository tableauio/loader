#include "hub/hub.h"

#include "hub.h"
#include "hub/custom/item/custom_item_conf.h"
#include "protoconf/logger.pc.h"

void LogWrite(std::ostream* os, const tableau::log::SourceLocation& loc, const tableau::log::LevelInfo& lvl,
              const std::string& content) {
  // clang-format off
  *os << tableau::log::NowStr() << " "
    // << std::this_thread::get_id() << "|"
    // << gettid() << " "
    << lvl.name << " [" 
    << loc.filename << ":" << loc.line << "][" 
    << loc.funcname << "]" 
    << content
    << std::endl << std::flush;
  // clang-format on
}

void Hub::Init() {
  // custom log
  tableau::log::DefaultLogger()->SetWriter(LogWrite);
  tableau::Registry::Init();
  InitCustomMessager();
}

void Hub::InitCustomMessager() { tableau::Registry::Register<CustomItemConf>(); }