package sbomtypes

// SbomField defines the key of sbom.
const SbomField = "moby.buildkit.sbom.v1"

// ImageConfig defines the structure of sbom
// inside image config.
type ImageConfig struct {
	Sbom string `json:"moby.buildkit.sbom.v1,omitempty"`
}
