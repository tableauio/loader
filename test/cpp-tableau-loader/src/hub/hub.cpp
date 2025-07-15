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

bool DefaultHubOptions::Filter(const std::string& name) {
  // all messagers
  return true;
}

std::shared_ptr<tableau::MessagerContainer> DefaultHubOptions::MessagerContainerProvider() {
  // default messager container
  return HubBase<DefaultHubOptions>::Instance().GetMessagerContainer();
}

const std::shared_ptr<tableau::HubOptions> DefaultHubOptions::GetOptions() {
  auto options = std::make_shared<tableau::HubOptions>();
  options->filter = DefaultHubOptions::Filter;
  options->provider = DefaultHubOptions::MessagerContainerProvider;
  return options;
}

void DefaultHubOptions::Init() {
  // custom log
  tableau::log::DefaultLogger()->SetWriter(LogWrite);
  InitCustomMessager();
}

void DefaultHubOptions::InitCustomMessager() { tableau::Registry::Register<CustomItemConf>(); }