package g

import (
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

// 生产获取上传凭证接口
func QiniuTokenAction(c *GContext) {
	token := c.engine.conf.StoreConf.GetQiniuToken()
	c.WebJsonSuccess(Map{
		"token":       token,
		"path":        time.Now().Format("/2006/01/02/"),
		"upload_host": c.engine.conf.StoreConf.ZoneHost,
		"cdn":         c.engine.conf.StoreConf.CdnDomain,
	})
}

// 获取七牛云上传token
func (c *_storeConf) GetQiniuToken(bucket ...string) string {
	if c.qini == nil {
		c.qini = auth.New(c.AccessKey, c.SecretKey)
	}
	bk := c.Bucket
	if len(bucket) > 0 {
		bk = bucket[0]
	}
	putPolicy := storage.PutPolicy{
		Scope:      bk,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)"}`,
	}
	putPolicy.Expires = 7200
	return putPolicy.UploadToken(c.qini)
}
