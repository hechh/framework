
SYSTEM=$(shell go env GOOS)

.PHONY: pb 

pb:
	@echo "building packet.proto"
	-rm -rf ./*.pb.go
ifeq (${SYSTEM}, windows)
	protoc.exe -I./ ./*.proto --go_out=. --go_opt=paths=source_relative
else # linux darwin(mac)
	protoc -I./ ./*.proto --go_out=. --go_opt=paths=source_relative
endif 


