package util

import (
	"encoding/json"
	"fmt"
	"github.com/kirinlabs/HttpRequest"
	"log"
)

const FlowSayDoUrl = "http://flowtest.aimango.net:5080/robot/say.do"
const FlowClosedUrl = "http://flowtest.aimango.net:5080/robot/closed.do"
const FlowInfoUrl = "http://114.116.103.13:8088/api/hly/flow/get"
const FlowTokenInfo = "Token 21c7d084b200a17c9641c83d4697fde9"

func FlowUtilSay(paramMap interface{}) (returnMap interface{}, err error) {
	/**
	 ** 开启流程会话接口
	 * 设置HTTP REST POST请求
	 * 1.使用http协议
	 * 2.调用/robot/say.do接口开启对话
	 * 4.设置必须请求参数：user_id，robot_id，input，token
	 */
	log.Print(paramMap.(map[string]interface{}))
	req := HttpRequest.NewRequest()
	req.SetHeaders(map[string]string{"Authorization": FlowTokenInfo})
	req.SetHeaders(map[string]string{"Content-Type": "application/json"})
	resp, err := req.Post(FlowSayDoUrl, paramMap)
	if err != nil {
		fmt.Printf(" post err:%s", err)
	}
	var dat map[string]interface{}
	body, err := resp.Body()
	_ = json.Unmarshal(body, &dat)
	//log.Printf("flow return :%s \n", dat)
	if _, ok := dat["successful"]; ok {
		returnMap = dat["info"].(map[string]interface{})
		return
	}

	//for k, v := range dat["info"].(map[string]interface{}) {
	//	fmt.Printf("%v=%v\n", k, v)
	//}
	return
}

func FlowUtilClosed(paramMap interface{}) {
	/**
	 ** 关闭流程会话接口
	 * 设置HTTP REST POST请求
	 * 1.使用http协议
	 * 2.调用/robot/closed.do接口开启对话
	 * 4.设置必须请求参数：user_id，robot_id，input，token
	 */
	req := HttpRequest.NewRequest()
	req.SetHeaders(map[string]string{"Authorization": FlowTokenInfo})
	req.SetHeaders(map[string]string{"Content-Type": "application/json"})
	resp, err := req.Post(FlowClosedUrl, paramMap)
	if err != nil {
		fmt.Printf("flow-closed post err:%s", err)
	}
	var dat map[string]interface{}
	body, err := resp.Body()
	_ = json.Unmarshal(body, &dat)
	log.Printf("flow return :%s \n", dat)
}

func FlowInfoByNumber(paramMap interface{}) (returnMap interface{}) {
	/**
	** 流程会话接口
	 * 设置HTTP REST POST请求
	 * 1.使用http协议
	 * 2.调用/robot/say.do接口开启对话
	 * 4.设置必须请求参数：user_id，robot_id，input，token
	*/
	//postData := map[string]interface{}{
	//	"mobile":  "18322693235",
	//}
	req := HttpRequest.NewRequest()
	req.SetHeaders(map[string]string{"Content-Type": "application/json"})
	resp, err := req.Post(FlowInfoUrl, paramMap)
	if err != nil {
		fmt.Printf("asr post err:%s", err)
	}
	var dat map[string]interface{}
	body, err := resp.Body()
	_ = json.Unmarshal(body, &dat)
	log.Printf("flow return :%s \n", dat)
	//returnMap := dat["info"].(map[string]interface{})
	//fmt.Print(returnMap)
	return dat
}
