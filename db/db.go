package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	sequencertypes "github.com/Lagrange-Labs/lagrange-node/sequencer/types/v2"
	"github.com/Lagrange-Labs/state-verifier/stateproof"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database interface {
    GetBatchStateProof(chainID, batchNumber int64) (*stateproof.StateProof, error)
    // UpdateLastProcessedBatch(chainID, batchNumber int64) error
    // AddFailedBatch(chainID, batchNumber int64) error
}

type MongoDatabase struct {
    client *mongo.Client
    db     *mongo.Database
}

func NewMongoDatabase(uri, dbName string) (*MongoDatabase, error) {
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
    if err != nil {
        return nil, fmt.Errorf("error connecting to DB: %s", err)
    }
    db := client.Database(dbName)
    return &MongoDatabase{client: client, db: db}, nil
}

func (m *MongoDatabase) GetBatchStateProof(chainId, batchNumber int64) (*stateproof.StateProof, error) {
    logger.Infof("Received GetBatchStateProofs request for chainId %d and batchNumber %d", chainId, batchNumber)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.D{
				{Key: "batch_header.batch_number", Value: batchNumber},
				{Key: "batch_header.chain_id", Value: chainId},
			}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "committee_roots"},
				{Key: "let", Value: bson.D{
					{Key: "current_committee", Value: "$committee_header.current_committee"},
				}},
				{Key: "pipeline", Value: mongo.Pipeline{
					{
						{Key: "$match", Value: bson.D{{Key: "$expr", Value: bson.D{
							{Key: "$and", Value: bson.A{
								bson.D{{Key: "$eq", Value: bson.A{"$current_committee_root", "$$current_committee"}}},
								bson.D{{Key: "$eq", Value: bson.A{"$chain_id", chainId}}},
							}},
						}}}},
					},
					{
						{Key: "$limit", Value: 1},
					},
				}},
				{Key: "as", Value: "committee"},
			}},
		},
		{
			{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$committee"},
				{Key: "preserveNullAndEmptyArrays", Value: false},
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "batch_header", Value: 1},
				{Key: "pub_keys", Value: 1},
				{Key: "operators", Value: 1},
				{Key: "committee_header", Value: 1},
				{Key: "agg_signature", Value: 1},
				{Key: "committee_info", Value: "$committee"},
			}},
		},
	}

	cursor, err := m.db.Collection("batches").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result map[string]interface{}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}

		stateProof := &stateproof.StateProof{}

		// Extract and assign values to StateProof fields
		if aggSig, ok := result["agg_signature"].(string); ok {
			stateProof.AggregatedSignature = aggSig
		}

		if batchHeader, ok := result["batch_header"]; ok {
			mBatchHeader := batchHeader.(map[string]interface{})
			stateProof.BatchHeader = &sequencertypes.BatchHeader{}
			stateProof.BatchHeader.BatchNumber = uint64(mBatchHeader["batch_number"].(int64))
			stateProof.BatchHeader.L1BlockNumber = uint64(mBatchHeader["l1_block_number"].(int64))
			stateProof.BatchHeader.L1TxHash = mBatchHeader["l1_tx_hash"].(string)
			stateProof.BatchHeader.L1TxIndex = uint32(mBatchHeader["l1_tx_index"].(int64))
			stateProof.BatchHeader.ChainId = uint32(mBatchHeader["chain_id"].(int64))
			stateProof.BatchHeader.L2Blocks = make([]*sequencertypes.BlockHeader, 0)
			for _, l2Block := range mBatchHeader["l2_blocks"].(primitive.A) {
				mL2Block := l2Block.(map[string]interface{})
				stateProof.BatchHeader.L2Blocks = append(stateProof.BatchHeader.L2Blocks, &sequencertypes.BlockHeader{
					BlockNumber: uint64(mL2Block["block_number"].(int64)),
					BlockHash:   mL2Block["block_hash"].(string),
				})
			}
		}

		if committeeHeader, ok := result["committee_header"]; ok {
			mCommitteeHeader := committeeHeader.(map[string]interface{})
			stateProof.CommitteeHeader = &sequencertypes.CommitteeHeader{}
			stateProof.CommitteeHeader.CurrentCommittee = mCommitteeHeader["current_committee"].(string)
			stateProof.CommitteeHeader.NextCommittee = mCommitteeHeader["next_committee"].(string)
			stateProof.CommitteeHeader.TotalVotingPower = uint64(mCommitteeHeader["total_voting_power"].(int64))
		}

		committeeInfo, ok := result["committee_info"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to parse committee info")
		}

		// Extract operators, public keys, and voting powers
		if operators, ok := committeeInfo["operators"].(primitive.A); ok {
			for _, operator := range operators {
				op, ok := operator.(map[string]interface{})
				if !ok {
					continue
				}
				if stakeAddress, ok := op["stake_address"].(string); ok {
					stateProof.OperatorAddresses = append(stateProof.OperatorAddresses, stakeAddress)
				}
				if publicKey, ok := op["public_key"].(string); ok {
					stateProof.BLSPublicKeys = append(stateProof.BLSPublicKeys, publicKey)
				}
				if vp, ok := op["voting_power"].(int64); ok {
					stateProof.VotingPowers = append(stateProof.VotingPowers, uint64(vp))
				}
			}
		}

		// Calculate aggregation bits
		stateProof.AggregationBits = make([]uint8, len(stateProof.BLSPublicKeys))
		batchOperatorAddresses, ok := result["operators"].(primitive.A)
		if !ok {
			logger.Errorf("Failed to retrieve 'operators' as a primitive.A slice ", result["operators"])
			return nil, fmt.Errorf("failed to parse operators")
		}
		if pubKeys, ok := result["pub_keys"].(primitive.A); ok {
			for i, publicKey := range stateProof.BLSPublicKeys {
				found := false
				for j, batchKey := range pubKeys {
					if bKey, ok := batchKey.(string); ok && bKey == publicKey && batchOperatorAddresses[j] == stateProof.OperatorAddresses[i] {
						found = true
						break
					}
				}
				if found {
					stateProof.AggregationBits[i] = 1
				} else {
					stateProof.AggregationBits[i] = 0
				}
			}
		}

		return stateProof, nil
	}

	return nil, fmt.Errorf("no results found for chainId %d and batchNumber %d", chainId, batchNumber)

}

// func (m *MongoDatabase) UpdateLastProcessedBatch(chainID, batchNumber int64) error {
//     // Implementation of updating last processed batch
// }

// func (m *MongoDatabase) AddFailedBatch(chainID, batchNumber int64) error {
//     // Implementation of adding failed batch
// }