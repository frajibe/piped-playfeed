package utils

import (
	"fmt"
)

func WrapError(msg string, err error) error {
	return fmt.Errorf("%s > %w", msg, err)
}
