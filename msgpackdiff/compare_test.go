package msgpackdiff

import (
	"encoding/base64"
	"testing"

	"github.com/algorand/msgp/msgp"
)

type CompareTest struct {
	Name         string
	FirstObject  string
	SecondObject string
	Expected     bool
}

func runTestsWithOptions(t *testing.T, tests []CompareTest, ignoreEmpty, ignoreOrder, flexibleTypes bool) {
	for _, test := range tests {
		runTest := func(t *testing.T) {
			firstObject, err := base64.StdEncoding.DecodeString(test.FirstObject)
			if err != nil {
				t.Fatalf("Could not decode first object \"%v\": %v\n", test.FirstObject, err)
			}

			secondObject, err := base64.StdEncoding.DecodeString(test.SecondObject)
			if err != nil {
				t.Fatalf("Could not decode second object \"%v\": %v\n", test.SecondObject, err)
			}

			result, err := Compare(firstObject, secondObject, false, ignoreEmpty, ignoreOrder, flexibleTypes)
			if err != nil {
				t.Fatalf("Unexpected error: %v\n", err)
			}

			stopEarlyResult, err := Compare(firstObject, secondObject, true, ignoreEmpty, ignoreOrder, flexibleTypes)
			if err != nil {
				t.Fatalf("Unexpected stop early error: %v\n", err)
			}

			if result.Equal != stopEarlyResult.Equal {
				t.Error("Different result with stopOnFirstDifference=true\n")
			}

			if result.Equal != test.Expected {
				t.Fatalf("Wrong result: got %v, expected %v\n", result.Equal, test.Expected)
			}
		}
		t.Run(test.Name, runTest)
	}
}

func TestCompareDefault(t *testing.T) {
	tests := []CompareTest{
		{
			Name:         "null",
			FirstObject:  "wA==", // {}
			SecondObject: "wA==", // {}
			Expected:     true,
		},
		{
			Name:         "empty ints",
			FirstObject:  "AA==", // 0
			SecondObject: "AA==", // 0
			Expected:     true,
		},
		{
			Name:         "nonempty ints",
			FirstObject:  "Yw==", // 99
			SecondObject: "Yw==", // 99
			Expected:     true,
		},
		{
			Name:         "different ints",
			FirstObject:  "AA==", // 0
			SecondObject: "Yw==", // 99
			Expected:     false,
		},
		{
			Name:         "empty float32s",
			FirstObject:  "ygAAAAA=", // 0.0
			SecondObject: "ygAAAAA=", // 0.0
			Expected:     true,
		},
		{
			Name:         "nonempty float32s",
			FirstObject:  "yj/AAAA=", // 1.5
			SecondObject: "yj/AAAA=", // 1.5
			Expected:     true,
		},
		{
			Name:         "different float32s",
			FirstObject:  "ygAAAAA=", // 0.0
			SecondObject: "yj/AAAA=", // 1.5
			Expected:     false,
		},
		{
			Name:         "empty float64s",
			FirstObject:  "ywAAAAAAAAAA", // 0.0
			SecondObject: "ywAAAAAAAAAA", // 0.0
			Expected:     true,
		},
		{
			Name:         "nonempty float64s",
			FirstObject:  "y0Az/752yLQ5", // 19.999
			SecondObject: "y0Az/752yLQ5", // 19.999
			Expected:     true,
		},
		{
			Name:         "different float64s",
			FirstObject:  "ywAAAAAAAAAA", // 0.0
			SecondObject: "y0Az/752yLQ5", // 19.999
			Expected:     false,
		},
		{
			Name:         "int and float32",
			FirstObject:  "AA==",     // 0
			SecondObject: "ygAAAAA=", // 0.0
			Expected:     false,
		},
		{
			Name:         "int and float64",
			FirstObject:  "AA==",         // 0
			SecondObject: "ywAAAAAAAAAA", // 0.0
			Expected:     false,
		},
		{
			Name:         "float32 and float64",
			FirstObject:  "ygAAAAA=",     // 0.0
			SecondObject: "ywAAAAAAAAAA", // 0.0
			Expected:     false,
		},
		{
			Name:         "true",
			FirstObject:  "ww==", // true
			SecondObject: "ww==", // true
			Expected:     true,
		},
		{
			Name:         "false",
			FirstObject:  "wg==", // false
			SecondObject: "wg==", // false
			Expected:     true,
		},
		{
			Name:         "empty objects",
			FirstObject:  "gA==", // {}
			SecondObject: "gA==", // {}
			Expected:     true,
		},
		{
			Name:         "nonempty objects",
			FirstObject:  "gqJpZBukbmFtZaVKYXNvbg==", // {"id": 27, "name": "Jason"}
			SecondObject: "gqJpZBukbmFtZaVKYXNvbg==", // {"id": 27, "name": "Jason"}
			Expected:     true,
		},
		{
			Name:         "different order objects",
			FirstObject:  "gqJpZBukbmFtZaVKYXNvbg==", // {"id": 27, "name": "Jason"}
			SecondObject: "gqRuYW1lpUphc29uomlkGw==", // {"name": "Jason", "id": 27}
			Expected:     false,
		},
		{
			Name:         "empty arrays",
			FirstObject:  "kA==", // []
			SecondObject: "kA==", // []
			Expected:     true,
		},
		{
			Name:         "nonempty arrays",
			FirstObject:  "kgEC", // [1, 2]
			SecondObject: "kgEC", // [1, 2]
			Expected:     true,
		},
		{
			Name:         "different order arrays",
			FirstObject:  "kgEC", // [1, 2]
			SecondObject: "kgIB", // [2, 1]
			Expected:     false,
		},
		{
			Name:         "different size arrays",
			FirstObject:  "kgEC", // [1, 2]
			SecondObject: "kQI=", // [2]
			Expected:     false,
		},
		{
			Name:         "empty strings",
			FirstObject:  "oA==", // ""
			SecondObject: "oA==", // ""
			Expected:     true,
		},
		{
			Name:         "nonempty strings",
			FirstObject:  "pHRlc3Q=", // "test"
			SecondObject: "pHRlc3Q=", // "test"
			Expected:     true,
		},
		{
			Name:         "different strings",
			FirstObject:  "pHRlc3Q=", // "test"
			SecondObject: "pXRlc3Qy", // "test2"
			Expected:     false,
		},
		{
			Name:         "empty binary strings",
			FirstObject:  "xAA=", // base64()
			SecondObject: "xAA=", // base64()
			Expected:     true,
		},
		{
			Name:         "nonempty binary strings",
			FirstObject:  "xAR0ZXN0", // base64(dGVzdA==)
			SecondObject: "xAR0ZXN0", // base64(dGVzdA==)
			Expected:     true,
		},
		{
			Name:         "different binary strings",
			FirstObject:  "xAR0ZXN0",     // base64(dGVzdA==)
			SecondObject: "xAV0ZXN0Mg==", // base64(dGVzdDI=)
			Expected:     false,
		},
	}

	runTestsWithOptions(t, tests, false, false, false)
}

func TestCompareIgnoreEmpty(t *testing.T) {
	tests := []CompareTest{
		{
			Name:         "null",
			FirstObject:  "gA==",         // {}
			SecondObject: "gaR1c2VywA==", // {"user": null}
			Expected:     true,
		},
		{
			Name:         "false",
			FirstObject:  "gA==",         // {}
			SecondObject: "gaVlcnJvcsI=", // {"error": false}
			Expected:     true,
		},
		{
			Name:         "true",
			FirstObject:  "gA==",         // {}
			SecondObject: "gaVlcnJvcsM=", // {"error": true}
			Expected:     false,
		},
		{
			Name:         "empty int",
			FirstObject:  "gA==",     // {}
			SecondObject: "gaJpZAA=", // {"id": 0}
			Expected:     true,
		},
		{
			Name:         "nonempty int",
			FirstObject:  "gA==",     // {}
			SecondObject: "gaJpZAE=", // {"id": 1}
			Expected:     false,
		},
		{
			Name:         "empty float32",
			FirstObject:  "gA==",             // {}
			SecondObject: "gaRjb3N0ygAAAAA=", // {"cost": 0.0}
			Expected:     true,
		},
		{
			Name:         "nonempty float32",
			FirstObject:  "gA==",             // {}
			SecondObject: "gaRjb3N0yj8AAAA=", // {"cost": 0.5}
			Expected:     false,
		},
		{
			Name:         "empty float64",
			FirstObject:  "gA==",                         // {}
			SecondObject: "gahxdWFudGl0ecsAAAAAAAAAAA==", // {"quantity": 0.0}
			Expected:     true,
		},
		{
			Name:         "nonempty float64",
			FirstObject:  "gA==",                         // {}
			SecondObject: "gahxdWFudGl0ecs+5Pi1iONo8Q==", // {"quantity": 0.00001}
			Expected:     false,
		},
		{
			Name:         "empty string",
			FirstObject:  "gA==",         // {}
			SecondObject: "gaRuYW1loA==", // {"name": ""}
			Expected:     true,
		},
		{
			Name:         "nonempty string",
			FirstObject:  "gA==",             // {}
			SecondObject: "gaRuYW1lpUphc29u", // {"name": "Jason"}
			Expected:     false,
		},
		{
			Name:         "empty binary string",
			FirstObject:  "gA==",         // {}
			SecondObject: "gaNrZXnEAA==", // {"key": base64()}
			Expected:     true,
		},
		{
			Name:         "nonempty binary string",
			FirstObject:  "gA==",             // {}
			SecondObject: "gaNrZXnEBHRlc3Q=", // {"key": base64(dGVzdA==)}
			Expected:     false,
		},
		{
			Name:         "empty map",
			FirstObject:  "gA==",             // {}
			SecondObject: "gadvcHRpb25zgA==", // {"options": {}}
			Expected:     true,
		},
		{
			Name:         "nonempty map",
			FirstObject:  "gA==",                     // {}
			SecondObject: "gadvcHRpb25zgaVzdGFydMM=", // {"options": {"start": true}}
			Expected:     false,
		},
		{
			Name:         "empty array",
			FirstObject:  "gA==",             // {}
			SecondObject: "gadmcmllbmRzkA==", // {"friends": []}
			Expected:     true,
		},
		{
			Name:         "nonempty array",
			FirstObject:  "gA==",                 // {}
			SecondObject: "gadmcmllbmRzkaRKb2hu", // {"friends": ["John"]}
			Expected:     false,
		},
		{
			Name:         "nonempty array with empty item",
			FirstObject:  "gA==",             // {}
			SecondObject: "gadmcmllbmRzkaA=", // {"friends": [""]}
			Expected:     true,
		},
		{
			Name:         "nonempty array with empty items",
			FirstObject:  "gA==",                 // {}
			SecondObject: "gadmcmllbmRzlKAAwpA=", // {"friends": ["", 0, false, []]}
			Expected:     true,
		},
		{
			Name:         "nonempty array with some empty items",
			FirstObject:  "gA==",                 // {}
			SecondObject: "gadmcmllbmRzlKAAwpEB", // {"friends": ["", 0, false, [1]]}
			Expected:     false,
		},
		{
			Name:         "multiple empty",
			FirstObject:  "gA==",                     // {}
			SecondObject: "gqR1c2VywKdwYXltZW50wA==", // {"user": null, "payment": null}
			Expected:     true,
		},
		{
			Name:         "one empty, one nonempty",
			FirstObject:  "gA==",                         // {}
			SecondObject: "gqR1c2VywKdwYXltZW50pXZhbGlk", // {"user": null, "payment": "valid"}
			Expected:     false,
		},
		{
			Name:         "empty in both",
			FirstObject:  "gaR1c2VywA==",     // {"user": null}
			SecondObject: "gadwYXltZW50wA==", // {"payment": null}
			Expected:     true,
		},
		{
			Name:         "empty and nonempty in both",
			FirstObject:  "gqJpZAekdXNlcsA=",     // {"id": 7, "user": null}
			SecondObject: "gqJpZAencGF5bWVudMA=", // {"id": 7, "payment": null}
			Expected:     true,
		},
	}

	runTestsWithOptions(t, tests, true, false, false)
}

func TestCompareIgnoreOrderTypes(t *testing.T) {
	tests := []CompareTest{
		{
			Name:         "same order",
			FirstObject:  "gqJpZBukbmFtZaVKYXNvbg==", // {"id": 27, "name": "Jason"}
			SecondObject: "gqJpZBukbmFtZaVKYXNvbg==", // {"id": 27, "name": "Jason"}
			Expected:     true,
		},
		{
			Name:         "different order 2",
			FirstObject:  "gqJpZBukbmFtZaVKYXNvbg==", // {"id": 27, "name": "Jason"}
			SecondObject: "gqRuYW1lpUphc29uomlkGw==", // {"name": "Jason", "id": 27}
			Expected:     true,
		},
		{
			Name:         "different order 3",
			FirstObject:  "g6NvbmUBo3R3bwKldGhyZWUD", // {"one": 1, "two": 2, "three": 3}
			SecondObject: "g6V0aHJlZQOjb25lAaN0d28C", // {"three": 3, "one": 1, "two": 2}
			Expected:     true,
		},
		{
			Name:         "different order same objects",
			FirstObject:  "gqV0b2RheYOjZGF5FqVtb250aAakeWVhcs0H5Kh0b21vcnJvd4OjZGF5F6Vtb250aAakeWVhcs0H5A==", // {"today": {"day": 22, "month": 6, "year": 2020}, "tomorrow": {"day": 23, "month": 6, "year": 2020}}
			SecondObject: "gqh0b21vcnJvd4OlbW9udGgGo2RheRekeWVhcs0H5KV0b2RheYOjZGF5FqR5ZWFyzQfkpW1vbnRoBg==", // {"tomorrow": {"month": 6, "day": 23, "year": 2020}, "today": {"day": 22, "year": 2020, "month": 6}}
			Expected:     true,
		},
		{
			Name:         "different order different objects",
			FirstObject:  "gqV0b2RheYOjZGF5FqVtb250aAakeWVhcs0H5Kh0b21vcnJvd4OjZGF5F6Vtb250aAakeWVhcs0H5A==", // {"today": {"day": 22, "month": 6, "year": 2020}, "tomorrow": {"day": 23, "month": 6, "year": 2020}}
			SecondObject: "gqh0b21vcnJvd4OlbW9udGgGo2RheRekeWVhcs0H5aV0b2RheYOjZGF5FqR5ZWFyzQfkpW1vbnRoBg==", // {"tomorrow": {"month": 6, "day": 23, "year": 2021}, "today": {"day": 22, "year": 2020, "month": 6}}
			Expected:     false,
		},
		{
			Name:         "different order arrays",
			FirstObject:  "kgEC", // [1, 2]
			SecondObject: "kgIB", // [2, 1]
			Expected:     false,
		},
	}

	runTestsWithOptions(t, tests, false, true, false)
}

func TestCompareFlexibleTypes(t *testing.T) {
	tests := []CompareTest{
		{
			Name:         "int and int",
			FirstObject:  "gaNudW17", // {"num": 123}
			SecondObject: "gaNudW17", // {"num": 123}
			Expected:     true,
		},
		{
			Name:         "int and float32",
			FirstObject:  "gaNudW17",         // {"num": 123}
			SecondObject: "gaNudW3KQvYAAA==", // {"num": 123.0}
			Expected:     true,
		},
		{
			Name:         "int and float64",
			FirstObject:  "gaNudW17",             // {"num": 123}
			SecondObject: "gaNudW3LQF7AAAAAAAA=", // {"num": 123.0}
			Expected:     true,
		},
		{
			Name:         "float64 and float32",
			FirstObject:  "gaNudW3LQF7AAAAAAAA=", // {"num": 123.0}
			SecondObject: "gaNudW3KQvYAAA==",     // {"num": 123.0}
			Expected:     true,
		},
		{
			Name:         "float32 and different float64",
			FirstObject:  "gaNudW3KQvYAAA==",     // {"num": 123.0}
			SecondObject: "gaNudW3LQF7AAAQxveg=", // {"num": 123.000001}
			Expected:     false,
		},
	}

	runTestsWithOptions(t, tests, false, false, true)
}

func TestLCSStrings(t *testing.T) {
	type LCSTest struct {
		Name      string
		FirstSeq  []string
		SecondSeq []string
		Expected  []string
	}

	tests := []LCSTest{
		{
			Name:      "both empty",
			FirstSeq:  []string{},
			SecondSeq: []string{},
			Expected:  []string{},
		},
		{
			Name:      "first empty",
			FirstSeq:  []string{},
			SecondSeq: []string{"A"},
			Expected:  []string{},
		},
		{
			Name:      "second empty",
			FirstSeq:  []string{"A"},
			SecondSeq: []string{},
			Expected:  []string{},
		},
		{
			Name:      "disjoint",
			FirstSeq:  []string{"A"},
			SecondSeq: []string{"B", "Q"},
			Expected:  []string{},
		},
		{
			Name:      "one has prefix",
			FirstSeq:  []string{"A", "B"},
			SecondSeq: []string{"B"},
			Expected:  []string{"B"},
		},
		{
			Name:      "both have prefix",
			FirstSeq:  []string{"A", "C", "D"},
			SecondSeq: []string{"0", "B", "C", "D"},
			Expected:  []string{"C", "D"},
		},
		{
			Name:      "one has suffix",
			FirstSeq:  []string{"A", "B", "C"},
			SecondSeq: []string{"A"},
			Expected:  []string{"A"},
		},
		{
			Name:      "both have suffix",
			FirstSeq:  []string{"A", "1", "2", "3"},
			SecondSeq: []string{"A", "B", "C"},
			Expected:  []string{"A"},
		},
		{
			Name:      "one has infix",
			FirstSeq:  []string{"A", "X", "Z", "B", "C"},
			SecondSeq: []string{"A", "B", "C"},
			Expected:  []string{"A", "B", "C"},
		},
		{
			Name:      "both have infix",
			FirstSeq:  []string{"A", "1", "B", "2", "C"},
			SecondSeq: []string{"A", "a", "B", "b", "C"},
			Expected:  []string{"A", "B", "C"},
		},
		{
			Name:      "subset",
			FirstSeq:  []string{"A", "B", "C", "D"},
			SecondSeq: []string{"C"},
			Expected:  []string{"C"},
		},
		{
			Name:      "wikipedia example",
			FirstSeq:  []string{"A", "G", "C", "A", "T"},
			SecondSeq: []string{"G", "A", "C"},
			Expected:  []string{"G", "A"},
		},
	}

	slicesEqual := func(s1 []string, s2 []string) bool {
		if len(s1) != len(s2) {
			return false
		}

		for i := range s1 {
			if s1[i] != s2[i] {
				return false
			}
		}

		return true
	}

	for _, test := range tests {
		runTest := func(t *testing.T) {
			result := lcsStrings(test.FirstSeq, test.SecondSeq)
			resultFlipped := lcsStrings(test.SecondSeq, test.FirstSeq)

			if !slicesEqual(result, resultFlipped) {
				t.Fatalf("Result differs based on order of arguments: got %v and %v\n", result, resultFlipped)
			}

			if !slicesEqual(result, test.Expected) {
				t.Fatalf("Wrong result: got %v, expected %v\n", result, test.Expected)
			}
		}
		t.Run(test.Name, runTest)
	}
}

func TestLCSObjects(t *testing.T) {
	type LCSTest struct {
		Name      string
		FirstSeq  []MsgpObject
		SecondSeq []MsgpObject
		Expected  [][2]int
	}

	tests := []LCSTest{
		{
			Name:      "both empty",
			FirstSeq:  []MsgpObject{},
			SecondSeq: []MsgpObject{},
			Expected:  [][2]int{},
		},
		{
			Name:     "first empty",
			FirstSeq: []MsgpObject{},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				},
			},
			Expected: [][2]int{},
		},
		{
			Name: "second empty",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				},
			},
			SecondSeq: []MsgpObject{},
			Expected:  [][2]int{},
		},
		{
			Name: "disjoint",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "Q",
				},
			},
			Expected: [][2]int{},
		},
		{
			Name: "one has prefix",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "B",
				},
			},
			Expected: [][2]int{{1, 0}},
		},
		{
			Name: "both have prefix",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				}, {
					Type:  msgp.StrType,
					Value: "D",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "0",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				}, {
					Type:  msgp.StrType,
					Value: "D",
				},
			},
			Expected: [][2]int{{1, 2}, {2, 3}},
		},
		{
			Name: "one has suffix",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				},
			},
			Expected: [][2]int{{0, 0}},
		},
		{
			Name: "both have suffix",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "1",
				}, {
					Type:  msgp.StrType,
					Value: "2",
				}, {
					Type:  msgp.StrType,
					Value: "3",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				},
			},
			Expected: [][2]int{{0, 0}},
		},
		{
			Name: "one has infix",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "X",
				}, {
					Type:  msgp.StrType,
					Value: "Z",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				},
			},
			Expected: [][2]int{{0, 0}, {3, 1}, {4, 2}},
		},
		{
			Name: "both have infix",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "1",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "2",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "a",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "b",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				},
			},
			Expected: [][2]int{{0, 0}, {2, 2}, {4, 4}},
		},
		{
			Name: "subset",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "B",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				}, {
					Type:  msgp.StrType,
					Value: "D",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "C",
				},
			},
			Expected: [][2]int{{2, 0}},
		},
		{
			Name: "wikipedia example",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "G",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				}, {
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "T",
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.StrType,
					Value: "G",
				}, {
					Type:  msgp.StrType,
					Value: "A",
				}, {
					Type:  msgp.StrType,
					Value: "C",
				},
			},
			Expected: [][2]int{{1, 0}, {3, 1}},
		},
		{
			Name: "numbers",
			FirstSeq: []MsgpObject{
				{
					Type:  msgp.IntType,
					Value: int64(1),
				}, {
					Type:  msgp.IntType,
					Value: int64(2),
				},
			},
			SecondSeq: []MsgpObject{
				{
					Type:  msgp.IntType,
					Value: int64(1),
				}, {
					Type:  msgp.IntType,
					Value: int64(2),
				},
			},
			Expected: [][2]int{{0, 0}, {1, 1}},
		},
	}

	slicesEqual := func(s1 [][2]int, s2 [][2]int) bool {
		if len(s1) != len(s2) {
			return false
		}

		for i := range s1 {
			pair1 := s1[i]
			pair2 := s2[i]
			if pair1[0] != pair2[0] || pair1[1] != pair2[1] {
				return false
			}
		}

		return true
	}

	for _, test := range tests {
		runTest := func(t *testing.T) {
			result := lcsObjects(test.FirstSeq, test.SecondSeq)

			if !slicesEqual(result, test.Expected) {
				t.Fatalf("Wrong result: got %v, expected %v\n", result, test.Expected)
			}
		}
		t.Run(test.Name, runTest)
	}
}
