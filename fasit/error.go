package fasit

import (
	"fmt"
)

type appError struct {
	OriginalError error
	Message       string
	StatusCode    int
}

func (e appError) Code() int {
	return e.StatusCode
}

func (e appError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("%s: %s (%d)", e.Message, e.OriginalError.Error(), e.StatusCode)
	}

	return fmt.Sprintf("%s (%d)", e.Message, e.StatusCode)
}
