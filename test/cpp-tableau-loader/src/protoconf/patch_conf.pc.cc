// Code generated by protoc-gen-cpp-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-cpp-tableau-loader v0.4.8
// - protoc                        v3.19.3
// source: patch_conf.proto

#include "patch_conf.pc.h"

namespace tableau {
const std::string PatchReplaceConf::kProtoName = "PatchReplaceConf";

bool PatchReplaceConf::Load(const std::string& dir, Format fmt, const LoadOptions* options /* = nullptr */) {
  bool ok = LoadMessage(data_, dir, fmt, options);
  return ok ? ProcessAfterLoad() : false;
}

const std::string PatchMergeConf::kProtoName = "PatchMergeConf";

bool PatchMergeConf::Load(const std::string& dir, Format fmt, const LoadOptions* options /* = nullptr */) {
  bool ok = LoadMessage(data_, dir, fmt, options);
  return ok ? ProcessAfterLoad() : false;
}

const protoconf::Item* PatchMergeConf::Get(uint32_t id) const {
  auto iter = data_.item_map().find(id);
  if (iter == data_.item_map().end()) {
    return nullptr;
  }
  return &iter->second;
}

}  // namespace tableau
