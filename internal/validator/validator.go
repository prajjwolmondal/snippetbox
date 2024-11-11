package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

// returns true if the FieldErrors map doesn't contain any entries.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// Adds an error message to the FieldErrors map (so long as no entry already exists
// for the given key).
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// Adds an error message to the FieldErrors map only if a validation check is not 'ok'.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// Returns true if the value contains no more than n characters.
func MaxChars(value string, n int) bool {
	// We’re using the utf8.RuneCountInString() function — not Go’s len() function. This
	// will count the number of Unicode code points in the title rather than the number of
	// bytes. For e.g. the string "Zoë" contains 3 Unicode code points, but 4 bytes
	// because of the umlauted ë character.
	return utf8.RuneCountInString(value) <= n
}

// Returns true if a value is in a list of specific permitted values.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Regex to sanity check the format of an email address. This returns a pointer to a
// 'compiled' regexp.Regexp type, or panics in the event of an error. Parsing this pattern
// once at startup and storing the compiled *regexp.Regexp in a variable is more performant
// than re-parsing the pattern each time we need it.
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Returns true if the given value contains at least n chars
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// Returns true if the given email is considered to be a valid email
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
