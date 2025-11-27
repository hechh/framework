
SYSTEM=$(shell go env GOOS)

.PHONY: pb 

pb:
	@echo "building packet.proto"
	-rm -rf ./*.pb.go
ifeq (${SYSTEM}, windows)
	protoc.exe -I./proto -I./packet ./proto/*.proto --go_out=. 
else # linux darwin(mac)
	protoc -I./proto -I./packet ./proto/*.proto --go_out=. 
endif 


