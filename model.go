package wemvc

import (
	"net/url"
	"reflect"
	"strconv"
)

// ModelParse convert the url values(request form data, url query) to model
func ModelParse(m interface{}, values url.Values) {
	mValue := reflect.ValueOf(m)
	mType := mValue.Elem().Type()
	mName := mType.Name()
	fieldNum := mType.NumField()
	for i := 0; i < fieldNum; i++ {
		field := mType.Field(i)
		fieldName, ok := field.Tag.Lookup("field")
		if !ok || len(fieldName) == 0 {
			fieldName, ok = field.Tag.Lookup("json")
			if !ok || len(fieldName) == 0 {
				fieldName = field.Name
			}
		}
		value := values.Get(fieldName)
		if len(value) == 0 && len(mName) > 0 {
			value = values.Get(strAdd(mName, ".", fieldName))
		}
		if len(value) == 0 {
			continue
		}
		fieldValue := mValue.Elem().Field(i)
		if fieldValue.IsValid() && fieldValue.CanSet() {
			switch field.Type.Kind() {
			case reflect.Bool:
				v, _ := strconv.ParseBool(value)
				fieldValue.SetBool(v)
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v, _ := strconv.ParseInt(value, 0, 64)
				fieldValue.SetInt(v)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				v, _ := strconv.ParseUint(value, 0, 64)
				fieldValue.SetUint(v)
			case reflect.Float32, reflect.Float64:
				v, _ := strconv.ParseFloat(value, 64)
				fieldValue.SetFloat(v)
			}
		}
	}
}
