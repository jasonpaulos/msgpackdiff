package msgpackdiff

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/algorand/msgp/msgp"
)

// MsgpMap represents an ordered map of strings to MsgpObjects
type MsgpMap struct {
	Order  []string
	Values map[string]MsgpObject
}

// MsgpObject contains a parsed MessagePack object and its type.
type MsgpObject struct {
	Type  msgp.Type
	Value interface{}
}

// IsEmpty checks if the value of a MessagePack object is the zero value for its type.
func (mo MsgpObject) IsEmpty() (empty bool) {
	switch mo.Type {
	case msgp.InvalidType:
		empty = true
	case msgp.StrType:
		empty = len(mo.Value.(string)) == 0
	case msgp.BinType:
		empty = len(mo.Value.([]byte)) == 0
	case msgp.MapType:
		valueMap := mo.Value.(MsgpMap)
		empty = true
		for _, value := range valueMap.Values {
			if !value.IsEmpty() {
				empty = false
				break
			}
		}
	case msgp.ArrayType:
		valueArray := mo.Value.([]MsgpObject)
		empty = true
		for _, item := range valueArray {
			if !item.IsEmpty() {
				empty = false
				break
			}
		}
	case msgp.Float32Type:
		empty = mo.Value.(float32) == 0.0
	case msgp.Float64Type:
		empty = mo.Value.(float64) == 0.0
	case msgp.BoolType:
		empty = !mo.Value.(bool)
	case msgp.IntType:
		empty = mo.Value.(int64) == 0
	case msgp.UintType:
		empty = mo.Value.(uint64) == 0
	case msgp.NilType:
		empty = true
	case msgp.Complex64Type:
		empty = mo.Value.(complex64) == 0
	case msgp.Complex128Type:
		empty = mo.Value.(complex128) == 0
	case msgp.TimeType:
		empty = mo.Value.(time.Time).IsZero()
	}
	return
}

func (mo MsgpObject) String() string {
	var str strings.Builder
	mo.Print(&str, "", 0)
	return str.String()
}

const indentation string = "  "

func (mo MsgpObject) Print(w io.Writer, prefix string, indent int) {
	indentStr := strings.Repeat(indentation, indent)

	switch mo.Type {
	case msgp.MapType:
		valueMap := mo.Value.(MsgpMap)
		if len(valueMap.Order) == 0 {
			fmt.Fprint(w, "{}")
			return
		}
		fmt.Fprint(w, "{\n")
		for index, key := range valueMap.Order {
			value := valueMap.Values[key]
			fmt.Fprintf(w, "%s%s%s%s: ", prefix, indentStr, indentation, escapeString(key))
			value.Print(w, prefix, indent+1)
			if index+1 < len(valueMap.Order) {
				fmt.Fprint(w, ",\n")
			} else {
				fmt.Fprint(w, "\n")
			}
		}
		fmt.Fprintf(w, "%s%s}", prefix, indentStr)
	case msgp.ArrayType:
		valueArray := mo.Value.([]MsgpObject)
		if len(valueArray) == 0 {
			fmt.Fprint(w, "[]")
			return
		}
		fmt.Fprint(w, "[\n")
		for index, item := range valueArray {
			fmt.Fprintf(w, "%s%s%s", prefix, indentStr, indentation)
			item.Print(w, prefix, indent+1)
			if index+1 < len(valueArray) {
				fmt.Fprint(w, ",\n")
			} else {
				fmt.Fprint(w, "\n")
			}
		}
		fmt.Fprintf(w, "%s%s]", prefix, indentStr)
	case msgp.NilType:
		fmt.Fprint(w, "null")
	case msgp.StrType:
		fmt.Fprint(w, escapeString(mo.Value.(string)))
	default:
		fmt.Fprintf(w, "%v", mo.Value)
	}
}

func (mo MsgpObject) PrintDiff(w io.Writer, context int, diffs []difference, indent int) {
	indentStr := strings.Repeat(indentation, indent)

	switch mo.Type {
	case msgp.MapType:
		levelZero := false
		for _, diff := range diffs {
			if len(diff.path) == 0 {
				sign := getSign(diff.isDeletion)

				fmt.Fprintf(w, "%s%s", sign, indentStr)
				mo.Print(w, sign, indent)
				levelZero = true
			}
		}
		if levelZero {
			return
		}

		valueMap := mo.Value.(MsgpMap)

		fmt.Fprint(w, "{\n")
		for start := 0; start < len(diffs); {
			diff := diffs[start]
			layer := diff.path[0]

			if len(diff.path) == 1 {
				sign := getSign(diff.isDeletion)

				fmt.Fprintf(w, "%s%s%s%s: ", sign, indentStr, indentation, escapeString(layer.currentKey))
				diff.object.Print(w, sign, indent+1)

				if start < len(diffs) {
					fmt.Fprint(w, ",\n")
				} else {
					fmt.Fprint(w, "\n")
				}

				start++
				continue
			}

			end := start + 1
			for j := start + 1; j < len(diffs); j++ {
				if layer.object == diffs[j].path[0].object {
					end = j + 1
				}
			}

			subdiffs := diffs[start:end]
			for i := 0; i < len(subdiffs); i++ {
				subdiffs[i].path = subdiffs[i].path[1:]
			}

			fmt.Fprintf(w, " %s%s%s: ", indentStr, indentation, escapeString(layer.currentKey))
			value, ok := valueMap.Values[layer.currentKey]
			if ok {
				value.PrintDiff(w, context, subdiffs, indent+1)
			} else {
				diff.object.PrintDiff(w, context, subdiffs, indent+1)
			}

			if end < len(diffs) {
				fmt.Fprint(w, ",\n")
			} else {
				fmt.Fprint(w, "\n")
			}

			start = end
		}
		fmt.Fprintf(w, " %s}", indentStr)
	case msgp.ArrayType:
		// todo
		fallthrough
	default:
		for _, diff := range diffs {
			sign := getSign(diff.isDeletion)

			if len(diff.path) == 0 {
				fmt.Fprintf(w, "%s%s", sign, indentStr)
				mo.Print(w, sign, indent)
				continue
			}

			if len(diff.path) == 1 {
				mo.Print(w, sign, indent)
				continue
			}

			panic("Unexpected path")
		}
	}
}

func getSign(isDeletion bool) string {
	if isDeletion {
		return "-"
	}
	return "+"
}

var stringEscaper *strings.Replacer = strings.NewReplacer("\t", "\\t", "\n", "\\n", "\\", "\\\\")

func escapeString(str string) string {
	return fmt.Sprintf("\"%s\"", stringEscaper.Replace(str))
}
