# loader

The configuration loader for Tableau.

## Prerequisites

- Development OS: linux
- Init protobuf: `bash init.sh`

## test

### C++

- Install: **CMake 3.22** or above
- Change dir: `cd test/cpp-tableau-loader`
- Generate protoconf: `bash ./gen.sh`
- Create build dir: `mkdir build && cd build`
- Run cmake: `cmake ../src/`
- Build: `make -j8`, then the **bin** dir will be generated at `test/cpp-tableau-loader/bin`.

### Golang

- Change dir: `cd test/go-tableau-loader`
- Generate protoconf: `bash ./gen.sh`
- Build: `go build`

## References

- [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src)
- [Protocol Buffers C++ Reference](https://protobuf.dev/reference/cpp/)
- [Protocol Buffers Go Reference](https://protobuf.dev/reference/go/)
