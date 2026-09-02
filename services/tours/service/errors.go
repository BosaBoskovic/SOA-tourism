package service

import "errors"

// ErrForbidden indicates the caller is authenticated but not allowed to
// perform this action (e.g. not the tour's author).
var ErrForbidden = errors.New("forbidden")

func isOwner(authorID, callerUsername, callerRole string) bool {
	return callerRole == "admin" || (callerUsername != "" && callerUsername == authorID)
}
