package client

import (
	"fmt"

	"github.com/pkg/errors"
)

type errConnectionFailed struct {
	host string
}

func (err errConnectionFailed) Error() string {
	if err.host == "" {
		return "Cannot connect to the Docker daemon. Is the docker daemon running on this host?"
	}
	return fmt.Sprintf("Cannot connect to the Docker daemon at %s. Is the docker daemon running?", err.host)
}

func isErrConnectionFailed(err error) bool {
	return errors.As(err, &errConnectionFailed{})
}

func errorConnectionFailed(host string) error {
	return errConnectionFailed{host: host}
}
