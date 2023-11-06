package g

import (
	"crypto/aes"
	"encoding/hex"
	"time"
)

var unixEpochTime = time.Unix(0, 0)

// 时间是否为空
func IsTimeZero(t time.Time) bool {
	return t.IsZero() || t == unixEpochTime
}

// Aes加密
func AesEncode(key, data string) string {
	a, e := aes.NewCipher([]byte(key))
	if e != nil {
		return ""
	}
	out := make([]byte, len(data))
	a.Encrypt(out, []byte(data))
	return hex.EncodeToString(out)
}

// Aes解密
func AesDecode(key, data string) string {
	d, e := hex.DecodeString(data)
	if e != nil {
		return ""
	}
	a, e := aes.NewCipher([]byte(key))
	if e != nil {
		return ""
	}
	out := make([]byte, len(d))
	a.Decrypt(out, d)
	return string(out)
}
