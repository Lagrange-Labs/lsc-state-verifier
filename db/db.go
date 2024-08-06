package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Lagrange-Labs/lagrange-node/logger"
	sequencertypes "github.com/Lagrange-Labs/lagrange-node/sequencer/types/v2"
	"github.com/Lagrange-Labs/lsc-state-verifier/stateproof"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database interface {
	GetBatchStateProof(chainID, batchNumber int64) (*stateproof.StateProof, error)
	UpdateLastProcessedBatch(chainID, batchNumber int64) error
	AddFailedBatch(chainID, batchNumber int64) error
	GetLastProcessedBatch(chainID int64) (int64, error)
}

type MongoDatabase struct {
	client *mongo.Client
}

func NewMongoDatabase(uri string) (*MongoDatabase, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("error connecting to DB: %s", err)
	}
	return &MongoDatabase{client: client}, nil
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

	cursor, err := m.client.Database("state").Collection("batches").Aggregate(ctx, pipeline)
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

func (m *MongoDatabase) UpdateLastProcessedBatch(chainID, batchNumber int64) error {
	logger.Infof("Updating last processed batch %d for chain %d in the database", batchNumber, chainID)
	filter := bson.M{"chain_id": chainID}
	update := bson.M{"$set": bson.M{"last_processed_batch": batchNumber}}
	_, err := m.client.Database("verifier").Collection("processed_batches").UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	return err
}

func (m *MongoDatabase) AddFailedBatch(chainID, batchNumber int64) error {
	logger.Infof("Adding failed batch %d for chain %d to the database", batchNumber, chainID)
	doc := bson.M{"chain_id": chainID, "batch_number": batchNumber}
	_, err := m.client.Database("verifier").Collection("batches").InsertOne(context.Background(), doc)
	return err
}

func (m *MongoDatabase) GetLastProcessedBatch(chainID int64) (int64, error) {
	logger.Infof("Fetching last processed batch for chain %d from the database", chainID)
	var result struct {
		LastProcessedBatch int64 `bson:"last_processed_batch"`
	}
	opts := options.FindOne().SetSort(bson.M{"last_processed_batch": -1})
	filter := bson.M{"chain_id": chainID}
	err := m.client.Database("verifier").Collection("processed_batches").FindOne(context.Background(), filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return 0, nil // No records found, start from the beginning
	} else if err != nil {
		return 0, err
	}
	logger.Infof("Last processed batch for chain %d is %d", chainID, result.LastProcessedBatch)
	return result.LastProcessedBatch, nil
}
