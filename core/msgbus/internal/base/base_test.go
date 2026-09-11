package base

import (
	"testing"

	"github.com/hechh/framework/packet"
)

// TestPacketHandler_RejectsInvalidPacket 验证 PacketHandler 对非法包（nil / 缺少 Head）兜底。
//
// 这类包在 NATS 订阅回调里可由"空 body 或不含 Head 字段的 body"反序列化得到（proto 反序列化返回 nil error），
// 若不拦截，下游会在异步投递协程中 nil 解引用 panic；该协程无 recover，会直接打崩整个进程。
func TestPacketHandler_RejectsInvalidPacket(t *testing.T) {
	called := false
	SetPacketFunc(func(*packet.Packet) { called = true })
	defer SetPacketFunc(nil)

	PacketHandler(nil)
	PacketHandler(&packet.Packet{})
	if called {
		t.Fatal("非法包不应投递给上层处理器")
	}

	PacketHandler(&packet.Packet{Head: &packet.Head{Cmd: 1}})
	if !called {
		t.Fatal("合法包应正常投递给上层处理器")
	}
}
