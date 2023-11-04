package g

import (
	"bytes"
	"net/http"
)

type GResponseWrite struct {
	header     http.Header
	statusCode int
	data       *bytes.Buffer
}

func (w *GResponseWrite) Header() http.Header {
	return w.header
}
func (w *GResponseWrite) Write(data []byte) (int, error) {
	return w.data.Write(data)
}
func (w *GResponseWrite) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
func (w *GResponseWrite) ReadAndClear() []byte {
	r := w.data.Bytes()
	w.data.Reset()
	return r
}
