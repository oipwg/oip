package oip042

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/json-iterator/go"

	"github.com/oipwg/oip/flo"
	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	oipSync "github.com/oipwg/oip/sync"
	"github.com/azer/logger"
	"gopkg.in/olivere/elastic.v6"
	jsonpatch "github.com/evanphx/json-patch"

	"github.com/pkg/errors"
//	"strings"
)

var editCommitMutex sync.Mutex

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

	// Lock the edit mutex to prevent multiple batches running at the same time (causing race conditions)
	editCommitMutex.Lock()
	defer editCommitMutex.Unlock()

	// Lookup edits that have not been completed yet
	edits, err := queryIncompleteEdits()
	if err != nil {
		log.Info("Error while querying for Edits!", logger.Attrs{"err": err})
		return
	}

	// Check if there are edits that need to be completed
	if len(edits) > 0 {
		// Make sure that we are only processing a single Edit for each OriginalTXID
		editMap := make(map[string]bool)
		filteredEdits := []*elasticOip042Edit{}

		for _, editRecord := range edits {
			if !editMap[editRecord.Meta.OriginalTxid] {
				editMap[editRecord.Meta.OriginalTxid] = true
				filteredEdits = append(filteredEdits, editRecord)
			}
		}

		preFilteredLen := len(edits)

		edits = filteredEdits

		log.Info("Processing %d Edits... (filtered out %d)", len(edits), (preFilteredLen - len(edits)))

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
	// Since we have verified we only have a single result, access it directly
	v := results.Hits.Hits[0]
	err = json.Unmarshal(*v.Source, &artifactRecord)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal Elastic result into OIP042 Record! %v", err)
	}
	
	return artifactRecord, nil

}

func processRecord(editRecord *elasticOip042Edit, artifactRecord *elasticOip042Artifact) (error) {
	// SANITY CHECK
	// Make sure that the txid of the Record exists
	if artifactRecord.Meta.Txid == "" {
		log.Info("Unable to process Edit Record! Latest OIP042 Record is empty!")
		err := errors.New("Unable to process Edit Record! Latest OIP042 Record is empty!")
		return err
	}

	// Convert the Record interface to JSON
	jsonArtRecord, err := json.Marshal(artifactRecord)
	if err != nil {
		return fmt.Errorf("Could not JSON Marshal latest Record! %v", err)
	}

	// Convert the Edit Record to JSON
	jsonEditRecord, err := json.Marshal(editRecord)
	if err != nil {
		return fmt.Errorf("Could not JSON Marshal Edit Record! %v", err)
	}

	// Verify the Edit is being signed with the Address that owns the Original Record
	floAddress := jsoniter.Get(jsonArtRecord, "artifact", "floAddress").ToString()
	signature := editRecord.Meta.Signature
	preImageArray := []string{artifactRecord.Meta.OriginalTxid, strconv.FormatInt(jsoniter.Get(jsonEditRecord, "edit", "timestamp").ToInt64(), 10)}
	preImage := strings.Join(preImageArray, "-")

	signatureOk, err := flo.CheckSignature(floAddress, signature, preImage)
	if !signatureOk {
		return fmt.Errorf("Edit has invalid Signature! Address: %v | Preimage: %v | Signature: %v | Error: %v", floAddress, preImage, signature, err)
	}

	// Grab the patch string
	patchString := string(editRecord.Patch)
	// Unsquash the patch into a standard JSON Edit patch
	editPatchString, err := UnSquashPatch(patchString)
	if err != nil {
		return fmt.Errorf("Could not unsquash Edit patch! %v", err)
	}

	// Attempt to decode the patch
	editPatch, err := jsonpatch.DecodePatch([]byte(editPatchString))
	if err != nil {
		return fmt.Errorf("Could not decode Edit patch! %v", err)
	}

	// Apply the patch to the serialized Record
	jsonModifiedArtRecord, err := editPatch.Apply(jsonArtRecord)
	if err != nil {
		return fmt.Errorf("Could not apply Edit patch! %v", err)
	}

	// Verify the updated signature of a Record is valid!
	signature = jsoniter.Get(jsonModifiedArtRecord, "artifact", "signature").ToString()
	preImageArray = []string{
		jsoniter.Get(jsonModifiedArtRecord, "artifact", "storage", "location").ToString(), floAddress,
		strconv.FormatInt(jsoniter.Get(jsonModifiedArtRecord, "artifact", "timestamp").ToInt64(), 10)}
	preImage = strings.Join(preImageArray, "-")

	signatureOk, err = flo.CheckSignature(floAddress, signature, preImage)
	if !signatureOk {
		return fmt.Errorf("Editted Record has invalid Signature! Address: %v | Preimage: %v | Signature: %v | Error: %v", floAddress, preImage, signature, err)
	}

	// Load the patched Record into the OIP042 Struct 
	var modifiedArtifactRecord *elasticOip042Artifact
	err = json.Unmarshal(jsonModifiedArtRecord, &modifiedArtifactRecord)
	if err != nil {
		return fmt.Errorf("Could not unmarshal the patched Record into an OIP042 Record Struct! %v", err)
	}

	// Final updates
	artifactRecord.Meta.Latest = false

	// Run updates to set "latest" to false on the previously latest Record
	cu := datastore.Client().Update().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(artifactRecord.Meta.Txid).Doc(artifactRecord).Refresh("true")
	_, err = cu.Do(context.TODO())
	if err != nil {
		log.Info("Could not update latest artifact", logger.Attrs{"err": err})
		return err
	}

	// Update the metadata fields
	modifiedArtifactRecord.Meta.Txid = editRecord.Meta.Txid
	modifiedArtifactRecord.Meta.Time = editRecord.Meta.Time

	// Store the patched Record
	ci := datastore.Client().Index().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(modifiedArtifactRecord.Meta.Txid).BodyJson(modifiedArtifactRecord).Refresh("true")
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