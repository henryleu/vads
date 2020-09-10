package util

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/kirinlabs/HttpRequest"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func AsrClient(filePath string) (content string) {
	/**
	**
	 * 设置HTTP REST POST请求
	 * 1.使用http协议
	 * 2.语音识别服务域名：nls-gateway.cn-shanghai.aliyuncs.com
	 * 3.语音识别接口请求路径：/stream/v1/asr
	 * 4.设置必须请求参数：appkey、format、sample_rate，
	 * 5.设置可选请求参数：enable_punctuation_prediction、enable_inverse_text_normalization、enable_voice_detection
	 * param: filePath 经vad 检测后额声音文件路径
	*/

	/**
	 * 读取声音文件,filePath为vad检测录音
	 */
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("read file fail", err)
		return ""
	}
	defer f.Close()
	fin, err := ioutil.ReadAll(f)
	finLen := len(fin)
	if err != nil {
		fmt.Println("read to fd fail", err)
		return ""
	}
	if len(fin) <= 0 {
		return ""
	}
	req := HttpRequest.NewRequest()
	/**
	拼接asr识别地址
	*/
	url := "http://nls-gateway.cn-shanghai.aliyuncs.com/stream/v1/asr"
	appkey := "?appkey=zHcTWH9zRW1RcKFr"
	format := "&format=wav"
	sampleRate := "&sample_rate=8000"
	httpUrl := url + appkey + format + sampleRate + ""
	//log.Print(httpUrl)
	/**
	 * 获取识别token
	 */
	client, err := sdk.NewClientWithAccessKey("cn-shanghai", "LTAIVbXZCLIwAnMg", "AXdRQ8jWfb1fGp2kZmZ6EKcaTfDPqc")
	if err != nil {
		panic(err)
	}
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Domain = "nls-meta.cn-shanghai.aliyuncs.com"
	request.ApiName = "CreateToken"
	request.Version = "2019-02-28"
	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		panic(err)
	}
	//fmt.Println(response.GetHttpStatus())
	//fmt.Println(response.GetHttpContentString())
	//json str 转map
	var dat map[string]interface{}
	_ = json.Unmarshal([]byte(response.GetHttpContentString()), &dat)
	mapTmp := dat["Token"].(map[string]interface{})
	accessToken := mapTmp["Id"].(string)
	//fmt.Printf("accessToken :%s \n", accessToken)
	/**
	 * 设置HTTP 头部字段
	 * 1.鉴权参数
	 * 2.Content-Type：application/octet-stream
	 */
	req.SetHeaders(map[string]string{"X-NLS-Token": accessToken})
	req.SetHeaders(map[string]string{"Content-Type": "application/octet-stream"})
	req.SetHeaders(map[string]string{"Content-Length": strconv.Itoa(finLen)})
	// POST 调用方法
	resp, err := req.Post(httpUrl, string(fin))
	if err != nil {
		fmt.Printf("asr post err:%s", err)
	}
	body, err := resp.Body()
	//定义asr识别数据结构体
	type asrResult struct {
		TaskId  string `json:"task_id"`
		Result  string
		Status  int
		Message string
	}
	res := &asrResult{}
	// 解析字符串为Json
	_ = json.Unmarshal([]byte(body), &res)
	log.Printf("asr result:%s", res.Result)
	if res.Message == "SUCCESS" {
		content = res.Result
	}
	return
}

/**
*	深思维语音识别调用方法如下：
*	url: ip:port/aicyber/asr/gpu/stream/gpu/recognise/
*	headers = {
*		'Content-Type': 'audio/wav;rate=%s' % framerate,
*		'Content-length': str(len(voice_data))
*	}
*   返回结果：
*	{
*	  "status": "ok",
*      "data": [{
*	    	"text": "你好我是私人助理小贼请问您打电话过来有什么事情吗。"
*	  }]
*    }
*
* */

func AsrByAicyber(filePath string) (content string) {
	/**
	 * TODO Step 001
	 * 读取声音文件,filePath为vad检测录音
	 */
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("read file fail", err)
		return ""
	}
	defer f.Close()
	fin, err := ioutil.ReadAll(f)
	finLen := len(fin)
	if err != nil {
		fmt.Println("read to fd fail", err)
		return ""
	}
	if len(fin) <= 0 {
		return ""
	}
	req := HttpRequest.NewRequest()
	/**
	 * TODO Step 002
	 * 拼接asr识别地址
	 */
	httpUrl := "http://192.168.2.200:8080/aicyber/asr/gpu/stream/gpu/recognise/"
	log.Print(httpUrl)
	/**
	 * 设置HTTP 头部字段
	 * 1.鉴权参数
	 * 2.Content-Type：application/octet-stream
	 */
	req.SetHeaders(map[string]string{"Content-Type": "application/octet-stream"})
	req.SetHeaders(map[string]string{"Content-Length": strconv.Itoa(finLen)})
	// POST 调用方法
	resp, err := req.Post(httpUrl, string(fin))
	if err != nil {
		fmt.Printf("asr post err:%s", err)
	}
	body, err := resp.Body()
	//定义asr识别数据结构体
	type asrData struct {
		Text string `json:"text"`
	}
	type asrResult struct {
		Status string    `json:"status"`
		Data   []asrData `json:"data"`
	}

	//var dat map[string]interface{}
	res := &asrResult{}
	if err := json.Unmarshal([]byte(body), &res); err == nil {
		log.Print(res)
		log.Print(res.Status)
		log.Print(res.Data[0].Text)
		if res.Status == "ok" {
			content = res.Data[0].Text
		}
	}
	return
}
