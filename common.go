package g

import "time"

var unixEpochTime = time.Unix(0, 0)

// 时间是否为空
func IsTimeZero(t time.Time) bool {
	return t.IsZero() || t == unixEpochTime
}
