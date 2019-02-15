package validators

import (
	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("*", "*", catchAll{})
}

type catchAll struct{}

var _ ArtifactValidator = catchAll{}

func (a catchAll) IsValid(art *jsoniter.Any) (Validity, error) {

	return Valid, nil
}
