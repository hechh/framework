system=$(shell go env GOOS)
protocol_path=.
gen_go_path=.

# 动态查找 protoc 插件路径（避免 PATH 问题）
PROTOC_GEN_GO := $(shell which protoc-gen-go)
PROTOC_GEN_GO_VTPROTO := $(shell which protoc-gen-go-vtproto)
PROTOC_GO_INJECT_TAG := $(shell which protoc-go-inject-tag)

pb:
	@echo "building core pb"
	@rm -rf ${gen_go_path}/*.pb.go
ifeq ($(system),windows)
	protoc.exe -I${protocol_path} ${protocol_path}/*.proto --go_opt paths=source_relative --go_out=${gen_go_path} --plugin=protoc-gen-go.exe --plugin=protoc-gen-go-vtproto.exe --go-vtproto_out=features=marshal+unmarshal+size+clone,paths=source_relative:${gen_go_path}
	protoc-go-inject-tag.exe -input=${gen_go_path}/*.pb.go -XXX_skip="state,sizeCache,unknownFields"
else
	protoc -I${protocol_path} ${protocol_path}/*.proto --go_opt paths=source_relative --go_out=${gen_go_path} --plugin=protoc-gen-go=$(PROTOC_GEN_GO) --plugin=protoc-gen-go-vtproto=$(PROTOC_GEN_GO_VTPROTO) --go-vtproto_out=features=marshal+unmarshal+size+clone,paths=source_relative:${gen_go_path}
	$(PROTOC_GO_INJECT_TAG) -input=${gen_go_path}/*.pb.go -XXX_skip="state,sizeCache,unknownFields"
endif
