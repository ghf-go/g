package g

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	cr "crypto/rand"
	"encoding/hex"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	T_DATE     = "2006-01-02"
	T_TIME     = "15:04:05.999"
	T_DATETIME = "2006-01-02 15:04:05.999"
)

var unixEpochTime = time.Unix(0, 0)

type Map map[any]any

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// 生成随机字符串
func RandomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

// 时间是否为空
func IsTimeZero(t time.Time) bool {
	return t.IsZero() || t == unixEpochTime
}

// 格式化日期
func FormatDate(t ...time.Time) string {
	if len(t) > 0 {
		return t[0].Format(T_DATE)
	}
	return time.Now().Format(T_DATE)
}

// 格式化时间
func FormatTime(t ...time.Time) string {
	if len(t) > 0 {
		return t[0].Format(T_TIME)
	}
	return time.Now().Format(T_TIME)
}

// 格式化日期时间
func FormatDateTime(t ...time.Time) string {
	if len(t) > 0 {
		return t[0].Format(T_DATETIME)
	}
	return time.Now().Format(T_DATETIME)
}

// 获取http请求的IP
func GetRequestIP(r *http.Request) string {
	ret := r.Header.Get("ipv4")
	if ret != "" {
		return ret
	}
	ret = r.Header.Get("X-Forwarded-For")
	if ret != "" {
		rs := strings.Split(ret, ",")
		if rs[0] != "" {
			return rs[0]
		}
	}
	ret = r.Header.Get("XForwardedFor")
	if ret != "" {
		rs := strings.Split(ret, ",")
		if rs[0] != "" {
			return rs[0]
		}
	}
	ret = r.Header.Get("X-Real-Ip")
	if ret != "" {
		rs := strings.Split(ret, ",")
		if rs[0] != "" {
			return rs[0]
		}
	}
	ret = r.Header.Get("X-Real-IP")
	if ret != "" {
		rs := strings.Split(ret, ",")
		if rs[0] != "" {
			return rs[0]
		}
	}
	ret = r.RemoteAddr
	if ret != "" {
		return ret
	}
	return "unknow"
}

// Md5
func Md5(src string) string {
	m5 := md5.New()
	return string(m5.Sum([]byte(src)))
}

// Aes加密
func AesEncode(key, data string) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}
	in := []byte(data)
	leng := len(data)
	if leng%16 != 0 {
		leng = leng/16*16 + 16
		leng = leng - len(data)
		for i := 0; i < leng; i++ {
			in = append(in, 0)
		}
		leng = len(in)
	}

	cipherText := make([]byte, aes.BlockSize+leng)
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(cr.Reader, iv); err != nil {
		return ""
	}
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(cipherText[aes.BlockSize:], in)
	return hex.EncodeToString(cipherText)
}

// Aes解密
func AesDecode(key, data string) string {
	ciphertext, err := hex.DecodeString(data)
	if err != nil {
		return ""
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}
	if len(ciphertext) < aes.BlockSize {
		return ""
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(ciphertext, ciphertext)
	return string(ciphertext)
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

// Map相关
func (m Map) GetString(key any, def ...string) string {
	if r, ok := m[key]; ok {
		return r.(string)
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}
func (m Map) GetInt(key any, def ...int) int {
	if r, ok := m[key]; ok {
		return r.(int)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetInt8(key any, def ...int8) int8 {
	if r, ok := m[key]; ok {
		return r.(int8)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetInt16(key any, def ...int16) int16 {
	if r, ok := m[key]; ok {
		return r.(int16)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetInt32(key any, def ...int32) int32 {
	if r, ok := m[key]; ok {
		return r.(int32)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetInt64(key any, def ...int64) int64 {
	if r, ok := m[key]; ok {
		return r.(int64)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetUint(key any, def ...uint) uint {
	if r, ok := m[key]; ok {
		return r.(uint)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetUint8(key any, def ...uint8) uint8 {
	if r, ok := m[key]; ok {
		return r.(uint8)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetUint16(key any, def ...uint16) uint16 {
	if r, ok := m[key]; ok {
		return r.(uint16)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetUint32(key any, def ...uint32) uint32 {
	if r, ok := m[key]; ok {
		return r.(uint32)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetUint64(key any, def ...uint64) uint64 {
	if r, ok := m[key]; ok {
		return r.(uint64)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetBool(key any, def ...bool) bool {
	if r, ok := m[key]; ok {
		return r.(bool)
	}
	if len(def) > 0 {
		return def[0]
	}
	return false
}
func (m Map) GetFloat32(key any, def ...float32) float32 {
	if r, ok := m[key]; ok {
		return r.(float32)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) GetFloat64(key any, def ...float64) float64 {
	if r, ok := m[key]; ok {
		return r.(float64)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (m Map) Get(key any, def ...any) any {
	if r, ok := m[key]; ok {
		return r
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}
