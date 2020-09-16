package msgpackdiff

import (
	"bytes"
	"io"
	"time"

	"github.com/algorand/msgp/msgp"
)

// CompareResult is the result of a comparison between two MsgpObjects.
type CompareResult struct {
	// If the objects are determined to be equal, this will be true true. Otherwise, false.
	Equal    bool
	Reporter Reporter
	// The two objects being compared.
	Objects [2]MsgpObject
}

// PrintReport prints a difference report of the CompareResult object to the io.Writer w.
func (result CompareResult) PrintReport(w io.Writer, context int) {
	if !result.Reporter.Brief && !result.Equal {
		result.Objects[0].PrintDiff(w, context, result.Reporter.Differences, 0, false)
	}
}

// CompareOptions are the options used in a call to Compare.
type CompareOptions struct {
	// Causes the comparison to exit as soon as a difference is detected and disables reporting the
	// comparison when true.
	Brief bool
	// Treats missing fields as empty objects for comparison when true.
	IgnoreEmpty bool
	// Ignores ordering of object keys for comparison when true.
	IgnoreOrder bool
	// Compares all numerical values regardless of their type when true. Some precision may be lost.
	FlexibleTypes bool
}

// Compare checks two MessagePack objects for equality. The first return value will be true if and
// only if the objects a and b are considered equivalent. If the second return value is a non-nil
// error, then the comparison could not be completed and the first return value should be ignored.
func Compare(a []byte, b []byte, options CompareOptions) (result CompareResult, err error) {
	result.Reporter.Brief = options.Brief

	result.Objects[0], _, err = Parse(a)
	if err != nil {
		return
	}

	result.Objects[1], _, err = Parse(b)
	if err != nil {
		return
	}

	result.Equal = compareObjects(&result.Reporter, result.Objects[0], result.Objects[1], options)

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

func compareObjects(reporter *Reporter, a MsgpObject, b MsgpObject, options CompareOptions) (equal bool) {
	if a.Type != b.Type {
		if options.FlexibleTypes && compareNumbers(a, b) {
			equal = true
			return
		}

		reporter.LogChange(a, b)

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
		if options.Brief && !options.IgnoreEmpty && len(mapA.Values) != len(mapB.Values) {
			equal = false
		} else if options.IgnoreOrder {
			equal = true

			for index, key := range mapA.Order {
				valueA := mapA.Values[key]
				valueB, ok := mapB.Values[key]

				reporter.SetKey(index, key)

				if !ok {
					if options.IgnoreEmpty && valueA.IsEmpty() {
						continue
					}

					reporter.LogDeletion(valueA)

					equal = false
					if options.Brief {
						break
					}
					continue
				}

				valuesEqual := compareObjects(reporter, valueA, valueB, options)
				if !valuesEqual {
					equal = false
					if options.Brief {
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

				if options.IgnoreEmpty && valueB.IsEmpty() {
					continue
				}

				reporter.SetKey(len(mapA.Order), key)
				reporter.LogAddition(valueB)

				equal = false

				if options.Brief {
					break
				}
			}
		} else {
			lcs := lcsStrings(mapA.Order, mapB.Order)
			if options.Brief && !options.IgnoreEmpty && (len(lcs) != len(mapA.Order) || len(lcs) != len(mapB.Order)) {
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

						if !options.IgnoreEmpty || !mapA.Values[keyA].IsEmpty() {
							reporter.SetKey(indexA, keyA)
							reporter.LogDeletion(mapA.Values[keyA])

							equal = false
						}
					}

					if options.Brief && !equal {
						break
					}

					for ; indexB < len(mapB.Order); indexB++ {
						keyB := mapB.Order[indexB]

						if keyB == keyLCS {
							indexB++
							break
						}

						if !options.IgnoreEmpty || !mapB.Values[keyB].IsEmpty() {
							reporter.SetKey(indexA-1, keyB)
							reporter.LogAddition(mapB.Values[keyB])

							equal = false
						}
					}

					if options.Brief && !equal {
						break
					}

					if inLCS {
						valueA := mapA.Values[keyLCS]
						valueB := mapB.Values[keyLCS]

						reporter.SetKey(indexA-1, keyLCS)

						valuesEqual := compareObjects(reporter, valueA, valueB, options)
						if !valuesEqual {
							equal = false
							if options.Brief {
								break
							}
						}
					}
				}

				// report differences for keys that occur after the last LCS key
				if !options.Brief || equal {
					for ; indexA < len(mapA.Order); indexA++ {
						keyA := mapA.Order[indexA]

						if !options.IgnoreEmpty || !mapA.Values[keyA].IsEmpty() {
							reporter.SetKey(indexA, keyA)
							reporter.LogDeletion(mapA.Values[keyA])

							equal = false
						}
					}

					for ; indexB < len(mapB.Order); indexB++ {
						keyB := mapB.Order[indexB]

						if !options.IgnoreEmpty || !mapB.Values[keyB].IsEmpty() {
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
		if options.Brief && len(arrayA) != len(arrayB) {
			equal = false
		} else {
			equal = true
			lcs := lcsObjects(arrayA, arrayB, options)

			indexA := 0
			indexB := 0

			for _, indexPair := range lcs {
				lcsIndexA := indexPair[0]
				lcsIndexB := indexPair[1]
				deleted := false

				for ; indexA < lcsIndexA; indexA++ {
					value := arrayA[indexA]
					if !options.IgnoreEmpty || !value.IsEmpty() {
						reporter.SetIndex(indexA)
						reporter.LogDeletion(value)
						equal = false
						deleted = true
					}
				}
				indexA++

				if deleted {
					reporter.SetIndex(lcsIndexA - 1)
				} else {
					reporter.SetIndex(lcsIndexA)
				}
				for ; indexB < lcsIndexB; indexB++ {
					value := arrayB[indexB]
					if !options.IgnoreEmpty || !value.IsEmpty() {
						reporter.LogAddition(value)
						equal = false
					}
				}
				indexB++
			}

			// report differences for keys that occur after the last LCS index
			if !options.Brief || equal {
				for ; indexA < len(arrayA); indexA++ {
					value := arrayA[indexA]

					if !options.IgnoreEmpty || !value.IsEmpty() {
						reporter.SetIndex(indexA)
						reporter.LogDeletion(value)

						equal = false
					}
				}

				for ; indexB < len(arrayB); indexB++ {
					value := arrayB[indexB]

					if !options.IgnoreEmpty || !value.IsEmpty() {
						reporter.SetIndex(indexA)
						reporter.LogAddition(value)

						equal = false
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
		reporter.LogChange(a, b)
	}

	return
}

// lcsStrings returns a solution to the longest subsequence problem for string slices a and b.
// Based on https://en.wikipedia.org/wiki/Longest_common_subsequence_problem#Solution_for_two_sequences
func lcsStrings(a []string, b []string) []string {
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

// lcsObjects returns a solution to the longest subsequence problem for MsgpObject slices a and b.
// Based on https://en.wikipedia.org/wiki/Longest_common_subsequence_problem#Solution_for_two_sequences
func lcsObjects(a []MsgpObject, b []MsgpObject, options CompareOptions) [][2]int {
	options.Brief = true
	prevRow := make([][][2]int, len(b)+1)
	currentRow := make([][][2]int, len(b)+1)

	reporter := Reporter{Brief: true}

	for indexA, itemA := range a {
		prevRow, currentRow = currentRow, prevRow

		for indexB, itemB := range b {
			if compareObjects(&reporter, itemA, itemB, options) {
				currentRow[indexB+1] = append(prevRow[indexB], [2]int{indexA, indexB})
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
