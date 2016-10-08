package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"page-cache/cache"
)

var (
	port        = flag.String("port", "3000", "set listen port")
	proxy       = flag.String("proxy", "", "proxy url")
	isConcatCss = flag.Bool("concatcss", false, "concat css true or false, default is false")
	redisHost   = flag.String("redishost", "127.0.0.1", "set redis host")
	redisPort   = flag.Int("redisprot", 6379, "set redis port")
	redisDB     = flag.Int("redisdb", 0, "set redis database, default 0")
	maxIdle     = flag.Int("maxidle", 1, "redis max idle time, unit is second")
	maxActive   = flag.Int("maxactive", 1000, "redis max connections")
)

func main() {
	flag.Parse()
	cache.Options = &cache.Option{
		Port:        *port,
		Proxy:       *proxy,
		IsConcatCss: *isConcatCss,
		RedisHost:   *redisHost,
		RedisPort:   *redisPort,
		RedisDB:     *redisDB,
		MaxIdle:     *maxIdle,
		MaxActive:   *maxActive,
	}
	cache.Options.CheckOption()
	cache.InitRedisPool()

	http.HandleFunc("/", handler)
	log.Println("Start serving on port ", cache.Options.Port)
	log.Fatal(http.ListenAndServe(":"+cache.Options.Port, nil))
	os.Exit(0)
}

func handler(w http.ResponseWriter, r *http.Request) {

	// 请求ID
	var id = cache.EncodeUrl(r.URL.String())
	// 判断缓存是否过期
	expDate, err := strconv.Atoi(r.Header.Get("EXPDATE"))
	if err != nil || expDate == 0 {
		expDate = 120
	}
	// 缓存最后存储的时间
	log.Printf("%s\t%s\t%s\t%s\n", r.RemoteAddr[0:strings.LastIndex(r.RemoteAddr, ":")], r.Method, r.Proto, r.URL)
	//从缓存中获取页面内容
	var c = cache.GetCache(int64(expDate), id, r)
	for k, v := range c.Header {
		w.Header().Add(k, v)
	}
	for _, c := range c.Cookies {
		w.Header().Add("Set-Cookie", c.Raw)
	}
	// 文件类型日志
	log.Println("Content-Type: " + c.ContentType)
	w.WriteHeader(c.StatusCode)
	w.Write(c.Data)
}
