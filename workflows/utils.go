package workflows

import (
	"errors"
	"math/rand/v2"
)

func failXOutOf10Times(x int) error {
	if rand.IntN(10) < x {
		return errors.New("service is down")
	}
	return nil
}
