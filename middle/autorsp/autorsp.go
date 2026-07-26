package autorsp

import (
	"github.com/hechh/framework/common/constant"
	"github.com/hechh/framework/common/define"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/middle/fun"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/logic"
	"github.com/hechh/library/base/uerror"
	"github.com/hechh/library/mlog"
	"google.golang.org/protobuf/proto"
)

func AutoRsp(ctx define.IContext, h handler.IHandler, head *packet.Head, rsp any, reterr error) {
	uerror.SetRspHead(rsp, reterr)

	irsp, ok := rsp.(proto.Message)
	if !ok {
		mlog.Errorf("跨服务转发只支持protobuf协议 func=%s", h.GetActorFuncName())
		return
	}

	var err error
	if head.Reply != "" {
		err = msgbus.Response(head.Reply, irsp)
	} else if head.Cmd > 0 {
		err = msgbus.Send(head, irsp, fun.SetRspClient, fun.CacheRouting)
	} else if head.Back != nil {
		err = msgbus.Send(head, irsp, fun.SetRspBack)
	}

	if err != nil {
		ctx.Error("自动回复失败", err, head, rsp)
	} else if !logic.Has(h.GetMask(), constant.LOG_MASK) {
		ctx.Trace("自动回复成功", err, head, rsp)
	}
}
