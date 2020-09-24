package upush

// Bool boolean type
type Bool string

const (
	// BoolTrue true
	BoolTrue Bool = "true"
	// BoolFalse false
	BoolFalse Bool = "false"
)

//RetCode 接口调用返回状态
type RetCode string

const (
	// Success 接口调用成功返回
	Success RetCode = "SUCCESS"
	// Fail 接口调用失败返回
	Fail RetCode = "FAIL"

	httpGatewayURL  = "http://msg.umeng.com"
	httpsGatewayURL = "https://msgapi.umeng.com"
)

// Result umeng api result
type Result interface {
	IsSuccess() bool
	Error() error
}
