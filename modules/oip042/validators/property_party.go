package validators

import (
	"encoding/json"

	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("property", "party", propertyParty{})
}

type propertyParty struct{}

var _ ArtifactValidator = propertyParty{}

type PartyDetails struct {
	Ns        string          `json:"ns,omitempty"`
	PartyRole string          `json:"partyRole,omitempty"`
	PartyType string          `json:"partyType,omitempty"`
	Tenures   []string        `json:"tenures,omitempty"`
	Groups    []string        `json:"groups,omitempty"`
	Members   []string        `json:"members,omitempty"`
	Attrs     json.RawMessage `json:"attrs,omitempty"`
}

func (r propertyParty) IsValid(art *jsoniter.Any) Validity {
	var pd PartyDetails
	(*art).Get("details").ToVal(&pd)

	// test validation on pd

	return Valid
}
