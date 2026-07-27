package config

import (
	"encoding/json"
	"testing"
)

type mockConvertor struct {
	names map[string]uint32
}

func (m *mockConvertor) ToUint32(s string) uint32 {
	if v, ok := m.names[s]; ok {
		return v
	}
	return 0
}

func (m *mockConvertor) ToString(i uint32) string {
	for k, v := range m.names {
		if v == i {
			return k
		}
	}
	return ""
}

func TestInit(t *testing.T) {
	conv := &mockConvertor{
		names: map[string]uint32{
			"ACCOUNTSRV": 1,
			"GATESRV":    1<<7 | 1,
		},
	}

	cfg, err := Init("config.yaml", 1, 1, conv)

	buf, _ := json.Marshal(cfg)
	t.Log(string(buf))

	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("Init 返回 nil")
	}

	if cfg.Mysql.UidModValue != 1000000000 {
		t.Errorf("Mysql.UidModValue = %d", cfg.Mysql.UidModValue)
	}
	if cfg.Redis.UidModValue != 1000000000 {
		t.Errorf("Redis.UidModValue = %d", cfg.Redis.UidModValue)
	}
	if cfg.Common.Env != "develop" {
		t.Errorf("Common.Env = %s", cfg.Common.Env)
	}
	if cfg.Logger == nil {
		t.Fatal("Logger 为空")
	}
	if cfg.Logger.Mode != "develop" {
		t.Errorf("Logger.Mode = %s", cfg.Logger.Mode)
	}
}
