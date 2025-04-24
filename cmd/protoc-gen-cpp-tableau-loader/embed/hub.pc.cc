#include "hub.pc.h"

#include <google/protobuf/stubs/logging.h>
#include <google/protobuf/stubs/status.h>
#include <google/protobuf/text_format.h>

#include <string>

#include "logger.pc.h"
#include "messager.pc.h"
#include "registry.pc.h"
#include "util.pc.h"

// Auto-generated includes below

namespace tableau {
bool Hub::Load(const std::string& dir, Format fmt /* = Format::kJSON */, const LoadOptions* options /* = nullptr */) {
  auto msger_map = InternalLoad(dir, fmt, options);
  if (!msger_map) {
    return false;
  }
  bool ok = Postprocess(msger_map);
  if (!ok) {
    return false;
  }
  SetMessagerMap(msger_map);
  return true;
}

bool Hub::AsyncLoad(const std::string& dir, Format fmt /* = Format::kJSON */,
                    const LoadOptions* options /* = nullptr */) {
  auto msger_map = InternalLoad(dir, fmt, options);
  if (!msger_map) {
    return false;
  }
  bool ok = Postprocess(msger_map);
  if (!ok) {
    return false;
  }
  sched_->Dispatch(std::bind(&Hub::SetMessagerMap, this, msger_map));
  return true;
}

int Hub::LoopOnce() { return sched_->LoopOnce(); }

void Hub::InitScheduler() {
  sched_ = new internal::Scheduler();
  sched_->Current();
}

std::shared_ptr<MessagerMap> Hub::InternalLoad(const std::string& dir, Format fmt /* = Format::kJSON */,
                                               const LoadOptions* options /* = nullptr */) const {
  // intercept protobuf error logs
  auto old_handler = google::protobuf::SetLogHandler(util::ProtobufLogHandler);
  auto msger_map = NewMessagerMap();
  for (auto iter : *msger_map) {
    auto&& name = iter.first;
    ATOM_DEBUG("loading %s", name.c_str());
    bool ok = iter.second->Load(dir, fmt, options);
    if (!ok) {
      ATOM_ERROR("load %s failed: %s", name.c_str(), GetErrMsg().c_str());
      // restore to old protobuf log handler
      google::protobuf::SetLogHandler(old_handler);
      return nullptr;
    }
    ATOM_DEBUG("loaded %s", name.c_str());
  }

  // restore to old protobuf log handler
  google::protobuf::SetLogHandler(old_handler);
  return msger_map;
}

std::shared_ptr<MessagerMap> Hub::NewMessagerMap() const {
  std::shared_ptr<MessagerMap> msger_map = std::make_shared<MessagerMap>();
  for (auto&& it : Registry::registrar) {
    if (!options_.filter || options_.filter(it.first)) {
      (*msger_map)[it.first] = it.second();
    }
  }
  return msger_map;
}

std::shared_ptr<MessagerMap> Hub::GetMessagerMap() const { return GetMessagerContainer()->msger_map_; }

void Hub::SetMessagerMap(std::shared_ptr<MessagerMap> msger_map) {
  // replace with thread-safe guarantee.
  std::unique_lock<std::mutex> lock(mutex_);
  msger_container_ = std::make_shared<MessagerContainer>(msger_map);
}

const std::shared_ptr<Messager> Hub::GetMessager(const std::string& name) const {
  auto msger_map = GetMessagerMap();
  if (msger_map) {
    auto it = msger_map->find(name);
    if (it != msger_map->end()) {
      return it->second;
    }
  }
  return nullptr;
}

bool Hub::Postprocess(std::shared_ptr<MessagerMap> msger_map) {
  // create a temporary hub with messager container for post process
  Hub tmp_hub;
  tmp_hub.SetMessagerMap(msger_map);

  // messager-level postprocess
  for (auto iter : *msger_map) {
    auto msger = iter.second;
    bool ok = msger->ProcessAfterLoadAll(tmp_hub);
    if (!ok) {
      SetErrMsg("hub call ProcessAfterLoadAll failed, messager: " + msger->Name());
      return false;
    }
  }
  return true;
}

std::time_t Hub::GetLastLoadedTime() const { return GetMessagerContainer()->last_loaded_time_; }

// Auto-generated template specializations below
MessagerContainer::MessagerContainer(std::shared_ptr<MessagerMap> msger_map /* = nullptr*/)
    : msger_map_(msger_map != nullptr ? msger_map : std::make_shared<MessagerMap>()),
      last_loaded_time_(std::time(nullptr)) {
  // Auto-generated initializations below
}
}  // namespace tableau