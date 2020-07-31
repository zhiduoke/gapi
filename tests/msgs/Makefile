all: bench.pb.go simple.pb.go

bench.pb.go: bench.proto
	protoc -I . -I ../../ --go_out=plugins=grpc:. --descriptor_set_out=bench.pd bench.proto

simple.pb.go: simple.proto
	protoc --go_out=. simple.proto
