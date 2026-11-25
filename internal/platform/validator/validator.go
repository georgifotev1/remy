package validator

import (
	"regexp"
	"slices"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool { return len(v.Errors) == 0 }

func (v *Validator) AddError(key, msg string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = msg
	}
}

func (v *Validator) Check(ok bool, key, msg string) {
	if !ok {
		v.AddError(key, msg)
	}
}

func NotBlank(s string) bool                   { return s != "" }
func MaxChars(s string, n int) bool            { return utf8.RuneCountInString(s) <= n }
func MinChars(s string, n int) bool            { return utf8.RuneCountInString(s) >= n }
func Matches(s string, rx *regexp.Regexp) bool { return rx.MatchString(s) }
func In[T comparable](v T, allowed ...T) bool  { return slices.Contains(allowed, v) }
