package validators

import (
	"errors"

	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("research", "tomogram", researchTomogram{})
}

type researchTomogram struct{}

var _ ArtifactValidator = researchTomogram{}

type TomogramDetails struct {
	Date           int64   `json:"date,omitempty"`
	NCBItaxID      int64   `json:"NCBItaxID,omitempty"`
	TypoNBCI       int64   `json:"NBCItaxID,omitempty"`
	ArtNotes       string  `json:"artNotes,omitempty"`
	ScopeName      string  `json:"scopeName,omitempty"`
	Roles          string  `json:"roles,omitempty"`
	SpeciesName    string  `json:"speciesName,omitempty"`
	Strain         string  `json:"strain,omitempty"`
	TiltSingleDual int64   `json:"tiltSingleDual,omitempty"`
	Defocus        float64 `json:"defocus,omitempty"`
	Dosage         float64 `json:"dosage,omitempty"`
	TiltConstant   float64 `json:"tiltConstant,omitempty"`
	TiltMin        float64 `json:"tiltMin,omitempty"`
	TiltMax        float64 `json:"tiltMax,omitempty"`
	TiltStep       float64 `json:"tiltStep,omitempty"`
	Magnification  float64 `json:"magnification,omitempty"`
	Emdb           string  `json:"emdb,omitempty"`
	Microscopist   string  `json:"microscopist,omitempty"`
	Institution    string  `json:"institution,omitempty"`
	Lab            string  `json:"lab,omitempty"`
	Sid            string  `json:"sid,omitempty"`
}

func (r researchTomogram) IsValid(art *jsoniter.Any) (Validity, error) {
	var details = (*art).Get("details")
	if details.Size() == 0 {
		return Invalid, errors.New("tomogram details missing")
	}
	var td TomogramDetails
	details.ToVal(&td)

	if len(td.SpeciesName) == 0 {
		m := "tomogram: missing species name"
		log.Error(m)
		return Invalid, errors.New(m)
	}
	if td.Date <= 0 {
		m := "tomogram: invalid Date"
		log.Error(m)
		return Invalid, errors.New(m)
	}
	// ToDo: add mutator support
	// if td.NCBItaxID == 0 && td.TypoNBCI != 0 {
	// 	// many artifacts were published with NCBI misspelled
	// 	// so we will transparently alias it to be correct
	// 	td.NCBItaxID = td.TypoNBCI
	// 	td.TypoNBCI = 0
	// }
	if td.NCBItaxID < 0 {
		m := "tomogram: invalid NCBItaxID"
		log.Error(m)
		return Invalid, errors.New(m)
	}

	return Valid, nil
}
