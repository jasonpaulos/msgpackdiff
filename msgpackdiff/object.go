package msgpackdiff

import (
	"time"

	"github.com/algorand/msgp/msgp"
)

// MsgpObject contains a parsed MessagePack object and its type.
type MsgpObject struct {
	Type   msgp.Type
	Object interface{}
}

// IsEmpty checks if the value of a MessagePack object is the zero value for its type.
func (obj MsgpObject) IsEmpty() (empty bool) {
	switch obj.Type {
	case msgp.InvalidType:
		empty = true
	case msgp.StrType:
		empty = len(obj.Object.(string)) == 0
	case msgp.BinType:
		empty = len(obj.Object.([]byte)) == 0
	case msgp.MapType:
		objMap := obj.Object.(map[string]MsgpObject)
		empty = true
		for _, value := range objMap {
			if !value.IsEmpty() {
				empty = false
				break
			}
		}
	case msgp.ArrayType:
		objArray := obj.Object.([]MsgpObject)
		empty = true
		for _, item := range objArray {
			if !item.IsEmpty() {
				empty = false
				break
			}
		}
	case msgp.Float32Type:
		empty = obj.Object.(float32) == 0.0
	case msgp.Float64Type:
		empty = obj.Object.(float64) == 0.0
	case msgp.BoolType:
		empty = !obj.Object.(bool)
	case msgp.IntType:
		empty = obj.Object.(int64) == 0
	case msgp.UintType:
		empty = obj.Object.(uint64) == 0
	case msgp.NilType:
		empty = true
	case msgp.Complex64Type:
		empty = obj.Object.(complex64) == 0
	case msgp.Complex128Type:
		empty = obj.Object.(complex128) == 0
	case msgp.TimeType:
		empty = obj.Object.(time.Time).IsZero()
	}
	return
}
