package upush

import (
	"crypto/md5"
	"fmt"
)

// Sign https://developer.umeng.com/docs/67966/detail/149296#h1--i-9
func Sign(method, url string, postBody []byte, appMasterSecret string) string {
	str := method + url + string(postBody) + appMasterSecret
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
