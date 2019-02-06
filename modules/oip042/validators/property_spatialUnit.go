package validators

import (
	"encoding/json"

	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("property", "spatialunit", propertySpatialUnit{})
}

type propertySpatialUnit struct{}

var _ ArtifactValidator = propertySpatialUnit{}

type SpatialUnitDetails struct {
	Ns           string          `json:"ns,omitempty"`
	Geometry     json.RawMessage `json:"geometry,omitempty"`
	SpatialType  string          `json:"spatialType,omitempty"`
	SpatialUnits []string        `json:"spatialUnits,omitempty"`
	BBox         []float64       `json:"bbox,omitempty"`
	Attrs        json.RawMessage `json:"attrs,omitempty"`
}

func (r propertySpatialUnit) IsValid(art *jsoniter.Any) Validity {
	var sud SpatialUnitDetails
	(*art).Get("details").ToVal(&sud)

	// test validation on sud

	return Valid
}
