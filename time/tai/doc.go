// Package tai provides time-related functions based on TAI.
//
// UTC conversion uses libtai's Unix epoch base and adds only the leap seconds
// beyond the initial ten-second TAI-UTC offset already included in that base.
//
// ParseTAIA and TAIA's text encoding methods read and write labels as @ followed
// by 16 hexadecimal seconds digits, eight nanoseconds digits and eight
// attoseconds digits. These are the same label fields as libtai's TAI64NA.
// Labels keep the seconds bias directly; they undergo no UTC conversion.
// Fractional fields must be below 1e9. A timestamp with negative seconds or
// seconds beyond uint64 cannot be written as a label, and fails with ErrTAIARange.
//
// String on TAI, TAIN and TAIA returns the lowercase label at its precision.
// Their signed 128-bit counts can exceed the label range. For those values,
// String returns diagnostic text such as TAIA(-1as), which is not a label and
// cannot be parsed by ParseTAIA. Use TAIA.MarshalText when writing a label
// requires an error for an out-of-range value.
package tai
