package services

import (
	types2 "closealerts/app/repositories/types"
	"context"
	"errors"
)

type Fakes struct {
	alerts chan types2.Alert
}

func NewFakes() Fakes {
	return Fakes{
		alerts: make(chan types2.Alert, 1),
	}
}

func (f Fakes) FakeAlert(_ context.Context, args string) error {
	select {
	case f.alerts <- types2.Alert{ID: args}:
	default:
		return errors.New("channel busy")
	}

	return nil
}

func (f Fakes) Alert(context.Context) (types2.Alert, bool) {
	select {
	case alert := <-f.alerts:
		return alert, true
	default:
		return types2.Alert{}, false
	}
}
