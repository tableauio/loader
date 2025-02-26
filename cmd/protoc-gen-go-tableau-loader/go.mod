module github.com/tableauio/loader/cmd/protoc-gen-go-tableau-loader

go 1.21

toolchain go1.23.4

require (
	github.com/iancoleman/strcase v0.3.0
	github.com/tableauio/loader v0.1.0
	github.com/tableauio/tableau v0.12.0
	google.golang.org/protobuf v1.36.5
)

require golang.org/x/exp v0.0.0-20230418202329-0354be287a23 // indirect

replace github.com/tableauio/loader => ../..
