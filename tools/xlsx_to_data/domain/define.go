package domain

import (
	_ "framework/configure/pb"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type Token int32

const (
	IDENT   Token = 0
	POINTER Token = 1
	ARRAY   Token = 2
	MAP     Token = 3
	GROUP   Token = 4
)

const (
	PROTO_PKG_NAME = "bit_casino_golang"
)

func GetFullName(name string) protoreflect.FullName {
	return protoreflect.FullName(PROTO_PKG_NAME + "." + name)
}

func GetMessageType(name string) (protoreflect.MessageType, error) {
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(GetFullName(name))
	if err == protoregistry.NotFound {
		err = nil
	}
	return msgType, err
}
