// Package models 数据模型
package models

import "github.com/link1st/gowebsocket/v2/common"

const (
	// MessageTypeText 文本类型消息
	MessageTypeText = "text"
	// MessageCmdEnter 用户进入类型消息
	MessageCmdEnter = "enter"
	// MessageCmdExit 用户退出类型消息
	MessageCmdExit = "exit"
)

// Message 消息的定义
type Message struct {
	Message       string `json:"message"`       //消息内容
	MessageType   string `json:"messageType"`   //消息类型
	RoomID        string `json:"roomID"`        //房间号
	FormImAccount string `json:"formImAccount"` // 发送者
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//
//// NewMsg 创建新的消息
//func NewMsg(from string, Msg string) (message *Message) {
//	message = &Message{
//		From: from,
//		Msg:  Msg,
//	}
//	return
//}

func getTextMsgData(cmd, uuID, msgID, message string) string {
	head := NewResponseHead(msgID, cmd, common.OK, common.GetErrorMessage(common.OK, "success"), message)
	return head.String()
}

// GetMsgData 文本消息
func GetMsgData(uuID, msgID, cmd, messageType, message string) string {
	return getTextMsgData(cmd, uuID, msgID, message)
}

// GetTextMsgData 文本消息
func GetTextMsgData(uuID, msgID, cmd, message string) string {
	return getTextMsgData(cmd, uuID, msgID, message)
}

// GetTextMsgDataEnter 用户进入消息
func GetTextMsgDataEnter(uuID, msgID, message string) string {
	return getTextMsgData("enter", uuID, msgID, message)
}

// GetTextMsgDataExit 用户退出消息
func GetTextMsgDataExit(uuID, msgID, message string) string {
	return getTextMsgData("exit", uuID, msgID, message)
}
