package ludumDareAPI

import "github.com/pkg/errors"

var (
	ErrNotAGame error = errors.New("requested element is not a game")
)
