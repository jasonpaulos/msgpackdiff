package msgpackdiff

import "github.com/algorand/msgp/msgp"

// MsgpObject contains a parsed MessagePack object and its type.
type MsgpObject struct {
	Type   msgp.Type
	Object interface{}
}
