package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/gomodule/redigo/redis"
	"time"
)

const (
	REDIS_MAX_IDLE         = 32
	REDIS_CONNECT_TIME_OUT = 3
	REDIS_READ_TIME_OUT    = 10
	REDIS_WRITE_TIME_OUT   = 5
	REDIS_IDLE_TIME_OUT    = 240
)

var (
	gRedisPool *redis.Pool
	gRawUrl    string
	gPassword  string
)

type RedisTimeTracker interface {
	SetModified(time.Time)
}

type RedisModel struct {
	base.Model
}

func InitConnection(url string, passowrd string) {
	logger.Warning("Try Connect...")
	gRawUrl = url
	gPassword = passowrd
	_, err := redis.DialURL(gRawUrl)

	if err != nil {
		panic(err)
	}

	if gRedisPool != nil {
		gRedisPool.Close()
		gRedisPool = nil
	}

	logger.Warning("Connect Success")
}

func (this *RedisModel) GetPool() *redis.Pool {
	if gRedisPool == nil {
		gRedisPool = &redis.Pool{
			Dial: func() (redis.Conn, error) {
				c, err := redis.DialURL(gRawUrl, redis.DialConnectTimeout(REDIS_CONNECT_TIME_OUT*time.Second),
					redis.DialReadTimeout(REDIS_READ_TIME_OUT*time.Second),
					redis.DialWriteTimeout(REDIS_WRITE_TIME_OUT*time.Second))

				if err != nil {
					logger.Error("DialURL %v error %v", gRawUrl, err)
					return nil, err
				}

				if _, err := c.Do("AUTH", gPassword); err != nil {
					logger.Error("AUTH error %v", err)
					return nil, err
				}

				return c, err
			},
			MaxIdle:     REDIS_MAX_IDLE,
			IdleTimeout: REDIS_IDLE_TIME_OUT * time.Second,
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}

				_, err := c.Do("PING")

				if err != nil {
					logger.Error("check connection error %v", err)
				}

				return err
			},
		}
	}

	return gRedisPool
}

func (this *RedisModel) LOG_RET_ERR(tableName string, tStart int64, op string, id interface{}, err error) error {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	duration := now - tStart
	var content string

	if err != nil {
		content = fmt.Sprintf("%v|%v|%v|%v|%v|%v", tableName, op, id,
			this.RequestID, duration, 1)
	} else {
		content = fmt.Sprintf("%v|%v|%v|%v|%v|%v", tableName, op, id,
			this.RequestID, duration, 0)
	}

	logger.Profile("REDIS|%v", content)

	return err
}

func (this *RedisModel) HGet(data base.IActiveRecord, id interface{}, hkey string) (base.IActiveRecord, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetPool().Get()
	defer session.Close()
	key := fmt.Sprintf("%v:%v", data.TableName(), id)
	jsonBytes, err := redis.Bytes(session.Do("HGET", key, hkey))

	if jsonBytes != nil && err == nil {
		err = json.Unmarshal(jsonBytes, data)

		if err != nil {
			logger.Warning("[%v] HGET %v decode failed %v", this.RequestID, key, err)
		} else {
			logger.Info("[%v] HGET %v-%v success", this.RequestID, key, hkey)
			logger.Info("[%v] %v", this.RequestID, data)
		}
	} else if err == redis.ErrNil {
		logger.Info("[%v] HGET %v no record", this.RequestID, key)
		err = nil
	} else {
		logger.Error("[%v] HGET %v failed %v", this.RequestID, key, err)
	}

	return data, this.LOG_RET_ERR(data.TableName(), req_start, "HGet", key, err)
}

func (this *RedisModel) HSet(data base.IActiveRecord, id interface{}, hkey string) error {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetPool().Get()
	defer session.Close()
	key := fmt.Sprintf("%v:%v", data.TableName(), id)
	if tracker, ok := data.(RedisTimeTracker); ok {
		tracker.SetModified(time.Now())
	}

	value_map := base.StructToMap(data, "json")
	value, err := json.Marshal(value_map)

	if err != nil {
		logger.Error("[%v] Save %v failed %v", this.RequestID, key, err)
		return err
	}

	if _, err := session.Do("HSET", key, hkey, value); err != nil {
		logger.Error("[%v] HSET %v failed %v", this.RequestID, key, err)
	} else {
		logger.Info("[%v] HSET %v success", this.RequestID, data)
	}

	return this.LOG_RET_ERR(data.TableName(), req_start, "HSet", key, err)
}

func (this *RedisModel) Expire(data base.IActiveRecord, id interface{}, expiration int64) error {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetPool().Get()
	defer session.Close()
	key := fmt.Sprintf("%v:%v", data.TableName(), id)
	_, err := session.Do("EXPIRE", key, expiration)

	if err != nil {
		logger.Error("[%v] EXPIRE %v failed %v", this.RequestID, key, err)
	}

	return this.LOG_RET_ERR(data.TableName(), req_start, "Expire", key, err)
}

func (this *RedisModel) FindOne(id interface{}) (base.IActiveRecord, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetPool().Get()
	defer session.Close()
	key := fmt.Sprintf("%v:%v", this.Data.TableName(), id)
	jsonBytes, err := redis.Bytes(session.Do("GET", key))

	if jsonBytes != nil && err == nil {
		err = json.Unmarshal(jsonBytes, this.Data)

		if err != nil {
			logger.Warning("[%v] Find %v decode failed %v", this.RequestID, key, err)
		} else {
			logger.Info("[%v] Find %v success", this.RequestID, key)
			logger.Info("[%v] %v", this.RequestID, this.Data)
			this.Exists = true
		}
	} else if err == redis.ErrNil {
		logger.Info("[%v] Find %v no record", this.RequestID, key)
		this.Exists = false
		err = nil
	} else {
		logger.Error("[%v] Find %v failed %v", this.RequestID, key, err)
	}

	return this.Data, this.LOG_RET_ERR(this.Data.TableName(), req_start, "FindOne", id, err)
}

func (this *RedisModel) Save() error {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetPool().Get()
	defer session.Close()
	key := fmt.Sprintf("%v:%v", this.Data.TableName(), this.Data.GetId())

	if tracker, ok := this.Data.(RedisTimeTracker); ok {
		tracker.SetModified(time.Now())
	}

	value_map := base.StructToMap(this.Data, "json")
	value, err := json.Marshal(value_map)

	if err != nil {
		logger.Error("[%v] Save %v failed %v", this.RequestID, this.Data.GetId(), err)
		return err
	}

	if !this.Exists {
		n, err := session.Do("SETNX", key, value)

		if err != nil {
			logger.Error("[%v] Save %v failed %v", this.RequestID, this.Data.GetId(), err)
			return err
		} else if n != int64(1) {
			logger.Error("[%v] Save %v duplicate", this.RequestID, this.Data.GetId())
			return errors.New("duplicate")
		} else {
			logger.Info("[%v] Insert %v success", this.RequestID, this.Data)
		}
	} else {
		_, err := session.Do("SET", key, value)

		if err != nil {
			logger.Error("[%v] Save %v failed %v", this.RequestID, this.Data.GetId(), err)
			return err
		} else {
			logger.Info("[%v] Update %v success", this.RequestID, this.Data)
		}
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Save", this.Data.GetId(), err)
}

func (this *RedisModel) Delete() error {
	if !this.Exists {
		return nil
	}

	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetPool().Get()
	defer session.Close()
	key := fmt.Sprintf("%v:%v", this.Data.TableName(), this.Data.GetId())

	if tracker, ok := this.Data.(RedisTimeTracker); ok {
		tracker.SetModified(time.Now())
	}

	_, err := session.Do("DEL", key)

	if err != nil {
		logger.Error("[%v] Delete %v failed %v", this.RequestID, this.Data.GetId(), err)
		return err
	} else {
		logger.Info("[%v] Delete %v success", this.RequestID, this.Data)
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Delete", this.Data.GetId(), err)
}
