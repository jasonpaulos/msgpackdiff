package msgpackdiff

import "testing"

func TestReporter(t *testing.T) {
	a, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGEB") // {"level":1,"data":{"level":2,"data":1}}
	b, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGEC") // {"level":1,"data":{"level":2,"data":2}}

	result, _ := Compare(a, b, false, false, false, false)

	if result {
		t.Error("Wrong result")
	}
}
