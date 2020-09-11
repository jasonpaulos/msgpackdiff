package msgpackdiff

import (
	"bytes"
	"io"
	"time"

	"github.com/algorand/msgp/msgp"
)

type CompareResult struct {
	Equal    bool
	Reporter Reporter
	Objects  [2]MsgpObject
}

func (result CompareResult) PrintReport(w io.Writer) {
	if !result.Reporter.Brief && !result.Equal {
		result.Objects[0].PrintDiff(w, 3, result.Reporter.differences, 0, false)
	}
}

// Compare checks two MessagePack objects for equality. The first return value will be true if and
// only if the objects a and b are considered equivalent. If the second return value is a non-nil
// error, then the comparison could not be completed and the first return value should be ignored.
func Compare(a []byte, b []byte, brief, ignoreEmpty, ignoreOrder, flexibleTypes bool) (result CompareResult, err error) {
	result.Reporter.Brief = brief

	result.Objects[0], _, err = Parse(a)
	if err != nil {
		return
	}

	result.Objects[1], _, err = Parse(b)
	if err != nil {
		return
	}

	result.Equal = compareObjects(&result.Reporter, result.Objects[0], result.Objects[1], ignoreOrder, brief, ignoreEmpty, flexibleTypes)

	return
}

func compareNumbers(a MsgpObject, b MsgpObject) (equal bool) {
	// make a have the smaller type so that the switch statement only has to check larger types for b
	if a.Type > b.Type {
		a, b = b, a
	}

	switch {
	case a.Type == msgp.Float64Type && b.Type == msgp.Float32Type:
		floatA := a.Value.(float64)
		floatB := b.Value.(float32)
		equal = floatA == float64(floatB)
	case a.Type == msgp.Float64Type && b.Type == msgp.IntType:
		floatA := a.Value.(float64)
		intB := b.Value.(int64)
		equal = floatA == float64(intB)
	case a.Type == msgp.Float64Type && b.Type == msgp.UintType:
		floatA := a.Value.(float64)
		intB := b.Value.(uint64)
		equal = floatA == float64(intB)
	case a.Type == msgp.Float64Type && b.Type == msgp.Complex64Type:
		floatA := a.Value.(float64)
		complexB := b.Value.(complex64)
		equal = complex(floatA, 0) == complex128(complexB)
	case a.Type == msgp.Float64Type && b.Type == msgp.Complex128Type:
		floatA := a.Value.(float64)
		complexB := b.Value.(complex128)
		equal = complex(floatA, 0) == complexB
	case a.Type == msgp.Float32Type && b.Type == msgp.IntType:
		floatA := a.Value.(float32)
		intB := b.Value.(int64)
		equal = floatA == float32(intB)
	case a.Type == msgp.Float32Type && b.Type == msgp.UintType:
		floatA := a.Value.(float32)
		intB := b.Value.(uint64)
		equal = floatA == float32(intB)
	case a.Type == msgp.Float32Type && b.Type == msgp.Complex64Type:
		floatA := a.Value.(float32)
		complexB := b.Value.(complex64)
		equal = complex(floatA, 0) == complexB
	case a.Type == msgp.Float32Type && b.Type == msgp.Complex128Type:
		floatA := a.Value.(float32)
		complexB := b.Value.(complex128)
		equal = complex(float64(floatA), 0) == complexB
	case a.Type == msgp.IntType && b.Type == msgp.UintType:
		intA := a.Value.(int64)
		intB := b.Value.(uint64)
		equal = intA >= 0 && uint64(intA) == intB
	case a.Type == msgp.IntType && b.Type == msgp.Complex64Type:
		intA := a.Value.(int64)
		complexB := b.Value.(complex64)
		equal = complex(float32(intA), 0) == complexB
	case a.Type == msgp.IntType && b.Type == msgp.Complex128Type:
		intA := a.Value.(int64)
		complexB := b.Value.(complex128)
		equal = complex(float64(intA), 0) == complexB
	case a.Type == msgp.UintType && b.Type == msgp.Complex64Type:
		intA := a.Value.(uint64)
		complexB := b.Value.(complex64)
		equal = complex(float32(intA), 0) == complexB
	case a.Type == msgp.UintType && b.Type == msgp.Complex128Type:
		intA := a.Value.(uint64)
		complexB := b.Value.(complex128)
		equal = complex(float64(intA), 0) == complexB
	case a.Type == msgp.Complex64Type && b.Type == msgp.Complex128Type:
		complexA := a.Value.(complex64)
		complexB := b.Value.(complex128)
		equal = complex128(complexA) == complexB
	default:
		// the arguments are not numbers so they can't be equal
		equal = false
	}
	return
}

func compareObjects(reporter *Reporter, a MsgpObject, b MsgpObject, ignoreOrder, brief, ignoreEmpty, flexibleTypes bool) (equal bool) {
	if a.Type != b.Type {
		if flexibleTypes && compareNumbers(a, b) {
			equal = true
			return
		}

		reporter.LogDifference(a, b)

		equal = false
		return
	}

	switch a.Type {
	case msgp.StrType:
		strA := a.Value.(string)
		strB := b.Value.(string)
		equal = strA == strB
	case msgp.BinType:
		bytesA := a.Value.([]byte)
		bytesB := b.Value.([]byte)
		equal = bytes.Equal(bytesA, bytesB)
	case msgp.MapType:
		mapA := a.Value.(MsgpMap)
		mapB := b.Value.(MsgpMap)
		reporter.EnterMap(a)
		defer reporter.LeaveMap()
		if brief && !ignoreEmpty && len(mapA.Values) != len(mapB.Values) {
			equal = false
		} else if ignoreOrder {
			equal = true

			for index, key := range mapA.Order {
				valueA := mapA.Values[key]
				valueB, ok := mapB.Values[key]

				reporter.SetKey(index, key)

				if !ok {
					if ignoreEmpty && valueA.IsEmpty() {
						continue
					}

					reporter.LogDeletion(valueA)

					equal = false
					if brief {
						break
					}
				}

				valuesEqual := compareObjects(reporter, valueA, valueB, ignoreOrder, brief, ignoreEmpty, flexibleTypes)
				if !valuesEqual {
					equal = false
					if brief {
						break
					}
				}
			}

			for _, key := range mapB.Order {
				_, ok := mapA.Values[key]
				valueB := mapB.Values[key]

				if ok {
					continue
				}

				if ignoreEmpty && valueB.IsEmpty() {
					continue
				}

				reporter.SetKey(len(mapA.Order), key)
				reporter.LogAddition(valueB)

				equal = false

				if brief {
					break
				}
			}
		} else {
			lcs := longestCommonSubsequence(mapA.Order, mapB.Order)
			if brief && !ignoreEmpty && (len(lcs) != len(mapA.Order) || len(lcs) != len(mapB.Order)) {
				equal = false
			} else {
				equal = true

				indexA := 0
				indexB := 0
				for _, keyLCS := range lcs {
					inLCS := false

					for ; indexA < len(mapA.Order); indexA++ {
						keyA := mapA.Order[indexA]

						if keyA == keyLCS {
							indexA++
							inLCS = true
							break
						}

						if !ignoreEmpty || !mapA.Values[keyA].IsEmpty() {
							reporter.SetKey(indexA, keyA)
							reporter.LogDeletion(mapA.Values[keyA])

							equal = false
						}
					}

					if brief && !equal {
						break
					}

					for ; indexB < len(mapB.Order); indexB++ {
						keyB := mapB.Order[indexB]

						if keyB == keyLCS {
							indexB++
							break
						}

						if !ignoreEmpty || !mapB.Values[keyB].IsEmpty() {
							reporter.SetKey(indexA-1, keyB)
							reporter.LogAddition(mapB.Values[keyB])

							equal = false
						}
					}

					if brief && !equal {
						break
					}

					if inLCS {
						valueA := mapA.Values[keyLCS]
						valueB := mapB.Values[keyLCS]

						reporter.SetKey(indexA-1, keyLCS)

						valuesEqual := compareObjects(reporter, valueA, valueB, ignoreOrder, brief, ignoreEmpty, flexibleTypes)
						if !valuesEqual {
							equal = false
							if brief {
								break
							}
						}
					}
				}

				// report differences for keys that occur after the last LCS key
				if !brief || equal {
					for ; indexA < len(mapA.Order); indexA++ {
						keyA := mapA.Order[indexA]

						if !ignoreEmpty || !mapA.Values[keyA].IsEmpty() {
							reporter.SetKey(indexA, keyA)
							reporter.LogDeletion(mapA.Values[keyA])

							equal = false
						}
					}

					for ; indexB < len(mapB.Order); indexB++ {
						keyB := mapB.Order[indexB]

						if !ignoreEmpty || !mapB.Values[keyB].IsEmpty() {
							reporter.SetKey(indexA, keyB)
							reporter.LogAddition(mapB.Values[keyB])

							equal = false
						}
					}
				}
			}
		}
	case msgp.ArrayType:
		arrayA := a.Value.([]MsgpObject)
		arrayB := b.Value.([]MsgpObject)
		reporter.EnterArray(a)
		defer reporter.LeaveArray()
		if brief && len(arrayA) != len(arrayB) {
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
			for i := 0; i < len(*largeArray); i++ {
				reporter.SetIndex(i)

				if i >= len(*smallArray) {
					// can assume brief=false here
					equal = false

					// report missing items from smallArray
					if smallArray == &arrayA {
						reporter.LogAddition(arrayB[i])
					} else {
						reporter.LogDeletion(arrayA[i])
					}

					continue
				}

				itemA := arrayA[i]
				itemB := arrayB[i]
				itemsEqual := compareObjects(reporter, itemA, itemB, ignoreOrder, brief, ignoreEmpty, flexibleTypes)
				if !itemsEqual {
					equal = false
					if brief {
						break
					}
				}
			}
		}
	case msgp.Float32Type:
		floatA := a.Value.(float32)
		floatB := b.Value.(float32)
		equal = floatA == floatB
	case msgp.Float64Type:
		floatA := a.Value.(float64)
		floatB := b.Value.(float64)
		equal = floatA == floatB
	case msgp.BoolType:
		boolA := a.Value.(bool)
		boolB := b.Value.(bool)
		equal = boolA == boolB
	case msgp.IntType:
		intA := a.Value.(int64)
		intB := b.Value.(int64)
		equal = intA == intB
	case msgp.UintType:
		intA := a.Value.(uint64)
		intB := b.Value.(uint64)
		equal = intA == intB
	case msgp.NilType:
		equal = true
	case msgp.Complex64Type:
		complexA := a.Value.(complex64)
		complexB := b.Value.(complex64)
		equal = complexA == complexB
	case msgp.Complex128Type:
		complexA := a.Value.(complex128)
		complexB := b.Value.(complex128)
		equal = complexA == complexB
	case msgp.TimeType:
		timeA := a.Value.(time.Time)
		timeB := b.Value.(time.Time)
		equal = timeA.Equal(timeB)
	}

	if !equal && a.Type != msgp.MapType && a.Type != msgp.ArrayType {
		reporter.LogDifference(a, b)
	}

	return
}

// longestCommonSubsequence returns a solution to the longest subsequence problem for a and b.
// Based on https://en.wikipedia.org/wiki/Longest_common_subsequence_problem#Solution_for_two_sequences
func longestCommonSubsequence(a []string, b []string) []string {
	// make b the smaller slice
	if len(a) < len(b) {
		a, b = b, a
	}

	prevRow := make([][]string, len(b)+1)
	currentRow := make([][]string, len(b)+1)

	for _, itemA := range a {
		prevRow, currentRow = currentRow, prevRow

		for indexB, itemB := range b {
			if itemA == itemB {
				currentRow[indexB+1] = append(prevRow[indexB], itemB)
			} else {
				above := prevRow[indexB+1]
				left := currentRow[indexB]
				if len(above) > len(left) {
					currentRow[indexB+1] = above
				} else {
					currentRow[indexB+1] = left
				}
			}
		}
	}

	return currentRow[len(b)]
}
