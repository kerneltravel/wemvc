package wemvc

import (
	"net/url"
	"reflect"
	"strconv"
)

// ModelParse convert the values(url.Values, map[string]interface{}) to model
func ModelParse(m interface{}, values interface{}) {
	if values == nil || m == nil {
		return
	}
	mValue := reflect.ValueOf(m)
	mType := mValue.Elem().Type()
	mName := mType.Name()
	fieldNum := mType.NumField()
	for i := 0; i < fieldNum; i++ {
		typeField := mType.Field(i)
		fieldName, ok := typeField.Tag.Lookup("field")
		if !ok || len(fieldName) == 0 {
			fieldName, ok = typeField.Tag.Lookup("json")
			if !ok || len(fieldName) == 0 {
				fieldName = typeField.Name
			}
		}
		value := getValue(fieldName, values);
		if value == nil {
			value = getValue(strAdd(mName, ".", fieldName), values)
		}
		if value == nil {
			continue
		}
		setFieldValue(typeField.Type.Kind(), mValue.Elem().Field(i), value)
	}
}

func getValue(key string, collection interface{}) interface{} {
	switch collection.(type) {
	case url.Values:
		return collection.(url.Values).Get(key)
	case *url.Values:
		return collection.(*url.Values).Get(key)
	case map[string]interface{}:
		return collection.(map[string]interface{})[key]
	}
	return nil
}

func setFieldValue(kind reflect.Kind, valueField reflect.Value, value interface{}) {
	if valueField.IsValid() && valueField.CanSet() && value != nil {
		switch kind {
		case reflect.Bool:
			var v,ok bool
			if v,ok = value.(bool);!ok {
				if valueStr, ok := value.(string); ok && len(valueStr) > 0 {
					v,_ = strconv.ParseBool(valueStr)
				}
			}
			valueField.SetBool(v)
		case reflect.String:
			if v,ok := value.(string); ok {
				valueField.SetString(v)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var v int64
			var ok bool
			if v,ok = value.(int64); !ok {
				if valueStr,ok := value.(string); ok && len(valueStr) > 0 {
					v, _ = strconv.ParseInt(valueStr, 0, 64)
				}
			}
			valueField.SetInt(v)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var v uint64
			var ok bool
			if v,ok = value.(uint64); !ok {
				if valueStr, ok := value.(string); ok && len(valueStr) > 0 {
					v, _ = strconv.ParseUint(valueStr, 0, 64)
				}
			}
			valueField.SetUint(v)
		case reflect.Float32, reflect.Float64:
			var v float64
			var ok bool
			if v,ok = value.(float64); !ok {
				if valueStr, ok := value.(string); ok && len(valueStr) > 0 {
					v, _ = strconv.ParseFloat(valueStr, 64)
				}
			}
			valueField.SetFloat(v)
		}
	}
}

// Model2Map convert model to map[string]interface{}
func Model2Map(m interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	mValue := reflect.ValueOf(m)
	mType := reflect.TypeOf(m)
	fieldNum := mType.NumField()
	data := make(map[string]interface{}, fieldNum)
	for i := 0; i < fieldNum; i++ {
		field := mType.Field(i)
		fieldName, ok := field.Tag.Lookup("field")
		if !ok || len(fieldName) == 0 {
			fieldName, ok = field.Tag.Lookup("json")
			if !ok || len(fieldName) == 0 {
				fieldName = field.Name
			}
		}
		fieldValue := mValue.Field(i).Interface()
		data[fieldName] = fieldValue
	}
	return data
}