package msgpackdiff

import (
	"encoding/base64"
	"testing"
)

type CompareTest struct {
	Name         string
	FirstObject  string
	SecondObject string
	Expected     bool
}

func runTestsWithOptions(t *testing.T, tests []CompareTest, stopOnFirstDifference, ignoreEmpty, ignoreOrder, flexibleTypes bool) {
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

			result, err := Compare(firstObject, secondObject, stopOnFirstDifference, ignoreEmpty, ignoreOrder, flexibleTypes)
			if err != nil {
				t.Fatalf("Unexpected error: %v\n", err)
			}

			if result != test.Expected {
				t.Fatalf("Wrong result: got %v, expected %v\n", result, test.Expected)
			}
		}
		t.Run(test.Name, runTest)
	}
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

	runTestsWithOptions(t, tests, false, true, true, false) // TODO: set ignoreOrder=false
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

	runTestsWithOptions(t, tests, false, false, true, true) // TODO: set ignoreOrder=false
}
