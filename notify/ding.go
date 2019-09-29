package notify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type dingMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

//发送钉钉通知
func Ding(url string, msg string) bool {
	jsonMsg, _ := json.Marshal(dingMsg{
		MsgType: "text",
		Text: struct {
			Content string `json:"content"`
		}{Content: msg},
	})
	payload := strings.NewReader(string(jsonMsg))
	//fmt.Println(payload)
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", "oapi.dingtalk.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	//fmt.Println(res)
	fmt.Println(string(body))
	//TODO 返回值判断,处理
	return true
}
