package wemvc

import (
	"net/url"
	"reflect"
	"strconv"
)

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
	case map[string]bool:
		return collection.(map[string]bool)[key]
	case map[string]string:
		return collection.(map[string]string)[key]
	case map[string]int:
		return collection.(map[string]int)[key]
	case map[string]int8:
		return collection.(map[string]int8)[key]
	case map[string]int16:
		return collection.(map[string]int16)[key]
	case map[string]int32:
		return collection.(map[string]int32)[key]
	case map[string]int64:
		return collection.(map[string]int64)[key]
	case map[string]uint:
		return collection.(map[string]uint)[key]
	case map[string]uint8:
		return collection.(map[string]uint8)[key]
	case map[string]uint16:
		return collection.(map[string]uint16)[key]
	case map[string]uint32:
		return collection.(map[string]uint32)[key]
	case map[string]uint64:
		return collection.(map[string]uint64)[key]
	case map[string]float32:
		return collection.(map[string]float32)[key]
	case map[string]float64:
		return collection.(map[string]float64)[key]
	}
	return nil
}

func setFieldInt64(valueField reflect.Value, value interface{}) {
	if valueStr,ok := value.(string); ok && len(valueStr) > 0 {
		v, _ := strconv.ParseInt(valueStr, 0, 64)
		valueField.SetInt(v)
	}
}

func setFieldUint64(valueField reflect.Value, value interface{}) {
	if valueStr, ok := value.(string); ok && len(valueStr) > 0 {
		v, _ := strconv.ParseUint(valueStr, 0, 64)
		valueField.SetUint(v)
	}
}

func setFieldFloat64(valueField reflect.Value, value interface{}) {
	if valueStr, ok := value.(string); ok && len(valueStr) > 0 {
		v, _ := strconv.ParseFloat(valueStr, 64)
		valueField.SetFloat(v)
	}
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
		case reflect.Int:
			var v int
			var ok bool
			if v,ok = value.(int); !ok {
				setFieldInt64(valueField, value)
			} else {
				valueField.SetInt(int64(v))
			}
		case reflect.Int8:
			var v int8
			var ok bool
			if v,ok = value.(int8); !ok {
				setFieldInt64(valueField, value)
			} else {
				valueField.SetInt(int64(v))
			}
		case reflect.Int16:
			var v int16
			var ok bool
			if v,ok = value.(int16); !ok {
				setFieldInt64(valueField, value)
			} else {
				valueField.SetInt(int64(v))
			}
		case reflect.Int32:
			var v int32
			var ok bool
			if v,ok = value.(int32); !ok {
				setFieldInt64(valueField, value)
			} else {
				valueField.SetInt(int64(v))
			}
		case reflect.Int64:
			var v int64
			var ok bool
			if v,ok = value.(int64); !ok {
				setFieldInt64(valueField, value)
			} else {
				valueField.SetInt(int64(v))
			}
		case reflect.Uint:
			var v uint
			var ok bool
			if v,ok = value.(uint); !ok {
				setFieldUint64(valueField, value)
			} else {
				valueField.SetUint(uint64(v))
			}
		case reflect.Uint8:
			var v uint8
			var ok bool
			if v,ok = value.(uint8); !ok {
				setFieldUint64(valueField, value)
			} else {
				valueField.SetUint(uint64(v))
			}
		case reflect.Uint16:
			var v uint16
			var ok bool
			if v,ok = value.(uint16); !ok {
				setFieldUint64(valueField, value)
			} else {
				valueField.SetUint(uint64(v))
			}
		case reflect.Uint32:
			var v uint32
			var ok bool
			if v,ok = value.(uint32); !ok {
				setFieldUint64(valueField, value)
			} else {
				valueField.SetUint(uint64(v))
			}
		case reflect.Uint64:
			var v uint64
			var ok bool
			if v,ok = value.(uint64); !ok {
				setFieldUint64(valueField, value)
			} else {
				valueField.SetUint(uint64(v))
			}
		case reflect.Float32:
			var v float32
			var ok bool
			if v,ok = value.(float32); !ok {
				setFieldFloat64(valueField, value)
			} else {
				valueField.SetFloat(uint64(v))
			}
		case reflect.Float64:
			var v float64
			var ok bool
			if v,ok = value.(float64); !ok {
				setFieldFloat64(valueField, value)
			} else {
				valueField.SetFloat(uint64(v))
			}
		default:
			valueField.Set(reflect.ValueOf(value))
		}
	}
}