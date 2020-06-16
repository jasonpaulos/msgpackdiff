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
func Parse(bytes []byte) (obj MsgpObject, err error) {
	switch msgp.NextType(bytes) {
	case msgp.StrType:
		// TODO
	case msgp.BinType:
		// TODO
	case msgp.MapType:
		// TODO
	case msgp.ArrayType:
		// TODO
	case msgp.Float32Type:
		//todo
	case msgp.Float64Type:
		// TODO
	case msgp.BoolType:
		// TODO
	case msgp.IntType:
		// TODO
	case msgp.UintType:
		// TODO
	case msgp.NilType:
		// TODO
	case msgp.Complex64Type:
		// TODO
	case msgp.Complex128Type:
		// TODO
	case msgp.TimeType:
		// TODO
	default:
		err = errors.New("Invalid MessagePack type")
	}
	return
}
