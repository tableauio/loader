# loader

The configuration loader for Tableau.

## Prerequisites

- Development OS: linux
- Init protobuf: `bash init.sh`

## test

### C++

- Install: **CMake 2.8+**
- Change into dir: `cd test/cpp-tableau-loader`
- Generate protoconf: `bash ./gen.sh`
- Create build dir: `mkdir build && cd build`
- Run cmake: `cmake ../src/`
- Build: `make -j8`, then the **bin** dir will be generated at `test/cpp-tableau-loader/bin`.

### Golang

- Change into dir: `cd test/go-tableau-loader`
- Generate protoconf: `bash ./gen.sh`

## TODO

- [ ] Log: support setting log handler

## References

- [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src)

