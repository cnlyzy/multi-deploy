package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

//TODO 文本日志 or Redis日志
var log = "Multi Deploy Runtime Log\n"

//运行时传路径参数 -p=/projectPath
var p string
var exclude bool

func init() {
	var v bool
	flag.StringVar(&p, "p", "", "set absolute project path (not have '/')")
	flag.BoolVar(&exclude, "exclude", false, "read exclude patterns from FILE (if you want to use please make sure project dir have 'multi_deploy_exclude.txt')")
	flag.BoolVar(&v, "v", false, "show version and exit")
	flag.Parse()
	if v {
		fmt.Println("version:0.1.2")
		os.Exit(0)
	}
	if !isDir(p) {
		flag.Usage()
		os.Exit(0)
	}
	//fmt.Println(flag.Args())
}

//定义配置文件解析后的结构
type Config struct {
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		Db       int    `json:"db"`
		Key      string `json:"key"`
		KeyTTL   int    `json:"keyTTL"`
	} `json:"redis"`
	Sync struct {
		UserName       string `json:"userName"`
		DestPathPrefix string `json:"destPathPrefix"`
		ExcludeFrom    string `json:"excludeFrom"`
	} `json:"sync"`
	Net struct {
		Port    string `json:"port"`
		Timeout int    `json:"timeout"`
	} `json:"net"`
	Log struct {
		Enable   bool   `json:"enable"`
		Type     string `json:"type"`
		KeyName  string `json:"keyName"`
		FilePath string `json:"filePath"`
	} `json:"log"`
	Notify struct {
		Type     string `json:"type"`
		RobotURL string `json:"robotUrl"`
	} `json:"notify"`
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}
}

func main() {
	conf := Config{}
	JsonParse := NewJsonStruct()
	//下面使用的是相对路径，config.json文件
	JsonParse.Load("./conf/config.json", &conf)
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password, // no password set
		DB:       conf.Redis.Db,       // use default DB
	})

	hostName, _ := os.Hostname()
	hostOS := runtime.GOOS
	pwd, _ := os.Getwd()
	fmt.Println("主机名称:", hostName)
	fmt.Println("操作系统:", hostOS)
	fmt.Println("运行目录:", pwd)
	if hostOS == "windows" {
		fmt.Println("Oops 本工具暂不支持Windows操作系统 请在Mac/Linux下使用 ~")
		os.Exit(0)
	}
	sIndex := strings.LastIndex(p, "/")
	eIndex := len(p)
	projectName := p[sIndex+1 : eIndex]
	fmt.Println("项目名称:", projectName)
	hostsNum, err := client.LLen(conf.Redis.Key).Result()
	if err != nil {
		fmt.Println("连接Redis失败:", err)
		return
	}

	fmt.Println("集群机器数量:", hostsNum)
	hostsAll, _ := client.LRange(conf.Redis.Key, 0, hostsNum-1).Result()
	fmt.Println("集群机器列表:", hostsAll)

	log += "主机名:" + hostName + "\n"
	log += "运行目录:" + pwd + "\n"

	wg.Add(int(hostsNum)) //声明计数 数量为集群机器数量
	for _, ip := range hostsAll {
		go doSync(ip, p, projectName, conf)
	}
	wg.Wait() //等待协程执行结束
	now := time.Now()
	client.HSet(conf.Log.KeyName+":"+projectName+":"+now.Format("20060102"), now.Format("1504.05"), log)
	ding(conf.Notify.RobotURL, log)
	fmt.Println("程序运行结束 Bye bye ~")
}

//开始同步,检查主机是否连通
func doSync(ip string, pwd string, projectName string, conf Config) {
	netStatus := tcp(ip+":"+conf.Net.Port, conf.Net.Timeout)
	if netStatus {
		//fmt.Println("网络连接成功:", ip)
		destPath := conf.Sync.DestPathPrefix + projectName
		syncCmd := ""
		if exclude {
			//目录排除 主要是日志等目录 每个项目不一致 需用户自行编写multi_deploy_exclude.txt文件 放在项目根目录
			syncCmd = "rsync -azpP --exclude-from=" + destPath + "/" + conf.Sync.ExcludeFrom + " -e 'ssh  -o PubkeyAuthentication=yes   -o stricthostkeychecking=no' " + pwd + "/ root@" + ip + ":" + destPath
		} else {
			syncCmd = "rsync -azpP -e 'ssh  -o PubkeyAuthentication=yes   -o stricthostkeychecking=no' " + pwd + "/ root@" + ip + ":" + destPath
		}
		//TODO 暂不支持Windows (目录路径处理)
		cmd := exec.Command("/bin/bash", "-c", syncCmd)
		syncResp, err := cmd.Output()
		if err != nil {
			fmt.Println(ip+"同步出错:", err)
			log += "同步出错:" + err.Error() + "\n"
		} else {
			fmt.Println(ip+"同步详情:", string(syncResp))
			log += ip + "同步详情:" + string(syncResp) + "\n"
		}
	} else {
		fmt.Println("网络连接失败:", ip)
		log += "网络连接失败:" + ip + "\n"
	}
	wg.Done() //计数减一
}

// 判断所给路径是否为文件夹
func isDir(path string) bool {
	if len(path) == 0 {
		return false
	}
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

//定义检测 tcp 服务的脚本，用到 net 包 主机连通性测试
func tcp(url string, timeOut int) bool {
	//_, err := net.Dial("tcp", url)
	_, err := net.DialTimeout("tcp", url, time.Duration(timeOut)*time.Millisecond)
	if err != nil {
		//fmt.Println(err)
		return false
	} else {
		return true
	}
}

type dingMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

//发送钉钉通知
func ding(url string, msg string) bool {
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
