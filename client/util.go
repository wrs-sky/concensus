package client

import (
	"encoding/json"
)

func ObjToJson(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(bytes)
}
