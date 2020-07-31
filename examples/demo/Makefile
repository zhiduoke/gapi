.PHONY: all protobuf

protobuf:
	cd api && \
	protoc -I. -I../../../ --go_out=plugins=grpc:. --descriptor_set_out=http.pd http.proto
