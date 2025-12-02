package context

import (
	"fmt"
	"framework/library/mlog"
	"framework/packet"
	"strings"
)

type Remote struct {
	Common
	head *packet.Head
}

func NewRemote(head *packet.Head, actorFunc string) *Remote {
	ret := &Remote{head: head}
	if pos := strings.Index(actorFunc, "."); pos >= 0 {
		ret.actorName = actorFunc[:pos]
		ret.actorName = actorFunc[pos+1:]
	}
	return ret
}

func (d *Remote) GetUid() uint64 {
	return d.head.Uid
}

func (d *Remote) GetActorId() uint64 {
	if d.head.ActorId <= 0 {
		return d.head.Uid
	}
	return d.head.ActorId
}

func (d *Remote) getformat(str string) string {
	src, dst := d.head.Src, d.head.Dst
	if d.head.ActorId > 0 {
		return fmt.Sprintf("[%d] Node(%d:%d) -> Node(%d:%d) %s.%s(%d)\t%s", src.NodeType, src.NodeId, dst.NodeType, dst.NodeId, d.head.Uid, d.actorName, d.funcName, d.head.ActorId, str)
	} else {
		return fmt.Sprintf("[%d] Node(%d:%d) -> Node(%d:%d) %s.%s(%d)\t%s", d.head.Uid, src.NodeType, src.NodeId, dst.NodeType, dst.NodeId, d.actorName, d.funcName, d.head.Uid, str)
	}
}

func (d *Remote) Tracef(format string, args ...any) {
	mlog.Trace(1, d.getformat(format), args...)
}

func (d *Remote) Debugf(format string, args ...any) {
	mlog.Debug(1, d.getformat(format), args...)
}

func (d *Remote) Warnf(format string, args ...any) {
	mlog.Warn(1, d.getformat(format), args...)
}

func (d *Remote) Infof(format string, args ...any) {
	mlog.Info(1, d.getformat(format), args...)
}

func (d *Remote) Errorf(format string, args ...any) {
	mlog.Error(1, d.getformat(format), args...)
}

func (d *Remote) Fatalf(format string, args ...any) {
	mlog.Fatal(1, d.getformat(format), args...)
}
