package main

import (
	"errors"
	"sort"

	"github.com/iancoleman/strcase"
	"github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader/helper"
	"github.com/tableauio/loader/internal/options"
	"github.com/tableauio/tableau/proto/tableaupb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func getAllOrderedFilesAndMessagers(gen *protogen.Plugin) (protofiles []string, fileMessagers map[string][]string) {
	fileMessagers = map[string][]string{}
	for _, f := range gen.Files {
		if !options.NeedGenFile(f) {
			continue
		}
		opts := f.Desc.Options().(*descriptorpb.FileOptions)
		workbook := proto.GetExtension(opts, tableaupb.E_Workbook).(*tableaupb.WorkbookOptions)
		if workbook == nil {
			continue
		}
		protofiles = append(protofiles, f.GeneratedFilenamePrefix)
		var messagers []string
		for _, message := range f.Messages {
			opts, ok := message.Desc.Options().(*descriptorpb.MessageOptions)
			if !ok {
				gen.Error(errors.New("get message options failed"))
			}
			worksheet, ok := proto.GetExtension(opts, tableaupb.E_Worksheet).(*tableaupb.WorksheetOptions)
			if !ok {
				gen.Error(errors.New("get worksheet extension failed"))
			}
			if worksheet != nil {
				messagerName := string(message.Desc.Name())
				messagers = append(messagers, messagerName)
			}
		}
		// sort messagers in one file to keep in order
		sort.Strings(messagers)
		fileMessagers[f.GeneratedFilenamePrefix] = messagers
	}
	// sort all files to keep in order
	sort.Strings(protofiles)
	return
}

// generateHub generates related hub files.
func generateHub(gen *protogen.Plugin) {
	if *shards <= 1 {
		protofiles, fileMessagers := getAllOrderedFilesAndMessagers(gen)

		hppFilename := "hub." + pcExt + ".h"
		g1 := gen.NewGeneratedFile(hppFilename, "")
		helper.GenerateCommonHeader(gen, g1, version)
		g1.P()
		g1.P(hubHpp)
		generateHubHppTplSpec(g1, protofiles, fileMessagers)
		g1.P(msgContainerHpp)
		generateHubHppMsgContainerMembers(g1, protofiles, fileMessagers)
		g1.P(registryHpp)
		g1.P(bottomHpp)

		cppFilename := "hub." + pcExt + ".cc"
		g2 := gen.NewGeneratedFile(cppFilename, "")
		helper.GenerateCommonHeader(gen, g2, version)
		g2.P()
		g2.P(hubCppHeader)
		generateHubCppHeader(g2, protofiles)
		g2.P(hubCpp)
		generateHubCppTplSpec(g2, protofiles, fileMessagers)
		g2.P(msgContainerCpp)
		generateHubCppMsgContainerCtor(g2, protofiles, fileMessagers)
		g2.P(registryCpp)
		generateHubCppRegistry(g2, protofiles, fileMessagers)
		g2.P(bottomCpp)
	} else {
		// sharding
		generateShardedHub(gen)
	}
}

func generateHubHppTplSpec(g *protogen.GeneratedFile, protofiles []string, fileMessagers map[string][]string) {
	for _, proto := range protofiles {
		for _, messager := range fileMessagers[proto] {
			g.P("class ", messager, ";")
			g.P("template <>")
			g.P("const std::shared_ptr<", messager, "> Hub::Get<", messager, ">() const;")
			g.P()
		}
	}
}

func generateHubHppMsgContainerMembers(g *protogen.GeneratedFile, protofiles []string, fileMessagers map[string][]string) {
	for _, proto := range protofiles {
		for _, messager := range fileMessagers[proto] {
			g.P(helper.Indent(1), "std::shared_ptr<", messager, "> ", strcase.ToSnake(messager), "_;")
		}
	}
}

func generateHubCppHeader(g *protogen.GeneratedFile, protofiles []string) {
	for _, proto := range protofiles {
		g.P(`#include "`, proto, ".", pcExt, `.h"`)
	}
	g.P()
}

func generateHubCppTplSpec(g *protogen.GeneratedFile, protofiles []string, fileMessagers map[string][]string) {
	for _, proto := range protofiles {
		for _, messager := range fileMessagers[proto] {
			g.P("template <>")
			g.P("const std::shared_ptr<", messager, "> Hub::Get<", messager, ">() const {;")
			g.P(helper.Indent(1), "return GetMessagerContainer()->", strcase.ToSnake(messager), "_;")
			g.P("}")
			g.P()
		}
	}
}

func generateHubCppMsgContainerCtor(g *protogen.GeneratedFile, protofiles []string, fileMessagers map[string][]string) {
	for _, proto := range protofiles {
		for _, messager := range fileMessagers[proto] {
			g.P(helper.Indent(1), strcase.ToSnake(messager), "_ = std::dynamic_pointer_cast<", messager, `>(GetMessager(`, messager, `::Name()));`)
		}
	}
}

func generateHubCppRegistry(g *protogen.GeneratedFile, protofiles []string, fileMessagers map[string][]string) {
	for _, proto := range protofiles {
		for _, messager := range fileMessagers[proto] {
			g.P(helper.Indent(2), "Register<", messager, ">();")
		}
	}
}

const hubHpp = `#pragma once
#include <ctime>
#include <functional>
#include <mutex>
#include <string>
#include <unordered_map>

#include "load.pc.h"
#include "scheduler.pc.h"

namespace tableau {
class MessagerContainer;
class Hub;

using MessagerMap = std::unordered_map<std::string, std::shared_ptr<Messager>>;
// FilterFunc filter in messagers if returned value is true.
// NOTE: name is the protobuf message name, e.g.: "message ItemConf{...}".
using Filter = std::function<bool(const std::string& name)>;
// MessagerContainerProvider provides a custom MessagerContainer for hub.
// If not specified, the hub's default MessagerContainer will be used.
// NOTE: This func must return non-nil MessagerContainer.
using MessagerContainerProvider = std::function<std::shared_ptr<MessagerContainer>(const Hub&)>;

struct HubOptions {
  // Filter can only filter in certain specific messagers based on the
  // condition that you provide.
  Filter filter;
  // Provide custom MessagerContainer. For keeping configuration access
  // consistent in a coroutine or a transaction.
  MessagerContainerProvider provider;
};

class Hub {
 public:
  Hub();

  // InitOnce inits the hub only once, and the subsequent calls will not take effect.
  void InitOnce(std::shared_ptr<const HubOptions> options);

  /***** Synchronous Loading *****/
  // Load fills messages (in MessagerContainer) from files in the specified directory and format.
  bool Load(const std::filesystem::path& dir, Format fmt = Format::kJSON, std::shared_ptr<const LoadOptions> options = nullptr);

  /***** Asynchronous Loading *****/
  // Load configs into temp MessagerContainer, and you should call LoopOnce() in you app's main loop,
  // in order to take the temp MessagerContainer into effect.
  bool AsyncLoad(const std::filesystem::path& dir, Format fmt = Format::kJSON,
                 std::shared_ptr<const LoadOptions> options = nullptr);
  int LoopOnce();
  // You'd better initialize the scheduler in the main thread.
  void InitScheduler();

  /***** MessagerMap *****/
  std::shared_ptr<MessagerMap> GetMessagerMap() const;
  void SetMessagerMap(std::shared_ptr<MessagerMap> msger_map);

  /***** MessagerContainer *****/
  // This function is exposed only for use in MessagerContainerProvider.
  std::shared_ptr<MessagerContainer> GetMessagerContainer() const { return msger_container_; }

  /***** Access APIs *****/
  template <typename T>
  const std::shared_ptr<T> Get() const;

  template <typename T, typename U, typename... Args>
  const U* Get(Args... args) const;

  template <typename T, typename U, typename... Args>
  const U* GetOrderedMap(Args... args) const;

  // GetLastLoadedTime returns the time when hub's msger_container_ was last set.
  inline std::time_t GetLastLoadedTime() const;

 private:
  std::shared_ptr<MessagerMap> InternalLoad(const std::string& dir, Format fmt = Format::kJSON,
                                            std::shared_ptr<const LoadOptions> options = nullptr) const;
  std::shared_ptr<MessagerMap> NewMessagerMap() const;
  std::shared_ptr<MessagerContainer> GetMessagerContainerWithProvider() const;
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const;

  bool Postprocess(std::shared_ptr<MessagerMap> msger_map);

 private:
  // For thread-safe guarantee during configuration updating.
  std::mutex mutex_;
  // All messagers' container.
  std::shared_ptr<MessagerContainer> msger_container_;
  // Loading scheduler.
  internal::Scheduler* sched_ = nullptr;
  // Init once
  std::once_flag init_once_;
  // Hub options
  std::shared_ptr<const HubOptions> options_;
};

template <typename T>
const std::shared_ptr<T> Hub::Get() const {
  auto msg = GetMessager(T::Name());
  return std::dynamic_pointer_cast<T>(msg);
}

template <typename T, typename U, typename... Args>
const U* Hub::Get(Args... args) const {
  auto msger = Get<T>();
  return msger ? msger->Get(args...) : nullptr;
}

template <typename T, typename U, typename... Args>
const U* Hub::GetOrderedMap(Args... args) const {
  auto msger = Get<T>();
  return msger ? msger->GetOrderedMap(args...) : nullptr;
}
`

const msgContainerHpp = `class MessagerContainer {
  friend class Hub;

 public:
  MessagerContainer(std::shared_ptr<MessagerMap> msger_map = nullptr);

 private:
  std::shared_ptr<MessagerMap> msger_map_;
  std::time_t last_loaded_time_;

 private:
  const std::shared_ptr<Messager> GetMessager(const std::string& name) const;`

const registryHpp = `};

using MessagerGenerator = std::function<std::shared_ptr<Messager>()>;
// messager name -> messager generator
using Registrar = std::unordered_map<std::string, MessagerGenerator>;
class Registry {
  friend class Hub;

 public:
  static void Init();

  template <typename T>
  static void Register();

 private:`

const bottomHpp = `  static std::once_flag once;
  static Registrar registrar;
};

template <typename T>
void Registry::Register() {
  registrar[T::Name()] = []() { return std::make_shared<T>(); };
}
}  // namespace tableau`

const hubCppHeader = `#include "hub.pc.h"

#include "load.pc.h"
#include "logger.pc.h"
#include "util.pc.h"`

const hubCpp = `
namespace tableau {
std::once_flag Registry::once;
Registrar Registry::registrar;

Hub::Hub() { tableau::Registry::Init(); }

void Hub::InitOnce(std::shared_ptr<const HubOptions> options) {
  std::call_once(init_once_, [&]() { options_ = options; });
}

bool Hub::Load(const std::filesystem::path& dir, Format fmt /* = Format::kJSON */,
               std::shared_ptr<const LoadOptions> options /* = nullptr */) {
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

bool Hub::AsyncLoad(const std::filesystem::path& dir, Format fmt /* = Format::kJSON */,
                    std::shared_ptr<const LoadOptions> options /* = nullptr */) {
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
                                               std::shared_ptr<const LoadOptions> options /* = nullptr */) const {
  // intercept protobuf error logs
  auto old_handler = google::protobuf::SetLogHandler(util::ProtobufLogHandler);
  auto opts = ParseLoadOptions(options);
  auto msger_map = NewMessagerMap();
  for (auto iter : *msger_map) {
    auto&& name = iter.first;
    ATOM_DEBUG("loading %s", name.c_str());
    auto mopts = ParseMessagerOptions(opts, name);
    bool ok = iter.second->Load(dir, fmt, mopts);
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
    if (options_ == nullptr || options_->filter == nullptr || options_->filter(it.first)) {
      (*msger_map)[it.first] = it.second();
    }
  }
  return msger_map;
}

std::shared_ptr<MessagerMap> Hub::GetMessagerMap() const { return GetMessagerContainerWithProvider()->msger_map_; }

void Hub::SetMessagerMap(std::shared_ptr<MessagerMap> msger_map) {
  // replace with thread-safe guarantee.
  std::unique_lock<std::mutex> lock(mutex_);
  msger_container_ = std::make_shared<MessagerContainer>(msger_map);
}

const std::shared_ptr<Messager> Hub::GetMessager(const std::string& name) const {
  return GetMessagerContainerWithProvider()->GetMessager(name);
}

std::shared_ptr<MessagerContainer> Hub::GetMessagerContainerWithProvider() const {
  if (options_ != nullptr && options_->provider != nullptr) {
    return options_->provider(*this);
  }
  return msger_container_;
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

std::time_t Hub::GetLastLoadedTime() const { return GetMessagerContainerWithProvider()->last_loaded_time_; }`

const msgContainerCpp = `
MessagerContainer::MessagerContainer(std::shared_ptr<MessagerMap> msger_map /* = nullptr*/)
    : msger_map_(msger_map != nullptr ? msger_map : std::make_shared<MessagerMap>()),
      last_loaded_time_(std::time(nullptr)) {`

const registryCpp = `}

const std::shared_ptr<Messager> MessagerContainer::GetMessager(const std::string& name) const {
  if (msger_map_) {
    auto it = msger_map_->find(name);
    if (it != msger_map_->end()) {
      return it->second;
    }
  }
  return nullptr;
}

void Registry::Init() {
  std::call_once(once, []() {`

const bottomCpp = `  });
}
}  // namespace tableau`
