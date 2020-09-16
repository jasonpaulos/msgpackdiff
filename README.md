# msgpackdiff

This is a command line tool written in Go that provides a way to compare two arbitrary
[MessagePack](https://msgpack.org/) encoded objects for equality. It uses the
[algorand/msgp](https://github.com/algorand/msgp) library to parse MessagePack objects.

## Build

To build the tool, clone the repo and run `make build`. This will place a copy of the `msgpackdiff`
binary in the `./bin` folder. You can place this binary somewhere on your system path to invoke it
from anywhere.

## Usage

The tool can be invoked from the command line like so:

```
msgpackdiff (flags) [A] [B]
```

`[A]` and `[B]` are required arguments and should be replaced with actual MessagePack objects.
`msgpackdiff` will parse each object and compare them for equality. If the objects are equal, the
program will exit with status code 0. If they are unequal, the progarm will exit with status code 1
and print the parts that differ to stdout.

### Object encoding
The objects `[A]` and `[B]` can be any of the following:
* A base64 encoded string of a MessagePack object.
* A path to a file that contains only the MessagePack object. The conents may be binary or a base64
  encoded string.

### Flags
* `--brief` enables quiet mode, which causes the program to refrain from outputting a detailed
  report if the objects are different. If the objects are equal, the program will output nothing.
* `--ignore-empty` causes the tool to ignore differences that can be explained by one MessagePack
  object omitting empty fields and the other keeping them. For example, if the objects
  `{"key": "val"}` and `{"key": "val", "extraKey": ""}` were encoded with MessagePack and compared with
  this flag enabled, they would be considered equivalent since the value for `extraKey` is the empty
  string, which is the default string value in Go.
* `--ignore-order` disables strict ordering of fields. Without this flag, each object must define
  the same fields in the same order, resulting in the following objects being considered different:
  `{"a": 1, "b": 2}`, `{"b": 2, "a": 1}`. However with this flag, the objects would be considered
  equivalent. This affects the order of all fields in objects, regardless of whether they are in the
  top level object or subobjects.
* `--flexible-types` disables strict type comparisons. Without this flag, int, uint, float32,
  float64, complex64, and complex128 types will never be compared to each other since they are
  assumed to be unequal due to their type. However with this flag, values belonging to these
  different numerical types will be cast to the larger type and compared using Go's == operator. In
  some cases this comparison will be inaccurate, such as between int64 and float32, since neither
  type can hold all values of the other.
  NOTE: This flag does not change the behavior of comparing different types within the int8/16/32/64
  family, which are always compared with each other regardless of what length they are. The same is
  true for the uint8/16/32/64 family, but not between the int and uint families.
* `--context` adjusts the number of nearby fields to show in difference reports. Defaults to 3.
