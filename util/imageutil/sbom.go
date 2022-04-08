package imageutil

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"io"

	sbomtypes "github.com/moby/buildkit/util/sbom/types"
	"github.com/pkg/errors"
)

// Sbom returns sbom from image config.
func Sbom(dt []byte) ([]byte, error) {
	if len(dt) == 0 {
		return nil, nil
	}

	var config sbomtypes.ImageConfig
	if err := json.Unmarshal(dt, &config); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal image config")
	}
	if len(config.Sbom) == 0 {
		return nil, nil
	}

	dtsbom, err := base64.StdEncoding.DecodeString(config.Sbom)
	if err != nil {
		return nil, err
	}

	var sbom bytes.Buffer
	b := bytes.NewReader(dtsbom)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(&sbom, r); err != nil {
		return nil, err
	}
	if err := r.Close(); err != nil {
		return nil, err
	}

	return sbom.Bytes(), nil
}
