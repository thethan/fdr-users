package draft

import "net/http"

type ErrorUpdateDraft struct {
	message string
}

func (e *ErrorUpdateDraft) Error() string {
	return "could not save draft result"
}

func (e *ErrorUpdateDraft) StatusCode() int {
	return http.StatusConflict
}