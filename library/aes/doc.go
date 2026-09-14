// Package aes 提供项目内使用的对称加解密工具。
//
// ⚠️ 各接口的强度与适用场景（选错等于没有保护，务必按需选择）：
//
//   - AesEncrypto / AesDecrypto：AES-ECB + PKCS7，**无 IV、无签名**，**仅作旧协议兼容保留**。
//     ECB 是确定性加密（相同明文恒产生相同密文，可做块级模式分析），且不带 MAC：
//     能改包的中间人可以重排、复制、拼接 16 字节块而不被发现——它只提供"看不懂"，
//     不提供完整性，**不能当作认证加密使用**。新代码不要再使用这一对。
//
//   - AesGcmEncrypto / AesGcmDecrypto：AES-GCM（AEAD，随机 nonce 前置、自带 tag 校验），
//     同时提供机密性与完整性，新协议请使用这一对。
//
//   - DecryptConfig / DecryptConfigFields：config.yaml 中密钥/口令字段的静态解密，
//     密文由 scripts/secret.sh 生成（当前仍是 ECB 格式，属旧协议兼容）。
//
// ⚠️ 迁移提示：把 AesEncrypto 换成 AesGcmEncrypto 属于**协议变更**，必须与对端
// （游戏客户端 / 浏览器前端）同步升级，只改单侧会导致对端全部解不开。
package aes
