package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	maxActive   = flag.Int("maxactive", 1, "redis max connections")
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
	var lastTime = cache.GetCacheStoredTime(id)
	cache.GetCache(id)
	println("now:", time.Now().Unix())
	println("lastTime:", lastTime)
	if time.Now().Unix()-lastTime < int64(expDate) {
		log.Printf("%s\t%s\t%s\t%s\t%s\n", "Cached", r.RemoteAddr[0:strings.LastIndex(r.RemoteAddr, ":")], r.Method, r.Proto, r.URL)
		// 缓存未过期，使用缓存
		//从缓存中获取页面内容
		var c = cache.GetCache(id)
		if err != nil {
			log.Println(err)
		}
		for k, v := range c.Header {
			w.Header().Add(k, v)
		}
		for _, c := range c.Cookies {
			w.Header().Add("Set-Cookie", c.Raw)
		}
		// 文件类型日志
		log.Println("Content-Type: " + c.ContentType)
		// 如果开启css文件合并，将页面中引入的css文本合并到当前页面中 goquery只支持utf8的页面解析
		w.WriteHeader(c.StatusCode)
		w.Write(c.Data)
	} else {
		log.Printf("%s\t%s\t%s\t%s\t%s\n", "No-Cache", r.RemoteAddr[0:strings.LastIndex(r.RemoteAddr, ":")], r.Method, r.Proto, r.URL)
		c := &cache.Cache{Id: id}
		header := make(map[string]string)
		var cookies []*http.Cookie
		// 缓存过期，重新获取页面，并保存到缓存
		resp, err := http.Get(cache.Options.Proxy + r.URL.String())
		cookies = resp.Cookies()
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			for _, vv := range v {
				header[k] = vv
				w.Header().Add(k, vv)
			}
		}
		for _, c := range resp.Cookies() {
			w.Header().Add("Set-Cookie", c.Raw)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil && err != io.EOF {
			log.Println(err)
		}
		// 文件类型日志
		contentType := http.DetectContentType(data)
		log.Println("Content-Type: ", contentType)
		// 如果开启css文件合并，将页面中引入的css文本合并到当前页面中 goquery只支持utf8的页面解析
		if cache.Options.IsConcatCss && "text/html; charset=utf-8" == contentType {
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			html, err := cache.ConcatCss(resp, cache.Options.Proxy)
			if err != nil {
				resp.StatusCode = http.StatusInternalServerError //解析错误时，修改返回状态为500
			}
			data = []byte(html)
		}
		c.Cookies = cookies
		c.Header = header
		c.Data = data
		c.ContentType = contentType
		c.StatusCode = resp.StatusCode
		c.Save()
		w.WriteHeader(resp.StatusCode)
		w.Write(data)
	}
}
