package base

import (
	"io"
	"time"

	"github.com/hechh/framework/pkg/mlog"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

// SetupEngine 配置引擎连接池参数（mysql/postgres 共用）
func SetupEngine(eg *xorm.EngineGroup) {
	eg.SetLogger(log.NewSimpleLogger(io.Discard))
	eg.SetMaxIdleConns(10)
	eg.SetMaxOpenConns(200)
	// 连接最大存活时间，避免使用已被数据库关闭的死连接
	eg.SetConnMaxLifetime(1 * time.Hour)
	// 空闲连接最大存活时间，超时后主动回收
	eg.SetConnMaxIdleTime(10 * time.Minute)
}

// SyncTables 检查并同步表结构。
// synced 标记首次同步是否已完成，避免重连时重复检查。
//
// 策略：
//   - 全部不存在 → 新环境，执行 Sync2 建表
//   - 全部已存在 → 已有库，跳过 Sync2
//   - 部分存在     → 警告并跳过，由手动迁移处理
func SyncTables(eg *xorm.EngineGroup, tables []any, synced *bool) error {
	if len(tables) == 0 || *synced {
		return nil
	}

	missingCount := 0
	existCount := 0
	for _, t := range tables {
		exist, err := eg.IsTableExist(t)
		if err != nil {
			return err
		}
		if exist {
			existCount++
		} else {
			missingCount++
		}
	}

	switch {
	case missingCount == 0:
		mlog.Debugf("所有表已存在，跳过 Sync2")
	case existCount == 0:
		if err := eg.Sync2(tables...); err != nil {
			return err
		}
		mlog.Infof("Sync2 表结构同步完成，共 %d 张表", len(tables))
	default:
		mlog.Warnf("检测到 %d/%d 张表已存在但并非全部，跳过自动 Sync2，请手动处理数据库迁移", existCount, len(tables))
	}

	*synced = true
	return nil
}
