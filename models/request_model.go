// Package models 数据模型
package models

// Request 通用请求数据格式
type Request struct {
	Seq  string `json:"seq"`            // 消息的唯一ID
	Cmd  string `json:"cmd"`            // 请求命令字
	Data string `json:"data,omitempty"` // 数据 json
}

// Login 登录请求数据
type Login struct {
	Token     string `json:"token"` // 验证用户是否登录
	AppID     uint32 `json:"appID"`
	ImAccount string `json:"imAccount"`
}

// HeartBeat 心跳请求数据
type HeartBeat struct {
}

// SendMsg 发送消息
type SendMsg struct {
	RoomID      string `json:"roomID"`
	MessageType string `json:"messageType"`
	Message     string `json:"message"`
}

// GetMsgLog 获取消息记录
type GetMsgLog struct {
	MaxID  int64  `json:"maxID"`
	Limit  int    `json:"limit,omitempty"`
	RoomID string `json:"roomID"`
}
