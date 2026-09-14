package aes

import (
	"bytes"
	"encoding/base64"
	"testing"
)

// TestAesGcmRoundTrip 正常路径：加密后可原样解回
func TestAesGcmRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef")
	plain := []byte(`{"username":"admin","password":"p@ssw0rd"}`)

	cipher, err := AesGcmEncrypto(plain, key)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	got, err := AesGcmDecrypto([]byte(cipher), key)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Errorf("解密结果=%q, 期望 %q", got, plain)
	}
}

// TestAesGcmNonDeterministic 相同明文两次加密必须不同（随机 nonce），
// 这正是 ECB（AesEncrypto）做不到的：ECB 相同明文恒产生相同密文。
func TestAesGcmNonDeterministic(t *testing.T) {
	key := []byte("0123456789abcdef")
	plain := []byte("same-plaintext")

	c1, err := AesGcmEncrypto(plain, key)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	c2, err := AesGcmEncrypto(plain, key)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if c1 == c2 {
		t.Errorf("两次加密结果相同, GCM 应使用随机 nonce: %s", c1)
	}

	// 对照：ECB 是确定性的（旧协议行为，仅作说明）
	e1, _ := AesEncrypto(plain, key)
	e2, _ := AesEncrypto(plain, key)
	if e1 != e2 {
		t.Errorf("AesEncrypto(ECB) 应为确定性加密, 实际两次不同")
	}
}

// TestAesGcmTamperDetected 密文被篡改（翻转比特 / 16 字节块重排 / 截断）必须解密失败。
// 这是 C18 的核心：ECB 无 MAC，改包者可以重排、复制块而不被发现；GCM 的 tag 会拦住。
func TestAesGcmTamperDetected(t *testing.T) {
	key := []byte("0123456789abcdef")
	plain := []byte("AAAAAAAAAAAAAAAABBBBBBBBBBBBBBBB") // 恰好 2 个 AES 块，便于块级重排

	cipher, err := AesGcmEncrypto(plain, key)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		t.Fatalf("Base64 解码失败: %v", err)
	}

	// ① 翻转最后一个字节（落在 tag 上）
	flipped := append([]byte(nil), raw...)
	flipped[len(flipped)-1] ^= 0x01
	if _, err := AesGcmDecrypto(flipped, key); err == nil {
		t.Errorf("被翻转比特的密文应校验失败")
	}

	// ② 交换两个 16 字节密文块的顺序（nonce 与 tag 保持原位）
	reordered := append([]byte(nil), raw...)
	body := reordered[GCM_NONCE_SIZE : len(reordered)-16]
	swap := make([]byte, 16)
	copy(swap, body[:16])
	copy(body[:16], body[16:32])
	copy(body[16:32], swap)
	if _, err := AesGcmDecrypto(reordered, key); err == nil {
		t.Errorf("密文块重排后应校验失败")
	}

	// ③ 复制第一个密文块覆盖第二个块
	duplicated := append([]byte(nil), raw...)
	dupBody := duplicated[GCM_NONCE_SIZE : len(duplicated)-16]
	copy(dupBody[16:32], dupBody[:16])
	if _, err := AesGcmDecrypto(duplicated, key); err == nil {
		t.Errorf("密文块复制后应校验失败")
	}

	// ④ 截断（去掉 tag 的最后一个字节）
	if _, err := AesGcmDecrypto(raw[:len(raw)-1], key); err == nil {
		t.Errorf("被截断的密文应校验失败")
	}
}

// TestAesGcmWrongKey 异密钥必须解密失败
func TestAesGcmWrongKey(t *testing.T) {
	cipher, err := AesGcmEncrypto([]byte("secret"), []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if _, err := AesGcmDecrypto([]byte(cipher), []byte("fedcba9876543210")); err == nil {
		t.Errorf("异密钥应解密失败")
	}
}

// TestAesGcmShortInput 长度不足（< nonce + tag）必须报错，且空输入不 panic
func TestAesGcmShortInput(t *testing.T) {
	key := []byte("0123456789abcdef")
	for _, in := range [][]byte{{}, []byte("x"), bytes.Repeat([]byte{0}, GCM_NONCE_SIZE+15)} {
		if _, err := AesGcmDecrypto(in, key); err == nil {
			t.Errorf("长度不足的输入 %d 字节应报错", len(in))
		}
	}
}

// TestAesGcmKeySizes 支持 16/24/32 字节密钥
func TestAesGcmKeySizes(t *testing.T) {
	for _, size := range []int{16, 24, 32} {
		key := bytes.Repeat([]byte{0x2a}, size)
		plain := []byte("payload")
		cipher, err := AesGcmEncrypto(plain, key)
		if err != nil {
			t.Fatalf("密钥长度 %d 加密失败: %v", size, err)
		}
		got, err := AesGcmDecrypto([]byte(cipher), key)
		if err != nil {
			t.Fatalf("密钥长度 %d 解密失败: %v", size, err)
		}
		if !bytes.Equal(got, plain) {
			t.Errorf("密钥长度 %d 解密结果=%q, 期望 %q", size, got, plain)
		}
	}
}

// TestAesGcmEmptyPlaintext 空明文也能正确往返（GCM 允许空明文，只有 tag）
func TestAesGcmEmptyPlaintext(t *testing.T) {
	key := []byte("0123456789abcdef")
	cipher, err := AesGcmEncrypto(nil, key)
	if err != nil {
		t.Fatalf("空明文加密失败: %v", err)
	}
	got, err := AesGcmDecrypto([]byte(cipher), key)
	if err != nil {
		t.Fatalf("空明文解密失败: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("空明文解密结果=%q, 期望空", got)
	}
}
