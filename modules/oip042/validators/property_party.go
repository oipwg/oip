package validators

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("property", "party", propertyParty{})
}

type propertyParty struct{}

var _ ArtifactValidator = propertyParty{}

type PartyRole struct {
	Party string `json:"party,omitempty"`
	Role  string `json:"role,omitempty"`
}

type PartyDetails struct {
	Ns           string          `json:"ns,omitempty"`
	ModifiedDate time.Time       `json:"modifiedDate,omitempty"`
	PartyType    string          `json:"partyType,omitempty"`
	Members      []PartyRole     `json:"members"`
	Attrs        json.RawMessage `json:"attrs,omitempty"`
}

func (r propertyParty) IsValid(art *jsoniter.Any) (Validity, error) {
	var details = (*art).Get("details")
	if details.Size() == 0 {
		return Invalid, errors.New("Party details missing")
	}
	var pd PartyDetails
	details.ToVal(&pd)

	if pd.ModifiedDate.IsZero() {
		return Invalid, errors.New("Party ModifiedDate missing")
	}

	if len(pd.Ns) == 0 {
		return Invalid, errors.New("Party Ns missing")
	}

	if len(pd.PartyType) == 0 {
		return Invalid, errors.New("Party PartyType missing")
	}

	return Valid, nil
}
