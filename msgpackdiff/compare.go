package msgpackdiff

import (
	"bytes"
	"io"
	"math"
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
		result.Objects[0].PrintDiff(w, context, result.Reporter.Differences, 0, false, true)
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

	objectsA := []MsgpObject{}

	for len(a) != 0 {
		var object MsgpObject
		object, a, err = Parse(a)
		if err != nil {
			return
		}
		objectsA = append(objectsA, object)
	}

	result.Objects[0] = MsgpObject{
		msgp.ArrayType,
		objectsA,
	}

	objectsB := []MsgpObject{}

	for len(b) != 0 {
		var object MsgpObject
		object, b, err = Parse(b)
		if err != nil {
			return
		}
		objectsB = append(objectsB, object)
	}

	result.Objects[1] = MsgpObject{
		msgp.ArrayType,
		objectsB,
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

					for ; indexA < len(mapA.Order); indexA++ {
						keyA := mapA.Order[indexA]

						if keyA == keyLCS {
							indexA++
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

			for _, member := range lcs {
				lcsIndexA := member.indexA
				lcsIndexB := member.indexB
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

				if options.Brief && !equal {
					break
				}

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

				if options.Brief && !equal {
					break
				}

				if len(member.diffs) > 0 {
					reporter.SetIndex(lcsIndexA)
					reporter.Accept(member.diffs)
					equal = false

					if options.Brief {
						break
					}
				}
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

type lcsMember struct {
	indexA int
	indexB int
	diffs  []Difference
}

// lcsObjects returns a solution to the longest subsequence problem for MsgpObject slices a and b.
// Based on https://en.wikipedia.org/wiki/Longest_common_subsequence_problem#Solution_for_two_sequences
func lcsObjects(a []MsgpObject, b []MsgpObject, options CompareOptions) []lcsMember {
	prevRow := make([][]lcsMember, len(b)+1)
	currentRow := make([][]lcsMember, len(b)+1)
	differences := make([][]Difference, len(b))

	for indexA, itemA := range a {
		prevRow, currentRow = currentRow, prevRow

		minDiffs := math.MaxInt32
		for indexB, itemB := range b {
			reporter := Reporter{
				Brief: options.Brief,
				// set Differences to an empty slice since we use nil as a special value below
				Differences: []Difference{},
			}
			if itemA.Type != itemB.Type {
				// items are different types so they can't be equal, don't even compare them
				if options.FlexibleTypes && compareNumbers(itemA, itemB) {
					// unless flexible types is enabled and the items are numbers
					differences[indexB] = []Difference{}
					minDiffs = 0
					continue
				}
				differences[indexB] = nil
				continue
			}
			isContainer := itemA.Type == msgp.ArrayType || itemA.Type == msgp.MapType
			equal := compareObjects(&reporter, itemA, itemB, options)
			if options.Brief || !isContainer {
				// if brief is enabled, then the diff count is meaningless
				// simiarly, if the items aren't containers but are different, ignore the diffs and
				// just mark them as different
				if equal {
					differences[indexB] = []Difference{}
					minDiffs = 0
				} else {
					differences[indexB] = nil
				}
				continue
			}

			differences[indexB] = reporter.Differences
			if len(differences[indexB]) < minDiffs {
				minDiffs = len(differences[indexB])
			}
		}

		for indexB := range b {
			// if differences[indexB] is nil, then the items are unquestionably different and diffs
			// don't apply
			// if len(differences[indexB]) <= minDiffs, then the items are relatively equal
			if differences[indexB] != nil && len(differences[indexB]) <= minDiffs {
				member := lcsMember{
					indexA: indexA,
					indexB: indexB,
					diffs:  differences[indexB],
				}
				currentRow[indexB+1] = append(prevRow[indexB], member)
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
