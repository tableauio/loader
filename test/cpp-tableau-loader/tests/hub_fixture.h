// Shared test fixture and constants for the C++ tableau loader tests.
//
// The unit tests are split into four blocks, one file each, mirroring the
// Go / C# / TS test layout:
//   - load_test.cpp        (Load, CustomConf, Bin, Patch)
//   - get_test.cpp         (point-lookup getters)
//   - ordered_map_test.cpp (GetOrderedMap traversal)
//   - index_test.cpp       (Find* index finders)
#pragma once

#include <gtest/gtest.h>

#include <memory>

#include "hub/hub.h"
#include "protoconf/hub.pc.h"
#include "protoconf/test_conf.pc.h"
#include "tests/test_paths.h"

// fruitType values match FruitConf.json (== protoconf::FruitType enum values).
constexpr int kFruitTypeApple = protoconf::FRUIT_TYPE_APPLE;
constexpr int kFruitTypeOrange = protoconf::FRUIT_TYPE_ORANGE;
constexpr int kFruitTypeBanana = protoconf::FRUIT_TYPE_BANANA;

// HubFixture loads the whole hub once per test, so test order doesn't matter
// (Hub::Instance() is a process-wide singleton). Mirrors the shared prepareHub
// helper in Go and the HubFixture in C#.
class HubFixture : public ::testing::Test {
 protected:
  void SetUp() override {
    Hub::Instance().InitOnce();
    auto options = std::make_shared<tableau::load::Options>();
    options->ignore_unknown_fields = true;
    auto mopts = std::make_shared<tableau::load::MessagerOptions>();
    mopts->path = (test::TestPaths::Conf() / "ItemConf.json").string();
    options->messager_options["ItemConf"] = mopts;

    bool ok = Hub::Instance().Load(test::TestPaths::Conf().string() + "/", tableau::Format::kJSON, options);
    ASSERT_TRUE(ok) << "hub load failed: " << tableau::GetErrMsg();
  }
};
