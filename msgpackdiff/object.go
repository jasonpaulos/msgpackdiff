package msgpackdiff

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/algorand/msgp/msgp"
	"github.com/ttacon/chalk"
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
	mo.Print(&str, "", 0, false)
	return str.String()
}

const indentation string = "  "

func (mo MsgpObject) Print(w io.Writer, prefix string, indent int, inline bool) {
	indentStr := strings.Repeat(indentation, indent)

	if !inline {
		fmt.Fprint(w, prefix)
		fmt.Fprint(w, indentStr)
		defer fmt.Fprintln(w)
	}

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
			value.Print(w, prefix, indent+1, true)
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
			item.Print(w, prefix, indent+1, true)
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
	case msgp.BinType:
		fmt.Fprintf(w, "base64(%s)", base64.StdEncoding.EncodeToString(mo.Value.([]byte)))
	default:
		fmt.Fprintf(w, "%v", mo.Value)
	}
}

func (mo MsgpObject) PrintDiff(w io.Writer, context int, diffs []Difference, indent int, inline bool) {
	indentStr := strings.Repeat(indentation, indent)
	levelZero := false
	embedded := false

	for _, diff := range diffs {
		if len(diff.Path) == 0 {
			sign := getSign(diff.Type)
			endSign := getSignEnd()

			diff.Object.Print(w, sign, indent, false)
			fmt.Fprint(w, endSign)
			levelZero = true
		} else {
			embedded = true
		}
	}

	if levelZero {
		if embedded {
			panic("Invalid diff layers")
		}
		return
	}

	if !inline {
		fmt.Fprint(w, " ")
		defer fmt.Fprintln(w)
	}

	switch mo.Type {
	case msgp.MapType:
		valueMap := mo.Value.(MsgpMap)

		fmt.Fprint(w, "{\n")
		lastContextIndex := 0
		for start := 0; start < len(diffs); {
			diff := diffs[start]
			layer := diff.Path[0]

			if layer.CurrentIndex-context > lastContextIndex {
				skipped := layer.CurrentIndex - context - lastContextIndex
				s := "s"
				if skipped == 1 {
					s = ""
				}
				fmt.Fprintf(w, " %s%s... %d skipped value%s\n", indentStr, indentation, skipped, s)
			}

			for i := context; i > 0; i-- {
				index := layer.CurrentIndex - i
				if index >= lastContextIndex {
					key := valueMap.Order[index]
					value := valueMap.Values[key]
					fmt.Fprintf(w, " %s%s%s: ", indentStr, indentation, escapeString(key))
					value.Print(w, " ", indent+1, true)
					fmt.Fprint(w, ",\n")
					lastContextIndex = index + 1
				}
			}

			nextLayerIndex := math.MaxInt32

			if len(diff.Path) == 1 {
				sign := getSign(diff.Type)
				endSign := getSignEnd()

				fmt.Fprintf(w, "%s%s%s%s: ", sign, indentStr, indentation, escapeString(layer.CurrentKey))
				diff.Object.Print(w, sign, indent+1, true)

				moreKeys := layer.CurrentIndex+1 < len(valueMap.Order)
				if diff.Type == Addition {
					moreKeys = layer.CurrentIndex < len(valueMap.Order) || start+1 < len(diffs)
				}

				if moreKeys {
					fmt.Fprintf(w, ",%s\n", endSign)
				} else {
					fmt.Fprintf(w, "%s\n", endSign)
				}

				start++

				if diff.Type == Deletion {
					lastContextIndex = layer.CurrentIndex + 1
				}

				if start < len(diffs) {
					nextLayerIndex = diffs[start].Path[0].CurrentIndex
				}
			} else {
				end := start + 1
				for j := start + 1; j < len(diffs); j++ {
					otherLayer := diffs[j].Path[0]
					if layer.CurrentIndex == otherLayer.CurrentIndex && layer.CurrentKey == otherLayer.CurrentKey {
						end = j + 1
					}
				}

				subdiffs := make([]Difference, end-start)
				copy(subdiffs, diffs[start:end])
				for i := 0; i < len(subdiffs); i++ {
					subdiffs[i].Path = subdiffs[i].Path[1:]
				}

				fmt.Fprintf(w, " %s%s%s: ", indentStr, indentation, escapeString(layer.CurrentKey))
				value, ok := valueMap.Values[layer.CurrentKey]
				if ok {
					value.PrintDiff(w, context, subdiffs, indent+1, true)
				} else {
					diff.Object.PrintDiff(w, context, subdiffs, indent+1, true)
				}

				if end < len(diffs) || layer.CurrentIndex+1 < len(valueMap.Order) {
					fmt.Fprint(w, ",\n")
				} else {
					fmt.Fprint(w, "\n")
				}

				start = end

				lastContextIndex = layer.CurrentIndex + 1

				if end < len(diffs) {
					nextLayerIndex = diffs[end].Path[0].CurrentIndex
				}
			}

			for i := 0; i <= context; i++ {
				index := layer.CurrentIndex + i
				if index >= nextLayerIndex {
					break
				}
				if index >= lastContextIndex && index < len(valueMap.Order) {
					key := valueMap.Order[index]
					value := valueMap.Values[key]
					fmt.Fprintf(w, " %s%s%s: ", indentStr, indentation, escapeString(key))
					value.Print(w, " ", indent+1, true)
					if index+1 < len(valueMap.Order) || start < len(diffs) {
						fmt.Fprint(w, ",")
					}
					fmt.Fprint(w, "\n")
					lastContextIndex = index + 1
				}
			}
		}

		if lastContextIndex < len(valueMap.Order) {
			skipped := len(valueMap.Order) - lastContextIndex
			s := "s"
			if skipped == 1 {
				s = ""
			}
			fmt.Fprintf(w, " %s%s... %d skipped value%s\n", indentStr, indentation, skipped, s)
		}

		fmt.Fprintf(w, " %s}", indentStr)
	case msgp.ArrayType:
		valueArray := mo.Value.([]MsgpObject)

		fmt.Fprint(w, "[\n")
		lastContextIndex := 0
		for start := 0; start < len(diffs); {
			diff := diffs[start]
			layer := diff.Path[0]

			if layer.CurrentIndex-context > lastContextIndex {
				skipped := layer.CurrentIndex - context - lastContextIndex
				s := "s"
				if skipped == 1 {
					s = ""
				}
				fmt.Fprintf(w, " %s%s... %d skipped value%s\n", indentStr, indentation, skipped, s)
			}

			for i := context; i > 0; i-- {
				index := layer.CurrentIndex - i
				if index >= lastContextIndex {
					value := valueArray[index]
					fmt.Fprintf(w, " %s%s", indentStr, indentation)
					value.Print(w, " ", indent+1, true)
					fmt.Fprint(w, ",\n")
					lastContextIndex = index + 1
				}
			}

			nextLayerIndex := math.MaxInt32

			if len(diff.Path) == 1 {
				sign := getSign(diff.Type)
				endSign := getSignEnd()

				fmt.Fprintf(w, "%s%s%s", sign, indentStr, indentation)
				diff.Object.Print(w, sign, indent+1, true)

				moreElements := layer.CurrentIndex+1 < len(valueArray)
				if diff.Type == Addition {
					moreElements = layer.CurrentIndex < len(valueArray) || start+1 < len(diffs)
				}

				if moreElements {
					fmt.Fprintf(w, ",%s\n", endSign)
				} else {
					fmt.Fprintf(w, "%s\n", endSign)
				}

				start++

				if diff.Type == Deletion {
					lastContextIndex = layer.CurrentIndex + 1
				}

				if start < len(diffs) {
					nextLayerIndex = diffs[start].Path[0].CurrentIndex
				}
			} else {
				end := start + 1
				for j := start + 1; j < len(diffs); j++ {
					if layer.Object == diffs[j].Path[0].Object {
						end = j + 1
					}
				}

				subdiffs := make([]Difference, end-start)
				copy(subdiffs, diffs[start:end])
				for i := 0; i < len(subdiffs); i++ {
					subdiffs[i].Path = subdiffs[i].Path[1:]
				}

				fmt.Fprintf(w, " %s%s", indentStr, indentation)
				if layer.CurrentIndex < len(valueArray) {
					valueArray[layer.CurrentIndex].PrintDiff(w, context, subdiffs, indent+1, true)
				} else {
					diff.Object.PrintDiff(w, context, subdiffs, indent+1, true)
				}

				if end < len(diffs) || layer.CurrentIndex+1 < len(valueArray) {
					fmt.Fprint(w, ",\n")
				} else {
					fmt.Fprint(w, "\n")
				}

				start = end

				lastContextIndex = layer.CurrentIndex + 1

				if end < len(diffs) {
					nextLayerIndex = diffs[end].Path[0].CurrentIndex
				}
			}

			for i := 0; i <= context; i++ {
				index := layer.CurrentIndex + i
				if index >= nextLayerIndex {
					break
				}
				if index >= lastContextIndex && index < len(valueArray) {
					value := valueArray[index]
					fmt.Fprintf(w, " %s%s", indentStr, indentation)
					value.Print(w, " ", indent+1, true)
					if index+1 < len(valueArray) || start < len(diffs) {
						fmt.Fprint(w, ",")
					}
					fmt.Fprint(w, "\n")
					lastContextIndex = index + 1
				}
			}
		}

		if lastContextIndex < len(valueArray) {
			skipped := len(valueArray) - lastContextIndex
			s := "s"
			if skipped == 1 {
				s = ""
			}
			fmt.Fprintf(w, " %s%s... %d skipped value%s\n", indentStr, indentation, skipped, s)
		}

		fmt.Fprintf(w, " %s]", indentStr)
	default:
		panic("Unexpected path")
	}
}

func getSign(diffType DifferenceType) string {
	if diffType == Deletion {
		return chalk.Red.String() + "-"
	}
	return chalk.Green.String() + "+"
}

func getSignEnd() string {
	return chalk.ResetColor.String()
}

func escapeString(str string) string {
	return strconv.QuoteToASCII(str)
}
