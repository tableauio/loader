# _lab/go127

Manual smoke module for Go 1.27 generic `Hub.Get[T]` / `MessagerContainer.Get[T]`
(plugin opt `go127=true`). Not wired into `make.py` or CI. Lives under `_lab/`
alongside the TypeScript PoC.

Protos are **not** copied here: [`buf.gen.yaml`](buf.gen.yaml) `inputs` points at
`../../test/proto/{item,common}_conf.proto` and managed mode overrides
`go_package` to this module.

## Prerequisites

- Go 1.27+
- [buf](https://buf.build/) CLI

## Generate

```bash
cd _lab/go127
buf generate
```

Regenerates `protoconf/*.pb.go` (gitignored) and `protoconf/loader/*.pc.go`
for ItemConf (+ `common_conf.proto` deps) only.

## Test

```bash
cd _lab/go127
go test ./...
```

Uses JSON fixtures under `../../test/testdata/conf/`.
