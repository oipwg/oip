package validators

import (
	"strings"

	"github.com/azer/logger"
	"github.com/json-iterator/go"
	"github.com/spf13/viper"
)

type Validity int

const (
	Valid   Validity = iota
	Invalid Validity = iota
)

type ArtifactValidator interface {
	IsValid(art *jsoniter.Any) Validity
}

var validators = make(map[string]ArtifactValidator)
var ignore = make(map[string]struct{})
var active = make(map[string]struct{})

// ToDo: extract functionality from init()
//       remove dependency upon github.com/bitspill/oip/config
func init() {
	log.Info("initializing active and ignore sets")
	i := viper.GetStringSlice("oip.artifact.ignore")
	for _, v := range i {
		ignore[v] = struct{}{}
	}
	a := viper.GetStringSlice("oip.artifact.active")
	for _, v := range a {
		active[v] = struct{}{}
	}
}

func RegisterArtifactValidator(artType, artSubType string, v ArtifactValidator) {
	k := getKey(artType, artSubType)

	_, ok := ignore[k]
	if ok {
		log.Info("validator is set to ignore", logger.Attrs{"validator": k, "ignore": ignore})
		return
	}
	_, ok = ignore[getKey(artType, "*")]
	if ok {
		log.Info("validator is set to ignore", logger.Attrs{"validator": k, "ignore": ignore})
		return
	}

	_, ok = active[k]
	if !ok {
		log.Info("validator is not active", logger.Attrs{"validator": k, "active": active})
		return
	}
	_, ok = validators[k]
	if ok {
		log.Info("existing validator being overwritten", logger.Attrs{"type": artType, "subtype": artSubType})
	}
	validators[k] = v
}

func IsValidArtifact(artType, artSubType string, art *jsoniter.Any, txid string) bool {
	logReturn := func(validity Validity) bool {
		if validity == Invalid {
			log.Error("artifact invalid", logger.Attrs{"txid": txid, "type": artType, "subtype": artSubType})
		}
		return validity == Valid
	}

	if value, ok := validators[getKey(artType, artSubType)]; ok {
		v := value.IsValid(art)
		return logReturn(v)
	}

	if value, ok := validators[getKey(artType, "*")]; ok {
		v := value.IsValid(art)
		return logReturn(v)
	}

	if value, ok := validators[getKey("*", "*")]; ok {
		v := value.IsValid(art)
		return logReturn(v)
	}

	if viper.GetBool("oip.artifact.strict") {
		return logReturn(Invalid)
	} else {
		return logReturn(Valid)
	}
}

func getKey(artType, artSubType string) string {
	return strings.ToLower(artType + "-" + artSubType)
}
