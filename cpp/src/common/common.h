#pragma once

#define SINGLETON(ClassName)            \
private:                                \
    ClassName() {}                      \
    ClassName(const ClassName&) {}      \
    void operator=(const ClassName&) {} \
                                        \
public:                                 \
    static ClassName& Instance()        \
    {                                   \
        static ClassName instance;      \
        return instance;                \
    }

namespace tableau {
enum class Format {
    kJSON,
    kProtowire,
    kPrototext,
};
}  // namespace tableau