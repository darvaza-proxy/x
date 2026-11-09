//go:build goexperiment.jsonv2

package num_test

// jsonStringOptionByKind reports whether encoding/json applies the
// string option of a json tag by the field's kind, asking a field of an
// integer kind for a string even when its type has its own
// UnmarshalJSON. The build with the jsonv2 experiment hands a type with
// its own methods to them, whatever the option.
const jsonStringOptionByKind = false
