package upush

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//UploadParam 文件上传调用参数
type UploadParam struct {
	AppKey    string `json:"appkey"`    // 必填, 应用唯一标识
	Timestamp string `json:"timestamp"` // 必填, 时间戳,10位或者13位均可,时间戳有效期为10分钟
	Content   string `json:"content"`   // 必填, 文件内容, 多个device_token/alias请用回车符"\n"分隔
}

//UploadResult 文件上传调用返回
type UploadResult struct {
	Ret  RetCode `json:"ret"`
	Data struct {
		// 任务类消息(type为broadcast、groupcast、filecast、customizedcast且file_id不为空)返回
		FileID string `json:"file_id,omitempty"`

		// 错误码, 当"ret"为"FAIL"时返回
		ErrorCode string `json:"error_code,omitempty"`
		// 错误消息, 当"ret"为"FAIL"时返回
		ErrorMsg string `json:"error_msg,omitempty"`
	}
}

// IsSuccess check if success
func (r *UploadResult) IsSuccess() bool {
	return r.Ret == Success
}

func (r *UploadResult) Error() error {
	if !r.IsSuccess() {
		return fmt.Errorf("%s, code=%s", r.Data.ErrorMsg, r.Data.ErrorCode)
	}
	return nil
}

/*
Upload 文件上传
文件上传接口支持两种应用场景：

发送类型为”filecast”的时候, 开发者批量上传device_token;
发送类型为”customizedcast”时, 开发者批量上传alias。
上传文件后获取file_id, 从而可以实现通过文件id来进行消息批量推送的目的。

文件自创建起，服务器会保存两个月。开发者可以在有效期内重复使用该file-id进行消息发送。

注意：上传的文件不超过10M。
*/
func (c *RawClient) Upload(ctx context.Context, p *UploadParam) (*UploadResult, error) {
	p.AppKey = c.appKey
	p.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	r := &UploadResult{}
	err := c.sendRequest(
		ctx,
		http.MethodPost,
		c.gatewayURL+"/upload",
		p,
		r,
	)
	return r, err
}

// Upload 文件上传
//
// tokens 多个device_token/alias
func (c *Client) Upload(ctx context.Context, tokens []string) (*UploadResult, error) {
	p := UploadParam{
		Content: strings.Join(tokens, "\n"),
	}
	return c.rc.Upload(ctx, &p)
}
