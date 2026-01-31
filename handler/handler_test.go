package handler

import (
	"fmt"
	"testing"

	"github.com/hechh/framework"
	"github.com/hechh/framework/handler/internal/entity"
)

func TestHandler(t *testing.T) {
	rr := entity.NewRpcHandler[any, struct{}](nil, 0, 0, "hch")
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
