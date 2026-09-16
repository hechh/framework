package dbpool

import (
	"fmt"

	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object  *DbPool
	Configs map[string]*Config
}

// 初始化
func (d *Component) Init() error {
	// 配置缺少 dbpool 段时 Configs 为空：此处拦截并给出明确提示；
	// 空 map 会让 Init 建出零连接池却不报错，问题被推迟到运行期才暴露
	if len(d.Configs) == 0 {
		err := fmt.Errorf("配置为空")
		mlog.Errorf("[dbpool] 初始化失败，error=%v", err)
		return err
	}
	// 模块初始化
	if err := d.Object.Init(d.Configs); err != nil {
		mlog.Errorf("[dbpool] 初始化失败，error=%v", err)
		return err
	}
	mlog.Infof("[dbpool] 初始化成功")
	SetObject(d.Object)
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	mlog.Infof("[dbpool] 关闭成功")
	d.Object = nil
}
