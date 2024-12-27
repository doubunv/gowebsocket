// Package websocket 处理
package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"time"

	"github.com/link1st/gowebsocket/v2/common"
	"github.com/link1st/gowebsocket/v2/lib/cache"
	"github.com/link1st/gowebsocket/v2/models"

	"github.com/redis/go-redis/v9"
)

// PingController ping
func PingController(client *Client, seq string, message []byte) (code uint32, msg string, data interface{}) {
	code = common.OK
	fmt.Println("webSocket_request ping接口", client.Addr, seq, message)
	data = "pong"
	return
}

// LoginController 用户登录
func LoginController(client *Client, seq string, message []byte) (code uint32, msg string, data interface{}) {
	code = common.OK
	currentTime := uint64(time.Now().Unix())
	request := &models.Login{}
	if err := json.Unmarshal([]byte(message), request); err != nil {
		code = common.ServerError
		fmt.Println("用户登录 解析数据失败", seq, err)
		return
	}
	fmt.Println("webSocket_request 用户登录", seq, "ServiceToken", request.Token)

	if request.ImAccount == "" {
		code = common.ServerError
		fmt.Println("用户登录 非法的用户", seq, request.ImAccount)
		return
	}
	if !InAppIDs(request.AppID) {
		code = common.Unauthorized
		fmt.Println("用户登录 不支持的平台", seq, request.AppID)
		return
	}
	if client.IsLogin() {
		fmt.Println("用户登录 用户已经登录", client.AppID, client.UserID, seq)
		code = common.ServerError
		return
	}

	rData := map[string]interface{}{
		"uid":   cast.ToInt64(request.ImAccount),
		"appId": cast.ToString(request.AppID),
	}

	type RED struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ImAccount      string `json:"imAccount,omitempty"`
			ImAccountToken string `json:"imAccountToken,omitempty"`
			ImAppKey       string `json:"imAppKey,omitempty"`
			ImRoomAddr     string `json:"imRoomAddr"`
			ImRoomId       string `json:"imRoomId"`
			Uid            int    `json:"uid,omitempty"`
		} `json:"data"`
	}

	apiHttpHost := viper.GetString("api.httpUrl")
	postRequest, err := PostRequest(apiHttpHost+"/customerChat/getChatRoom", rData)
	if postRequest == nil || err != nil {
		fmt.Println("请求业务API失败", err.Error())
		return common.ServerError, "Request api fail", nil
	}

	rep := RED{}
	json.Unmarshal(postRequest, &rep)
	if rep.Code != 200 {
		return uint32(rep.Code), rep.Message, rep.Data
	}

	client.Login(request.AppID, rep.Data.ImAccount, currentTime)
	// 存储数据
	userOnline := models.UserLogin(serverIp, serverPort, request.AppID, rep.Data.ImAccount, client.Addr, currentTime)
	err = cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	if err != nil {
		code = common.ServerError
		fmt.Println("用户登录 SetUserOnlineInfo", seq, err)
		return
	}

	// 用户登录
	login := &login{
		AppID:  request.AppID,
		UserID: rep.Data.ImAccount,
		Client: client,
	}
	clientManager.Login <- login
	fmt.Println("用户登录 成功", seq, client.Addr, request.ImAccount)

	rep.Data.ImAccount = ""
	rep.Data.Uid = 0
	rep.Data.ImAccountToken = ""
	rep.Data.ImAppKey = ""
	marshal, err := json.Marshal(rep.Data)
	return uint32(rep.Code), rep.Message, string(marshal)
}

func AdminLoginController(client *Client, seq string, message []byte) (code uint32, msg string, data interface{}) {
	code = common.OK
	data = "{}"
	currentTime := uint64(time.Now().Unix())
	request := &models.Login{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ServerError
		fmt.Println("用户登录 解析数据失败", seq, err)
		return
	}

	// 本项目只是演示，所以直接过去客户端传入的用户ID
	if request.ImAccount == "" {
		code = common.ServerError
		fmt.Println("用户登录 非法的用户", seq, request.ImAccount)
		return
	}
	if !InAppIDs(request.AppID) {
		code = common.Unauthorized
		fmt.Println("用户登录 不支持的平台", seq, request.AppID)
		return
	}
	if client.IsLogin() {
		fmt.Println("用户登录 用户已经登录", client.AppID, client.UserID, seq)
		code = common.ServerError
		return
	}

	//todo 登录token健全
	fmt.Println("webSocket_request admin 用户登录", seq, "ServiceToken", request.Token)

	client.Login(request.AppID, request.ImAccount, currentTime)
	// 存储数据
	userOnline := models.UserLogin(serverIp, serverPort, request.AppID, request.ImAccount, client.Addr, currentTime)
	err := cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	if err != nil {
		code = common.ServerError
		fmt.Println("用户登录 SetUserOnlineInfo", seq, err)
		return
	}

	// 用户登录
	login := &login{
		AppID:  request.AppID,
		UserID: request.ImAccount,
		Client: client,
	}
	clientManager.Login <- login
	fmt.Println("admin 用户登录 成功", seq, client.Addr, request.ImAccount)
	return
}

// HeartbeatController 心跳接口
func HeartbeatController(client *Client, seq string, message []byte) (code uint32, msg string, data interface{}) {
	code = common.OK
	data = "{}"
	currentTime := uint64(time.Now().Unix())
	request := &models.HeartBeat{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ServerError
		fmt.Println("心跳接口 解析数据失败", seq, err)
		return
	}
	fmt.Println("webSocket_request 心跳接口", client.AppID, client.UserID)
	if !client.IsLogin() {
		fmt.Println("心跳接口 用户未登录", client.AppID, client.UserID, seq)
		code = common.NotLoggedIn
		return
	}
	userOnline, err := cache.GetUserOnlineInfo(client.GetKey())
	if err != nil {
		if errors.Is(err, redis.Nil) {
			code = common.NotLoggedIn
			fmt.Println("心跳接口 用户未登录", seq, client.AppID, client.UserID)
			return
		} else {
			code = common.ServerError
			fmt.Println("心跳接口 GetUserOnlineInfo", seq, client.AppID, client.UserID, err)
			return
		}
	}
	client.Heartbeat(currentTime)
	userOnline.Heartbeat(currentTime)
	err = cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	if err != nil {
		code = common.ServerError
		fmt.Println("心跳接口 SetUserOnlineInfo", seq, client.AppID, client.UserID, err)
		return
	}
	return
}

func SendMsgController(client *Client, seq string, message []byte) (code uint32, msg string, data interface{}) {
	code = common.OK
	data = "{}"
	request := &models.SendMsg{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ServerError
		fmt.Println("发送消息 解析数据失败", seq, err)
		return
	}

	rData := map[string]interface{}{
		"imAccount":   client.UserID,
		"appId":       cast.ToString(client.AppID),
		"roomId":      request.RoomID,
		"message":     request.Message,
		"messageType": request.MessageType,
	}

	apiHttpHost := viper.GetString("api.httpUrl")
	postRequest, err := PostRequest(apiHttpHost+"/customerChat/sendChatMsg", rData)
	if postRequest == nil || err != nil {
		fmt.Println("请求业务API失败", err.Error())
		return common.ServerError, "Request api fail", nil
	}

	rep := models.APIResponse{}
	json.Unmarshal(postRequest, &rep)

	marshal, err := json.Marshal(rep.Data)
	return uint32(rep.Code), rep.Message, string(marshal)
}

func GetMsgLogController(client *Client, seq string, message []byte) (code uint32, msg string, data interface{}) {
	code = common.OK
	request := &models.GetMsgLog{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ServerError
		fmt.Println("获取消息记录 解析数据失败", seq, err)
		return
	}

	rData := map[string]interface{}{
		"appId":  cast.ToString(client.AppID),
		"roomId": request.RoomID,
		"limit":  request.Limit,
		"maxId":  request.MaxID,
	}
	apiHttpHost := viper.GetString("api.httpUrl")
	postRequest, err := PostRequest(apiHttpHost+"/customerChat/getChatMsgByMaxId", rData)
	if postRequest == nil || err != nil {
		fmt.Println("请求业务API失败", err.Error())
		return common.ServerError, "Request api fail", nil
	}

	rep := models.APIResponse{}
	json.Unmarshal(postRequest, &rep)

	marshal, err := json.Marshal(rep.Data)
	return uint32(rep.Code), rep.Message, string(marshal)
}
