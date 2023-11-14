package g

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// POST JSON 数据
func PostJSON(url string, jsonData any, ret any) error {
	pdata, e := json.Marshal(jsonData)
	if e != nil {
		return e
	}
	dd := strings.NewReader(string(pdata))
	req, e := http.NewRequest(http.MethodPost, url, dd)
	if e != nil {
		return e
	}
	rep, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer rep.Body.Close()
	rd, e := io.ReadAll(rep.Body)
	if e != nil {
		return e
	}
	return json.Unmarshal(rd, ret)
}

// Get 获取Json数据
func GetJSON(url string, ret any) error {
	req, e := http.NewRequest(http.MethodGet, url, strings.NewReader(string("")))
	if e != nil {
		return e
	}
	rep, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer rep.Body.Close()
	rd, e := io.ReadAll(rep.Body)
	if e != nil {
		return e
	}
	return json.Unmarshal(rd, ret)
}

// 上传文件使用字节的方式
func PostFileByteJSON(url, postname, filename string, fdata []byte, ret any) error {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	w2, e := w.CreateFormFile(postname, filename)
	if e != nil {
		return e
	}
	_, e = w2.Write(fdata)
	if e != nil {
		return e
	}

	req, e := http.NewRequest(http.MethodPost, url, b)
	if e != nil {
		return e
	}
	rep, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer rep.Body.Close()
	rd, e := io.ReadAll(rep.Body)
	if e != nil {
		return e
	}
	return json.Unmarshal(rd, ret)
}
