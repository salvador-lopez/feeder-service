package mongo

import (
	"context"
	"errors"
	"feeder-service/internal/sku/domain"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionName = "sku"

type SkuRepository struct {
	collection *mongo.Collection
	hydrator   *domain.Hydrator
}

var ErrMongoDBNil = fmt.Errorf("mongoDB is not defined")
func NewSkuRepository(db *mongo.Database, hydrator *domain.Hydrator) (*SkuRepository, error) {
	if db == (nil) {
		return nil, ErrMongoDBNil
	}
	return &SkuRepository{collection: db.Collection(collectionName), hydrator: hydrator}, nil
}

var ErrFind = fmt.Errorf("error during find execution")

func (r *SkuRepository) Find(ctx context.Context, id *domain.SkuId) (*domain.Sku, error) {
	var skuDTO *domain.SkuDTO

	result := r.collection.FindOne(ctx, bson.M{"_id": id.Value()})
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, result.Err()
	}

	err := result.Decode(&skuDTO)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFind, err.Error())
	}
	return r.hydrator.Hydrate(skuDTO), nil
}

var ErrSave = fmt.Errorf("error during save execution")

func (r *SkuRepository) Save(ctx context.Context, sku *domain.Sku) error {
	skuDTO := r.hydrator.Dehydrate(sku)
	_, err := r.collection.InsertOne(ctx, skuDTO)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", domain.ErrSkuAlreadyExists, err.Error())
		}
		return fmt.Errorf("%w: %s", ErrSave, err.Error())
	}

	return nil
}
