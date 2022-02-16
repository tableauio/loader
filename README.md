# loader

Configuration loader for Tableau.

## Prerequisites

- Development OS: linux
- Init protobuf: `bash init.sh`

### C++ Loader

### cpp/

- Install: **CMake 2.8+**
- Change into cpp dir: `cd cpp/`
- Generate protoconf: `bash ./gen.sh`
- Create build dir: `mkdir build && cd build`
- Run cmake: `cmake ../src/`
- Build: `make -j8`, then the **bin** dir will be generated at `cpp/bin`.

### plugin/cmd/protoc-gen-cpp-tableau-loader/

- Install: **CMake 2.8+**
- Change into cpp dir: `cd plugin/cmd/protoc-gen-cpp-tableau-loader/test`
- Generate protoconf: `bash ./gen.sh`
- Create build dir: `mkdir build && cd build`
- Run cmake: `cmake ../src/`
- Build: `make -j8`, then the **bin** dir will be generated at `cpp/bin`.


## References

- [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src)

