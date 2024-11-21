package password

import (
	"regexp"
)

var rePassword = regexp.MustCompile(`^[A-Za-z\d!@#$%^&*]{8,20}$`)

func VerifyPassword(password string) bool {
	return rePassword.MatchString(password)
}
