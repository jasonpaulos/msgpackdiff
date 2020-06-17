package msgpackdiff

import (
	"encoding/base64"
	"errors"
	"io/ioutil"

	"github.com/algorand/msgp/msgp"
)

// GetBinary gathers the binary content of a string that represents a MessagePack object. The string
// may be a base64 encoded binary object, or the path to a binary file that contains the object as
// its only content.
func GetBinary(object string) ([]byte, error) {
	decoded, b64err := base64.StdEncoding.DecodeString(object)
	if b64err == nil {
		return decoded, nil
	}

	content, fileErr := ioutil.ReadFile(object)
	if fileErr != nil {
		return []byte{}, fileErr
	}

	// TODO: what if content is not binary, but base64 encoded?
	return content, nil
}

// Parse parses a MessagePack encoded binary object into an in-memory data structure.
func Parse(bytes []byte) (parsed MsgpObject, remaining []byte, err error) {
	parsed.Type = msgp.NextType(bytes)
	switch parsed.Type {
	case msgp.StrType:
		parsed.Object, bytes, err = msgp.ReadStringBytes(bytes)
	case msgp.BinType:
		parsed.Object, bytes, err = msgp.ReadBytesBytes(bytes, nil)
	case msgp.MapType:
		var size int
		size, _, bytes, err = msgp.ReadMapHeaderBytes(bytes)
		objectMap := make(map[string]MsgpObject)
		for size > 0 {
			size--
			var key string
			key, bytes, err = msgp.ReadStringBytes(bytes)
			if err != nil {
				break
			}
			objectMap[key], bytes, err = Parse(bytes)
			if err != nil {
				break
			}
		}
		parsed.Object = objectMap
	case msgp.ArrayType:
		var size int
		size, _, bytes, err = msgp.ReadArrayHeaderBytes(bytes)
		objectArray := make([]MsgpObject, size)
		for i := 0; i < size; i++ {
			objectArray[i], bytes, err = Parse(bytes)
			if err != nil {
				break
			}
		}
		parsed.Object = objectArray
	case msgp.Float32Type:
		parsed.Object, bytes, err = msgp.ReadFloat32Bytes(bytes)
	case msgp.Float64Type:
		parsed.Object, bytes, err = msgp.ReadFloat64Bytes(bytes)
	case msgp.BoolType:
		parsed.Object, bytes, err = msgp.ReadBoolBytes(bytes)
	case msgp.IntType:
		parsed.Object, bytes, err = msgp.ReadInt64Bytes(bytes)
	case msgp.UintType:
		parsed.Object, bytes, err = msgp.ReadUint64Bytes(bytes)
	case msgp.NilType:
		parsed.Object = nil
	case msgp.Complex64Type:
		parsed.Object, bytes, err = msgp.ReadComplex64Bytes(bytes)
	case msgp.Complex128Type:
		parsed.Object, bytes, err = msgp.ReadComplex128Bytes(bytes)
	case msgp.TimeType:
		parsed.Object, bytes, err = msgp.ReadTimeBytes(bytes)
	default:
		err = errors.New("Invalid MessagePack type")
	}
	remaining = bytes
	return
}
