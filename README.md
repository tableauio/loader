# loader

Configuration loader for Tableau.

## Prerequisites

- Development OS: linux
- Init protobuf: `bash init.sh`

### C++ Loader

- Install: **CMake 2.8+**
- Change into cpp dir: `cd cpp`
- Generate protoconf: `bash ./scripts/gen_pb.sh`
- Create build dir: `mkdir build && cd build`
- Run cmake: `cmake ../src/`
- Build: `make`, then the **bin** dir will be generated at `cpp/bin`.

## References

- [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src)

