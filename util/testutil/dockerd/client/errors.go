package client

import (
	"fmt"

	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/errdefs"
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

func IsErrConnectionFailed(err error) bool {
	return errors.As(err, &errConnectionFailed{})
}

func ErrorConnectionFailed(host string) error {
	return errConnectionFailed{host: host}
}

func IsErrNotFound(err error) bool {
	if errdefs.IsNotFound(err) {
		return true
	}
	var e errdefs.ErrNotFound
	return errors.As(err, &e)
}

type objectNotFoundError struct {
	object string
	id     string
}

func (e objectNotFoundError) NotFound() {}

func (e objectNotFoundError) Error() string {
	return fmt.Sprintf("Error: No such %s: %s", e.object, e.id)
}

func (cli *Client) NewVersionError(APIrequired, feature string) error {
	if cli.version != "" && versions.LessThan(cli.version, APIrequired) {
		return fmt.Errorf("%q requires API version %s, but the Docker daemon API version is %s", feature, APIrequired, cli.version)
	}
	return nil
}
