package utils

import (
	"net"
	"os"
	"time"
)

//判断所给路径是否为文件夹
func IsDir(path string) bool {
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
func Tcp(url string, timeOut int) bool {
	//_, err := net.Dial("tcp", url)
	_, err := net.DialTimeout("tcp", url, time.Duration(timeOut)*time.Millisecond)
	if err != nil {
		//fmt.Println(err)
		return false
	} else {
		return true
	}
}
