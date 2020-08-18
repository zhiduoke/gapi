.PHONY: all protobuf

protobuf:
	protoc -I. -I../../../../ --go_out=. --descriptor_set_out=test.pd test.proto
