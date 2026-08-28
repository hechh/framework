package kafkago

import (
	"crypto/tls"
	"testing"

	"github.com/hechh/framework/pkg/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGo 构造带指定配置的适配器（不建立真实连接）。
func newTestGo(cfg *kafka.Config) *KafkaGo {
	return &KafkaGo{cfg: cfg}
}

// TestSaslMechanism 验证 SASL 机制选择与大小写兼容。
func TestSaslMechanism(t *testing.T) {
	cases := []struct {
		name     string
		cfg      *kafka.Config
		wantNil  bool
		wantName string
	}{
		{"未配置", &kafka.Config{}, true, ""},
		{"PLAIN", &kafka.Config{SaslMechanism: "PLAIN", SaslUsername: "u", SaslPassword: "p"}, false, "PLAIN"},
		{"SCRAM-SHA-512 小写", &kafka.Config{SaslMechanism: "scram-sha-512", SaslUsername: "u", SaslPassword: "p"}, false, "SCRAM-SHA-512"},
		{"SCRAM-SHA-256", &kafka.Config{SaslMechanism: "SCRAM-SHA-256", SaslUsername: "u", SaslPassword: "p"}, false, "SCRAM-SHA-256"},
		{"未知机制", &kafka.Config{SaslMechanism: "GSSAPI"}, true, ""},
		{"nil 配置", nil, true, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := newTestGo(c.cfg).saslMechanism()
			if c.wantNil {
				assert.Nil(t, m)
				return
			}
			require.NotNil(t, m)
			assert.Equal(t, c.wantName, m.Name())
		})
	}
}

// TestUseTLS 验证安全协议到 TLS 的映射。
func TestUseTLS(t *testing.T) {
	cases := []struct {
		name string
		cfg  *kafka.Config
		want bool
	}{
		{"空协议", &kafka.Config{}, false},
		{"PLAINTEXT", &kafka.Config{SecurityProtocol: "PLAINTEXT"}, false},
		{"SASL_PLAINTEXT", &kafka.Config{SecurityProtocol: "SASL_PLAINTEXT"}, false},
		{"SSL", &kafka.Config{SecurityProtocol: "SSL"}, true},
		{"SASL_SSL 小写", &kafka.Config{SecurityProtocol: "sasl_ssl"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, newTestGo(c.cfg).useTLS())
		})
	}
}

// TestTransport 验证生产端 Writer 的 Transport 注入逻辑。
func TestTransport(t *testing.T) {
	t.Run("无认证返回 nil", func(t *testing.T) {
		assert.Nil(t, newTestGo(&kafka.Config{}).transport())
	})

	t.Run("SASL_PLAINTEXT 仅注入 SASL", func(t *testing.T) {
		tr := newTestGo(&kafka.Config{
			SecurityProtocol: "SASL_PLAINTEXT",
			SaslMechanism:    "PLAIN",
			SaslUsername:     "u",
			SaslPassword:     "p",
		}).transport()
		require.NotNil(t, tr)
		assert.NotNil(t, tr.SASL)
		assert.Nil(t, tr.TLS)
	})

	t.Run("SSL 仅注入 TLS", func(t *testing.T) {
		tr := newTestGo(&kafka.Config{SecurityProtocol: "SSL"}).transport()
		require.NotNil(t, tr)
		assert.Nil(t, tr.SASL)
		require.NotNil(t, tr.TLS)
		assert.Equal(t, uint16(tls.VersionTLS12), tr.TLS.MinVersion)
	})

	t.Run("SASL_SSL 同时注入 SASL 与 TLS", func(t *testing.T) {
		tr := newTestGo(&kafka.Config{
			SecurityProtocol: "SASL_SSL",
			SaslMechanism:    "SCRAM-SHA-512",
			SaslUsername:     "u",
			SaslPassword:     "p",
		}).transport()
		require.NotNil(t, tr)
		assert.NotNil(t, tr.SASL)
		assert.NotNil(t, tr.TLS)
	})
}

// TestDialer 验证消费端 Reader 的 Dialer 注入逻辑。
func TestDialer(t *testing.T) {
	t.Run("无认证为普通拨号器", func(t *testing.T) {
		dl := newTestGo(&kafka.Config{}).dialer()
		require.NotNil(t, dl)
		assert.Nil(t, dl.SASLMechanism)
		assert.Nil(t, dl.TLS)
	})

	t.Run("SASL_SSL 注入 SASL 与 TLS", func(t *testing.T) {
		dl := newTestGo(&kafka.Config{
			SecurityProtocol: "SASL_SSL",
			SaslMechanism:    "SCRAM-SHA-512",
			SaslUsername:     "u",
			SaslPassword:     "p",
		}).dialer()
		require.NotNil(t, dl)
		assert.NotNil(t, dl.SASLMechanism)
		assert.NotNil(t, dl.TLS)
	})
}
