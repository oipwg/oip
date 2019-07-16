package oip042

import (
	"context"
	"encoding/json"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	oipSync "github.com/oipwg/oip/sync"
	"github.com/azer/logger"
	"gopkg.in/olivere/elastic.v6"
	jsonpatch "github.com/evanphx/json-patch"

	"github.com/pkg/errors"
//	"strings"
)

func init() {
	log.Info("init edit")
	// Subscribe to the datastore event emitter, run our edit processing on each datastore
	events.SubscribeAsync("datastore:commit", onDatastoreCommit, false)
}

func onDatastoreCommit() {
	// If we are still working on the initial sync, don't attempt to apply edits yet
	if oipSync.IsInitialSync {
		return
	}

	// Lookup edits that have not been completed yet
	edits, err := queryIncompleteEdits()
	if err != nil {
		log.Info("Error while querying for Edits!", logger.Attrs{"err": err})
		return
	}

	// Check if there are edits that need to be completed
	if len(edits) > 0 {
		log.Info("Processing %d Edits...", len(edits))

		// Iterate through each edit record and process each edit one at a time
		for _, editRecord := range edits {
			// First, lookup the latest record held in ElasticSearch
			latestRecord, err := queryArtifact(editRecord.Meta.OriginalTxid)
			if err != nil {
				log.Info("Error querying latest Record with txid %v for Edit %v! Error: %v", editRecord.Meta.OriginalTxid, editRecord.Meta.Txid, err)
				// If there was an error, go ahead and log the error but then attempt to continue processing the next edit
				continue
			}
			// Then, attempt to process the edit
			err = processRecord(editRecord, latestRecord)
			if err != nil {
				log.Info("Error processing Edit %v! Error: %v", editRecord.Meta.Txid, err)
				// Move on and attempt to process the next edit
				continue
			}

			log.Info("Edit %v on Record %v Successfully Processed!", editRecord.Meta.Txid, editRecord.Meta.OriginalTxid)
		}
	}
}

func queryIncompleteEdits() ([]*elasticOip042Edit, error) {
	// Create a search query for Edits that are not completed
	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.completed", false),
	)

	// Search for pending edits, sort by the given edit timestamp
	search := datastore.Client().
		Search(datastore.Index(oip042EditIndex)).
		Type("_doc").
		Query(q).
		Size(1000).
		Sort("edit.timestamp", true)

	// Perform the search
	results, err := search.Do(context.TODO())
	// Check for and return error
	if err != nil {
		log.Info("Error while querying for Incomplete Edits!", logger.Attrs{"err": err})
		return nil, err
	}

	// Create an array of OIP Edits
	edits := []*elasticOip042Edit{}
	// Iterage through each of the search results and attempt to "deserialize" it
	for _, v := range results.Hits.Hits {
		var editRecord *elasticOip042Edit
		err := json.Unmarshal(*v.Source, &editRecord)
		if err != nil {
			log.Info("Failed to unmarshal Elastic result into Edit Record!", logger.Attrs{"err": err})
			continue
		}
		
		// Append the latest edit record on top of the others
		edits = append(edits, editRecord)
	}

	return edits, nil
}

func queryArtifact(txid string) (*elasticOip042Artifact, error) {
	// Search for the latest record that has that original txid
	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.latest", true),
		elastic.NewTermQuery("meta.originalTxid", txid),
	)

	// Build the search
	search := datastore.Client().
		Search(datastore.Index(oip042ArtifactIndex)).
		Type("_doc").
		Query(q).
		Sort("meta.time", true)

	// Run the search
	results, err := search.Do(context.TODO())
	if err != nil {
		return nil, err
	}

	// SANITY CHECKS
	// Check if there were no search results
	if len(results.Hits.Hits) == 0 {
		log.Info("Failed to find OIP Record %v while processing Edits", txid)
		err := errors.New("Failed to lookup OIP Record!")
		return nil, err
	}
	// Check if we have more than one latest result (which should hopefully never happen)
	if len(results.Hits.Hits) > 1 {
		log.Info("Found more than one (%d) latest OIP Records for %v while processing Edits!", len(results.Hits.Hits), txid)
		err := errors.New("Found multiple latest OIP Records!")
		return nil, err
	}

	// Create the struct
	var artifactRecord *elasticOip042Artifact
	// Iterate through each hit, though we only expect a single result.
	for _, v := range results.Hits.Hits {
		err := json.Unmarshal(*v.Source, &artifactRecord)
		if err != nil {
			log.Info("Failed to unmarshal Elastic result into OIP042 Record!", logger.Attrs{"err": err})
			return nil, err
		}
		
		return artifactRecord, nil
	}

	// If we got here, then throw an error
	return nil, errors.New("Unknown error while attempting to query latest Record while processing Edits!")
}

func processRecord(editRecord *elasticOip042Edit, artifactRecord *elasticOip042Artifact) (error) {
	// SANITY CHECK
	// Make sure that the txid of the Record exists
	if artifactRecord.Meta.Txid == "" {
		log.Info("Unable to process Edit Record! Latest OIP042 Record is empty!")
		err := errors.New("Unable to process Edit Record! Latest OIP042 Record is empty!")
		return err
	}

	// Serialize the Record
	byteArtRecord, err := json.Marshal(artifactRecord)
	if err != nil {
		log.Info("Could not JSON Marshal latest Record!", logger.Attrs{"err": err})
		return err
	}

	// Grab the patch string
	spatchString := string(editRecord.Patch)
	// Unsquash the patch into a standard JSON Edit patch
	editPatchString, err := UnSquashPatch(spatchString)
	if err != nil {
		log.Info("Could not unsquash Edit patch!", logger.Attrs{"err": err})
		return err
	}

	// Attempt to decode the patch
	editPatch, err := jsonpatch.DecodePatch([]byte(editPatchString))
	if err != nil {
		log.Info("Could not decode Edit patch!", logger.Attrs{"err": err})
		return err
	}

	// Apply the patch to the serialized Record
	byteModifiedArtRecord, err := editPatch.Apply(byteArtRecord)
	if err != nil {
		log.Info("Could not apply Edit patch!", logger.Attrs{"err": err})
		return err
	}

	// Load the patched Record into the OIP042 Struct 
	var modifiedArtifactRecord *elasticOip042Artifact
	err = json.Unmarshal(byteModifiedArtRecord, &modifiedArtifactRecord)
	if err != nil {
		log.Info("Could not unmarshal the patched Record into an OIP042 Record Struct", logger.Attrs{"err": err})
		return err
	}

	// Final updates
	artifactRecord.Meta.Latest = false

	// Run updates to set "latest" to false on the previously latest Record
	cu := datastore.Client().Update().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(artifactRecord.Meta.Txid).Doc(artifactRecord).Refresh("false")
	_, err = cu.Do(context.TODO())
	if err != nil {
		log.Info("Could not update latest artifact", logger.Attrs{"err": err})
		return err
	}

	// Update the metadata fields
	modifiedArtifactRecord.Meta.Txid = editRecord.Meta.Txid
	modifiedArtifactRecord.Meta.Time = editRecord.Meta.Time

	// Store the patched Record
	ci := datastore.Client().Index().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(modifiedArtifactRecord.Meta.Txid).BodyJson(modifiedArtifactRecord).Refresh("wait_for")
	_, err = ci.Do(context.TODO())
	if err != nil {
		log.Info("Could not create modified record", logger.Attrs{"err": err})
		return err
	}

	// Set the Edit to be Complete
	editRecord.Meta.Completed = true

	// Update the Edit Record to be completed
	cu = datastore.Client().Update().Index(datastore.Index(oip042EditIndex)).Type("_doc").Id(editRecord.Meta.Txid).Doc(editRecord).Refresh("true")
	_, err = cu.Do(context.TODO())
	if err != nil {
		log.Info("Could update edit record", logger.Attrs{"err": err})
		return err
	}

	// Return nil if everything was successful
	return nil
}

type SquashPatch struct {
	Remove    []string       								 `json:"remove"`
	Replace   map[string]*json.RawMessage    `json:"replace"`
	Add   		map[string]*json.RawMessage    `json:"add"`
}

func UnSquashPatch(squashedPatchString string) (string, error) {
	// Create var to store squashedPatch
	var squashedPatch SquashPatch
	// Create unsquashed patch json object
	var up jsonpatch.Patch

	// Attempt to unmarshal the squashed patch
	err := json.Unmarshal([]byte(squashedPatchString), &squashedPatch)
	if err != nil {
		log.Info("Unable to Unsquash Patch! Patch Str: %v", squashedPatchString)
		return "", err
	}

	// Check if we have remove operations
	if len(squashedPatch.Remove) > 0 {
		// For each path in the string array, add it to the json patch
		for _, rmPath := range squashedPatch.Remove {
			var row = make(map[string]*json.RawMessage)
			o := json.RawMessage([]byte(`"remove"`))
			row["op"] = &o
			pp := json.RawMessage([]byte(`"/artifact` + rmPath + `"`))
			row["path"] = &pp
			up = append(up, row)
		}
	}

	// Check if we have replace operations
	if len(squashedPatch.Replace) > 0 {
		for path, value := range squashedPatch.Replace {
			var row = make(map[string]*json.RawMessage)
			o := json.RawMessage([]byte(`"replace"`))
			row["op"] = &o
			pp := json.RawMessage([]byte(`"/artifact` + path + `"`))
			row["path"] = &pp
			row["value"] = value
			up = append(up, row)
		}
	}

	// Check if we have add operations
	if len(squashedPatch.Add) > 0 {
		for path, value := range squashedPatch.Add {
			var row = make(map[string]*json.RawMessage)
			o := json.RawMessage([]byte(`"add"`))
			row["op"] = &o
			pp := json.RawMessage([]byte(`"/artifact` + path + `"`))
			row["path"] = &pp
			row["value"] = value
			up = append(up, row)
		}
	}

	// todo, handle `test`, `move`, and `copy` JSON Patch Operations

	// Attempt to turn unsquashed patch into a json string
	usp, err := json.Marshal(&up)
	if err != nil {
		return "", err
	}

	return string(usp), nil
}