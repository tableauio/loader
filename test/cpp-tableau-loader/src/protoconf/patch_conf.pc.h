// Code generated by protoc-gen-cpp-tableau-loader. DO NOT EDIT.
// versions:
// - protoc-gen-cpp-tableau-loader v0.6.0
// - protoc                        v3.19.3
// source: patch_conf.proto

#pragma once
#include <string>

#include "hub.pc.h"
#include "patch_conf.pb.h"

namespace tableau {
class PatchReplaceConf : public Messager {
 public:
  static const std::string& Name() { return kProtoName; };
  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) override;
  const protoconf::PatchReplaceConf& Data() const { return data_; };


 private:
  static const std::string kProtoName;
  protoconf::PatchReplaceConf data_;
};

class PatchMergeConf : public Messager {
 public:
  static const std::string& Name() { return kProtoName; };
  virtual bool Load(const std::string& dir, Format fmt, const LoadOptions* options = nullptr) override;
  const protoconf::PatchMergeConf& Data() const { return data_; };

 public:
  const protoconf::Item* Get(uint32_t id) const;

 private:
  static const std::string kProtoName;
  protoconf::PatchMergeConf data_;
};

}  // namespace tableau

namespace protoconf {
// Here are some type aliases for easy use.
using PatchReplaceConfMgr = tableau::PatchReplaceConf;
using PatchMergeConfMgr = tableau::PatchMergeConf;
}  // namespace protoconf
