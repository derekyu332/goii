package sql

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/frame/cache"
	"github.com/derekyu332/goii/helper/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"reflect"
	"time"
)

var (
	gEngine *xorm.Engine
)

type ISqlRecord interface {
	PrimaryKey() core.PK
}

type TableLocker interface {
	OptimisticLock() string
}

func init() {

}

func InitEngine(driver string, dbUri string, maxIdelConns int) {
	var err error
	logger.Warning("Try Connect...")
	gEngine, err = xorm.NewEngine(driver, dbUri)

	if err != nil {
		panic(err)
	}

	gEngine.SetMaxIdleConns(maxIdelConns)
	gEngine.ShowSQL(false)
	gEngine.SetLogLevel(core.LOG_WARNING)
	go keepAlive(10 * time.Second)
	logger.Warning("Connect Success")
}

func keepAlive(d time.Duration) {
	ticker := time.NewTicker(d)

	for range ticker.C {
		if err := gEngine.Ping(); err != nil {
			logger.Warning("Engine Ping failed %v", err.Error())
		}
	}
}

type SqlModel struct {
	base.Model
	oldAttr map[string]interface{}
}

func (this *SqlModel) GetEngine() *xorm.Engine {
	return gEngine
}

func (this *SqlModel) LOG_RET_ERR(tableName string, tStart int64, op string, query interface{}, err error) error {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	duration := now - tStart
	var content string

	if err != nil {
		content = fmt.Sprintf("%v|%v|%v|%v|%v|%v", tableName, op, query,
			this.RequestID, duration, 1)
	} else {
		content = fmt.Sprintf("%v|%v|%v|%v|%v|%v", tableName, op, query,
			this.RequestID, duration, 0)
	}

	logger.Bill("SQL|%v", content)

	return err
}

func (this *SqlModel) RefreshOldAttr() {
	this.oldAttr = base.StructToMap(this.Data, "sql")
}

func (this *SqlModel) GetDirtyCols() []string {
	if this.Data == nil {
		return nil
	}

	var cols []string
	attr := base.StructToMap(this.Data, "sql")

	for name, oldValue := range this.oldAttr {
		value, ok := attr[name]

		if ok {
			obj := reflect.TypeOf(oldValue)

			if obj.Kind() == reflect.Struct || obj.Kind() == reflect.Slice {
				if !reflect.DeepEqual(value, oldValue) {
					cols = append(cols, name)
				}
			} else {
				if oldValue != value {
					cols = append(cols, name)
				}
			}
		}
	}

	return cols
}

func (this *SqlModel) Get(id interface{}) (base.IActiveRecord, error) {
	if gEngine == nil {
		return nil, errors.New("Unexpected error")
	}

	if cache_record, ok := this.Data.(cache.ICacheRecord); ok {
		if record := cache.GetCacheRecord(fmt.Sprintf("%v@%v", id, this.Data.TableName())); record != nil {
			if cache_record.ReadOnly() {
				this.Data = record
				this.Exists = true
				return this.Data, nil
			} else {
				this.Data = reflect.New(reflect.ValueOf(record).Elem().Type()).Interface().(base.IActiveRecord)
				this.RefreshOldAttr()
				this.Exists = true
				return this.Data, nil
			}
		}
	}

	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	has, err := gEngine.ID(id).Get(this.Data)

	if err == nil {
		if has {
			logger.Info("[%v] Get %v success", this.RequestID, id)
			logger.Info("[%v] %v", this.RequestID, this.Data)
			this.RefreshOldAttr()
			this.Exists = true

			if _, ok := this.Data.(cache.ICacheRecord); ok {
				logger.Info("[%v] Set %v To Cache", this.RequestID, id)
				cache.SetCacheRecord(this.Data)
			}
		} else {
			logger.Info("[%v] Get %v no record", this.RequestID, id)
			this.Exists = false
		}
	} else {
		logger.Error("[%v] Get %v failed %v", this.RequestID, id, err)
	}

	return this.Data, this.LOG_RET_ERR(this.Data.TableName(), req_start, "Get", id, err)
}

func (this *SqlModel) Count(query interface{}, args ...interface{}) (int64, error) {
	if gEngine == nil {
		return 0, errors.New("Unexpected error")
	}

	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	total, err := gEngine.Where(query, args...).Count(this.Data)

	if err != nil {
		logger.Error("[%v] Find %v failed %v", this.RequestID, query, err)
	}

	return total, this.LOG_RET_ERR(this.Data.TableName(), req_start, "Count", query, err)
}

type FindAllSelector struct {
	Cond  string
	Page  int
	Limit int
	Cols  []string
	Asc   string
	Desc  string
}

func (this *SqlModel) FindAll(result interface{}, selector FindAllSelector) error {
	if selector.Asc == "" && selector.Desc == "" {
		return errors.New("selector.Asc == ''")
	}

	if gEngine == nil {
		return errors.New("Unexpected error")
	}

	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	var session *xorm.Session

	if selector.Asc != "" {
		session = gEngine.Asc(selector.Asc)
	}

	if selector.Desc != "" {
		session = gEngine.Desc(selector.Desc)
	}

	if selector.Cols != nil {
		session = session.Cols(selector.Cols...)
	}

	var err error

	if selector.Page < 0 || selector.Limit <= 0 {
		if selector.Cond == "" {
			err = session.Find(result)
		} else {
			err = session.Where(selector.Cond).Find(result)
		}
	} else {
		offset := selector.Page * selector.Limit

		if selector.Cond == "" {
			err = session.Limit(selector.Limit, offset).Find(result)
		} else {
			err = session.Where(selector.Cond).Limit(selector.Limit, offset).Find(result)
		}
	}

	if err != nil {
		logger.Error("[%v] FindAll %v failed %v", this.RequestID, selector.Cond, err)
	} else {
		logger.Info("[%v] FindAll %v Success.", this.RequestID, selector.Cond)
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "FindAll", selector.Cond, err)
}

func (this *SqlModel) FindOne(query interface{}, args ...interface{}) (base.IActiveRecord, error) {
	if gEngine == nil {
		return nil, errors.New("Unexpected error")
	}

	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	has, err := gEngine.Where(query, args...).Get(this.Data)

	if err == nil {
		if has {
			logger.Info("[%v] Find %v success", this.RequestID, query)
			logger.Info("[%v] %v", this.RequestID, this.Data)
			this.RefreshOldAttr()
			this.Exists = true
		} else {
			logger.Info("[%v] Find %v no record", this.RequestID, query)
			this.Exists = false
		}
	} else {
		logger.Error("[%v] Find %v failed %v", this.RequestID, query, err)
	}

	return this.Data, this.LOG_RET_ERR(this.Data.TableName(), req_start, "FindOne", query, err)
}

func (this *SqlModel) Delete() error {
	if gEngine == nil || !this.Exists {
		return errors.New("Unexpected error")
	}

	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	var err error

	if this.Exists {
		record, ok := this.Data.(ISqlRecord)

		if !ok {
			logger.Warning("[%v] ISqlRecord not implemented")
			return errors.New("Unexpected error")
		}

		var pk interface{}

		if len(record.PrimaryKey()) <= 1 {
			pk = this.Data.GetId()
		} else {
			pk = record.PrimaryKey()
		}

		var affected int64
		affected, err = gEngine.ID(pk).Delete(this.Data)

		if err != nil {
			logger.Warning("[%v] Delete %v failed %v", this.RequestID, pk, err.Error())
		} else if affected == 0 {
			logger.Warning("[%v] Delete %v not affected", this.RequestID, pk)
			return errors.New("Unexpected error")
		} else {
			logger.Info("[%v] Delete %v success", this.RequestID, pk)
		}
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Delete", this.Data.GetId(), err)
}

func (this *SqlModel) Save() error {
	if gEngine == nil {
		return errors.New("Unexpected error")
	}

	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	var err error
	var affected int64

	if this.Exists {
		dirty_cols := this.GetDirtyCols()

		if len(dirty_cols) <= 0 {
			logger.Info("[%v] Update %v No Change", this.RequestID, this.Data.GetId())
			return nil
		}

		record, ok := this.Data.(ISqlRecord)

		if !ok {
			logger.Warning("[%v] ISqlRecord not implemented")
			return errors.New("Unexpected error")
		}

		/*if locker, ok := this.Data.(TableLocker); ok {
			lock_name := locker.OptimisticLock()
			dirty_attr[lock_name] = this.oldAttr[lock_name]
		}*/

		var pk interface{}

		if len(record.PrimaryKey()) <= 1 {
			pk = this.Data.GetId()
		} else {
			pk = record.PrimaryKey()
		}

		affected, err = gEngine.ID(pk).Cols(dirty_cols...).Update(this.Data)

		if err != nil {
			logger.Warning("[%v] Update %v failed %v", this.RequestID, pk, err.Error())
		} else if affected == 0 {
			logger.Warning("[%v] Update %v not affected", this.RequestID, pk)
			return errors.New("Unexpected error")
		} else {
			logger.Info("[%v] Update %v success", this.RequestID, pk)

			if _, ok := this.Data.(cache.ICacheRecord); ok {
				logger.Info("[%v] Set %v To Cache", this.RequestID, pk)
				cache.SetCacheRecord(this.Data)
			}

			this.RefreshOldAttr()
		}
	} else {
		affected, err = gEngine.Insert(this.Data)

		if err != nil {
			logger.Warning("[%v] Insert %v failed %v", this.RequestID, this.Data, err.Error())
		} else {
			logger.Info("[%v] Insert %v success", this.RequestID, this.Data)
			this.RefreshOldAttr()
		}
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Save", this.Data.GetId(), err)
}
