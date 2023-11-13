package g

import (
	"encoding/json"
	"io"
	"net/http"
)

// 绑定Post的JSON内容
func bindRequestBodyJson(r *http.Request, obj any) error {
	body := r.Body
	// if e != nil {
	// 	return e
	// }
	defer body.Close()
	data, e := io.ReadAll(body)
	if e != nil {
		return e
	}
	return json.Unmarshal(data, obj)
}
