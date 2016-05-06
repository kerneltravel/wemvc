package utils

import (
	"encoding/xml"
	"io/ioutil"
	"encoding/json"
)

func Data2Json(data interface{}) string {
	if data == nil {
		return ""
	}
	bytes, err := json.Marshal(data)
	if err == nil {
		return ""
	}
	return string(bytes)
}

func Data2Xml(data interface{}) string {
	if data == nil {
		return ""
	}
	bytes, err := xml.Marshal(data)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func File2Xml(fpath string, v interface{}) error {
	bytes, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(bytes, v)
	if err != nil {
		return err
	}
	return nil
}