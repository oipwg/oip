package oip042

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/json-iterator/go"

	"github.com/azer/logger"
	jsonpatch "github.com/evanphx/json-patch"
	"gopkg.in/olivere/elastic.v6"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/flo"
	oipSync "github.com/oipwg/oip/sync"
)

var editCommitMutex sync.Mutex
var previousEditLength int

func init() {
	log.Info("init edit")
	// Subscribe to the datastore event emitter, run our edit processing on each datastore
	events.SubscribeAsync("datastore:commit", onDatastoreCommit, false)
}

func onDatastoreCommit() {
	// If we are still working on the initial sync and completing multiparts, 
	// don't attempt to apply edits yet.
	if !oipSync.MultipartSyncComplete {
		return
	}

	// Lock the edit mutex to prevent multiple batches running at the same time (causing race conditions)
	editCommitMutex.Lock()
	defer editCommitMutex.Unlock()

moreEdits:
	// Lookup edits that have not been completed yet
	edits, err := queryIncompleteEdits()
	if err != nil {
		log.Error("Error while querying for Edits!", logger.Attrs{"err": err})
		return
	}

	// Check if there are edits that need to be completed
	if len(edits) > 0 {
		oipSync.EditSyncComplete = false
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
				// If there was an error, go ahead and log the error but then attempt to continue processing the next edit
				log.Error("Error while querying latest Record with txid %v for Edit %v! Error: %v", editRecord.Meta.OriginalTxid, editRecord.Meta.Txid, err)

				// Check if we should mark this edit as invalid (if all multiparts are complete, and we still can't find the Record, then the Edit
				// txid is likely invalid and the Edit should be marked as invalid)
				if oipSync.MultipartSyncComplete {
					err = markEditInvalid(editRecord)
					if err != nil {
						log.Error("Error while marking Edit (%v) as invalid! Error: %v", editRecord.Meta.Txid, err)
					}
				}
				continue
			}
			// Then, attempt to process the edit
			err = processRecord(editRecord, latestRecord)
			if err != nil {
				log.Error("Error while processing Edit %v! Error: %v", editRecord.Meta.Txid, err)
				// Mark as invalid to prevent processing again in the future
				err = markEditInvalid(editRecord)
				if err != nil {
					log.Error("Error while marking Edit (%v) as invalid! Error: %v", editRecord.Meta.Txid, err)
				}

				// Move on and attempt to process the next edit
				continue
			}

			log.Info("Edit %v on Record %v Successfully Processed!", editRecord.Meta.Txid, editRecord.Meta.OriginalTxid)
		}

		// Check if there are any edits that got filtered, and also check to make sure that we have less edits to perform than the last time
		// if the number of edits this time is exactly the same as last time, don't loop
		if (preFilteredLen - len(edits) > 0 && previousEditLength != len(edits)) {
			// Since we have some edits we filtered out, go ahead and recursively loop back to complete them
			goto moreEdits
		}
	} else {
		oipSync.EditSyncComplete = true
	}
}

func queryIncompleteEdits() ([]*elasticOip042Edit, error) {
	// Create a search query for Edits that are not completed
	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.completed", false),
		elastic.NewTermQuery("meta.invalid", false),
	)

	// Search for pending edits, sort by the given edit timestamp
	search := datastore.Client().
		Search(datastore.Index(oip042EditIndex)).
		Type("_doc").
		Query(q).
		Size(10000).
		Sort("edit.timestamp", true)

	// Perform the search
	results, err := search.Do(context.TODO())
	// Check for and return error
	if err != nil {
		return nil, fmt.Errorf("Error while querying for Incomplete Edits! %v", err)
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
		return nil, fmt.Errorf("Failed to find OIP Record %v while processing Edits", txid)
	}
	// Check if we have more than one latest result (which should hopefully never happen)
	if len(results.Hits.Hits) > 1 {
		return nil, fmt.Errorf("Found more than one (%d) latest OIP Records for %v while processing Edits!", len(results.Hits.Hits), txid)
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

type Invalid struct {
	Invalid bool `json:"invalid"`
}
type MetaInvalid struct {
	Meta Invalid `json:"meta"`
}
type Latest struct {
	Latest bool `json:"latest"`
}
type MetaLatest struct {
	Meta Latest `json:"meta"`
}

func markEditInvalid(editRecord *elasticOip042Edit) error {
	cu := datastore.Client().Update().Index(datastore.Index(oip042EditIndex)).Type("_doc").Id(editRecord.Meta.Txid).Doc(MetaInvalid{Invalid{true}}).Refresh("true")
	_, err := cu.Do(context.TODO())
	if err != nil {
		return fmt.Errorf("Could not mark edit as invalid! %v", err)
	}

	log.Info("Marked Edit %v as Invalid!", editRecord.Meta.Txid)

	// Return nil if successful!
	return nil
}

func processRecord(editRecord *elasticOip042Edit, artifactRecord *elasticOip042Artifact) error {
	// SANITY CHECK
	// Make sure that the txid of the Record exists
	if artifactRecord.Meta.Txid == "" {
		return fmt.Errorf("Unable to process Edit Record! Latest OIP042 Record is empty!")
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

	// Attempt to decode the patch
	editPatch, err := jsonpatch.DecodePatch([]byte(editRecord.Patch))
	if err != nil {
		return fmt.Errorf("Could not decode Edit patch! %v", err)
	}

	// Prepend on "artifact" at the start of the edit patch to decend to the correct level
	for i, operation := range editPatch {
		quotedPath := *operation["path"]
		pathPrefix := []byte("\"/artifact")

		newPath := make([]byte, len(pathPrefix)+len(quotedPath)-1)
		copy(newPath, pathPrefix)
		copy(newPath[len(pathPrefix):], quotedPath[1:])

		editPatch[i]["path"] = (*json.RawMessage)(&newPath)
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

	// Run updates to set "latest" to false on the previously latest Record
	cu := datastore.Client().Update().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(artifactRecord.Meta.Txid).Doc(MetaLatest{Latest{false}}).Refresh("true")
	_, err = cu.Do(context.TODO())
	if err != nil {
		return fmt.Errorf("Could not update latest artifact! %v", err)
	}

	// Update the metadata fields
	modifiedArtifactRecord.Meta.Txid = editRecord.Meta.Txid
	modifiedArtifactRecord.Meta.Time = editRecord.Meta.Time

	// Store the patched Record
	ci := datastore.Client().Index().Index(datastore.Index(oip042ArtifactIndex)).Type("_doc").Id(modifiedArtifactRecord.Meta.Txid).BodyJson(modifiedArtifactRecord).Refresh("true")
	_, err = ci.Do(context.TODO())
	if err != nil {
		return fmt.Errorf("Could not create modified record! %v", err)
	}

	// Set the Edit to be Complete
	editRecord.Meta.Completed = true

	// Update the Edit Record to be completed
	cu = datastore.Client().Update().Index(datastore.Index(oip042EditIndex)).Type("_doc").Id(editRecord.Meta.Txid).Doc(editRecord).Refresh("true")
	_, err = cu.Do(context.TODO())
	if err != nil {
		return fmt.Errorf("Could update edit record! %v", err)
	}

	// Return nil if everything was successful
	return nil
}
