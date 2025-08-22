#include "hub/hub.h"

#include "hub/custom/item/custom_item_conf.h"
#include "protoconf/logger.pc.h"

void LogWrite(std::ostream* os, const tableau::log::SourceLocation& loc, const tableau::log::LevelInfo& lvl,
              const std::string& content) {
  *os << tableau::log::Now() << " " << lvl.name << " [" << loc.filename << ":" << loc.line << "][" << loc.funcname
      << "]" << content << std::endl
      << std::flush;
}

bool DefaultFilter(const std::string& name) {
  // all messagers except TaskConf
  return name != "TaskConf";
}

std::shared_ptr<tableau::MessagerContainer> DefaultMessagerContainerProvider(const tableau::Hub& hub) {
  // default messager container
  return hub.GetMessagerContainer();
}

void Hub::InitOnce() {
  // custom log
  tableau::log::DefaultLogger()->SetWriter(LogWrite);
  // call base hub's InitOnce
  auto options = std::make_shared<tableau::HubOptions>();
  options->filter = DefaultFilter;
  options->provider = DefaultMessagerContainerProvider;
  tableau::Hub::InitOnce(options);
  // register custom messagers
  InitCustomMessager();
}

void Hub::InitCustomMessager() { tableau::Registry::Register<CustomItemConf>(); }