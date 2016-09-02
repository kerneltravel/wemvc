package wemvc

import (
	"encoding/xml"
	"io/ioutil"
	//"encoding/json"
)

/*
func data2Json(data interface{}) []byte {
	if data == nil {
		return nil
	}
	bytes, err := json.Marshal(data)
	if err == nil {
		return nil
	}
	return bytes
}

func data2Xml(data interface{}) []byte {
	if data == nil {
		return nil
	}
	bytes, err := xml.Marshal(data)
	if err != nil {
		return nil
	}
	return bytes
}
*/
func file2Xml(path string, v interface{}) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(bytes, v)
	if err != nil {
		return err
	}
	return nil
}
