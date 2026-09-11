package websocket

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hechh/framework/core/network/internal/domain"
	"github.com/hechh/framework/core/network/internal/frame"
	"github.com/hechh/framework/library/queue"
	"github.com/hechh/framework/packet"
)

// TestConcurrentSendFramesIntact 验证多协程并发 Send 时帧不被写坏。
//
// 背景：连接帧的写入来自多个协程（NATS 派发、actor 任务、连接读循环），而 gorilla 要求
// 应用自行保证同一连接只有一个协程在写。修复前 Send 直接调用 conn.WriteMessage，
// 并发时会命中 gorilla 的 "concurrent write to websocket connection" panic，
// 或两个协程写坏同一块组帧缓冲区导致帧交错（客户端解帧错乱）。
func TestConcurrentSendFramesIntact(t *testing.T) {
	domain.SetEncodeFunc(frame.Encode)
	domain.SetDecodeFunc(frame.Decode)

	server := NewServer()
	defer server.Close()

	clientCh := make(chan *OptimizeClient, 1)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewOptimizeClient(server, conn, "127.0.0.1")
		server.Add(client)
		clientCh <- client
		client.Start() // 阻塞在读循环，连接关闭后返回
	}))
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	cli, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("连接 WebSocket 失败: %v", err)
	}
	defer cli.Close()

	var client *OptimizeClient
	select {
	case client = <-clientCh:
	case <-time.After(3 * time.Second):
		t.Fatal("等待服务端连接就绪超时")
	}

	const (
		senders   = 8  // 并发写协程数
		perSender = 64 // 每个协程发送帧数
	)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for id := 0; id < senders; id++ {
		wg.Add(1)
		go func(senderId int) {
			defer wg.Done()
			<-start // 同时起跑，最大化并发写重叠窗口
			for seq := 0; seq < perSender; seq++ {
				// 负载长度随 senderId/seq 变化，覆盖不同组帧长度分支
				body := bytes.Repeat([]byte{byte('a' + senderId)}, bodyLen(senderId, seq))
				head := &packet.Head{Uid: uint64(senderId + 1), Cmd: uint32(200000 + senderId*2), Seq: uint32(seq)}
				if err := client.Send(head, body); err != nil {
					t.Errorf("Send 失败: senderId=%d, seq=%d, err=%v", senderId, seq, err)
					return
				}
			}
		}(id)
	}
	close(start)
	wg.Wait()

	// 逐帧校验：长度、负载内容、cmd 与 uid 的对应关系全部要正确（错帧/交错会在此暴露）
	total := senders * perSender
	cli.SetReadDeadline(time.Now().Add(10 * time.Second))
	for n := 0; n < total; n++ {
		_, data, err := cli.ReadMessage()
		if err != nil {
			t.Fatalf("读取第 %d/%d 帧失败（帧可能已错乱）: %v", n+1, total, err)
		}
		pack, err := domain.DecodeFrame(data)
		if err != nil {
			t.Fatalf("第 %d/%d 帧解码失败（帧已错乱）: %v", n+1, total, err)
		}
		senderId := int(pack.Head.Uid) - 1
		if senderId < 0 || senderId >= senders {
			t.Fatalf("第 %d 帧 uid 异常: uid=%d", n+1, pack.Head.Uid)
		}
		if want := uint32(200000 + senderId*2); pack.Head.Cmd != want {
			t.Fatalf("第 %d 帧 cmd 与 uid 不匹配: uid=%d, cmd=%d, want=%d", n+1, pack.Head.Uid, pack.Head.Cmd, want)
		}
		wantLen := bodyLen(senderId, int(pack.Head.Seq))
		if len(pack.Body) != wantLen {
			t.Fatalf("第 %d 帧负载长度不符: uid=%d, seq=%d, want=%d, got=%d",
				n+1, pack.Head.Uid, pack.Head.Seq, wantLen, len(pack.Body))
		}
		if want := bytes.Repeat([]byte{byte('a' + senderId)}, wantLen); !bytes.Equal(pack.Body, want) {
			t.Fatalf("第 %d 帧负载内容不符: uid=%d, seq=%d", n+1, pack.Head.Uid, pack.Head.Seq)
		}
	}
}

// bodyLen 测试用负载长度：随 senderId/seq 变化，覆盖不同长度分支
func bodyLen(senderId, seq int) int {
	return 32 + senderId*8 + seq%97
}

// TestSendQueueFullReturnsError 验证待发队列达到软上限时 Send 显式返回错误：
// 慢客户端导致队列堆积时，宁可让调用方感知失败，也不让待发帧无界堆积撑爆内存。
func TestSendQueueFullReturnsError(t *testing.T) {
	domain.SetEncodeFunc(frame.Encode)

	client := &OptimizeClient{
		sendQ:    queue.NewQueue[[]byte](),
		sendWake: make(chan struct{}, 1),
	}
	client.status.Store(true)                    // 模拟连接活跃
	client.sendBytes.Store(SEND_QUEUE_MAX_BYTES) // 模拟队列已达上限
	if err := client.Send(&packet.Head{Cmd: 200010, Uid: 1}, []byte("x")); err == nil {
		t.Fatal("队列达到软上限时 Send 应返回错误")
	}
}
