package validators

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("property", "spatialunit", propertySpatialUnit{})
}

type propertySpatialUnit struct{}

var _ ArtifactValidator = propertySpatialUnit{}

type SpatialUnitDetails struct {
	Ns           string          `json:"ns,omitempty"`
	ModifiedDate time.Time       `json:"modifiedDate,omitempty"`
	SpatialType  string          `json:"spatialType,omitempty"`
	SpatialData  string          `json:"spatialData,omitempty"`
	TextualData  string          `json:"textualData,omitempty"`
	AddressData  string          `json:"addressData,omitempty"`
	OfficialId   string          `json:"officialID,omitempty"`
	ParentId     string          `json:"parentID,omitempty"`
	Attrs        json.RawMessage `json:"attrs,omitempty"`
}

func (r propertySpatialUnit) IsValid(art *jsoniter.Any) (Validity, error) {
	var details = (*art).Get("details")
	if details.Size() == 0 {
		return Invalid, errors.New("Spatial details missing")
	}
	var sd SpatialUnitDetails
	details.ToVal(&sd)

	if sd.ModifiedDate.IsZero() {
		return Invalid, errors.New("Spatial ModifiedDate missing")
	}

	if len(sd.Ns) == 0 {
		return Invalid, errors.New("Spatial Ns missing")
	}

	if len(sd.SpatialType) == 0 {
		return Invalid, errors.New("Spatial SpatialType missing")
	}

	return Valid, nil

}
