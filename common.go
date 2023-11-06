package g

import (
	"crypto/aes"
	"encoding/hex"
	"strconv"
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
func String2Int64(src string) int64 {
	r, e := strconv.ParseInt(src, 10, 64)
	if e != nil {
		return 0
	}
	return r
}
func String2Int32(src string) int32 {
	r, e := strconv.ParseInt(src, 10, 32)
	if e != nil {
		return 0
	}
	return int32(r)
}
func String2Int16(src string) int16 {
	r, e := strconv.ParseInt(src, 10, 16)
	if e != nil {
		return 0
	}
	return int16(r)
}
func String2Int8(src string) int8 {
	r, e := strconv.ParseInt(src, 10, 8)
	if e != nil {
		return 0
	}
	return int8(r)
}
func String2Int(src string) int {
	r, e := strconv.Atoi(src)
	if e == nil {
		return r
	}
	return 0
}
func String2Uint64(src string) uint64 {
	r, e := strconv.ParseUint(src, 10, 64)
	if e != nil {
		return 0
	}
	return r
}
func String2Uint32(src string) uint32 {
	r, e := strconv.ParseUint(src, 10, 32)
	if e != nil {
		return 0
	}
	return uint32(r)
}
func String2Uint16(src string) uint16 {
	r, e := strconv.ParseUint(src, 10, 16)
	if e != nil {
		return 0
	}
	return uint16(r)
}
func String2Uint8(src string) uint8 {
	r, e := strconv.ParseUint(src, 10, 8)
	if e != nil {
		return 0
	}
	return uint8(r)
}
func String2Uint(src string) uint {
	r, e := strconv.ParseUint(src, 10, 32)
	if e != nil {
		return 0
	}
	return uint(r)
}
func String2Float64(src string) float64 {
	r, e := strconv.ParseFloat(src, 64)
	if e != nil {
		return 0
	}
	return r
}
func String2Float32(src string) float32 {
	r, e := strconv.ParseFloat(src, 32)
	if e != nil {
		return 0
	}
	return float32(r)
}
func String2Bool(src string) bool {
	r, e := strconv.ParseBool(src)
	if e != nil {
		return false
	}
	return r
}
