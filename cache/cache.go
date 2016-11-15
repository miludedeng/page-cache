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
	"errors"
)


type Cache struct {
	Id          string
	ContentType string
	StatusCode  int
	Data        []byte
	Header      map[string]string
	Cookies     []*http.Cookie
}

var cacheMap = make(map[string]string)

func (c *Cache) Save() {
	cacheJ, _ := json.Marshal(c)
	cacheMap[c.Id] =  string(cacheJ)
	cacheMap[c.Id+"_date"] = strconv.FormatInt(int64(time.Now().Unix()),10)

}

func Get(key string) (string,error) {
	data := cacheMap[key]
	if data == ""{
		return "", errors.New("no-cache")
	}
	return data,nil
}

func GetCache(expdate int64, id string, r *http.Request) *Cache {
	// 参数中含有nocache=true,则不使用缓存功能
	if nocache, err := strconv.ParseBool(r.URL.Query().Get("nocache")); err == nil && nocache {
		log.Println("nocache=true")
		c := GetCacheByUrl(r)
		c.Id = id
		return c
	}

	data,err := Get(id)
	if err != nil {
		log.Println(err)
	}
	c := &Cache{}
	if err = json.Unmarshal([]byte(data), c); err != nil {
		// 防止重复缓存
		if Q.Contains(id) {
			c.Data = []byte("<script>请刷新后重试！</script>")
		} else {
			Q.Add(id)
			c = GetCacheByUrl(r)
			if c == nil {
				return c
			}
			c.Id = id
			c.Save()
			Q.Remove(id)
		}
	} else {
		log.Println("By-Cache")
		var lastTime = GetCacheStoredTime(id)
		// 缓存过期， 刷新缓存
		if time.Now().Unix()-lastTime > expdate {
			if !Q.Contains(id) {
				Q.Add(id)
				log.Println("Refresh-Cache")
				go func() {
					c = GetCacheByUrl(r)
					if c == nil {
						return
					}
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
	if resp == nil {
		return nil
	}
	cookies = resp.Cookies()
	//defer resp.Body.Close()

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
	date,err := Get(id+"_date")
	if err != nil {
		log.Println(err)
		return 0
	}
	dateI,err :=  strconv.ParseInt(date,10,64)
	if err != nil {
		log.Println(err)
		return 0
	}
	return dateI

}
