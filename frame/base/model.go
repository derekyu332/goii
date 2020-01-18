package base

import (
	"github.com/derekyu332/goii/helper/extend"
	"reflect"
	"strings"
)

const (
	DEFAULT_SCENARIO = "default"
)

type IMapable interface {
	Fields() map[string]interface{}
}

type IActiveRecord interface {
	GetId() interface{}
	TableName() string
}

type IFieldMethod interface {
	Attr(string) interface{}
}

type Model struct {
	Data          IActiveRecord //interface{}
	Exists        bool
	RequestID     int64
	scenarioName  string
	scenariosMaps map[string][]string
}

func StructToMap(obj interface{}, tagName string) map[string]interface{} {
	obj1 := reflect.TypeOf(obj)

	if obj1.Kind() == reflect.Ptr {
		obj1 = obj1.Elem()
	}

	obj2 := reflect.ValueOf(obj)

	if obj2.Kind() == reflect.Ptr {
		obj2 = obj2.Elem()
	}

	var data = make(map[string]interface{})

	for i := 0; i < obj1.NumField(); i++ {
		tempField := obj1.Field(i)

		if fieldName := tempField.Tag.Get(tagName); fieldName != "" {
			fieldArray := strings.Split(fieldName, ",")

			if fieldArray[0] != "-" && fieldArray[0] != "" {
				data[fieldArray[0]] = obj2.Field(i).Interface()
			}
		} else {
			data[strings.ToLower(tempField.Name)] = obj2.Field(i).Interface()
		}
	}

	return data
}

func (this *Model) Fields() map[string]interface{} {
	fieldsMap := StructToMap(this.Data, "attr")
	return this.FilterScenario(fieldsMap)
}

func (this *Model) Scenarios(scenario string) (string, []string) {
	switch scenario {
	case DEFAULT_SCENARIO:
		{
			obj1 := reflect.TypeOf(this.Data)
			defaultScenario := make([]string, obj1.NumField())

			for i := 0; i < obj1.NumField(); i++ {
				defaultScenario[i] = strings.ToLower(obj1.Field(i).Name)
			}

			return scenario, defaultScenario
		}
	}

	return "", nil
}

func (this *Model) SetScenario(name string, fields []string) {
	if this.scenariosMaps == nil {
		this.scenariosMaps = make(map[string][]string)
	}

	this.scenariosMaps[name] = fields
	this.scenarioName = name
}

func (this *Model) FilterScenario(fieldsMap map[string]interface{}) map[string]interface{} {
	scenario, ok := this.scenariosMaps[this.scenarioName]

	if ok {
		for name, _ := range fieldsMap {
			if extend.InStringArray(name, scenario) < 0 {
				delete(fieldsMap, name)
			}
		}

		if fieldMethod, ok := this.Data.(IFieldMethod); ok {
			for _, name := range scenario {
				if v := fieldMethod.Attr(name); v != nil {
					fieldsMap[name] = v
				}
			}
		}
	}

	return fieldsMap
}
