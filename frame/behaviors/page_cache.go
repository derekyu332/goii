package behaviors

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type DependencyFunc func(*gin.Context) bool

const (
	CACHE_FILE_SUFFIX = ".bin"
)

type PageCache struct {
	CachePath  string
	Variations map[string][]string
	Dependency DependencyFunc
	Duration   int64
	Version    string
	cached     bool
}

func (this *PageCache) BeforeAction(c *gin.Context) error {
	if this.Duration <= 0 || this.Variations == nil {
		return nil
	} else if this.Dependency != nil && this.Dependency(c) {
		return nil
	}

	variation, ok := this.Variations[c.Request.URL.Path]

	if ok {
		key := this.calculateKey(c, variation)
		filename := this.getCacheFile(key)

		if fi, err := os.Stat(filename); err == nil {
			if fi.ModTime().Unix()+this.Duration > time.Now().Unix() {
				if contents, err := ioutil.ReadFile(filename); err == nil {
					var data map[string]interface{}
					err := json.Unmarshal([]byte(contents), &data)

					if err == nil {
						c.Set(base.KEY_RESPONSE, data)
						this.cached = true
						logger.Notice("[%v] Get Cache %v", c.GetInt64(base.KEY_REQUEST_ID), filename)
					}
				}
			} else {
				os.Remove(filename)
			}
		}
	}

	return nil
}

func (this *PageCache) calculateKey(c *gin.Context, variation []string) string {
	var keys []string
	keys = append(keys, c.Request.URL.Path, this.Version)
	queries := c.Request.URL.Query()
	forms := c.Request.PostForm

	for _, name := range variation {
		if value, ok := queries[name]; ok {
			keys = append(keys, name, value[0])
		} else if value, ok := forms[name]; ok {
			keys = append(keys, name, value[0])
		}
	}

	joinStr := strings.Join(keys, "-")
	src := sha1.Sum([]byte(joinStr))

	return hex.EncodeToString(src[:])
}

func (this *PageCache) getCacheFile(key string) string {
	if this.CachePath == "" {
		this.CachePath = "../cache"
	}

	return this.CachePath + "/" + key + CACHE_FILE_SUFFIX
}

func (this *PageCache) AfterAction(c *gin.Context) error {
	if this.Duration <= 0 || this.Variations == nil {
		return nil
	}

	variation, ok := this.Variations[c.Request.URL.Path]

	if !ok {
		return nil
	}

	data, exist := c.Get(base.KEY_RESPONSE)

	if !exist || this.cached {
		return nil
	}

	key := this.calculateKey(c, variation)
	filename := this.getCacheFile(key)

	if fi, err := os.Stat(filename); err == nil {
		if fi.ModTime().Unix()+this.Duration > time.Now().Unix() {
			return nil
		}
	}

	content, err := json.Marshal(data)

	if err != nil {
		return nil
	}

	ioutil.WriteFile(filename, content, 0644)
	logger.Notice("[%v] Set Cache %v", c.GetInt64(base.KEY_REQUEST_ID), filename)

	return nil
}
