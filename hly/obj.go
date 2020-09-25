package hly

import (
	"bytes"
	"encoding/base64"
)

/*
	基本信息
	Port:6000
	数据格式：2字节（网络字节顺序，表示json长度）+json数据串
	连接方式：websocket
	备注：语音格式（单声道，采样率16K,位深16bit）
*/

/*
	客户端请求
	第一次发送http请求内容
	格式：
	{
		"cid": "01010101010",
		"rate": "16000",
		"business": {
			"uid": "1331114444 abcd",
			"province": "beijing",
			"channel": "03"
		}
	}

	参数说明
	字段				类型			说明
	----------------------------------------
	cid				string		连接会话唯一标识
	rate			string		采样率(目前仅支持16k,后期可修改）
	uid				string		用户唯一标识，建议由字母大小写及数字组成，一定要保证一个UID代表一个用户的终身ID
	province	string		用户号码省份
	channel		string		渠道

*/

// Request is the session request
type Request struct {
	CID      string    `json:"cid"`
	Rate     string    `json:"rate"`
	Business *Business `json:"business"`
}

// Message creates a request Message
func (o *Request) Message() *Message {
	return &Message{
		Type:    RequestType,
		Payload: o,
	}
}

// NewErrorResponse creates and returns a new response with error result
func (o *Request) NewErrorResponse(detail string) *Response {
	return &Response{
		CID: o.CID,
		Result: &Result{
			Code:   0, //处理结果 1：成功; 0：失败;
			Detail: detail,
			Return: &Return{
				Control: &Control{
					// 挂机处置方式：0：继续对话；1：结束对话；
					Status: 1,
				},
				Recognition: &Recognition{},
			},
		},
	}
}

// NewResponse creates and returns a new response with error result
func (o *Request) NewSuccessResponse(status int, recognition *Recognition) *Response {
	return &Response{
		CID: o.CID,
		Result: &Result{
			Code:   1, //处理结果 1：成功; 0：失败;
			Detail: "success",
			Return: &Return{
				Control: &Control{
					// 挂机处置方式：0：继续对话；1：结束对话；
					Status: status,
				},
				Recognition: recognition,
			},
		},
	}
}

// Business is the biz info in the inbound session message
type Business struct {
	UID      string `json:"uid"`
	Province string `json:"province"`
	Channel  string `json:"channel"`
	Called   string `json:"called"`
}

/*
	格式：
	{
		"cid": "01010101010",
		"chunk": 1,
		"audio": "xxxxxxx"
	}

	参数说明
	字段				类型			说明
	----------------------------------------
	chunk				int			该分片在所有分片中的编号
	audio				string	音频内容（每次时长100ms,采用base64编码）

	服务器端返回结果：
	备注：要求client第二次收到内容后断开

*/

// Chunk is the chunk data of the inbound voice in the session
type Chunk struct {
	CID   string `json:"cid"`
	NO    int    `json:"chunk"`
	Audio string `json:"audio"`
	Data  []byte `json:"-"`
}

// EncodeAudio encodes audio bytes to base64 string
func (o *Chunk) EncodeAudio() {
	o.Audio = base64.StdEncoding.EncodeToString(o.Data)
}

// DecodeAudio decodes audio base64 string back to bytes
func (o *Chunk) DecodeAudio() error {
	data, err := base64.StdEncoding.DecodeString(o.Audio)
	if err != nil {
		return err
	}
	o.Data = data
	return nil
}

// ValidateCoding compares data with data->encoded->decoded
func (o *Chunk) ValidateCoding() bool {
	encoded := base64.StdEncoding.EncodeToString(o.Data)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}
	return bytes.Compare(o.Data, decoded) == 0
}

// Message creates a chunk Message
func (o *Chunk) Message() *Message {
	return &Message{
		Type:    ChunkType,
		Payload: o,
	}
}

/*
	客户端请求
	第一次发送http请求内容
	格式：
	{
		"cid": "121321212",
		"result": {
			"code": 0,
			"detail": "success",
			“ret”:{
				"control": {
					"status": 1
				},
				"recog": {
					"audio_text": "什么是4g",
					"answer_text": "4g is xxx",
					"audio_num ": "001"
				}
			}
		}
	}

	字段说明
	字段					类型			说明
	----------------------------------------
	cid					string		连接会话唯一标识
	code				int				处理结果 1：成功; 0：失败;
	status			int				挂机处置方式：0：继续对话；1：结束对话；
	detail			string		失败原因描述
	audio_text	string		语音识别文本
	answer_text	string		应答结果文本
	audio_num		string		应答语音编号
*/

// Response is the session response
type Response struct {
	CID    string  `json:"cid"`
	Result *Result `json:"result"`
}

// Message creates a response Message
func (o *Response) Message() *Message {
	return &Message{
		Type:    ResponseType,
		Payload: o,
	}
}

// Result descries the result infos of the response
type Result struct {
	Code   int     `json:"code"`
	Detail string  `json:"detail"`
	Return *Return `json:"ret"`
}

// Return descries the return info of the response
type Return struct {
	Control     *Control     `json:"control"`
	Recognition *Recognition `json:"recog"`
}

// Control descries the control info of the response
type Control struct {
	Status int `json:"status"`
}

// Recognition descries the recognition info of the response
type Recognition struct {
	AudioText  string `json:"audio_text"`
	AnswerText string `json:"answer_text"`
	AudioNum   string `json:"audio_num"`
}

/**
	flow 请求数据
	{
		"user_id":"ihu0rHkuZV",
		"robot_id":"690340eeadf7638d8e7a916a6160778d",
		"input":"1231",
		"parameter":{
			"option-dict":{
				"fa5101658de401d18e2f157de6371b12_-44":"正在开会",
				"d2f299baa005c3dae83e3b93736bb6ed_-22":"放在前台",
				"0961779b1a537a1fba781663a408ef6b_-22":"代收点"
			},
			"entity-parameter":{
			}
		},
		"token":"f86fcc23404392239bb985a75cf38f57"
	}
字段说明
	字段					类型			说明
	----------------------------------------
	user_id				string		连接会话唯一标识
	robot_id			string		机器人id
	input				string		asr识别结果
	token				string		机器人token
	parameter			json		机器人扩展信息
*/
type FlowParam struct {
	UserId    string     `json:"user_id"`
	RobotId   string     `json:"robot_id"`
	Input     string     `json:"input"`
	Parameter *Parameter `json:"parameter"`
	Token     string     `json:"token"`
}

type Parameter struct {
}

type Client struct {
	Mobile string `json:"mobile"`
}
