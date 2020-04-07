package mongo

import (
	"errors"
	"fmt"
	"github.com/derekyu332/goii/frame/base"
	"github.com/derekyu332/goii/helper/extend"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"reflect"
	"strings"
	"time"
)

var (
	gDialInfo    *mgo.DialInfo
	gDbName      string
	gMainSession *mgo.Session
)

type CollectionLocker interface {
	OptimisticLock() string
}

type CollectionTimeTracker interface {
	SetModified(time.Time)
}

type MongoModel struct {
	base.Model
	oldAttr    map[string]interface{}
	update_seq int64
}

func init() {
}

func InitConnection(initDialInfo *mgo.DialInfo, initDbName string) {
	gDialInfo = initDialInfo
	gDbName = initDbName
	logger.Warning("Try Connect...")

	if gMainSession != nil {
		gMainSession.Close()
	}

	var err error
	gMainSession, err = mgo.DialWithInfo(gDialInfo)

	if err != nil {
		panic(err)
	}

	logger.Warning("Connect Success")
}

func (this *MongoModel) GetDbName() string {
	return gDbName
}

func (this *MongoModel) GetSession() *mgo.Session {
	if gMainSession == nil {
		var err error
		gMainSession, err = mgo.DialWithInfo(gDialInfo)

		if err != nil {
			return nil
		}

		gMainSession.SetMode(mgo.Strong, false)
	}

	session := gMainSession.Copy()

	return session
}

func (this *MongoModel) LOG_RET_ERR(colName string, tStart int64, op string, cond bson.M, err error) error {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	duration := now - tStart
	var content string

	if err != nil {
		content = fmt.Sprintf("%v|%v|%v|%v|%v|%v", colName, op, cond,
			this.RequestID, duration, 1)
	} else {
		content = fmt.Sprintf("%v|%v|%v|%v|%v|%v", colName, op, cond,
			this.RequestID, duration, 0)
	}

	logger.Profile("MONGO|%v", content)

	return err
}

func (this *MongoModel) GetAutoIncrement() (int64, error) {
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C("col_counters")
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"auto_increment": 1}},
		Upsert:    true,
		ReturnNew: true,
	}
	doc := struct{ Auto_increment int64 }{}
	_, err := collection.Find(bson.M{"_id": this.Data.TableName()}).Apply(change, &doc)

	if err != nil {
		logger.Error("[%v] get AutoIncrement failed %v", this.RequestID, err)
		return 0, err
	}

	logger.Info("[%v] get AutoIncrement %v", this.RequestID, doc.Auto_increment)

	return doc.Auto_increment, nil
}

func (this *MongoModel) Count(cond bson.M) (int, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(this.GetDbName()).C(this.Data.TableName())
	c, err := collection.Find(cond).Count()

	return c, this.LOG_RET_ERR(this.Data.TableName(), req_start, "Count", cond, err)
}

func (this *MongoModel) GroupCount(key string) ([]bson.M, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(this.GetDbName()).C(this.Data.TableName())
	result := make([]bson.M, 0)
	err := collection.Pipe([]bson.M{
		bson.M{"$group": bson.M{"_id": fmt.Sprintf("$%v", key), "count": bson.M{"$sum": 1}}},
	}).All(&result)

	if err != nil {
		logger.Error("[%v] GroupCount %v failed %v", this.RequestID, key, err)
	}

	return result, this.LOG_RET_ERR(this.Data.TableName(), req_start, "GroupCount", bson.M{"key": key}, err)
}

func (this *MongoModel) Max(cond bson.M, key string) (int, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(this.GetDbName()).C(this.Data.TableName())
	result := make([]bson.M, 0)
	var err error
	var sum int64

	if cond != nil {
		err = collection.Pipe([]bson.M{
			bson.M{"$match": cond},
			bson.M{"$project": bson.M{key: 1}},
			bson.M{"$group": bson.M{"_id": nil, "max": bson.M{"$max": fmt.Sprintf("$%v", key)}}},
		}).All(&result)
	} else {
		err = collection.Pipe([]bson.M{
			bson.M{"$project": bson.M{key: 1}},
			bson.M{"$group": bson.M{"_id": nil, "max": bson.M{"$max": fmt.Sprintf("$%v", key)}}},
		}).All(&result)
	}

	if err != nil {
		logger.Error("[%v] Sum %v %v failed %v", this.RequestID, cond, key, err)
	} else if len(result) > 0 {
		v, _ := result[0]["max"]
		sum, _ = extend.InterfaceToInt64(v)
		logger.Info("[%v] %v", this.RequestID, result)
	}

	return int(sum), this.LOG_RET_ERR(this.Data.TableName(), req_start, "Max", bson.M{"key": key}, err)
}

func (this *MongoModel) Sum(cond bson.M, key string) (int, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(this.GetDbName()).C(this.Data.TableName())
	result := make([]bson.M, 0)
	var err error
	var sum int64

	if cond != nil {
		err = collection.Pipe([]bson.M{
			bson.M{"$match": cond},
			bson.M{"$project": bson.M{key: 1}},
			bson.M{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": fmt.Sprintf("$%v", key)}}},
		}).All(&result)
	} else {
		err = collection.Pipe([]bson.M{
			bson.M{"$project": bson.M{key: 1}},
			bson.M{"$group": bson.M{"_id": nil, "sum": bson.M{"$sum": fmt.Sprintf("$%v", key)}}},
		}).All(&result)
	}

	if err != nil {
		logger.Error("[%v] Sum %v %v failed %v", this.RequestID, cond, key, err)
	} else if len(result) > 0 {
		v, _ := result[0]["sum"]
		sum, _ = extend.InterfaceToInt64(v)
		logger.Info("[%v] %v", this.RequestID, result)
	}

	return int(sum), this.LOG_RET_ERR(this.Data.TableName(), req_start, "Sum", bson.M{"key": key}, err)
}

func (this *MongoModel) DistinctCount(cond bson.M, key string) (int, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(this.GetDbName()).C(this.Data.TableName())
	result := make([]bson.M, 0)
	var err error
	var count int64

	if cond != nil {
		err = collection.Pipe([]bson.M{
			bson.M{"$match": cond},
			bson.M{"$project": bson.M{key: 1}},
			bson.M{"$group": bson.M{"_id": fmt.Sprintf("$%v", key)}},
			bson.M{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
		}).All(&result)
	} else {
		err = collection.Pipe([]bson.M{
			bson.M{"$project": bson.M{key: 1}},
			bson.M{"$group": bson.M{"_id": fmt.Sprintf("$%v", key)}},
			bson.M{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}}},
		}).All(&result)
	}

	if err != nil {
		logger.Error("[%v] DistinctCount %v %v failed %v", this.RequestID, cond, key, err)
	} else if len(result) > 0 {
		v, _ := result[0]["count"]
		count, _ = extend.InterfaceToInt64(v)
		logger.Info("[%v] %v", this.RequestID, result)
	}

	return int(count), this.LOG_RET_ERR(this.Data.TableName(), req_start, "DistinctCount", bson.M{"key": key}, err)
}

func (this *MongoModel) FindOne(cond bson.M) (base.IActiveRecord, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	err := collection.Find(cond).One(this.Data)

	if err == nil {
		logger.Info("[%v] Find %v success", this.RequestID, cond)
		logger.Info("[%v] %v", this.RequestID, this.Data)
		this.RefreshOldAttr()
		this.Exists = true
	} else if err == mgo.ErrNotFound {
		logger.Info("[%v] Find %v no record", this.RequestID, cond)
		this.Exists = false
		err = nil
	} else {
		logger.Error("[%v] Find %v failed %v", this.RequestID, cond, err)
	}

	return this.Data, this.LOG_RET_ERR(this.Data.TableName(), req_start, "FindOne", cond, err)
}

type FindAllSelector struct {
	Cond  bson.M
	Col   bson.M
	Page  int
	Limit int
	Sort  string
}

func (this *MongoModel) FindAll(result interface{}, selector FindAllSelector) error {
	if selector.Cond == nil {
		return errors.New("selector.Cond == nil")
	}

	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	var iter *mgo.Iter
	var sort string

	if selector.Sort == "" {
		sort = "_id"
	} else {
		sort = selector.Sort
	}

	sort_slics := strings.Split(sort, ",")

	if selector.Page < 0 || selector.Limit <= 0 {
		if selector.Col != nil {
			iter = collection.Find(selector.Cond).Sort(sort_slics...).Select(selector.Col).Iter()
		} else {
			iter = collection.Find(selector.Cond).Sort(sort_slics...).Iter()
		}
	} else {
		offset := selector.Page * selector.Limit

		if selector.Col != nil {
			iter = collection.Find(selector.Cond).Sort(sort_slics...).Select(selector.Col).Skip(offset).Limit(selector.Limit).Iter()
		} else {
			iter = collection.Find(selector.Cond).Sort(sort_slics...).Skip(offset).Limit(selector.Limit).Iter()
		}
	}

	var err error

	if err = iter.All(result); err != nil {
		logger.Error("[%v] FindAll %v failed %v", this.RequestID, selector.Cond, err)
	} else {
		logger.Info("[%v] FindAll %v Success.", this.RequestID, selector.Cond)
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "FindAll", selector.Cond, err)
}

func (this *MongoModel) ForceIncrement(cond bson.M, incAttr string) error {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	err := collection.Update(cond, bson.M{"$inc": bson.M{incAttr: 1}})

	if err != nil {
		logger.Warning("[%v] Inc %v failed %v", this.RequestID, incAttr, err.Error())
	} else {
		logger.Info("[%v] Inc %v success", this.RequestID, incAttr)
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "ForceIncrement", cond, err)
}

func (this *MongoModel) ForceMinus(cond bson.M, minusAttr string) error {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	err := collection.Update(cond, bson.M{"$inc": bson.M{minusAttr: -1}})

	if err != nil {
		logger.Warning("[%v] Minus %v failed %v", this.RequestID, minusAttr, err.Error())
	} else {
		logger.Info("[%v] Minus %v success", this.RequestID, minusAttr)
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "ForceMinus", cond, err)
}

func (this *MongoModel) RefreshOldAttr() {
	this.oldAttr = base.StructToMap(this.Data, "bson")
	this.update_seq = 0
}

func (this *MongoModel) GetDirtyAttr() map[string]interface{} {
	if this.Data == nil {
		return nil
	}

	attr := base.StructToMap(this.Data, "bson")

	for name, oldValue := range this.oldAttr {
		value, ok := attr[name]

		if ok {
			obj := reflect.TypeOf(oldValue)

			if obj.Kind() == reflect.Struct || obj.Kind() == reflect.Slice {
				if reflect.DeepEqual(value, oldValue) {
					delete(attr, name)
				}
			} else {
				if oldValue == value {
					delete(attr, name)
				}
			}
		}
	}

	return attr
}

func (this *MongoModel) Save() error {
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	var err error
	req_start := time.Now().UnixNano() / int64(time.Millisecond)

	if this.Exists {
		if tracker, ok := this.Data.(CollectionTimeTracker); ok {
			tracker.SetModified(time.Now())
		}

		dirty_attr := this.GetDirtyAttr()

		if len(dirty_attr) <= 0 {
			logger.Info("[%v] Update %v No Change", this.RequestID, this.Data.GetId())

			return nil
		}

		if locker, ok := this.Data.(CollectionLocker); ok {
			lock_name := locker.OptimisticLock()
			lock_seq, _ := extend.InterfaceToInt64(this.oldAttr[lock_name])
			err = collection.Update(bson.D{{Name: "_id", Value: this.Data.GetId()}, {Name: lock_name, Value: lock_seq + this.update_seq}},
				bson.M{"$set": dirty_attr, "$inc": bson.M{lock_name: 1}})
		} else {
			err = collection.UpdateId(this.Data.GetId(), bson.M{"$set": dirty_attr})
		}

		if err != nil {
			logger.Warning("[%v] Update %v failed %v", this.RequestID, dirty_attr, err.Error())
		} else {
			logger.Info("[%v] Update %v success", this.RequestID, dirty_attr)
			new_update_seq := this.update_seq + 1
			this.RefreshOldAttr()
			this.update_seq = new_update_seq
		}

		return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Update", bson.M{"_id": this.Data.GetId()}, err)
	} else {
		err = collection.Insert(this.Data)

		if err != nil {
			logger.Warning("[%v] Insert %v failed %v", this.RequestID, this.Data, err.Error())
		} else {
			logger.Info("[%v] Insert %v success", this.RequestID, this.Data)
			this.Exists = true
			this.RefreshOldAttr()
		}

		return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Insert", bson.M{"_id": this.Data.GetId()}, err)
	}
}

func (this *MongoModel) Remove() error {
	if !this.Exists {
		return nil
	}

	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	var err error
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	err = collection.RemoveId(this.Data.GetId())

	if err != nil {
		logger.Warning("[%v] Remove %v failed %v", this.RequestID, this.Data.GetId(), err.Error())
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Remove", bson.M{"_id": this.Data.GetId()}, err)
}

func (this *MongoModel) RemoveAll(cond bson.M) (int, error) {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	info, err := collection.RemoveAll(cond)
	var removed int

	if err != nil {
		logger.Warning("[%v] RemoveAll %v failed %v", this.RequestID, cond, err.Error())
	} else {
		removed = info.Removed
	}

	return removed, this.LOG_RET_ERR(this.Data.TableName(), req_start, "RemoveAll", cond, err)
}

func (this *MongoModel) UpdateAll(cond bson.M, update interface{}) error {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	_, err := collection.UpdateAll(cond, update)

	if err != nil {
		logger.Warning("[%v] UpdateAll %v failed %v", this.RequestID, cond, err.Error())
	} else {
		logger.Info("[%v] UpdateAll %v success", this.RequestID, cond)
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "UpdateAll", cond, err)
}

func (this *MongoModel) Update(cond bson.M, update interface{}, apply bool) error {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())

	if cond == nil {
		if !this.Exists {
			return errors.New("!this.Exists")
		}

		cond = bson.M{"_id": this.Data.GetId()}
	}

	if apply == true {
		change := mgo.Change{
			Update:    update,
			Upsert:    true,
			ReturnNew: true,
		}

		info, err := collection.Find(cond).Apply(change, this.Data)

		if err != nil {
			logger.Warning("[%v] Update %v failed %v", this.RequestID, cond, err.Error())
		} else if info.Updated <= 0 {
			logger.Warning("[%v] Update %v No change", this.RequestID, cond)
		} else {
			logger.Info("[%v] Update %v success", this.RequestID, cond)
		}

		return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Apply", cond, err)
	} else {
		err := collection.Update(cond, update)

		if err != nil {
			logger.Warning("[%v] Update %v failed %v", this.RequestID, cond, err.Error())
		} else {
			logger.Info("[%v] Update %v success", this.RequestID, cond)
		}

		return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Update", cond, err)
	}
}

func (this *MongoModel) Upsert(cond bson.M, update interface{}) error {
	req_start := time.Now().UnixNano() / int64(time.Millisecond)
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	_, err := collection.Upsert(cond, update)

	if err != nil {
		logger.Warning("[%v] Upsert %v failed %v", this.RequestID, cond, err.Error())
	} else {
		logger.Info("[%v] Upsert %v success", this.RequestID, cond)
	}

	return this.LOG_RET_ERR(this.Data.TableName(), req_start, "Upsert", cond, err)
}

func (this *MongoModel) EnsureIndex(index mgo.Index) error {
	session := this.GetSession()
	defer session.Close()
	collection := session.DB(gDbName).C(this.Data.TableName())
	return collection.EnsureIndex(index)
}
