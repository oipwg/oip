package validators

import (
	"github.com/json-iterator/go"
)

func init() {
	RegisterArtifactValidator("video", "*", videoWildcard{})
}

type videoWildcard struct{}

var _ ArtifactValidator = videoWildcard{}

func (v videoWildcard) IsValid(art *jsoniter.Any) Validity {

	return Valid
}
