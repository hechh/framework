package linkpool

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/core/network/internal/domain"
	"github.com/hechh/library/base/safe"
	"github.com/hechh/library/mlog"
)

type LinkPool struct {
	idGenerator atomic.Uint32
	uids        map[uint64]uint32
	links       map[uint32]domain.IClient
	exitCh      chan struct{}
	wait        sync.WaitGroup
	mutex       sync.RWMutex
}

func NewLinkPool() *LinkPool {
	return &LinkPool{
		exitCh: make(chan struct{}, 1),
		uids:   make(map[uint64]uint32),
		links:  make(map[uint32]domain.IClient),
	}
}

func (d *LinkPool) Init() {
	d.wait.Add(1)
	tt := time.NewTicker(10 * time.Second)
	safe.SafeGo(mlog.Fatalf, func() {
		defer func() {
			tt.Stop()
			d.wait.Done()
		}()

		for {
			select {
			case now := <-tt.C:
				var expired []domain.IClient
				d.mutex.RLock()
				for _, client := range d.links {
					if now.Unix()-client.GetUpdateTime() >= domain.IDLE_INTERVAL {
						expired = append(expired, client)
					}
				}
				d.mutex.RUnlock()

				if len(expired) > 0 {
					d.mutex.Lock()
					for _, cli := range expired {
						delete(d.uids, cli.GetUid())
						delete(d.links, cli.GetId())
						mlog.Infof("清理过期连接. socketId=%d, uid=%d", cli.GetId(), cli.GetUid())
					}
					d.mutex.Unlock()
				}
			case <-d.exitCh:
				return
			}
		}
	})
}

// Close 关闭所有连接
func (d *LinkPool) Close() {
	close(d.exitCh)
	d.wait.Wait()

	// 读取所有连接
	d.mutex.RLock()
	clients := make([]domain.IClient, 0, len(d.links))
	for _, client := range d.links {
		clients = append(clients, client)
	}
	d.mutex.RUnlock()

	// 关闭所有连接
	for _, client := range clients {
		client.Close()
	}
}

func (d *LinkPool) GenSocketId() uint32 {
	return d.idGenerator.Add(1)
}

// Bind 建立 uid ↔ socketId 双向绑定
func (d *LinkPool) Bind(socketId uint32, uid uint64) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	client, ok := d.links[socketId]
	if !ok {
		return false
	}

	if client.SetUid(uid) {
		d.uids[uid] = socketId
		return true
	}
	return false
}

// Add 添加连接
func (d *LinkPool) Add(client domain.IClient) {
	d.mutex.Lock()
	d.links[client.GetId()] = client
	d.mutex.Unlock()
}

// Get 根据 uid 获取连接
func (d *LinkPool) Get(val any) domain.IClient {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	switch vv := val.(type) {
	case uint32:
		if cli, ok := d.links[vv]; ok {
			return cli
		}
	case uint64:
		if sid, ok1 := d.uids[vv]; ok1 {
			if cli, ok := d.links[sid]; ok {
				return cli
			}
		}
	}
	return nil
}

// Remove 移除并关闭连接
func (d *LinkPool) Del(val any) {
	if cli := d.Get(val); cli != nil {
		d.mutex.Lock()
		delete(d.uids, cli.GetUid())
		delete(d.links, cli.GetId())
		d.mutex.Unlock()
		cli.Close()
		mlog.Infof("删除连接. socketId=%d, uid=%d", cli.GetId(), cli.GetUid())
	}
}
