package testutils_test

import (
	"errors"

	"darvaza.org/core"
)

// Named errors for testing.
var (
	ErrValueEmpty         = errors.New("value is empty")
	ErrValueCannotBeEmpty = errors.New("value cannot be empty")
	ErrCountNegative      = errors.New("count cannot be negative")
	ErrDivisionByZero     = errors.New("division by zero")
	ErrMinGreaterThanMax  = errors.New("min cannot be greater than max")
	ErrNegativeValues     = errors.New("negative values not allowed")
	ErrMinLengthNegative  = errors.New("minimum length cannot be negative")
)

// Test struct for demonstrations.
type TestStruct struct {
	value   string
	count   int
	enabled bool
	err     error
}

func validateTestStruct(_ core.T, ts *TestStruct) bool {
	// NOTE: No nil check needed - factory logic handles nil/not-nil automatically
	return ts.value != "" && ts.count >= 0
}

// Method examples that take *TestStruct as parameter.
func (ts *TestStruct) GetValue() string                     { return ts.value }
func (ts *TestStruct) GetCount() int                        { return ts.count }
func (ts *TestStruct) IsEnabled() bool                      { return ts.enabled }
func (ts *TestStruct) GetValueWithArg(suffix string) string { return ts.value + suffix }
func (ts *TestStruct) GetCountOK() (int, bool)              { return ts.count, ts.count > 0 }
func (ts *TestStruct) FindValueOK(search string) (string, bool) {
	if ts.value == search {
		return ts.value, true
	}
	return "", false
}
func (ts *TestStruct) GetValueError() (string, error) {
	if ts.err != nil {
		return "", ts.err
	}
	return ts.value, nil
}

func (ts *TestStruct) ParseValueError(input string) (string, error) {
	if input == "" {
		return "", ErrValueEmpty
	}
	if input == "invalid" {
		return "", ErrValueCannotBeEmpty
	}
	return ts.value + ":" + input, nil
}

func (ts *TestStruct) Validate() error {
	if ts.value == "" {
		return ErrValueEmpty
	}
	return nil
}

func (ts *TestStruct) ValidateOne(value string) error {
	if value == "" {
		return ErrValueEmpty
	}
	if ts.count < 0 {
		return ErrCountNegative
	}
	return nil
}

// Error-only wrapper methods for systematic testing.
func (ts *TestStruct) ValidateTwoArgs(minVal, maxVal int) error {
	if minVal > maxVal {
		return ErrMinGreaterThanMax
	}
	return nil
}

func (ts *TestStruct) ValidateThreeArgs(a, b, c int) error {
	if a < 0 || b < 0 || c < 0 {
		return ErrNegativeValues
	}
	sum := a + b + c
	if sum != ts.count {
		return ErrDivisionByZero
	}
	return nil
}

func (ts *TestStruct) ValidateFourArgs(a, b, c, d int) error {
	if a < 0 || b < 0 || c < 0 || d < 0 {
		return ErrNegativeValues
	}
	sum := a + b + c + d
	if sum != ts.count {
		return ErrDivisionByZero
	}
	return nil
}

func (ts *TestStruct) ValidateFiveArgs(a, b, c, d, e int) error {
	if a < 0 || b < 0 || c < 0 || d < 0 || e < 0 {
		return ErrNegativeValues
	}
	sum := a + b + c + d + e
	if sum != ts.count {
		return ErrDivisionByZero
	}
	return nil
}

// Two and three argument methods for testing.
func (ts *TestStruct) CombineValues(a, b string) string {
	return ts.value + a + b
}

func (ts *TestStruct) CheckRange(minVal, maxVal int) (inRange, hasPositiveCount bool) {
	inRange = ts.count >= minVal && ts.count <= maxVal
	hasPositiveCount = ts.count > 0
	return inRange, hasPositiveCount
}

func (ts *TestStruct) ValidateRange(minVal, maxVal int) (bool, error) {
	if minVal > maxVal {
		return false, ErrMinGreaterThanMax
	}
	return ts.count >= minVal && ts.count <= maxVal, nil
}

func (ts *TestStruct) ComplexMethod(a, b, c string) string {
	return ts.value + ":" + a + "+" + b + "=" + c
}

func (ts *TestStruct) CheckThreeValues(a, b, c int) (int, bool) {
	sum := a + b + c
	return sum, sum == ts.count
}

func (*TestStruct) ValidateThree(a, b, c int) (int, error) {
	if a < 0 || b < 0 || c < 0 {
		return 0, ErrNegativeValues
	}
	return a + b + c, nil
}

// Four and five argument methods for testing.
func (ts *TestStruct) CombineFourValues(a, b, c, d string) string {
	return ts.value + ":" + a + "-" + b + "-" + c + "-" + d
}

func (ts *TestStruct) CheckFourValues(a, b, c, d int) (int, bool) {
	sum := a + b + c + d
	return sum, sum == ts.count
}

func (*TestStruct) ValidateFourValues(a, b, c, d int) (int, error) {
	if a < 0 || b < 0 || c < 0 || d < 0 {
		return 0, ErrNegativeValues
	}
	return a + b + c + d, nil
}

func (ts *TestStruct) CombineFiveValues(a, b, c, d, e string) string {
	return ts.value + ":" + a + "/" + b + "/" + c + "/" + d + "/" + e
}

func (ts *TestStruct) CheckFiveValues(a, b, c, d, e int) (int, bool) {
	sum := a + b + c + d + e
	return sum, sum == ts.count
}

func (*TestStruct) ValidateFiveValues(a, b, c, d, e int) (int, error) {
	if a < 0 || b < 0 || c < 0 || d < 0 || e < 0 {
		return 0, ErrNegativeValues
	}
	return a + b + c + d + e, nil
}

// Factory examples that return *TestStruct.

// ============================================================================
// Factory variants (return *T)
// ============================================================================

func NewTestStruct() *TestStruct {
	return &TestStruct{value: "default", count: 1, enabled: true}
}

func NewTestStructNil() *TestStruct {
	// Factory that returns nil for testing expectNil cases
	return nil
}

func NewTestStructValid() *TestStruct {
	// Factory that returns valid struct for testing
	return &TestStruct{value: "valid", count: 5, enabled: true}
}

func NewTestStructInvalid() *TestStruct {
	// Factory that returns invalid struct for testing validation failures
	return &TestStruct{value: "", count: -1, enabled: false}
}

func NewTestStructOneArg(value string) *TestStruct {
	return &TestStruct{value: value, count: 1, enabled: true}
}

func NewTestStructTwoArgs(value string, count int) *TestStruct {
	return &TestStruct{value: value, count: count, enabled: true}
}

func NewTestStructThreeArgs(value string, count int, enabled bool) *TestStruct {
	return &TestStruct{value: value, count: count, enabled: enabled}
}

func NewTestStructFourArgs(value string, count int, enabled bool, err error) *TestStruct {
	return &TestStruct{value: value, count: count, enabled: enabled, err: err}
}

func NewTestStructFiveArgs(value string, count int, enabled bool, err error, extra int) *TestStruct {
	// Incorporate extra parameter in count for testing
	return &TestStruct{value: value, count: count + extra, enabled: enabled, err: err}
}

// ============================================================================
// FactoryOK variants (return (*T, bool))
// ============================================================================

func NewTestStructOK() (*TestStruct, bool) {
	// Factory that always succeeds for no-arg variant
	return &TestStruct{value: "default", count: 1, enabled: true}, true
}

func NewTestStructOKFalse() (*TestStruct, bool) {
	// Factory that always returns false for no-arg variant
	return nil, false
}

func NewTestStructOKTrue() (*TestStruct, bool) {
	// Factory that always returns true for no-arg variant
	return &TestStruct{value: "success", count: 5, enabled: true}, true
}

func NewTestStructOKOneArg(value string) (*TestStruct, bool) {
	if value == "" {
		return nil, false
	}
	return &TestStruct{value: value, count: 1, enabled: true}, true
}

func NewTestStructOKTwoArgs(value string, count int) (*TestStruct, bool) {
	if value == "" || count < 0 {
		return nil, false
	}
	return &TestStruct{value: value, count: count, enabled: true}, true
}

func NewTestStructOKThreeArgs(value string, count int, enabled bool) (*TestStruct, bool) {
	if value == "" || count < 0 {
		return nil, false
	}
	return &TestStruct{value: value, count: count, enabled: enabled}, true
}

func NewTestStructOKFourArgs(value string, count int, enabled bool, err error) (*TestStruct, bool) {
	if value == "" || count < 0 {
		return nil, false
	}
	return &TestStruct{value: value, count: count, enabled: enabled, err: err}, true
}

func NewTestStructOKFiveArgs(value string, count int, enabled bool, err error, extra int) (*TestStruct, bool) {
	if value == "" || count < 0 {
		return nil, false
	}
	return &TestStruct{value: value, count: count + extra, enabled: enabled, err: err}, true
}

// ============================================================================
// FactoryError variants (return (*T, error))
// ============================================================================

func NewTestStructError() (*TestStruct, error) {
	// Factory that always succeeds for no-arg variant
	return &TestStruct{value: "default", count: 1, enabled: true}, nil
}

func NewTestStructErrorSuccess() (*TestStruct, error) {
	// Factory that always succeeds for no-arg variant
	return &TestStruct{value: "success", count: 10, enabled: true}, nil
}

func NewTestStructErrorFail() (*TestStruct, error) {
	// Factory that always fails for no-arg variant
	return nil, ErrValueCannotBeEmpty
}

func NewTestStructErrorDifferentFail() (*TestStruct, error) {
	// Factory that fails with different error for no-arg variant
	return nil, ErrCountNegative
}

func NewTestStructErrorOneArg(value string) (*TestStruct, error) {
	if value == "" {
		return nil, ErrValueCannotBeEmpty
	}
	return &TestStruct{value: value, count: 1, enabled: true}, nil
}

func NewTestStructErrorTwoArgs(value string, count int) (*TestStruct, error) {
	if value == "" {
		return nil, ErrValueCannotBeEmpty
	}
	if count < 0 {
		return nil, ErrCountNegative
	}
	return &TestStruct{value: value, count: count, enabled: true}, nil
}

func NewTestStructErrorThreeArgs(value string, count int, enabled bool) (*TestStruct, error) {
	if value == "" {
		return nil, ErrValueCannotBeEmpty
	}
	if count < 0 {
		return nil, ErrCountNegative
	}
	return &TestStruct{value: value, count: count, enabled: enabled}, nil
}

func NewTestStructErrorFourArgs(value string, count int, enabled bool, err error) (*TestStruct, error) {
	if value == "" {
		return nil, ErrValueCannotBeEmpty
	}
	if count < 0 {
		return nil, ErrCountNegative
	}
	return &TestStruct{value: value, count: count, enabled: enabled, err: err}, nil
}

func NewTestStructErrorFiveArgs(value string, count int, enabled bool, err error, extra int) (*TestStruct, error) {
	if value == "" {
		return nil, ErrValueCannotBeEmpty
	}
	if count < 0 {
		return nil, ErrCountNegative
	}
	return &TestStruct{value: value, count: count + extra, enabled: enabled, err: err}, nil
}
