package test

import (
	"testing"

	_ "framework/configure/pb"

	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestTool(t *testing.T) {
	msgType, err := protoregistry.GlobalTypes.FindMessageByName("bit_casino_golang.test")
	if err == protoregistry.NotFound {
		err = nil
	}
	t.Log(msgType, err)
}
