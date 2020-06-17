package msgpackdiff

import (
	"bytes"
	"errors"
	"time"

	"github.com/algorand/msgp/msgp"
)

// Compare checks two MessagePack objects for equality. The first return value will be true if and
// only if the objects a and b are considered equivalent. If the second return value is a non-nil
// error, then the comparison could not be completed and the first return value should be ignored.
func Compare(a []byte, b []byte, stopOnFirstDifference, ignoreEmpty, ignoreOrder bool) (bool, error) {
	objA, _, err := Parse(a)
	if err != nil {
		return false, err
	}

	objB, _, err := Parse(b)
	if err != nil {
		return false, err
	}

	if !ignoreOrder {
		return false, errors.New("Strict order has not been implemented yet")
	}

	return compareObjects(objA, objB, stopOnFirstDifference, ignoreEmpty), nil
}

func compareObjects(a MsgpObject, b MsgpObject, stopOnFirstDifference, ignoreEmpty bool) (equal bool) {
	if a.Type != b.Type {
		equal = false
		return
	}

	switch a.Type {
	case msgp.StrType:
		strA := a.Object.(string)
		strB := b.Object.(string)
		equal = strA == strB
	case msgp.BinType:
		bytesA := a.Object.([]byte)
		bytesB := b.Object.([]byte)
		equal = bytes.Equal(bytesA, bytesB)
	case msgp.MapType:
		mapA := a.Object.(map[string]MsgpObject)
		mapB := b.Object.(map[string]MsgpObject)
		if stopOnFirstDifference && len(mapA) != len(mapB) {
			equal = false
		} else {
			allKeys := make(map[string]bool)
			for key := range mapA {
				allKeys[key] = true
			}
			for key := range mapB {
				allKeys[key] = true
			}

			equal = true
			for key := range allKeys {
				valueA, okA := mapA[key]
				valueB, okB := mapB[key]

				if !okA || !okB {
					equal = false
					if stopOnFirstDifference {
						break
					}
				}

				valuesEqual := compareObjects(valueA, valueB, stopOnFirstDifference, ignoreEmpty)
				if !valuesEqual {
					equal = false
					if stopOnFirstDifference {
						break
					}
				}
			}
		}
	case msgp.ArrayType:
		arrayA := a.Object.([]MsgpObject)
		arrayB := b.Object.([]MsgpObject)
		if stopOnFirstDifference && len(arrayA) != len(arrayB) {
			equal = false
		} else {
			var largeArray *[]MsgpObject
			var smallArray *[]MsgpObject
			if len(arrayA) >= len(arrayB) {
				largeArray = &arrayA
				smallArray = &arrayB
			} else {
				largeArray = &arrayB
				smallArray = &arrayA
			}

			equal = true
			for i := 0; i <= len(*largeArray); i++ {
				if i >= len(*smallArray) {
					// can assume stopOnFirstDifference=false here
					equal = false
					// TODO: add missing items from smallArray to report
					continue
				}

				itemA := arrayA[i]
				itemB := arrayB[i]
				itemsEqual := compareObjects(itemA, itemB, stopOnFirstDifference, ignoreEmpty)
				if !itemsEqual {
					equal = false
					if stopOnFirstDifference {
						break
					}
				}
			}
		}
	case msgp.Float32Type:
		floatA := a.Object.(float32)
		floatB := b.Object.(float32)
		equal = floatA == floatB
	case msgp.Float64Type:
		floatA := a.Object.(float64)
		floatB := b.Object.(float64)
		equal = floatA == floatB
	case msgp.BoolType:
		boolA := a.Object.(bool)
		boolB := b.Object.(bool)
		equal = boolA == boolB
	case msgp.IntType:
		intA := a.Object.(int64)
		intB := b.Object.(int64)
		equal = intA == intB
	case msgp.UintType:
		intA := a.Object.(uint64)
		intB := b.Object.(uint64)
		equal = intA == intB
	case msgp.NilType:
		equal = true
	case msgp.Complex64Type:
		complexA := a.Object.(complex64)
		complexB := b.Object.(complex64)
		equal = complexA == complexB
	case msgp.Complex128Type:
		complexA := a.Object.(complex128)
		complexB := b.Object.(complex128)
		equal = complexA == complexB
	case msgp.TimeType:
		timeA := a.Object.(time.Time)
		timeB := b.Object.(time.Time)
		equal = timeA.Equal(timeB)
	}
	return
}
