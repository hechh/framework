GCFLAGS=-gcflags "all=-N -l"
SYSTEM=$(shell go env GOOS)
OUTPUT=./output


.PHONY: pb xlsx_to_proto

pb:
	@echo "Building pb"
	@rm -rf ./configure/pb && mkdir -p ./configure/pb
ifeq (${SYSTEM}, windows)
	protoc.exe -I./configure/protocol ./configure/protocol/*.proto --go_opt paths=source_relative --go_out=./configure/pb
else 
	protoc -I./configure/protocol ./configure/protocol/*.proto paths=source_relative --go_out=./configure/pb
endif 

xlsx_to_proto:
	@echo "xlsx to proto"
	@rm -rf ./configure/protocol/*.gen.proto
	@go run ./tools/xlsx_to_proto/main.go -src=./configure/table -dst=./configure/protocol

