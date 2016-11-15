package cache

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	// 定义常量
	RedisClient *redis.Pool
	REDIS_HOST string
	REDIS_DB int
)

func InitRedisPool() {
	// 从配置文件获取redis的ip以及db
	REDIS_HOST = Options.RedisHost + ":" + strconv.Itoa(Options.RedisPort)
	REDIS_DB = Options.RedisDB
	// 建立连接池
	RedisClient = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     Options.MaxIdle,
		MaxActive:   Options.MaxActive,
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", REDIS_HOST)
			if err != nil {
				return nil, err
			}
			// 选择db
			c.Do("SELECT", REDIS_DB)
			return c, nil
		},
	}
}

type Cache struct {
	Id          string
	ContentType string
	StatusCode  int
	Data        []byte
	Header      map[string]string
	Cookies     []*http.Cookie
}

func (c *Cache) Save() {
	cacheJ, _ := json.Marshal(c)
	// 从池里获取连接
	rc := RedisClient.Get()
	defer rc.Close()
	log.Println("ID", c.Id)
	if _, err := rc.Do("SET", c.Id, string(cacheJ)); err != nil {
		log.Println(err)
	}
	if _, err := rc.Do("SET", c.Id + "_date", time.Now().Unix()); err != nil {
		log.Println(err)
	}
}

func GetCache(expdate int64, id string, r *http.Request) *Cache {
	// 参数中含有nocache=true,则不使用缓存功能
	if nocache, err := strconv.ParseBool(r.URL.Query().Get("nocache")); err == nil && nocache {
		log.Println("nocache=true")
		c := GetCacheByUrl(r)
		c.Id = id
		return c
	}
	rc := RedisClient.Get()
	defer rc.Close()
	data, err := redis.String(rc.Do("GET", id))
	if err != nil {
		log.Println(err)
	}
	c := &Cache{}
	if err = json.Unmarshal([]byte(data), c); err != nil {
		log.Println("No-Cache")
		// 防止重复缓存
		if Q.Contains(id) {
			c.Data = []byte("<script>请刷新后重试！</script>")
		} else {
			Q.Add(id)
			c = GetCacheByUrl(r)
			c.Id = id
			c.Save()
			Q.Remove(id)
		}
	} else {
		log.Println("Cached")
		var lastTime = GetCacheStoredTime(id)
		// 缓存过期， 刷新缓存
		if time.Now().Unix() - lastTime > expdate {
			if !Q.Contains(id) {
				Q.Add(id)
				log.Println("Refresh-Cache")
				go func() {
					c = GetCacheByUrl(r)
					c.Id = id
					c.Save()
					Q.Remove(id)
				}()
			}
		}
	}
	return c
}

func GetCacheByUrl(r *http.Request) *Cache {
	c := &Cache{}
	header := make(map[string]string)
	var cookies []*http.Cookie
	// 缓存过期，重新获取页面，并保存到缓存
	resp, err := http.Get(Options.Proxy + r.URL.String())
	if err != nil {
		log.Println(err)
	}
	cookies = resp.Cookies()
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			header[k] = vv
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		log.Println(err)
	}
	// 文件类型日志
	contentType := http.DetectContentType(data)
	// 如果开启css文件合并，将页面中引入的css文本合并到当前页面中 goquery只支持utf8的页面解析
	if Options.IsConcatCss && "text/html; charset=utf-8" == contentType {
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		html, err := ConcatCss(resp, Options.Proxy)
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
	return c
}

func GetCacheStoredTime(id string) int64 {
	rc := RedisClient.Get()
	defer rc.Close()
	date, err := redis.Int64(rc.Do("GET", id + "_date"))
	if err != nil {
		return 0
	}
	return date
}
