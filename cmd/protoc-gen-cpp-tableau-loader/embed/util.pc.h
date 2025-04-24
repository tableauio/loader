#pragma once

#include <chrono>
#include <functional>
#include <string>

namespace tableau {
namespace util {
// Combine hash values
//
// References:
//  - https://stackoverflow.com/questions/2590677/how-do-i-combine-hash-values-in-c0x
//  - https://stackoverflow.com/questions/17016175/c-unordered-map-using-a-custom-class-type-as-the-key
inline void HashCombine(std::size_t& seed) {}

template <typename T, typename... O>
inline void HashCombine(std::size_t& seed, const T& v, O... others) {
  std::hash<T> hasher;
  seed ^= hasher(v) + 0x9e3779b9 + (seed << 6) + (seed >> 2);
  HashCombine(seed, others...);
}

template <typename T, typename... O>
inline std::size_t SugaredHashCombine(const T& v, O... others) {
  std::size_t seed = 0;  // start with a hash value 0
  HashCombine(seed, v, others...);
  return seed;
}

// Mkdir makes dir recursively.
int Mkdir(const std::string& path);
// GetDir returns all but the last element of path, typically the path's
// directory.
std::string GetDir(const std::string& path);
// GetExt returns the file name extension used by path.
// The extension is the suffix beginning at the final dot
// in the final element of path; it is empty if there is
// no dot.
std::string GetExt(const std::string& path);

class TimeProfiler {
 protected:
  std::chrono::time_point<std::chrono::steady_clock> last_;

 public:
  TimeProfiler() { Start(); }
  void Start() { last_ = std::chrono::steady_clock::now(); }
  // Calculate duration between the last time point and now,
  // and update last time point to now.
  std::chrono::microseconds Elapse() {
    auto now = std::chrono::steady_clock::now();
    auto duration = now - last_;  // This is of type std::chrono::duration
    last_ = now;
    return std::chrono::duration_cast<std::chrono::microseconds>(duration);
  }
};

}  // namespace util
}  // namespace tableau