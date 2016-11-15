package cache

import (
	"fmt"
	"os"
	"strings"
)

var Options *Option

type Option struct {
	Port        string
	Proxy       string
	IsConcatCss bool

	RedisHost string //redis地址
	RedisPort int    //redis 端口
	RedisDB   int    //redis database
	MaxIdle   int    //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态。
	MaxActive int    //最大的激活连接数，表示同时最多有N个连接
}

func (o *Option) CheckOption() {
	if o.Proxy == "" {
		fmt.Println("proxy can not be empty!")
		os.Exit(1)
	}
	if !strings.HasPrefix(o.Proxy, "http://") {
		fmt.Println("proxy must starts with http://")
		os.Exit(1)
	}
}
