package handler

import (
	"testing"

	"github.com/hechh/framework/handler/internal/entity"
)

func TestHandler(t *testing.T) {
	rr := entity.NewRpcHandler[any, struct{}](nil, 0, 0, "hch")
	aaa := rr.New(0)
	t.Log("------>", aaa)
}
