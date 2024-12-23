// Package common 通用函数
package common

const (
	OK           = 200 // Success
	NotLoggedIn  = 401 // 未登录
	Unauthorized = 402 // 未授权
	ServerError  = 500 // 系统错误
)

// GetErrorMessage 根据错误码 获取错误信息
func GetErrorMessage(code uint32, message string) string {
	var codeMessage string
	codeMap := map[uint32]string{
		OK:           "success",
		NotLoggedIn:  "未登录",
		Unauthorized: "未授权",
		ServerError:  "系统错误",
	}

	if message == "" {
		if value, ok := codeMap[code]; ok {
			// 存在
			codeMessage = value
		} else {
			codeMessage = "未定义错误类型!"
		}
	} else {
		codeMessage = message
	}

	return codeMessage
}
