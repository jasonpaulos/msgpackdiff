package msgpackdiff

import (
	"fmt"
	"time"

	"github.com/algorand/msgp/msgp"
)

// MsgpObject contains a parsed MessagePack object and its type.
type MsgpObject struct {
	Type  msgp.Type
	Value interface{}
}

func (obj MsgpObject) String() string {
	if obj.Type == msgp.NilType {
		return "null"
	}
	return fmt.Sprintf("%v(%v)", obj.Type, obj.Value)
}

// IsEmpty checks if the value of a MessagePack object is the zero value for its type.
func (obj MsgpObject) IsEmpty() (empty bool) {
	switch obj.Type {
	case msgp.InvalidType:
		empty = true
	case msgp.StrType:
		empty = len(obj.Value.(string)) == 0
	case msgp.BinType:
		empty = len(obj.Value.([]byte)) == 0
	case msgp.MapType:
		valueMap := obj.Value.(map[string]MsgpObject)
		empty = true
		for _, value := range valueMap {
			if !value.IsEmpty() {
				empty = false
				break
			}
		}
	case msgp.ArrayType:
		valueArray := obj.Value.([]MsgpObject)
		empty = true
		for _, item := range valueArray {
			if !item.IsEmpty() {
				empty = false
				break
			}
		}
	case msgp.Float32Type:
		empty = obj.Value.(float32) == 0.0
	case msgp.Float64Type:
		empty = obj.Value.(float64) == 0.0
	case msgp.BoolType:
		empty = !obj.Value.(bool)
	case msgp.IntType:
		empty = obj.Value.(int64) == 0
	case msgp.UintType:
		empty = obj.Value.(uint64) == 0
	case msgp.NilType:
		empty = true
	case msgp.Complex64Type:
		empty = obj.Value.(complex64) == 0
	case msgp.Complex128Type:
		empty = obj.Value.(complex128) == 0
	case msgp.TimeType:
		empty = obj.Value.(time.Time).IsZero()
	}
	return
}
