package cache

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	// 定义常量
	RedisClient *redis.Pool
	REDIS_HOST  string
	REDIS_DB    int
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
			log.Println(REDIS_DB)
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
	if _, err := rc.Do("SET", c.Id+"_date", time.Now().Unix()); err != nil {
		log.Println(err)
	}
}

func GetCache(id string) *Cache {
	rc := RedisClient.Get()
	defer rc.Close()
	data, err := redis.String(rc.Do("GET", id))
	if err != nil {
		log.Println(err)
	}
	c := &Cache{}
	if err = json.Unmarshal([]byte(data), c); err != nil {
		log.Println(err)
		return nil
	}
	return c
}

func GetCacheStoredTime(id string) int64 {
	rc := RedisClient.Get()
	defer rc.Close()
	date, err := redis.Int64(rc.Do("GET", id+"_date"))
	if err != nil {
		log.Println(err)
		return 0
	}
	return date
}
