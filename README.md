# msgpackdiff

This is a command line tool written in Go that provides a way to compare two arbitrary
[MessagePack](https://msgpack.org/) encoded objects for equality. It uses the
[algorand/msgp](https://github.com/algorand/msgp) library to parse MessagePack objects.

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
 