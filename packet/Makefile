system=$(shell go env GOOS)
protocol_path=.
gen_go_path=.

pb:
	@echo "building core pb"
	@rm -rf ${gen_go_path}/*.pb.go
ifeq ($(system),windows)
	protoc.exe -I${protocol_path} ${protocol_path}/*.proto --go_opt paths=source_relative --go_out=${gen_go_path} --plugin=protoc-gen-go.exe --plugin=protoc-gen-go-vtproto.exe --go-vtproto_out=features=marshal+unmarshal+size+clone,paths=source_relative:${gen_go_path}
	protoc-go-inject-tag.exe -input=${gen_go_path}/*.pb.go -XXX_skip="state,sizeCache,unknownFields"
	goprotoyamltag.exe --filename=${gen_go_path}/yaml.pb.go --write
else
	protoc -I${protocol_path} ${protocol_path}/*.proto --go_opt paths=source_relative --go_out=${gen_go_path} --plugin=protoc-gen-go --plugin=protoc-gen-go-vtproto --go-vtproto_out=features=marshal+unmarshal+size+clone,paths=source_relative:${gen_go_path}
	protoc-go-inject-tag -input=${gen_go_path}/*.pb.go -XXX_skip="state,sizeCache,unknownFields"
	goprotoyamltag --filename=${gen_go_path}/yaml.pb.go --write
endif
