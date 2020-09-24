package upush

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// SendType types of send
type SendType string

const (
	// SendTypeUnicast 单播
	SendTypeUnicast SendType = "unicast"
	// SendTypeListcast 列播，要求不超过500个device_token
	SendTypeListcast SendType = "listcast"
	// SendTypeFilecast 文件播，多个device_token可通过文件形式批量发送
	SendTypeFilecast SendType = "filecast"
	//SendTypeBroadcast 广播
	SendTypeBroadcast SendType = "broadcast"
	// SendTypeCustomizedcast 通过alias进行推送，包括以下两种case:
	//     - alias: 对单个或者多个alias进行推送
	//     - file_id: 将alias存放到文件后，根据file_id来推送
	SendTypeCustomizedcast SendType = "customizedcast"
)

// ReceiptType .
type ReceiptType string

const (
	// ReceiptTypeReceived 送达
	ReceiptTypeReceived ReceiptType = "1"
	// ReceiptTypeClicked 点击
	ReceiptTypeClicked ReceiptType = "2"
	//ReceiptTypeBoth 送达+点击
	ReceiptTypeBoth ReceiptType = "3"
)

// SendParam 消息发送调用参数
type SendParam struct {
	AppKey    string   `json:"appKey"`    // 必填，应用唯一标识
	Timestamp string   `json:"timestamp"` // 必填，时间戳，10位或者13位均可，时间戳有效期为10分钟
	Type      SendType `json:"type"`      // 必填，消息发送类型
	/*
		当type=unicast时, 必填, 表示指定的单个设备

		当type=listcast时, 必填, 要求不超过500个, 以英文逗号分隔
	*/
	DeviceTokens string `json:"device_tokens,omitempty"`
	/*
		当type=customizedcast时, 必填。

		alias的类型, alias_type可由开发者自定义, 开发者在SDK中
		调用setAlias(alias, alias_type)时所设置的alias_type
	*/
	AliasType string `json:"alias_type,omitempty"`
	/*
		当type=customizedcast时, 选填(此参数和file_id二选一)

		开发者填写自己的alias, 要求不超过500个alias, 多个alias以英文逗号间隔，
		在SDK中调用setAlias(alias, alias_type)时所设置的alias
	*/
	Alias string `json:"alias,omitempty"`
	/*
		当type=filecast时，必填，file内容为多条device_token，以回车符分割

		当type=customizedcast时，选填(此参数和alias二选一)

		file内容为多条alias，以回车符分隔。注意同一个文件内的alias所对应
		的alias_type必须和接口参数alias_type一致。

		使用文件播需要先调用文件上传接口获取file_id，参照"文件上传"
	*/
	FileID  string      `json:"file_id,omitempty"`
	Filter  string      `json:"filter,omitempty"` // 当type=groupcast时,必填,用户筛选条件,如用户标签、渠道等,参考附录G, filter的内容长度最大为3000B
	Payload interface{} `json:"payload"`          // 必填,JSON格式,具体消息内容(iOS最大为2012B,Android最大为1840B)
	Policy  SendPolicy  `json:"policy,omitempty"` // 可选,发送策略

	// 可选,正式/测试模式,默认为true
	// 测试模式只对“广播”、“组播”类消息生效,其他类型的消息任务（如“文件播”）不会走测试模式
	// 测试模式只会将消息发给测试设备,测试设备需要到web上添加
	ProductionMode Bool   `json:"production_mode,omitempty"`
	Description    string `json:"description,omitempty"` // 可选,发送消息描述,建议填写

	Mipush     Bool   `json:"mipush,omitempty"`      // 可选,默认为false,当为true时,表示MIUI、EMUI、Flyme系统设备离线转为系统下发
	MiActivity string `json:"mi_activity,omitempty"` // 可选,mipush值为true时生效,表示走系统通道时打开指定页面acitivity的完整包路径

	ReceiptURL  string      `json:"receipt_url,omitempty"`  // 开发者接受数据的地址,最大长度256字节
	ReceiptType ReceiptType `json:"receipt_type,omitempty"` // 回执数据类型,1:送达回执;2:点击回执;3:送达和点击回执,默认为3
}

// SendPolicy .
type SendPolicy struct {
	// 可选,定时发送时间,若不填写表示立即发送
	// 定时发送时间不能小于当前时间
	// 格式: "yyyy-MM-dd HH:mm:ss"
	// 注意,start_time只对任务生效
	StartTime string `json:"start_time,omitempty"`

	// 可选,消息过期时间,其值不可小于发送时间或者
	// start_time(如果填写了的话),
	// 如果不填写此参数,默认为3天后过期,格式同start_time
	ExpireTime string `json:"expire_time,omitempty"`

	// 可选,发送限速,每秒发送的最大条数,最小值1000
	// 开发者发送的消息如果有请求自己服务器的资源,可以考虑此参数
	MaxSendNum int64 `json:"max_send_num,omitempty"`

	// 可选,开发者对消息的唯一标识,服务器会根据这个标识避免重复发送
	// 有些情况下（例如网络异常）开发者可能会重复调用API导致
	// 消息多次下发到客户端,如果需要处理这种情况,可以考虑此参数
	// 注意,out_biz_no只对任务生效
	OutBizNo string `json:"out_biz_no,omitempty"`

	// 可选,多条带有相同apns_collapse_id的消息,iOS设备仅展示
	// 最新的一条,字段长度不得超过64bytes
	ApnsCollapseID string `json:"apns_collapse_id,omitempty"`
}

// DisplayType 消息类型
type DisplayType string

const (
	// DisplayTypeNotification 通知
	DisplayTypeNotification = "notification"
	//DisplayTypeMessage 消息
	DisplayTypeMessage = "message"
)

// AndroidPayload android notification payload
type AndroidPayload struct {
	DisplayType DisplayType `json:"display_type"` // 必填,消息类型: notification(通知)、message(消息)

	// 必填,消息体
	// 当display_type=message时,body的内容只需填写custom字段
	// 当display_type=notification时,body包含如下参数:
	Body AndroidPayloadBody `json:"body"`

	// 可选,JSON格式,用户自定义key-value,只对"通知"
	// (display_type=notification)生效
	// 可以配合通知到达后,打开App/URL/Activity使用
	Extra interface{} `json:"extra,omitempty"`
}

// AfterOpen AfterOpen in android payload body
type AfterOpen string

const (
	// AfterOpenGoApp go_app
	AfterOpenGoApp = "go_app"
	// AfterOpenGoURL go_url
	AfterOpenGoURL = "go_url"
	// AfterOpenGoActivity go_activity
	AfterOpenGoActivity = "go_activity"
	// AfterOpenGoCustom go_custom
	AfterOpenGoCustom = "go_custom"
)

// AndroidPayloadBody .
type AndroidPayloadBody struct {
	Ticker string `json:"ticker"` // 必填,通知栏提示文字
	Title  string `json:"title"`  // 必填,通知标题
	Text   string `json:"text"`   // 必填,通知文字描述

	// 可选,状态栏图标ID,R.drawable.[smallIcon],
	// 如果没有,默认使用应用图标
	// 图片要求为24*24dp的图标,或24*24px放在drawable-mdpi下
	// 注意四周各留1个dp的空白像素
	Icon string `json:"icon,omitempty"`

	// 可选,通知栏拉开后左侧图标ID,R.drawable.[largeIcon],
	// 图片要求为64*64dp的图标,
	// 可设计一张64*64px放在drawable-mdpi下,
	// 注意图片四周留空,不至于显示太拥挤
	LargeIcon string `json:"largeIcon,omitempty"`

	// 可选,通知栏大图标的URL链接,该字段的优先级大于largeIcon
	// 该字段要求以http或者https开头
	Img string `json:"img,omitempty"`

	// 可选,通知声音,R.raw.[sound]
	// 如果该字段为空,采用SDK默认的声音,即res/raw/下的
	// umeng_push_notification_default_sound声音文件,如果
	// SDK默认声音文件不存在,则使用系统默认Notification提示音
	Sound string `json:"sound,omitempty"`

	BuilderID   string `json:"builder_id,omitempty"`   // 可选,默认为0,用于标识该通知采用的样式,使用该参数时,开发者必须在SDK里面实现自定义通知栏样式
	PlayVibrate Bool   `json:"play_vibrate,omitempty"` // 可选,收到通知是否震动,默认为"true"
	PlayLights  Bool   `json:"play_lights,omitempty"`  // 可选,收到通知是否闪灯,默认为"true"
	PlaySound   Bool   `json:"play_sound,omitempty"`   // 可选,收到通知是否发出声音,默认为"true"

	// 点击"通知"的后续行为,默认为打开app
	// 可选,默认为"go_app",值可以为:
	//   "go_app": 打开应用
	//   "go_url": 跳转到URL
	//   "go_activity": 打开特定的activity
	//   "go_custom": 用户自定义内容
	AfterOpen AfterOpen `json:"after_open,omitempty"`

	// 当after_open=go_url时,必填
	// 通知栏点击后跳转的URL,要求以http或者https开头
	URL string `json:"url,omitempty"`

	// 当after_open=go_activity时,必填
	// 通知栏点击后打开的Activity
	Activity string `json:"activity,omitempty"`

	// 当display_type=message时, 必填
	// 当display_type=notification且
	// after_open=go_custom时,必填
	Custom interface{} `json:"custom,omitempty"`
}

// IOSPayload iOS notification payload
// "key1":"value1",       // 可选,用户自定义内容, "d","p"为友盟保留字段, key不可以是"d","p"
// "key2":"value2",
type IOSPayload map[string]interface{}

//IOSPayloadAps aps in ios payload
type IOSPayloadAps struct {
	// 当content-available=1时(静默推送),可选; 否则必填
	// 可为JSON类型和字符串类型
	Alert            IOSPayloadAlert `json:"alert,omitempty"`
	Badge            int64           `json:"badge,omitempty"`
	Sound            string          `json:"sound,omitempty"`
	ContentAvailable int64           `json:"content-available,omitempty"` // 可选,代表静默推送
	Category         string          `json:"category,omitempty"`          // 可选,注意: ios8才支持该字段
	MutableContent   int             `json:"mutable-content,omitempty"`   // 添加图片需要设置为1
	Image            string          `json:"image,omitempty"`
}

// IOSPayloadAlert alert in aps
type IOSPayloadAlert struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Body     string `json:"body,omitempty"`
}

// SendResult 消息发送调用返回
type SendResult struct {
	Ret  RetCode `json:"ret"`
	Data struct {
		/*
		 当"ret"为"SUCCESS"时，包含如下参数:

		 单播类消息(type为unicast、listcast、customizedcast且不带file_id)返回
		*/
		MsgID string `json:"msg_id,omitempty"`
		// 任务类消息(type为broadcast、groupcast、filecast、customizedcast且file_id不为空)返回
		TaskID string `json:"task_id,omitempty"`

		// 错误码, 当"ret"为"FAIL"时返回
		ErrorCode string `json:"error_code,omitempty"`
		// 错误消息, 当"ret"为"FAIL"时返回
		ErrorMsg string `json:"error_msg,omitempty"`
	}
}

// IsSuccess check if success
func (r *SendResult) IsSuccess() bool {
	return r.Ret == Success
}

func (r *SendResult) Error() error {
	if !r.IsSuccess() {
		return fmt.Errorf("%s, code=%s", r.Data.ErrorMsg, r.Data.ErrorCode)
	}
	return nil
}

// Send 消息发送
// 功能说明
// 开发者通过此接口，可向指定用户(单播)、所有用户(广播)或满足特定条件的用户群(组播)，发送通知或消息。此外，该接口还支持开发者使用自有的账号系统(alias) 来发送消息给指定的账号或者账号群。
// 注意：iOS推送的相关协议，请严格按照APNs的协议来填写，友盟完全遵循APNs协议。
// https://developer.umeng.com/docs/67966/detail/68343#h1-u6D88u606Fu53D1u90014
func (c *RawClient) Send(ctx context.Context, p *SendParam) (*SendResult, error) {
	p.AppKey = c.appKey
	p.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	r := &SendResult{}
	err := c.sendRequest(
		ctx,
		http.MethodPost,
		c.gatewayURL+"/api/send",
		p,
		r,
	)
	return r, err
}

// SendFilecast filecast
func (c *Client) SendFilecast(ctx context.Context, fileID string, payload interface{}, params ...*SendParam) (*SendResult, error) {
	var p *SendParam
	if len(params) > 0 {
		p = params[0]
	} else {
		p = &SendParam{}
	}
	p.FileID = fileID
	p.Payload = payload
	p.Type = SendTypeFilecast
	return c.rc.Send(ctx, p)
}
