GCFLAGS=-gcflags "all=-N -l"
SYSTEM=$(shell go env GOOS)
OUTPUT=./output


.PHONY: msgtool pb

pb:
	@echo "Building pb"
	@rm -rf ./configure/pb && mkdir -p ./configure/pb
ifeq (${SYSTEM}, windows)
	protoc.exe -I./configure/protocol ./configure/protocol/*.proto --go_opt paths=source_relative --go_out=./configure/pb
else 
	protoc -I./configure/protocol ./configure/protocol/*.proto paths=source_relative --go_out=./configure/pb
endif 

msgtool:
	@echo "xlsx to proto"
	@rm -rf ./configure/protocol/*.gen.proto
	@go run ./tools/msgtool/main.go -src=./configure/table -dst=./configure/protocol

