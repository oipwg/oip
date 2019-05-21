package oip042

import (
	"context"
	"encoding/json"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	//"sync"
	"fmt"
	"github.com/azer/logger"
	"gopkg.in/olivere/elastic.v6"
	jsonpatch "github.com/evanphx/json-patch"
//	"strings"
)

func init() {
	log.Info("init edit")
	events.SubscribeAsync("datastore:commit", onDatastoreCommit, false)
}

func onDatastoreCommit() {
	log.Info("Process Edits")

	edits, err := queryIncompleteEdits()
	if err != nil {
		log.Info("could not load edits", logger.Attrs{"err": err})
		return
	}

	for _, editRecord := range edits {
		artifactRecord, err := queryArtifact(editRecord.Meta.OriginalTxid)
		if err != nil {
			log.Info("Error getting original artifact", logger.Attrs{"err": err})
			log.Info(fmt.Sprintf("%+v\n", editRecord.Meta.OriginalTxid))
			continue
		}
		err = processRecord(editRecord, artifactRecord)
		if err != nil {
			log.Info("Error processing edit record", logger.Attrs{"err": err})
			log.Info(fmt.Sprintf("%+v\n", editRecord.Meta.Txid))
		}
		log.Info("Complete processing edit record")
		log.Info(fmt.Sprintf("%+v\n", editRecord.Meta.Txid))
	}
}

func queryIncompleteEdits() ([]*elasticOip042Edit, error) {
	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.completed", false),
	)

	search := datastore.Client().
		Search(datastore.Index(oip042EditIndex)).
		Type("_doc").
		Query(q).
		Sort("meta.time", true)

	results, err := search.Do(context.TODO())

	if err != nil {
		return nil, err
	}

	edits := []*elasticOip042Edit{}

	for _, v := range results.Hits.Hits {
		var editRecord *elasticOip042Edit
		err := json.Unmarshal(*v.Source, &editRecord)
		if err != nil {
			log.Info("failed to unmarshal elastic hit", logger.Attrs{"err": err})
			continue
		}
		
		edits = append(edits, editRecord)
	}

	return edits, nil
}

func queryArtifact(txid string) (*elasticOip042Artifact, error) {
	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.latest", true),
		elastic.NewTermQuery("meta.originalTxid", txid),
	)

	search := datastore.Client().
		Search(datastore.Index(oip042ArtifactIndex)).
		Type("_doc").
		Query(q).
		Sort("meta.time", true)

	results, err := search.Do(context.TODO())
	if err != nil {
		return nil, err
	}

	var artifactRecord *elasticOip042Artifact

	for _, v := range results.Hits.Hits {
		
		err := json.Unmarshal(*v.Source, &artifactRecord)

		if err != nil {
			log.Info("failed to unmarshal elastic hit", logger.Attrs{"err": err})
			return nil, err
		}
		
		break
	}

	return artifactRecord, nil	
}

func processRecord(editRecord *elasticOip042Edit, artifactRecord *elasticOip042Artifact) (error) {

	byteArtRecord, _ := json.Marshal(artifactRecord)

	spatchString := string(editRecord.Patch)

	editPatchString, err := UnSquashPatch(spatchString)

	editPatch, err := jsonpatch.DecodePatch([]byte(editPatchString))

	byteModifiedArtRecord, err := editPatch.Apply(byteArtRecord)

	var modifiedArtifactRecord *elasticOip042Artifact
	err = json.Unmarshal(byteModifiedArtRecord, &modifiedArtifactRecord)
	if err != nil {
		log.Info("could not apply patch", logger.Attrs{"err": err})
		return err
	}

	// Final updates
	artifactRecord.Meta.Latest = false

	cu := datastore.Client().Update().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(artifactRecord.Meta.Txid).Doc(artifactRecord).Refresh("wait_for")
	_, err = cu.Do(context.TODO())

	if err != nil {
		log.Info("Could not update latest artifact", logger.Attrs{"err": err})
		return err
	}

	modifiedArtifactRecord.Meta.Txid = editRecord.Meta.Txid
	modifiedArtifactRecord.Meta.Time = editRecord.Meta.Time

	ci := datastore.Client().Index().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(modifiedArtifactRecord.Meta.Txid).BodyJson(modifiedArtifactRecord).Refresh("wait_for")
	_, err = ci.Do(context.TODO())

	if err != nil {
		log.Info("Could not create modified record", logger.Attrs{"err": err})
		return err
	}

	editRecord.Meta.Completed = true

	cu = datastore.Client().Update().Index(datastore.Index(oip042EditIndex)).Type("_doc").Id(editRecord.Meta.Txid).Doc(editRecord).Refresh("wait_for")
	_, err = cu.Do(context.TODO())
	if err != nil {
		log.Info("Could update edit record", logger.Attrs{"err": err})
		return err
	}

	return nil
}

func UnSquashPatch(sp string) (string, error) {

	var p map[string]map[string]*json.RawMessage
	var up jsonpatch.Patch

	err := json.Unmarshal([]byte(sp), &p)
	if err != nil {
		return "", err
	}

	for op, updates := range p {
		for path, value := range updates {
			var row = make(map[string]*json.RawMessage)
			o := json.RawMessage([]byte(`"` + op + `"`))
			row["op"] = &o
			pp := json.RawMessage([]byte(`"/artifact` + path + `"`))
			row["path"] = &pp
			row["value"] = value
			up = append(up, row)
		}
	}

	usp, err := json.Marshal(&up)
	if err != nil {
		return "", err
	}

	return string(usp), nil
}