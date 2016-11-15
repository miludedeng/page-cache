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
