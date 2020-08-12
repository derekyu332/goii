package behaviors

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/frame/redis"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/gin-gonic/gin"
	"strings"
)

type ActionLock struct {
	redis.RedisModel
	VaryByRoute bool
	Variations  map[string][]string
	Expiration  int64
	lockKey     string
}

func (this *ActionLock) BeforeAction(c *gin.Context) error {
	if this.Expiration < 0 || this.Variations == nil {
		return nil
	}

	variation, ok := this.Variations[c.Request.URL.Path]

	if ok {
		session := this.GetPool().Get()
		defer session.Close()
		key := this.calculateKey(c, variation)

		if key == "" {
			return nil
		}

		if this.Expiration > 0 {
			if n, err := session.Do("SET", key, "1", "EX", this.Expiration, "NX"); err != nil {
				logger.Warning("[%v] lock %v failed %v", this.RequestID, key, err.Error())
				return base.LockActionHttpError(c, err.Error())
			} else if n != "OK" {
				logger.Warning("[%v] lock %v failed %v", this.RequestID, key, n)
				return base.LockActionHttpError(c, "")
			}
		} else {
			if n, err := session.Do("SETNX", key, "1"); err != nil {
				logger.Warning("[%v] lock %v failed %v", this.RequestID, key, err.Error())
				return base.LockActionHttpError(c, err.Error())
			} else if n != int64(1) {
				logger.Warning("[%v] lock %v failed %v", this.RequestID, key, n)
				return base.LockActionHttpError(c, "")
			}
		}

		this.lockKey = key
		logger.Notice("[%v] Lock %v", c.GetInt64(base.KEY_REQUEST_ID), this.lockKey)
	}

	return nil
}

func (this *ActionLock) calculateKey(c *gin.Context, variation []string) string {
	var keys []string

	if this.VaryByRoute {
		keys = append(keys, c.Request.URL.Path)
	}

	queries := c.Request.URL.Query()
	forms := c.Request.PostForm

	for _, name := range variation {
		if value, ok := queries[name]; ok {
			keys = append(keys, name, value[0])
		} else if value, ok := forms[name]; ok {
			keys = append(keys, name, value[0])
		}
	}

	if len(keys) <= 0 {
		return ""
	}

	joinStr := strings.Join(keys, "-")
	src := sha1.Sum([]byte(joinStr))

	return hex.EncodeToString(src[:])
}

func (this *ActionLock) AfterAction(c *gin.Context) error {
	if this.Expiration < 0 || this.Variations == nil {
		return nil
	}

	if this.lockKey != "" {
		session := this.GetPool().Get()
		defer session.Close()

		if _, err := session.Do("DEL", this.lockKey); err != nil {
			return base.LockActionHttpError(c, err.Error())
		}

		logger.Notice("[%v] Unlock %v", c.GetInt64(base.KEY_REQUEST_ID), this.lockKey)
	}

	return nil
}
