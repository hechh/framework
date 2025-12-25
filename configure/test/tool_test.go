package test

import (
	"testing"

	_ "framework/configure/pb"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestPb(t *testing.T) {
	aryType, err := protoregistry.GlobalTypes.FindMessageByName("bit_casino_golang.CurrencyExchangeConfigAry")
	if err != nil {
		t.Fatalf("找不到类型: %v", err)
	}

	cfgType, err := protoregistry.GlobalTypes.FindMessageByName("bit_casino_golang.CurrencyExchangeConfig")
	if err != nil {
		t.Fatalf("找不到类型: %v", err)
	}

	// 创建配置消息
	cfg := dynamicpb.NewMessage(cfgType.Descriptor())
	if idField := cfgType.Descriptor().Fields().ByName("Id"); idField != nil {
		cfg.Set(idField, protoreflect.ValueOfUint32(123))
	}

	// 创建数组消息
	ary := dynamicpb.NewMessage(aryType.Descriptor())
	if aryField := aryType.Descriptor().Fields().ByName("Ary"); aryField != nil {
		list := ary.Mutable(aryField).List()
		list.Append(protoreflect.ValueOfMessage(cfg)) // 使用 ValueOfMessage
	}

	buf, err := protojson.Marshal(ary)
	t.Log(err, "--->", string(buf))
}
