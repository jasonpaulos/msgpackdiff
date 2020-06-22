package msgpackdiff

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/algorand/msgp/msgp"
)

func TestGetBinary(t *testing.T) {
	type GetBinaryTest struct {
		Name     string
		Input    string
		Expected []byte
	}

	algoTxn, err := ioutil.ReadFile("../test/algo_txn_binary")
	if err != nil {
		t.Fatal(err)
	}

	tests := []GetBinaryTest{
		{
			Name:     "inline base64",
			Input:    "gaJwactACSH7VEQtGA==", // {"pi": 3.141592653589793}
			Expected: []byte{129, 162, 112, 105, 203, 64, 9, 33, 251, 84, 68, 45, 24},
		},
		{
			Name:     "binary file",
			Input:    "../test/algo_txn_binary",
			Expected: algoTxn,
		},
		{
			Name:     "base64 file",
			Input:    "../test/algo_txn_base64",
			Expected: algoTxn,
		},
	}

	for _, test := range tests {
		runTest := func(t *testing.T) {
			result, err := GetBinary(test.Input)
			if err != nil {
				t.Fatalf("Unexpected error: %v\n", err)
			}

			if !bytes.Equal(result, test.Expected) {
				t.Fatalf("Invalid binary: got %v, expected %v\n", result, test.Expected)
			}
		}
		t.Run(test.Name, runTest)
	}
}

func TestParse(t *testing.T) {
	type ParseTest struct {
		Name     string
		Input    string
		Expected MsgpObject
	}

	tests := []ParseTest{
		{
			Name:  "{}",
			Input: "gA==",
			Expected: MsgpObject{
				msgp.MapType,
				map[string]MsgpObject{},
			},
		},
		{
			Name:  "[]",
			Input: "kA==",
			Expected: MsgpObject{
				msgp.ArrayType,
				[]MsgpObject{},
			},
		},
		{
			Name:  "\"\"",
			Input: "oA==",
			Expected: MsgpObject{
				msgp.StrType,
				"",
			},
		},
		{
			Name:  "null",
			Input: "wA==",
			Expected: MsgpObject{
				msgp.NilType,
				nil,
			},
		},
		{
			Name:  "true",
			Input: "ww==",
			Expected: MsgpObject{
				msgp.BoolType,
				true,
			},
		},
		{
			Name:  "false",
			Input: "wg==",
			Expected: MsgpObject{
				msgp.BoolType,
				false,
			},
		},
		{
			Name:  "0",
			Input: "AA==",
			Expected: MsgpObject{
				msgp.IntType,
				int64(0),
			},
		},
		{
			Name:  "123456789",
			Input: "zgdbzRU=",
			Expected: MsgpObject{
				msgp.UintType,
				uint64(123456789),
			},
		},
		{
			Name:  "0.5",
			Input: "yj8AAAA=",
			Expected: MsgpObject{
				msgp.Float32Type,
				float32(0.5),
			},
		},
		{
			Name:  "0.99999999",
			Input: "yz/v///6oZxH",
			Expected: MsgpObject{
				msgp.Float64Type,
				float64(0.99999999),
			},
		},
		{
			Name:  "\"just_a_string\"",
			Input: "rWp1c3RfYV9zdHJpbmc=",
			Expected: MsgpObject{
				msgp.StrType,
				"just_a_string",
			},
		},
		{
			Name:  "[1,2,3,4,5]",
			Input: "lQECAwQF",
			Expected: MsgpObject{
				msgp.ArrayType,
				[]MsgpObject{
					{msgp.IntType, int64(1)},
					{msgp.IntType, int64(2)},
					{msgp.IntType, int64(3)},
					{msgp.IntType, int64(4)},
					{msgp.IntType, int64(5)},
				},
			},
		},
		{
			Name:  "[[1],[2],[\"three\"],4]",
			Input: "lJEBkQKRpXRocmVlBA==",
			Expected: MsgpObject{
				msgp.ArrayType,
				[]MsgpObject{
					{msgp.ArrayType, []MsgpObject{{msgp.IntType, int64(1)}}},
					{msgp.ArrayType, []MsgpObject{{msgp.IntType, int64(2)}}},
					{msgp.ArrayType, []MsgpObject{{msgp.StrType, "three"}}},
					{msgp.IntType, int64(4)},
				},
			},
		},
		{
			Name:  "{\"a\":1}",
			Input: "gaFhAQ==",
			Expected: MsgpObject{
				msgp.MapType,
				map[string]MsgpObject{
					"a": {msgp.IntType, int64(1)},
				},
			},
		},
		{
			Name:  "{\"longer_key\":2,\"null_key\":null}",
			Input: "gqpsb25nZXJfa2V5AqhudWxsX2tlecA=",
			Expected: MsgpObject{
				msgp.MapType,
				map[string]MsgpObject{
					"longer_key": {msgp.IntType, int64(2)},
					"null_key":   {msgp.NilType, nil},
				},
			},
		},
		{
			Name:  "{\"pi\":3.141592653589793}",
			Input: "gaJwactACSH7VEQtGA==",
			Expected: MsgpObject{
				msgp.MapType,
				map[string]MsgpObject{
					"pi": {msgp.Float64Type, float64(3.141592653589793)},
				},
			},
		},
		{
			Name:  "{\"32 bit float\":1.5}",
			Input: "gawzMiBiaXQgZmxvYXTKP8AAAA==",
			Expected: MsgpObject{
				msgp.MapType,
				map[string]MsgpObject{
					"32 bit float": {msgp.Float32Type, float32(1.5)},
				},
			},
		},
		{
			Name:  "{\"first_null_key\":null,\"second_null_key\":null}",
			Input: "gq5maXJzdF9udWxsX2tlecCvc2Vjb25kX251bGxfa2V5wA==",
			Expected: MsgpObject{
				msgp.MapType,
				map[string]MsgpObject{
					"first_null_key":  {msgp.NilType, nil},
					"second_null_key": {msgp.NilType, nil},
				},
			},
		},
		{
			Name:  "{\"txn\":{\"amt\":5000000,\"fee\":1000,\"fv\":6000000,\"gen\":\"mainnet-v1.0\",\"gh\":\"wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=\",\"lv\":6001000,\"note\":\"SGVsbG8gV29ybGQ=\",\"rcv\":\"GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A\",\"snd\":\"EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4\",\"type\":\"pay\"}}",
			Input: "gaN0eG6Ko2FtdM4ATEtAo2ZlZc0D6KJmds4AW42Ao2dlbqxtYWlubmV0LXYxLjCiZ2jZLHdHSEUyUHdkdmQ3UzEyQkw1RmFPUDIwRUdZZXNONzNrdGlDMXF6a2tpdDg9omx2zgBbkWikbm90ZbBTR1ZzYkc4Z1YyOXliR1E9o3Jjdtk6R0Q2NFlJWTNUV0dETUNOUFA1NTNEWlBQUjZMRFVTRlFPSUpWRkRQUFhXRUczRlZPSkNDREJCSFU1QaNzbmTZOkVXNjRHQzZGMjRNN05EU0M1UjNFUzRZVVZFM1pYWE5NQVJKSERDQ0NMSUhaVTZUQkVPQzdYUlNCRzSkdHlwZaNwYXk=",
			Expected: MsgpObject{
				msgp.MapType,
				map[string]MsgpObject{
					"txn": {msgp.MapType, map[string]MsgpObject{
						"amt":  {msgp.UintType, uint64(5000000)},
						"fee":  {msgp.UintType, uint64(1000)},
						"fv":   {msgp.UintType, uint64(6000000)},
						"gen":  {msgp.StrType, "mainnet-v1.0"},
						"gh":   {msgp.StrType, "wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8="},
						"lv":   {msgp.UintType, uint64(6001000)},
						"note": {msgp.StrType, "SGVsbG8gV29ybGQ="},
						"rcv":  {msgp.StrType, "GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A"},
						"snd":  {msgp.StrType, "EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4"},
						"type": {msgp.StrType, "pay"},
					}},
				},
			},
		},
	}

	for _, test := range tests {
		runTest := func(t *testing.T) {
			decoded, err := base64.StdEncoding.DecodeString(test.Input)
			if err != nil {
				t.Fatalf("Could not decode input \"%v\": %v\n", test.Input, err)
			}

			result, _, err := Parse(decoded)
			if err != nil {
				t.Fatalf("Unexpected error: %v\n", err)
			}

			if result.Type != test.Expected.Type {
				t.Fatalf("Wrong type: got %v, expected %v\n", result.Type, test.Expected.Type)
			}

			if !reflect.DeepEqual(result.Value, test.Expected.Value) {
				t.Fatalf("Objects unequal: got %+v, expected %+v\n", result.Value, test.Expected.Value)
			}
		}
		t.Run(test.Name, runTest)
	}
}
