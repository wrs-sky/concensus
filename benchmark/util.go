package benchmark

import (
	"encoding/json"
)

func ObjToString(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(bytes)
}
