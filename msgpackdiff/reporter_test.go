package msgpackdiff

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ttacon/chalk"
)

func TestIntLevel0(t *testing.T) {
	a, _ := GetBinary("AQ==") // 1
	b, _ := GetBinary("Ag==") // 2

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(`%s-1
%s%s+2
%s`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestIntLevel1(t *testing.T) {
	a, _ := GetBinary("gqVsZXZlbAGkZGF0YQE=") // {"level":1,"data":1}
	b, _ := GetBinary("gqVsZXZlbAGkZGF0YQI=") // {"level":1,"data":2}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   "level": 1,
%s-  "data": 1,%s
%s+  "data": 2,%s
 }
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestIntLevel2(t *testing.T) {
	a, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGEB") // {"level":1,"data":{"level":2,"data":1}}
	b, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGEC") // {"level":1,"data":{"level":2,"data":2}}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   "level": 1,
   "data": {
     "level": 2,
%s-    "data": 1,%s
%s+    "data": 2,%s
   }
 }
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestIntLevel3(t *testing.T) {
	a, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGGCpWxldmVsA6RkYXRhAQ==") // {"level":1,"data":{"level":2,"data":{"level":3,"data":1}}}
	b, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGGCpWxldmVsA6RkYXRhAg==") // {"level":1,"data":{"level":2,"data":{"level":3,"data":2}}}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   "level": 1,
   "data": {
     "level": 2,
     "data": {
       "level": 3,
%s-      "data": 1,%s
%s+      "data": 2,%s
     }
   }
 }
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestObjectDeletion(t *testing.T) {
	a, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGGDpWxldmVsA6RkYXRhAaNlbmTD") // {"level":1,"data":{"level":2,"data":{"level":3,"data":1,"end":true}}}
	b, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGGCpWxldmVsA6NlbmTD")         // {"level":1,"data":{"level":2,"data":{"level":3,"end":true}}}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   "level": 1,
   "data": {
     "level": 2,
     "data": {
       "level": 3,
%s-      "data": 1,%s
       "end": true,
     }
   }
 }
`, chalk.Red.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestObjectAddition(t *testing.T) {
	a, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGGCpWxldmVsA6NlbmTD")         // {"level":1,"data":{"level":2,"data":{"level":3,"end":true}}}
	b, _ := GetBinary("gqVsZXZlbAGkZGF0YYKlbGV2ZWwCpGRhdGGDpWxldmVsA6RkYXRhAaNlbmTD") // {"level":1,"data":{"level":2,"data":{"level":3,"data":1,"end":true}}}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   "level": 1,
   "data": {
     "level": 2,
     "data": {
       "level": 3,
%s+      "data": 1,%s
       "end": true,
     }
   }
 }
`, chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestObjectContextSingle(t *testing.T) {
	a, _ := GetBinary("iqFhAaFiAqFjA6FkBKFlBaFmBqFnB6FoCKFpCaFqCg==") // {"a":1,"b":2,"c":3,"d":4,"e":5,"f":6,"g":7,"h":8,"i":9,"j":10}
	b, _ := GetBinary("iqFhAaFiAqFjA6FkBKFlMqFmBqFnB6FoCKFpCaFqCg==") // {"a":1,"b":2,"c":3,"d":4,"e":50,"f":6,"g":7,"h":8,"i":9,"j":10}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   ...
   "b": 2,
   "c": 3,
   "d": 4,
%s-  "e": 5,%s
%s+  "e": 50,%s
   "f": 6,
   "g": 7,
   "h": 8,
   ...
 }
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestObjectContextOverlap(t *testing.T) {
	a, _ := GetBinary("jKFhAaFiAqFjA6FkBKFlBaFmBqFnB6FoCKFpCaFqCqFrC6FsDA==") // {"a":1,"b":2,"c":3,"d":4,"e":5,"f":6,"g":7,"h":8,"i":9,"j":10,"k":11,"l":12}
	b, _ := GetBinary("jKFhAaFiAqFjA6FkBKFlMqFmBqFnRqFoCKFpCaFqCqFrC6FsDA==") // {"a":1,"b":2,"c":3,"d":4,"e":50,"f":6,"g":70,"h":8,"i":9,"j":10,"k":11,"l":12}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   ...
   "b": 2,
   "c": 3,
   "d": 4,
%s-  "e": 5,%s
%s+  "e": 50,%s
   "f": 6,
%s-  "g": 7,%s
%s+  "g": 70,%s
   "h": 8,
   "i": 9,
   "j": 10,
   ...
 }
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String(), chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestObjectContextAdjacent(t *testing.T) {
	a, _ := GetBinary("3gAQoWEBoWICoWMDoWQEoWUFoWYGoWcHoWgIoWkJoWoKoWsLoWwMoW0NoW4OoW8PoXAQ") // {"a":1,"b":2,"c":3,"d":4,"e":5,"f":6,"g":7,"h":8,"i":9,"j":10,"k":11,"l":12,"m":13,"n":14,"o":15,"p":16}
	b, _ := GetBinary("3gAQoWEBoWICoWMDoWQEoWUyoWYGoWcHoWgIoWkJoWoKoWsLoWx4oW0NoW4OoW8PoXAQ") // {"a":1,"b":2,"c":3,"d":4,"e":50,"f":6,"g":7,"h":8,"i":9,"j":10,"k":11,"l":120,"m":13,"n":14,"o":15,"p":16}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   ...
   "b": 2,
   "c": 3,
   "d": 4,
%s-  "e": 5,%s
%s+  "e": 50,%s
   "f": 6,
   "g": 7,
   "h": 8,
   "i": 9,
   "j": 10,
   "k": 11,
%s-  "l": 12,%s
%s+  "l": 120,%s
   "m": 13,
   "n": 14,
   "o": 15,
   ...
 }
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String(), chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestObjectContextSeparate(t *testing.T) {
	a, _ := GetBinary("3gAQoWEBoWICoWMDoWQEoWUFoWYGoWcHoWgIoWkJoWoKoWsLoWwMoW0NoW4OoW8PoXAQ")     // {"a":1,"b":2,"c":3,"d":4,"e":5,"f":6,"g":7,"h":8,"i":9,"j":10,"k":11,"l":12,"m":13,"n":14,"o":15,"p":16}
	b, _ := GetBinary("3gAQoWEBoWICoWMDoWQEoWUyoWYGoWcHoWgIoWkJoWoKoWsLoWwMoW0NoW7MjKFvD6FwEA==") // {"a":1,"b":2,"c":3,"d":4,"e":50,"f":6,"g":7,"h":8,"i":9,"j":10,"k":11,"l":12,"m":13,"n":140,"o":15,"p":16}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   ...
   "b": 2,
   "c": 3,
   "d": 4,
%s-  "e": 5,%s
%s+  "e": 50,%s
   "f": 6,
   "g": 7,
   "h": 8,
   ...
   "k": 11,
   "l": 12,
   "m": 13,
%s-  "n": 14,%s
%s+  "n": 140,%s
   "o": 15,
   "p": 16,
 }
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String(), chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayDeletionEmpty(t *testing.T) {
	a, _ := GetBinary("kQc=") // [7]
	b, _ := GetBinary("kA==") // []

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
%s-  7,%s
 ]
`, chalk.Red.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayDeletion(t *testing.T) {
	a, _ := GetBinary("lKFhoWKhY6Fk") // ["a","b","c","d"]
	b, _ := GetBinary("k6FhoWKhZA==") // ["a","b","d"]

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
   "a",
   "b",
%s-  "c",%s
   "d"
 ]
`, chalk.Red.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayAdditionEmpty(t *testing.T) {
	a, _ := GetBinary("kA==") // []
	b, _ := GetBinary("kQc=") // [7]

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
%s+  7,%s
 ]
`, chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayAddition(t *testing.T) {
	a, _ := GetBinary("k6FhoWKhZA==") // ["a","b","d"]
	b, _ := GetBinary("lKFhoWKhY6Fk") // ["a","b","c","d"]

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
   "a",
   "b"
%s+  "c",%s
   "d"
 ]
`, chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayChange(t *testing.T) {
	a, _ := GetBinary("kQY=") // [6]
	b, _ := GetBinary("kQc=") // [7]

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
%s-  6,%s
%s+  7,%s
 ]
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayContextSingle(t *testing.T) {
	a, _ := GetBinary("mgECAwQFBgcICQo=") // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
	b, _ := GetBinary("mgECAwQyBgcICQo=") // [1, 2, 3, 4, 50, 6, 7, 8, 9, 10]

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
   ...
   2,
   3,
   4,
%s-  5,%s
%s+  50,%s
   6,
   7,
   8,
   ...
 ]
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayContextOverlap(t *testing.T) {
	a, _ := GetBinary("nAECAwQFBgcICQoLDA==") // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12]
	b, _ := GetBinary("nAECAwQyBkYICQoLDA==") // [1, 2, 3, 4, 50, 6, 70, 8, 9, 10, 11, 12]

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
   ...
   2,
   3,
   4,
%s-  5,%s
%s+  50,%s
   6,
%s-  7,%s
%s+  70,%s
   8,
   9,
   10,
   ...
 ]
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String(), chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayContextAdjacent(t *testing.T) {
	a, _ := GetBinary("3AAQAQIDBAUGBwgJCgsMDQ4PEA==") // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]
	b, _ := GetBinary("3AAQAQIDBDIGBwgJCgt4DQ4PEA==") // [1, 2, 3, 4, 50, 6, 7, 8, 9, 10, 11, 120, 13, 14, 15, 16]

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
   ...
   2,
   3,
   4,
%s-  5,%s
%s+  50,%s
   6,
   7,
   8,
   9,
   10,
   11,
%s-  12,%s
%s+  120,%s
   13,
   14,
   15,
   ...
 ]
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String(), chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestArrayContextSeparate(t *testing.T) {
	a, _ := GetBinary("3AAQAQIDBAUGBwgJCgsMDQ4PEA==") // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]
	b, _ := GetBinary("3AAQAQIDBDIGBwgJCgsMDcyMDxA=") // [1, 2, 3, 4, 50, 6, 7, 8, 9, 10, 11, 12, 13, 140, 15, 16]

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` [
   ...
   2,
   3,
   4,
%s-  5,%s
%s+  50,%s
   6,
   7,
   8,
   ...
   11,
   12,
   13,
%s-  14,%s
%s+  140,%s
   15,
   16,
 ]
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String(), chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestEmbeddedArray(t *testing.T) {
	a, _ := GetBinary("gqVsZXZlbAGkZGF0YZIBAg==") // {"level":1,"data":[1, 2]}
	b, _ := GetBinary("gqVsZXZlbAGkZGF0YZICAQ==") // {"level":1,"data":[2, 1]}

	result, _ := Compare(a, b, false, false, false, false)

	if result.Equal {
		t.Error("Wrong result")
	}

	var builder strings.Builder
	result.PrintReport(&builder)

	expected := fmt.Sprintf(` {
   "level": 1,
   "data": [
%s-    1,%s
%s+    2,%s
%s-    2,%s
%s+    1,%s
   ]
 }
`, chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String(), chalk.Red.String(), chalk.ResetColor.String(), chalk.Green.String(), chalk.ResetColor.String())
	actual := builder.String()

	if expected != actual {
		t.Fatalf("Invalid report:\nExpected:\n%s\nGot:\n%s\n", expected, actual)
	}
}
