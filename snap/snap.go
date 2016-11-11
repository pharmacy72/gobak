package snap

import (
	"github.com/go-errors/errors"
)

var (
	ErrNotImplement = errors.New("SnapItem func Action not implement")
)

type Snapper interface {
	Send() error
}

type SnapItem struct {
	action func() error
}

func (s *SnapItem) Send() error {
	if s.action == nil {
		panic(ErrNotImplement)
	}
	return s.action()
}
