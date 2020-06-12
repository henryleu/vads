package app

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
	----------------------------------------
	字段				类型			说明
	cid				string		连接会话唯一标识
	rate			string		采样率(目前仅支持16k,后期可修改）
	uid				string		用户唯一标识，建议由字母大小写及数字组成，一定要保证一个UID代表一个用户的终身ID
	province	string		用户号码省份
	channel		string		渠道

*/

type Session struct {
}

type Chunk struct {
}
