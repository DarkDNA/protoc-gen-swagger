package protobuf

//go:generate protoc -I .:/usr/local/include --go_out plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. darkdna/api/api.proto
//go:generate protoc -I .:/usr/local/include --go_out plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. google/api/http.proto google/api/annotations.proto

type foo struct {}