package validators

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("property", "tenure", propertyTenure{})
}

type propertyTenure struct{}

var _ ArtifactValidator = propertyTenure{}

type TenureDetails struct {
	Ns             string          `json:"ns,omitempty"`
	ModifiedDate   time.Time       `json:"modifiedDate,omitempty"`
	TenureType     string          `json:"tenureType,omitempty"`
	Parties        []PartyRole     `json:"parties"`
	SpatialUnits   []string        `json:"spatialUnits"`
	EffectiveDate  time.Time       `json:"effectiveDate,omitempty"`
	ExpirationDate time.Time       `json:"expirationDate,omniempty"`
	Attrs          json.RawMessage `json:"attrs,omitempty"`
}

func (r propertyTenure) IsValid(art *jsoniter.Any) (Validity, error) {
	var details = (*art).Get("details")
	if details.Size() == 0 {
		return Invalid, errors.New("Tenure details missing")
	}
	var td TenureDetails
	details.ToVal(&td)

	if td.ModifiedDate.IsZero() {
		return Invalid, errors.New("Tenure ModifiedDate missing")
	}

	if len(td.Ns) == 0 {
		return Invalid, errors.New("Tenure Ns missing")
	}

	return Valid, nil
}
