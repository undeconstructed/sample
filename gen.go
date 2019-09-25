package common

// for generating code

//go:generate protoc -I common -I config --go_out=plugins=grpc:common config/config.proto
//go:generate protoc -I common -I store --go_out=plugins=grpc:common store/store.proto
//go:generate protoc -I common -I user --go_out=plugins=grpc:common user/user.proto
