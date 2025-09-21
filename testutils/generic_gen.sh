#!/bin/sh
# shellcheck disable=SC1007,SC3043 # empty assignments and local usage
# cspell:ignore GOPACKAGE GOFILE

set -eu

cd "$(dirname "$0")"

TAG=go:generate
GOFILE=${GOFILE:-generic_gen.go}
GOPACKAGE=${GOPACKAGE:-testutils}
SCRIPT="./$(basename "$0")"

TAB="$(printf '\t')"

TMPFILE="$(mktemp -t "tmp.${GOPACKAGE}.${GOFILE}.XXXX")"
# shellcheck disable=SC2064  # filename is safe, comes from template
trap "rm -f '$TMPFILE'" EXIT

# Helper function to format field declarations with proper alignment using column -t
# and apply indentation with single space output separator.
# Usage: ... | tab "$indent"
tab() {
	local indent="$1"
	column -t -s "$TAB" -o " " | sed "s/^/$indent/"
}

# Generate argument fields for TestCase structs
# Usage: gen_arg_fields <num_args>
gen_arg_fields() {
	local num_args="$1"
	local i

	for i in $(seq 1 "$num_args"); do
		echo "arg$i${TAB}A$i"
	done
}

# Generate constructor parameters for arguments
# Usage: gen_arg_constructor_params <num_args>
gen_arg_constructor_params() {
	local num_args="$1"
	local i

	for i in $(seq 1 "$num_args"); do
		echo "arg$i A$i,"
	done
}

# Generate constructor field assignments for arguments
# Usage: gen_arg_constructor_fields <num_args>
gen_arg_constructor_fields() {
	local num_args="$1"
	local i

	for i in $(seq 1 "$num_args"); do
		echo "arg$i:${TAB}arg$i,"
	done
}

# Convert type parameters to TestCase field references using sed.
# Usage: convert_args_to_fields <type_params>
# Example: "A1, A2" -> "tc.arg1, tc.arg2"
convert_args_to_fields() {
	echo "$1" | sed 's/A\([0-9]\)/tc.arg\1/g'
}

# Generate a struct definition with pre-formatted fields.
# Usage: gen_struct <struct_name> <type_decl> <fields> <description>
gen_struct() {
	local struct_name="$1" type_decl="$2" fields="$3" description="$4"

	cat <<EOT

// ${struct_name} is a generic test case for testing
// ${description}
type ${struct_name}${type_decl} struct {
$(echo "$fields" | tab "$TAB")
}
EOT
}

# Generate a Name() method.
# Usage: gen_name_method <struct_name> <type_ref>
gen_name_method() {
	local struct_name="$1" type_ref="$2"

	cat <<EOT

// Name returns the test case name.
func (tc ${struct_name}${type_ref}) Name() string {
	return tc.name
}
EOT
}

# Generate a constructor function.
# Usage: gen_constructor <struct_name> <type_decl> <type_ref> <params> <assignments>
gen_constructor() {
	local struct_name="$1" type_decl="$2" type_ref="$3" params="$4" assignments="$5"

	cat <<EOT

// New${struct_name} creates a new ${struct_name} instance.
func New${struct_name}${type_decl}(
$(echo "$params" | sed "s/^/${TAB}/")
) ${struct_name}${type_ref} {
	return ${struct_name}${type_ref}{
$(echo "$assignments" | tab "$TAB$TAB")
	}
}
EOT
}

# Generate a verification line for TestCase interface.
# Usage: gen_verification <base_name> <suffix> <type_params>
gen_verification() {
	local base_name="$1" suffix="$2" type_params="$3"
	local struct_name="${base_name}${suffix}TestCase"
	echo "var _ core.TestCase = (*${struct_name}[$type_params])(nil)"
}

# Generate verification for method types (T + args + V).
# Usage: gen_method_verification <base_name> <suffix> <arg_type_params>
gen_method_verification() {
	gen_verification "$1" "$2" "any${3:+, $3}, int"
}

# Generate verification for factory types (T + args).
# Usage: gen_factory_verification_base <base_name> <suffix> <arg_type_params>
gen_factory_verification_base() {
	gen_verification "$1" "$2" "any${3:+, $3}"
}

# Generate verification for function types (args + V).
# Usage: gen_function_verification_base <base_name> <suffix> <arg_type_params>
gen_function_verification_base() {
	gen_verification "$1" "$2" "${3:+$3, }int"
}

# Generate a type definition with common structure.
# Usage: gen_type_definition <name> <suffix> <type_params> <arg_desc> <input_signature> <output_signature> <generic_constraints>
gen_type_definition() {
	local name="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local input_sig="$5" output_sig="$6" generic_constraints="$7"

	# Build the description based on whether it's a method or function
	local desc
	if [ -n "$input_sig" ]; then
		# It's a method (has *T input signature)
		desc="represents a method that takes $input_sig${arg_desc:+, $arg_desc,} and returns $output_sig"
	else
		# It's a function
		desc="represents a function that ${arg_desc:+takes $arg_desc and }returns $output_sig"
	fi

	cat <<EOT

// ${name}${suffix} $desc
type ${name}${suffix}[${generic_constraints}] func($input_sig${type_params:+${input_sig:+, }$type_params}) $output_sig
EOT
}

# Generate a method type (takes *T as first parameter).
# Usage: gen_method_type <base_name> <return_sig> <suffix> <type_params> <arg_desc>
gen_method_type() {
	local base_name="$1" return_sig="$2"
	local suffix="$3" type_params="$4" arg_desc="$5"

	local generic_constraints="T${type_params:+, $type_params} any${return_sig:+, V comparable}"
	gen_type_definition "${base_name}Method" "$suffix" "$type_params" "$arg_desc" "*T" "$return_sig" "$generic_constraints"
}

# Generate a factory type (returns *T).
# Usage: gen_factory_func_type <base_name> <return_sig> <suffix> <type_params> <arg_desc>
gen_factory_func_type() {
	local base_name="$1" return_sig="$2"
	local suffix="$3" type_params="$4" arg_desc="$5"

	local generic_constraints="T${type_params:+, $type_params} any"
	gen_type_definition "$base_name" "$suffix" "$type_params" "$arg_desc" "" "$return_sig" "$generic_constraints"
}

# Generate a function type (pure function).
# Usage: gen_function_func_type <base_name> <return_sig> <suffix> <type_params> <arg_desc>
gen_function_func_type() {
	local base_name="$1" return_sig="$2"
	local suffix="$3" type_params="$4" arg_desc="$5"

	local generic_constraints="${type_params:+$type_params any, }V comparable"
	gen_type_definition "$base_name" "$suffix" "$type_params" "$arg_desc" "" "$return_sig" "$generic_constraints"
}

# Generate type variations for a generator function.
# Usage: generate_type_loop <generator_function>
generate_type_loop() {
	local generator_func="$1"

	# Call generator for each argument count (0, 1, 2, 3, 4, 5)
	$generator_func "" "" ""
	$generator_func "OneArg" "A1" "one argument"
	$generator_func "TwoArgs" "A1, A2" "two arguments"
	$generator_func "ThreeArgs" "A1, A2, A3" "three arguments"
	$generator_func "FourArgs" "A1, A2, A3, A4" "four arguments"
	$generator_func "FiveArgs" "A1, A2, A3, A4, A5" "five arguments"
}

# Generate Getter type
gen_getter_type() {
	gen_method_type "Getter" "V" "$@"
}

# Generate GetterOK type
gen_getter_ok_type() {
	gen_method_type "GetterOK" "(V, bool)" "$@"
}

# Generate GetterError type
gen_getter_error_type() {
	gen_method_type "GetterError" "(V, error)" "$@"
}

# Generate Error type
gen_error_type() {
	local suffix="$1" type_params="$2" arg_desc="$3"
	local generic_constraints="T${type_params:+, $type_params} any"
	gen_type_definition "ErrorMethod" "$suffix" "$type_params" "$arg_desc" "*T" "error" "$generic_constraints"
}

# Generate Factory type
gen_factory_type() {
	gen_factory_func_type "Factory" "*T" "$@"
}

# Generate FactoryOK type
gen_factory_ok_type() {
	gen_factory_func_type "FactoryOK" "(*T, bool)" "$@"
}

# Generate FactoryError type
gen_factory_error_type() {
	gen_factory_func_type "FactoryError" "(*T, error)" "$@"
}

# Generate Function type
gen_function_type() {
	gen_function_func_type "Function" "V" "$@"
}

# Generate FunctionOK type
gen_function_ok_type() {
	gen_function_func_type "FunctionOK" "(V, bool)" "$@"
}

# Generate FunctionError type
gen_function_error_type() {
	gen_function_func_type "FunctionError" "(V, error)" "$@"
}

# Generate verification variations for a generator function.
# Usage: generate_verification_loop <generator_function>
generate_verification_loop() {
	local generator_func="$1"
	local suffix type_params

	# Call generator for each argument count (0, 1, 2, 3, 4, 5)
	$generator_func "" ""
	$generator_func "OneArg" "any"
	$generator_func "TwoArgs" "any, any"
	$generator_func "ThreeArgs" "any, any, any"
	$generator_func "FourArgs" "any, any, any, any"
	$generator_func "FiveArgs" "any, any, any, any, any"
}

# Generate Getter verification
gen_getter_verification() {
	gen_method_verification "Getter" "$@"
}

# Generate GetterOK verification
gen_getter_ok_verification() {
	gen_method_verification "GetterOK" "$@"
}

# Generate GetterError verification
gen_getter_error_verification() {
	gen_method_verification "GetterError" "$@"
}

# Generate Error verification
gen_error_verification() {
	gen_factory_verification_base "Error" "$@"
}

# Generate Factory verification
gen_factory_verification() {
	gen_factory_verification_base "Factory" "$@"
}

# Generate FactoryOK verification
gen_factory_ok_verification() {
	gen_factory_verification_base "FactoryOK" "$@"
}

# Generate FactoryError verification
gen_factory_error_verification() {
	gen_factory_verification_base "FactoryError" "$@"
}

# Generate Function verification
gen_function_verification() {
	gen_function_verification_base "Function" "$@"
}

# Generate FunctionOK verification
gen_function_ok_verification() {
	gen_function_verification_base "FunctionOK" "$@"
}

# Generate FunctionError verification
gen_function_error_verification() {
	gen_function_verification_base "FunctionError" "$@"
}

# Generate TestCase variations for a generator function.
# Usage: generate_testcase_loop <generator_function>
generate_testcase_loop() {
	local generator_func="$1"
	local suffix type_params arg_desc

	# Call generator for each argument count (0, 1, 2, 3, 4, 5)
	$generator_func 0 "" "" ""
	$generator_func 1 "OneArg" "A1" "one argument"
	$generator_func 2 "TwoArgs" "A1, A2" "two arguments"
	$generator_func 3 "ThreeArgs" "A1, A2, A3" "three arguments"
	$generator_func 4 "FourArgs" "A1, A2, A3, A4" "four arguments"
	$generator_func 5 "FiveArgs" "A1, A2, A3, A4, A5" "five arguments"
}

# Generate Getter testcases.
gen_getter_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="Getter${suffix}TestCase"
	local method_type="GetterMethod${suffix}"

	# Type parameters for Getter (has V)
	local type_decl="[T${type_params:+, $type_params} any, V comparable]"
	local type_ref="[T${type_params:+, $type_params}, V]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
method${TAB}${method_type}${type_ref}
methodName${TAB}string
item${TAB}*T${arg_fields:+
$arg_fields}
expected${TAB}V"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
method:${TAB}method,
methodName:${TAB}methodName,
item:${TAB}item,${arg_constructor_fields:+
$arg_constructor_fields}
expected:${TAB}expected,"

	# Build constructor params
	local constructor_params="name string,
method ${method_type}${type_ref},
methodName string,
item *T,${arg_constructor_params:+
$arg_constructor_params}
expected V,"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "getter methods that return V${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	actual := tc.method(tc.item${args:+, $args})
	core.AssertEqual(t, tc.expected, actual, tc.methodName)
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate GetterOK testcases.
gen_getter_ok_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="GetterOK${suffix}TestCase"
	local method_type="GetterOKMethod${suffix}"

	# Type parameters for GetterOK (has V)
	local type_decl="[T${type_params:+, $type_params} any, V comparable]"
	local type_ref="[T${type_params:+, $type_params}, V]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
method${TAB}${method_type}${type_ref}
methodName${TAB}string
item${TAB}*T${arg_fields:+
$arg_fields}
expected${TAB}V
expectOK${TAB}bool"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
method:${TAB}method,
methodName:${TAB}methodName,
item:${TAB}item,${arg_constructor_fields:+
$arg_constructor_fields}
expected:${TAB}expected,
expectOK:${TAB}expectOK,"

	# Build constructor params
	local constructor_params="name string,
method ${method_type}${type_ref},
methodName string,
item *T,${arg_constructor_params:+
$arg_constructor_params}
expected V,
expectOK bool,"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "getter methods that return (V, bool)${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	actual, ok := tc.method(tc.item${args:+, $args})
	core.AssertEqual(t, tc.expectOK, ok, tc.methodName+":ok")
	if tc.expectOK {
		core.AssertEqual(t, tc.expected, actual, tc.methodName)
	}
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate GetterError testcases.
gen_getter_error_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="GetterError${suffix}TestCase"
	local method_type="GetterErrorMethod${suffix}"

	# Type parameters for GetterError (has V)
	local type_decl="[T${type_params:+, $type_params} any, V comparable]"
	local type_ref="[T${type_params:+, $type_params}, V]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
method${TAB}${method_type}${type_ref}
methodName${TAB}string
item${TAB}*T${arg_fields:+
$arg_fields}
expected${TAB}V
expectError${TAB}bool
errorIs${TAB}error"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
method:${TAB}method,
methodName:${TAB}methodName,
item:${TAB}item,${arg_constructor_fields:+
$arg_constructor_fields}
expected:${TAB}expected,
expectError:${TAB}expectError,
errorIs:${TAB}errorIs,"

	# Build constructor params
	local constructor_params="name string,
method ${method_type}${type_ref},
methodName string,
item *T,${arg_constructor_params:+
$arg_constructor_params}
expected V,
expectError bool,
errorIs error,"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "getter methods that return (V, error)${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	actual, err := tc.method(tc.item${args:+, $args})
	switch {
	case !tc.expectError:
		core.AssertNoError(t, err, tc.methodName+":error")
		core.AssertEqual(t, tc.expected, actual, tc.methodName)
	case tc.errorIs == nil:
		core.AssertError(t, err, tc.methodName+":error")
	default:
		core.AssertErrorIs(t, tc.errorIs, err, tc.methodName+":error")
	}
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate Error testcases.
gen_error_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="Error${suffix}TestCase"
	local method_type="ErrorMethod${suffix}"

	# Type parameters for Error (no V)
	local type_decl="[T${type_params:+, $type_params} any]"
	local type_ref="[T${type_params:+, $type_params}]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
method${TAB}${method_type}${type_ref}
methodName${TAB}string
item${TAB}*T${arg_fields:+
$arg_fields}
expectError${TAB}bool
errorIs${TAB}error"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
method:${TAB}method,
methodName:${TAB}methodName,
item:${TAB}item,${arg_constructor_fields:+
$arg_constructor_fields}
expectError:${TAB}expectError,
errorIs:${TAB}errorIs,"

	# Build constructor params
	local constructor_params="name string,
method ${method_type}${type_ref},
methodName string,
item *T,${arg_constructor_params:+
$arg_constructor_params}
expectError bool,
errorIs error,"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "methods that return error${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	err := tc.method(tc.item${args:+, $args})
	switch {
	case !tc.expectError:
		core.AssertNoError(t, err, tc.methodName)
	case tc.errorIs == nil:
		core.AssertError(t, err, tc.methodName)
	default:
		core.AssertErrorIs(t, tc.errorIs, err, tc.methodName)
	}
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate Factory testcases.
gen_factory_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="Factory${suffix}TestCase"
	local factory_type="Factory${suffix}"

	# Type parameters for Factory (no V)
	local type_decl="[T${type_params:+, $type_params} any]"
	local type_ref="[T${type_params:+, $type_params}]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
fn${TAB}${factory_type}${type_ref}
funcName${TAB}string${arg_fields:+
$arg_fields}
expectNil${TAB}bool
typeTest${TAB}TypeTestFunc[T]"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
fn:${TAB}fn,
funcName:${TAB}funcName,${arg_constructor_fields:+
$arg_constructor_fields}
expectNil:${TAB}expectNil,
typeTest:${TAB}typeTest,"

	# Build constructor params
	local constructor_params="name string,
fn ${factory_type}${type_ref},
funcName string,${arg_constructor_params:+
$arg_constructor_params}
expectNil bool,
typeTest TypeTestFunc[T],"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "factory functions that return *T${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	actual := tc.fn(${args})
	if tc.expectNil {
		core.AssertNil(t, actual, tc.funcName)
		return
	}

	core.AssertMustNotNil(t, actual, tc.funcName)
	if tc.typeTest != nil {
		valid := tc.typeTest(t, actual)
		core.AssertTrue(t, valid, tc.funcName+":valid")
	}
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate FactoryOK testcases.
gen_factory_ok_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="FactoryOK${suffix}TestCase"
	local factory_type="FactoryOK${suffix}"

	# Type parameters for FactoryOK (no V)
	local type_decl="[T${type_params:+, $type_params} any]"
	local type_ref="[T${type_params:+, $type_params}]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
fn${TAB}${factory_type}${type_ref}
funcName${TAB}string${arg_fields:+
$arg_fields}
expectOK${TAB}bool
typeTest${TAB}TypeTestFunc[T]"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
fn:${TAB}fn,
funcName:${TAB}funcName,${arg_constructor_fields:+
$arg_constructor_fields}
expectOK:${TAB}expectOK,
typeTest:${TAB}typeTest,"

	# Build constructor params
	local constructor_params="name string,
fn ${factory_type}${type_ref},
funcName string,${arg_constructor_params:+
$arg_constructor_params}
expectOK bool,
typeTest TypeTestFunc[T],"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "factory functions that return (*T, bool)${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	actual, ok := tc.fn(${args})
	core.AssertEqual(t, tc.expectOK, ok, tc.funcName+":ok")
	if !tc.expectOK {
		core.AssertMustNil(t, actual, tc.funcName)
		return
	}

	core.AssertMustNotNil(t, actual, tc.funcName)
	if tc.typeTest != nil {
		valid := tc.typeTest(t, actual)
		core.AssertTrue(t, valid, tc.funcName+":valid")
	}
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate FactoryError testcases.
gen_factory_error_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="FactoryError${suffix}TestCase"
	local factory_type="FactoryError${suffix}"

	# Type parameters for FactoryError (no V)
	local type_decl="[T${type_params:+, $type_params} any]"
	local type_ref="[T${type_params:+, $type_params}]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
fn${TAB}${factory_type}${type_ref}
funcName${TAB}string${arg_fields:+
$arg_fields}
expectError${TAB}bool
errorIs${TAB}error
typeTest${TAB}TypeTestFunc[T]"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
fn:${TAB}fn,
funcName:${TAB}funcName,${arg_constructor_fields:+
$arg_constructor_fields}
expectError:${TAB}expectError,
errorIs:${TAB}errorIs,
typeTest:${TAB}typeTest,"

	# Build constructor params
	local constructor_params="name string,
fn ${factory_type}${type_ref},
funcName string,${arg_constructor_params:+
$arg_constructor_params}
expectError bool,
errorIs error,
typeTest TypeTestFunc[T],"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "factory functions that return (*T, error)${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	actual, err := tc.fn(${args})
	switch {
	case !tc.expectError:
		core.AssertNoError(t, err, tc.funcName)
		core.AssertMustNotNil(t, actual, tc.funcName)
		if tc.typeTest == nil {
			return
		}

		valid := tc.typeTest(t, actual)
		core.AssertTrue(t, valid, tc.funcName+":valid")
	case tc.errorIs == nil:
		core.AssertMustError(t, err, tc.funcName)
	default:
		core.AssertErrorIs(t, err, tc.errorIs, tc.funcName+":type")
	}
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate Function testcases.
gen_function_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="Function${suffix}TestCase"
	local function_type="Function${suffix}"

	# Type parameters for Function (has V, no T)
	local type_decl="[${type_params:+$type_params any, }V comparable]"
	local type_ref="[${type_params:+$type_params, }V]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
fn${TAB}${function_type}${type_ref}
funcName${TAB}string${arg_fields:+
$arg_fields}
expected${TAB}V"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
fn:${TAB}fn,
funcName:${TAB}funcName,${arg_constructor_fields:+
$arg_constructor_fields}
expected:${TAB}expected,"

	# Build constructor params
	local constructor_params="name string,
fn ${function_type}${type_ref},
funcName string,${arg_constructor_params:+
$arg_constructor_params}
expected V,"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "functions that return V${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	core.AssertMustNotNil(t, tc.fn, tc.funcName+":function")
	actual := tc.fn(${args})
	core.AssertEqual(t, tc.expected, actual, tc.funcName)
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate FunctionOK testcases.
gen_function_ok_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="FunctionOK${suffix}TestCase"
	local function_type="FunctionOK${suffix}"

	# Type parameters for FunctionOK (has V, no T)
	local type_decl="[${type_params:+$type_params any, }V comparable]"
	local type_ref="[${type_params:+$type_params, }V]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
fn${TAB}${function_type}${type_ref}
funcName${TAB}string${arg_fields:+
$arg_fields}
expected${TAB}V
expectOK${TAB}bool"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
fn:${TAB}fn,
funcName:${TAB}funcName,${arg_constructor_fields:+
$arg_constructor_fields}
expected:${TAB}expected,
expectOK:${TAB}expectOK,"

	# Build constructor params
	local constructor_params="name string,
fn ${function_type}${type_ref},
funcName string,${arg_constructor_params:+
$arg_constructor_params}
expected V,
expectOK bool,"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "functions that return (V, bool)${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	actual, ok := tc.fn(${args})
	core.AssertEqual(t, tc.expectOK, ok, tc.funcName+":ok")
	if tc.expectOK {
		core.AssertEqual(t, tc.expected, actual, tc.funcName)
	}
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

# Generate FunctionError testcases.
gen_function_error_testcase() {
	local num_args="$1" suffix="$2" type_params="$3" arg_desc="$4"
	local struct_name="FunctionError${suffix}TestCase"
	local function_type="FunctionError${suffix}"

	# Type parameters for FunctionError (has V, no T)
	local type_decl="[${type_params:+$type_params any, }V comparable]"
	local type_ref="[${type_params:+$type_params, }V]"

	# Generate argument fields and call parameters
	# shellcheck disable=SC2155,SC2046
	local arg_fields=$(gen_arg_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_params=$(gen_arg_constructor_params "$num_args")
	# shellcheck disable=SC2155,SC2046
	local arg_constructor_fields=$(gen_arg_constructor_fields "$num_args")
	# shellcheck disable=SC2155,SC2046
	local args=$(convert_args_to_fields "$type_params")

	# Build field list
	local all_fields="name${TAB}string
fn${TAB}${function_type}${type_ref}
funcName${TAB}string${arg_fields:+
$arg_fields}
expected${TAB}V
expectError${TAB}bool
errorIs${TAB}error"

	# Build constructor field assignments
	local constructor_field_assignments="name:${TAB}name,
fn:${TAB}fn,
funcName:${TAB}funcName,${arg_constructor_fields:+
$arg_constructor_fields}
expected:${TAB}expected,
expectError:${TAB}expectError,
errorIs:${TAB}errorIs,"

	# Build constructor params
	local constructor_params="name string,
fn ${function_type}${type_ref},
funcName string,${arg_constructor_params:+
$arg_constructor_params}
expected V,
expectError bool,
errorIs error,"

	# Generate struct
	gen_struct "$struct_name" "$type_decl" "$all_fields" "functions that return (V, error)${arg_desc:+ with $arg_desc}"

	# Generate Name method
	gen_name_method "$struct_name" "$type_ref"

	# Generate Test method
	cat <<EOT

// Test executes the test case.
func (tc ${struct_name}${type_ref}) Test(t *testing.T) {
	t.Helper()
	actual, err := tc.fn(${args})
	switch {
	case !tc.expectError:
		core.AssertNoError(t, err, tc.funcName+":error")
		core.AssertEqual(t, tc.expected, actual, tc.funcName)
	case tc.errorIs == nil:
		core.AssertError(t, err, tc.funcName+":error")
	default:
		core.AssertErrorIs(t, tc.errorIs, err, tc.funcName+":error")
	}
}
EOT

	# Generate constructor
	gen_constructor "$struct_name" "$type_decl" "$type_ref" "$constructor_params" "$constructor_field_assignments"
}

generate() {
	local name=

	# Generate header
	cat <<EOT
package $GOPACKAGE

// Code generated by $SCRIPT; DO NOT EDIT

//$TAG $SCRIPT

//revive:disable:flag-parameter,argument-limit,confusing-naming

import (
	"testing"

	"darvaza.org/core"
)
EOT

	# Generate type aliases
	for name; do generate_type_loop "gen_${name}_type"; done

	# Generate interface verifications
	cat <<EOT

// Compile-time verification that test case types implement TestCase interface.
EOT
	for name; do generate_verification_loop "gen_${name}_verification"; done

	# Generate TestCase types
	for name; do generate_testcase_loop "gen_${name}_testcase"; done
}

generate \
	getter getter_ok getter_error error \
	factory factory_ok factory_error \
	function function_ok function_error \
	> "$TMPFILE"

# Only replace if content differs
if ! cmp -s "$TMPFILE" "$GOFILE"; then
	mv "$TMPFILE" "$GOFILE"
else
	rm "$TMPFILE"
fi
