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
	decoded, err := base64.StdEncoding.DecodeString(object)
	if err == nil {
		return decoded, nil
	}

	content, err := ioutil.ReadFile(object)
	if err != nil {
		return []byte{}, err
	}

	// attempt to decode from base64
	maxLen := base64.StdEncoding.DecodedLen(len(content))
	decoded = make([]byte, maxLen)

	n, err := base64.StdEncoding.Decode(decoded, content)
	if err == nil {
		return decoded[:n], nil
	}

	return content, nil
}

// Parse parses a MessagePack encoded binary object into an in-memory data structure.
func Parse(bytes []byte) (parsed MsgpObject, remaining []byte, err error) {
	parsed.Type = msgp.NextType(bytes)
	switch parsed.Type {
	case msgp.StrType:
		parsed.Value, bytes, err = msgp.ReadStringBytes(bytes)
	case msgp.BinType:
		parsed.Value, bytes, err = msgp.ReadBytesBytes(bytes, nil)
	case msgp.MapType:
		var size int
		size, _, bytes, err = msgp.ReadMapHeaderBytes(bytes)
		valueMap := MsgpMap{
			Order:  make([]string, size),
			Values: make(map[string]MsgpObject, size),
		}
		for i := 0; i < size; i++ {
			var key string
			key, bytes, err = msgp.ReadStringBytes(bytes)
			if err != nil {
				break
			}
			if _, ok := valueMap.Values[key]; ok {
				err = errors.New("Object has duplicate key")
				break
			}
			valueMap.Order[i] = key
			valueMap.Values[key], bytes, err = Parse(bytes)
			if err != nil {
				break
			}
		}
		parsed.Value = valueMap
	case msgp.ArrayType:
		var size int
		size, _, bytes, err = msgp.ReadArrayHeaderBytes(bytes)
		valueArray := make([]MsgpObject, size)
		for i := 0; i < size; i++ {
			valueArray[i], bytes, err = Parse(bytes)
			if err != nil {
				break
			}
		}
		parsed.Value = valueArray
	case msgp.Float32Type:
		parsed.Value, bytes, err = msgp.ReadFloat32Bytes(bytes)
	case msgp.Float64Type:
		parsed.Value, bytes, err = msgp.ReadFloat64Bytes(bytes)
	case msgp.BoolType:
		parsed.Value, bytes, err = msgp.ReadBoolBytes(bytes)
	case msgp.IntType:
		parsed.Value, bytes, err = msgp.ReadInt64Bytes(bytes)
	case msgp.UintType:
		parsed.Value, bytes, err = msgp.ReadUint64Bytes(bytes)
	case msgp.NilType:
		parsed.Value = nil
		bytes, err = msgp.ReadNilBytes(bytes)
	case msgp.Complex64Type:
		parsed.Value, bytes, err = msgp.ReadComplex64Bytes(bytes)
	case msgp.Complex128Type:
		parsed.Value, bytes, err = msgp.ReadComplex128Bytes(bytes)
	case msgp.TimeType:
		parsed.Value, bytes, err = msgp.ReadTimeBytes(bytes)
	default:
		err = errors.New("Invalid MessagePack type")
	}
	remaining = bytes
	return
}
