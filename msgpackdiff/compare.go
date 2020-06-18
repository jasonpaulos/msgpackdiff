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
func Compare(a []byte, b []byte, stopOnFirstDifference, ignoreEmpty, ignoreOrder, flexibleTypes bool) (bool, error) {
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

	return compareObjects(objA, objB, stopOnFirstDifference, ignoreEmpty, flexibleTypes), nil
}

func compareNumbers(a MsgpObject, b MsgpObject) (equal bool) {
	// make a have the smaller type so that the switch statement only has to check larger types for b
	if a.Type > b.Type {
		a, b = b, a
	}

	switch {
	case a.Type == msgp.Float64Type && b.Type == msgp.Float32Type:
		floatA := a.Object.(float64)
		floatB := b.Object.(float32)
		equal = floatA == float64(floatB)
	case a.Type == msgp.Float64Type && b.Type == msgp.IntType:
		floatA := a.Object.(float64)
		intB := b.Object.(int64)
		equal = floatA == float64(intB)
	case a.Type == msgp.Float64Type && b.Type == msgp.UintType:
		floatA := a.Object.(float64)
		intB := b.Object.(uint64)
		equal = floatA == float64(intB)
	case a.Type == msgp.Float64Type && b.Type == msgp.Complex64Type:
		floatA := a.Object.(float64)
		complexB := b.Object.(complex64)
		equal = complex(floatA, 0) == complex128(complexB)
	case a.Type == msgp.Float64Type && b.Type == msgp.Complex128Type:
		floatA := a.Object.(float64)
		complexB := b.Object.(complex128)
		equal = complex(floatA, 0) == complexB
	case a.Type == msgp.Float32Type && b.Type == msgp.IntType:
		floatA := a.Object.(float32)
		intB := b.Object.(int64)
		equal = floatA == float32(intB)
	case a.Type == msgp.Float32Type && b.Type == msgp.UintType:
		floatA := a.Object.(float32)
		intB := b.Object.(uint64)
		equal = floatA == float32(intB)
	case a.Type == msgp.Float32Type && b.Type == msgp.Complex64Type:
		floatA := a.Object.(float32)
		complexB := b.Object.(complex64)
		equal = complex(floatA, 0) == complexB
	case a.Type == msgp.Float32Type && b.Type == msgp.Complex128Type:
		floatA := a.Object.(float32)
		complexB := b.Object.(complex128)
		equal = complex(float64(floatA), 0) == complexB
	case a.Type == msgp.IntType && b.Type == msgp.UintType:
		intA := a.Object.(int64)
		intB := b.Object.(uint64)
		equal = intA >= 0 && uint64(intA) == intB
	case a.Type == msgp.IntType && b.Type == msgp.Complex64Type:
		intA := a.Object.(int64)
		complexB := b.Object.(complex64)
		equal = complex(float32(intA), 0) == complexB
	case a.Type == msgp.IntType && b.Type == msgp.Complex128Type:
		intA := a.Object.(int64)
		complexB := b.Object.(complex128)
		equal = complex(float64(intA), 0) == complexB
	case a.Type == msgp.UintType && b.Type == msgp.Complex64Type:
		intA := a.Object.(uint64)
		complexB := b.Object.(complex64)
		equal = complex(float32(intA), 0) == complexB
	case a.Type == msgp.UintType && b.Type == msgp.Complex128Type:
		intA := a.Object.(uint64)
		complexB := b.Object.(complex128)
		equal = complex(float64(intA), 0) == complexB
	case a.Type == msgp.Complex64Type && b.Type == msgp.Complex128Type:
		complexA := a.Object.(complex64)
		complexB := b.Object.(complex128)
		equal = complex128(complexA) == complexB
	default:
		// the arguments are not numbers so they can't be equal
		equal = false
	}
	return
}

func compareObjects(a MsgpObject, b MsgpObject, stopOnFirstDifference, ignoreEmpty, flexibleTypes bool) (equal bool) {
	if a.Type != b.Type {
		if flexibleTypes && compareNumbers(a, b) {
			equal = true
			return
		}

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
		if stopOnFirstDifference && !ignoreEmpty && len(mapA) != len(mapB) {
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
					if ignoreEmpty && ((okA && valueA.IsEmpty()) || (okB && valueB.IsEmpty())) {
						// one map does not have an object for this field, but the other map has an
						// empty object for the field, so they are treated as equal with ignoreEmpty
						continue
					}
					equal = false
					if stopOnFirstDifference {
						break
					}
				}

				valuesEqual := compareObjects(valueA, valueB, stopOnFirstDifference, ignoreEmpty, flexibleTypes)
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
				itemsEqual := compareObjects(itemA, itemB, stopOnFirstDifference, ignoreEmpty, flexibleTypes)
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
