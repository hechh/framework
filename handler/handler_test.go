package handler

import (
	"fmt"
	"poker_server/common/pb"
	"reflect"
	"testing"

	"github.com/hechh/framework"
	"github.com/hechh/framework/handler/internal/entity"
	"google.golang.org/protobuf/proto"
)

func TestHandler(t *testing.T) {
	rr := entity.NewRpc[any, struct{}](nil, 0, 0, "hch")
	aaa := rr.New(0)
	t.Log("------>", aaa)
}

func Print(uids ...uint64) {
	uids = append(uids, 110)
}

func PrintCmd(cmd framework.IEnum) {
	if cmd == nil {
		fmt.Println("----->", nil)
	}
}

func TestPrint(t *testing.T) {
	uids := []uint64{1, 2}
	Print(uids...)
	t.Log(uids)

	PrintCmd(nil)
}

func TestReflect(t *testing.T) {
	pbType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	rspType := reflect.TypeOf((*framework.IResponse)(nil)).Elem()
	nType := reflect.TypeOf((*pb.GenRoomIdReq)(nil))
	rType := reflect.TypeOf((*pb.GenRoomIdRsp)(nil))
	t.Log("======>", nType.Implements(pbType))
	t.Log("======>", rType.Implements(rspType))
}
