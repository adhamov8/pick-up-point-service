package usecase

import (
	"errors"
	"fmt"
)

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrInvalidInput          = errors.New("invalid input")
	ErrOrderAlreadyDelivered = errors.New("order already delivered")
	ErrOrderCannotBeRemoved  = errors.New("order cannot be removed")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrOrderExpired          = errors.New("order has expired")
	ErrReturnPeriodExpired   = errors.New("return period has expired")
	ErrOrderNotDelivered     = errors.New("order was not delivered")
	ErrCacheMiss             = fmt.Errorf("cache miss")
)
