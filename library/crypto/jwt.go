package crypto

import (
	"framework/library/uerror"

	"github.com/golang-jwt/jwt/v5"
)

// 生成token
func JwtEncrypto(token jwt.Claims, secret string) (string, error) {
	// 1. 生成Token
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, token)

	// 2. 签名（使用环境变量获取密钥更安全）
	return tok.SignedString(secret)
}

// 解析token
func JwtDecrypto(str string, secret string, token jwt.Claims) (jwt.Claims, error) {
	tok, err := jwt.ParseWithClaims(str, token, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, uerror.New(-1, "JWT签名验证错误:%v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, uerror.New(-1, "Token is invalid")
	}
	return tok.Claims, nil
}
