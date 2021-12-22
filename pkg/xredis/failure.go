package xredis

import "encoding/json"

type XFailure struct {
	Err     error
	Payload XGenericMap
}
type XGenericMap map[string]interface{}

func (x *XGenericMap) String() string {
	b, _ := json.Marshal(x)
	return string(b)
}
