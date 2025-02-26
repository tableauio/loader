module github.com/tableauio/loader/cmd/protoc-gen-cpp-tableau-loader

go 1.21

replace github.com/tableauio/loader => ../..

require (
	github.com/iancoleman/strcase v0.3.0
	github.com/tableauio/loader v0.0.0-00010101000000-000000000000
	github.com/tableauio/tableau v0.12.0
	google.golang.org/protobuf v1.36.5
)

require golang.org/x/exp v0.0.0-20230418202329-0354be287a23 // indirect
