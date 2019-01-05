package validators

import (
	"encoding/json"

	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("property", "tenure", propertyTenure{})
}

type propertyTenure struct{}

var _ ArtifactValidator = propertyTenure{}

type TenureDetails struct {
	Ns           string          `json:"ns"`
	TenureType   string          `json:"tenureType"`
	Tenures      []string        `json:"tenures"`
	SpatialUnits []string        `json:"spatialUnits"`
	Attrs        json.RawMessage `json:"attrs"`
}

func (r propertyTenure) IsValid(art *jsoniter.Any) Validity {
	var td TenureDetails
	(*art).Get("details").ToVal(&td)

	// test validation on td

	return Valid
}
