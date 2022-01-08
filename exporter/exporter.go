package exporter

import (
	"context"

	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/cache"
	"github.com/moby/buildkit/solver"
)

type Exporter interface {
	Resolve(context.Context, map[string]string) (ExporterInstance, error)
}

type ExporterInstance interface {
	Name() string
	Config() Config
	Export(ctx context.Context, src Source, sessionID string) (*controlapi.ExporterResponse, error)
}

type Source struct {
	Ref      cache.ImmutableRef
	Refs     map[string]cache.ImmutableRef
	Metadata map[string][]byte
}

type Config struct {
	Compression solver.CompressionOpt
}
