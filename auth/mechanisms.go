package auth

import (
	"encoding/base64"
	"errors"
	"strings"
)

// DecodePLAIN decodes a PLAIN hash into user/pass
//
// It assumes `authorization-id` to be unused, returning `authentication-id``
// as the username.
func DecodePLAIN(s string) (user, pass string, err error) {
	val, _ := base64.StdEncoding.DecodeString(s)
	bits := strings.Split(string(val), string(rune(0)))

	if len(bits) < 3 {
		err = errors.New("Badly formed parameter")
		return
	}

	user, pass = bits[1], bits[2]
	return
}
