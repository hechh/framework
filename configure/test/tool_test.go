package test

import (
	"fmt"
	"testing"

	"framework/configure/pb"
	_ "framework/configure/pb"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
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

func TestStrcase(t *testing.T) {
	name := "Provider"
	t.Log(strcase.ToLowerCamel(name))

	str := `test%d`
	t.Log(fmt.Sprintf(str, 123))
}

func TestGameConfigAry(t *testing.T) {
	ary := &pb.GameConfigAry{}
	str := `Ary: {
  Id: 1
  IsHot: true
  Sort: 1
  Version: "1.2.00"
  Platforms: PlatformTypeNone
  Currencys: CurrencyNone
  Currencys: CurrencyNone
  ProviderIcon: "spribe.png"
  GameIcon: "aviator.png"
}
Ary: {
  Id: 2
  IsHot: true
  Sort: 2
  Version: "1.3.00"
  Platforms: PlatformTypeNone
  Currencys: CurrencyNone
  Currencys: CurrencyNone
  ProviderIcon: "plinko.png"
  GameIcon: "plinko.png"
}`

	err := prototext.Unmarshal([]byte(str), ary)
	t.Log(err, ary)
}
