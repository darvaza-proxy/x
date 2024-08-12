package forms

// Code generated by ./value_gen.sh; DO NOT EDIT

//go:generate ./value_gen.sh

import (
	"net/http"

	"darvaza.org/core"
)

// FormValueSigned reads a field from [http.Request#Form], after populating it if needed,
// and returns a [core.Signed] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func FormValueSigned[T core.Signed](req *http.Request, field string,
	base int) (value T, found bool, err error) {
	//
	s, found, err := FormValue[string](req, field)
	if err == nil && found {
		value, err = ParseSigned[T](s, base)
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// FormValuesSigned reads a field from [http.Request#Form], after populating it if needed,
// and returns its [core.Signed] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func FormValuesSigned[T core.Signed](req *http.Request, field string,
	base int) (values []T, found bool, err error) {
	//
	ss, found, err := FormValues[string](req, field)
	if err == nil && found {
		values = make([]T, 0, len(ss))

		for _, s := range ss {
			v, err := ParseSigned[T](s, base)
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}

// FormValueSignedInRange reads a field from [http.Request#Form], after populating it if needed,
// and returns a [core.Signed] value, an indicator saying if it was actually,
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type or if it's outside the specified
// boundaries.
func FormValueSignedInRange[T core.Signed](req *http.Request, field string, base int,
	min, max T) (value T, found bool, err error) {
	//
	value, found, err = FormValueSigned[T](req, field, base)
	if err == nil && found {
		if value < min || value > max {
			err = errRange("ParseInt", FormatSigned(value, 10))
		}
	}
	return value, found, err
}

// FormValueUnsigned reads a field from [http.Request#Form], after populating it if needed,
// and returns a [core.Unsigned] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func FormValueUnsigned[T core.Unsigned](req *http.Request, field string,
	base int) (value T, found bool, err error) {
	//
	s, found, err := FormValue[string](req, field)
	if err == nil && found {
		value, err = ParseUnsigned[T](s, base)
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// FormValuesUnsigned reads a field from [http.Request#Form], after populating it if needed,
// and returns its [core.Unsigned] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func FormValuesUnsigned[T core.Unsigned](req *http.Request, field string,
	base int) (values []T, found bool, err error) {
	//
	ss, found, err := FormValues[string](req, field)
	if err == nil && found {
		values = make([]T, 0, len(ss))

		for _, s := range ss {
			v, err := ParseUnsigned[T](s, base)
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}

// FormValueUnsignedInRange reads a field from [http.Request#Form], after populating it if needed,
// and returns a [core.Unsigned] value, an indicator saying if it was actually,
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type or if it's outside the specified
// boundaries.
func FormValueUnsignedInRange[T core.Unsigned](req *http.Request, field string, base int,
	min, max T) (value T, found bool, err error) {
	//
	value, found, err = FormValueUnsigned[T](req, field, base)
	if err == nil && found {
		if value < min || value > max {
			err = errRange("ParseUint", FormatUnsigned(value, 10))
		}
	}
	return value, found, err
}

// FormValueFloat reads a field from [http.Request#Form], after populating it if needed,
// and returns a [core.Float] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func FormValueFloat[T core.Float](req *http.Request, field string) (value T, found bool, err error) {
	//
	s, found, err := FormValue[string](req, field)
	if err == nil && found {
		value, err = ParseFloat[T](s)
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// FormValuesFloat reads a field from [http.Request#Form], after populating it if needed,
// and returns its [core.Float] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func FormValuesFloat[T core.Float](req *http.Request, field string) (values []T, found bool, err error) {
	//
	ss, found, err := FormValues[string](req, field)
	if err == nil && found {
		values = make([]T, 0, len(ss))

		for _, s := range ss {
			v, err := ParseFloat[T](s)
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}

// FormValueFloatInRange reads a field from [http.Request#Form], after populating it if needed,
// and returns a [core.Float] value, an indicator saying if it was actually,
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type or if it's outside the specified
// boundaries.
func FormValueFloatInRange[T core.Float](req *http.Request, field string,
	min, max T) (value T, found bool, err error) {
	//
	value, found, err = FormValueFloat[T](req, field)
	if err == nil && found {
		if value < min || value > max {
			err = errRange("ParseFloat", FormatFloat(value, 'f', -1))
		}
	}
	return value, found, err
}

// FormValueBool reads a field from [http.Request#Form], after populating it if needed,
// and returns a [core.Bool] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func FormValueBool[T core.Bool](req *http.Request, field string) (value T, found bool, err error) {
	//
	s, found, err := FormValue[string](req, field)
	if err == nil && found {
		value, err = ParseBool[T](s)
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// FormValuesBool reads a field from [http.Request#Form], after populating it if needed,
// and returns its [core.Bool] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func FormValuesBool[T core.Bool](req *http.Request, field string) (values []T, found bool, err error) {
	//
	ss, found, err := FormValues[string](req, field)
	if err == nil && found {
		values = make([]T, 0, len(ss))

		for _, s := range ss {
			v, err := ParseBool[T](s)
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}

// PostFormValueSigned reads a field from [http.Request#PostForm], after populating it if needed,
// and returns a [core.Signed] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func PostFormValueSigned[T core.Signed](req *http.Request, field string,
	base int) (value T, found bool, err error) {
	//
	s, found, err := PostFormValue[string](req, field)
	if err == nil && found {
		value, err = ParseSigned[T](s, base)
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// PostFormValuesSigned reads a field from [http.Request#PostForm], after populating it if needed,
// and returns its [core.Signed] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func PostFormValuesSigned[T core.Signed](req *http.Request, field string,
	base int) (values []T, found bool, err error) {
	//
	ss, found, err := PostFormValues[string](req, field)
	if err == nil && found {
		values = make([]T, 0, len(ss))

		for _, s := range ss {
			v, err := ParseSigned[T](s, base)
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}

// PostFormValueSignedInRange reads a field from [http.Request#PostForm], after populating it if needed,
// and returns a [core.Signed] value, an indicator saying if it was actually,
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type or if it's outside the specified
// boundaries.
func PostFormValueSignedInRange[T core.Signed](req *http.Request, field string, base int,
	min, max T) (value T, found bool, err error) {
	//
	value, found, err = PostFormValueSigned[T](req, field, base)
	if err == nil && found {
		if value < min || value > max {
			err = errRange("ParseInt", FormatSigned(value, 10))
		}
	}
	return value, found, err
}

// PostFormValueUnsigned reads a field from [http.Request#PostForm], after populating it if needed,
// and returns a [core.Unsigned] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func PostFormValueUnsigned[T core.Unsigned](req *http.Request, field string,
	base int) (value T, found bool, err error) {
	//
	s, found, err := PostFormValue[string](req, field)
	if err == nil && found {
		value, err = ParseUnsigned[T](s, base)
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// PostFormValuesUnsigned reads a field from [http.Request#PostForm], after populating it if needed,
// and returns its [core.Unsigned] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func PostFormValuesUnsigned[T core.Unsigned](req *http.Request, field string,
	base int) (values []T, found bool, err error) {
	//
	ss, found, err := PostFormValues[string](req, field)
	if err == nil && found {
		values = make([]T, 0, len(ss))

		for _, s := range ss {
			v, err := ParseUnsigned[T](s, base)
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}

// PostFormValueUnsignedInRange reads a field from [http.Request#PostForm], after populating it if needed,
// and returns a [core.Unsigned] value, an indicator saying if it was actually,
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type or if it's outside the specified
// boundaries.
func PostFormValueUnsignedInRange[T core.Unsigned](req *http.Request, field string, base int,
	min, max T) (value T, found bool, err error) {
	//
	value, found, err = PostFormValueUnsigned[T](req, field, base)
	if err == nil && found {
		if value < min || value > max {
			err = errRange("ParseUint", FormatUnsigned(value, 10))
		}
	}
	return value, found, err
}

// PostFormValueFloat reads a field from [http.Request#PostForm], after populating it if needed,
// and returns a [core.Float] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func PostFormValueFloat[T core.Float](req *http.Request, field string) (value T, found bool, err error) {
	//
	s, found, err := PostFormValue[string](req, field)
	if err == nil && found {
		value, err = ParseFloat[T](s)
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// PostFormValuesFloat reads a field from [http.Request#PostForm], after populating it if needed,
// and returns its [core.Float] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func PostFormValuesFloat[T core.Float](req *http.Request, field string) (values []T, found bool, err error) {
	//
	ss, found, err := PostFormValues[string](req, field)
	if err == nil && found {
		values = make([]T, 0, len(ss))

		for _, s := range ss {
			v, err := ParseFloat[T](s)
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}

// PostFormValueFloatInRange reads a field from [http.Request#PostForm], after populating it if needed,
// and returns a [core.Float] value, an indicator saying if it was actually,
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type or if it's outside the specified
// boundaries.
func PostFormValueFloatInRange[T core.Float](req *http.Request, field string,
	min, max T) (value T, found bool, err error) {
	//
	value, found, err = PostFormValueFloat[T](req, field)
	if err == nil && found {
		if value < min || value > max {
			err = errRange("ParseFloat", FormatFloat(value, 'f', -1))
		}
	}
	return value, found, err
}

// PostFormValueBool reads a field from [http.Request#PostForm], after populating it if needed,
// and returns a [core.Bool] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func PostFormValueBool[T core.Bool](req *http.Request, field string) (value T, found bool, err error) {
	//
	s, found, err := PostFormValue[string](req, field)
	if err == nil && found {
		value, err = ParseBool[T](s)
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// PostFormValuesBool reads a field from [http.Request#PostForm], after populating it if needed,
// and returns its [core.Bool] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func PostFormValuesBool[T core.Bool](req *http.Request, field string) (values []T, found bool, err error) {
	//
	ss, found, err := PostFormValues[string](req, field)
	if err == nil && found {
		values = make([]T, 0, len(ss))

		for _, s := range ss {
			v, err := ParseBool[T](s)
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}
