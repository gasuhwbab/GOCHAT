package proto

import "regexp"

var nick = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,20}$`)

func ValidNick(s string) bool {
	return nick.MatchString(s)
}
