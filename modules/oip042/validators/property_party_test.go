package validators

import (
	_ "github.com/bitspill/oip/config"
	"github.com/json-iterator/go"
	"testing"
)

func TestEmptyDetails(t *testing.T) {

	b := []byte(`{"oip042":{"publish":{"artifact":{"floAddress":"FTSTq8xx8yWUKJA5E3bgXLzZqqG9V6dvnr","timestamp":1525138120,"type":"property","subtype":"party","info":{"title":"the title","description":" "}
,"signature":"phonysignature"}}}}`)

	art := jsoniter.ConfigDefault.Get(b, "oip042","publish","artifact")
	v := IsValidArtifact("property", "party", &art, "")

	if v {
		t.Error("Expected invalid artifact")
	}

}

func TestEmptyModifiedDate(t *testing.T) {
	b := []byte(`{"oip042":{"publish":{"artifact":{"floAddress":"FTSTq8xx8yWUKJA5E3bgXLzZqqG9V6dvnr","timestamp":1525138120,"type":"property","subtype":"party","info":{"title":"the title","description":" "}
,"details":{"ns":"somethingnamespace","partyType":"INDIVIDUAL"},"signature":"phonysignature"}}}}`)

	art := jsoniter.ConfigDefault.Get(b, "oip042","publish","artifact")
	v := IsValidArtifact("property", "party", &art, "")

	if v {
		t.Error("Expected invalid artifact")
	}
}

