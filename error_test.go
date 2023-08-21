package inject

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestError(t *testing.T) {
	err := fmt.Errorf("%w: %v", ErrValueNotFound, reflect.TypeOf(""))
	expect(t, errors.Is(err, ErrValueNotFound), true)
	err = fmt.Errorf("%w: %v", ErrValueCanNotSet, reflect.TypeOf(""))
	expect(t, errors.Is(err, ErrValueCanNotSet), true)
}
