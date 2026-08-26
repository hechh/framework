package pprof

import (
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/hechh/framework/pkg/mlog"
)

type Config struct {
	Addr string `yaml:"pprof_addr,omitempty"` // ip 地址
}

type Pprof struct {
	mu     sync.Mutex
	server *http.Server
	port   int
}

func (p *Pprof) Init(cfg *Config) error {
	p.server = &http.Server{
		Addr:    cfg.Addr,
		Handler: http.DefaultServeMux, // 复用 DefaultServeMux（net/http/pprof 已注册）
	}
	go func() {
		mlog.Infof("[pprof] 启动性能分析服务: http://%s/debug/pprof/", cfg.Addr)
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mlog.Errorf("[pprof] 启动失败: %v", err)
		}
	}()
	return nil
}

func (p *Pprof) Close() {
	if p.server != nil {
		p.mu.Lock()
		defer p.mu.Unlock()
		if err := p.server.Close(); err != nil {
			mlog.Errorf("[pprof] 关闭失败: %v", err)
		} else {
			mlog.Infof("[pprof] 已关闭性能分析服务")
		}
	}
}
