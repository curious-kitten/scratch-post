package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client creates a DB client
func Client(ctx context.Context, address string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(address))
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Collection creates a collection object for the DB
func Collection(dbName, collectionName string, client *mongo.Client) (*Data, error) {
	coll := client.Database(dbName).Collection(collectionName)
	_, err := coll.Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "identity.id", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{
					{Key: "projectId", Value: 1},
					{Key: "name", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return &Data{coll: coll}, nil
}

// Data is used to manipulate the collections
type Data struct {
	coll *mongo.Collection
}

// AddOne adds an item to the data collection
func (d *Data) AddOne(ctx context.Context, data interface{}) error {
	_, err := d.coll.InsertOne(ctx, data)
	return err
}

// GetAll returns all the items from a collection
func (d *Data) GetAll(ctx context.Context, items interface{}) error {
	cursor, err := d.coll.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, items); err != nil {
		return err
	}
	return nil
}

// Get returns a single item based on the item ID
func (d *Data) Get(ctx context.Context, id string, item interface{}) error {
	cursor := d.coll.FindOne(ctx, bson.M{"identity.id": id})
	if err := cursor.Decode(item); err != nil {
		return err
	}
	return nil
}

// Delete an item based on the item ID
func (d *Data) Delete(ctx context.Context, id string) error {
	res, err := d.coll.DeleteOne(ctx, bson.M{"identity.id": id})
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return err
}

// Update replaces the item with the given item ID with the provided one
func (d *Data) Update(ctx context.Context, id string, item interface{}) error {
	_, err := d.coll.ReplaceOne(ctx, bson.M{"identity.id": id}, item)
	return err
}

// IsNotFoundError checks if an error is no ducument error
func IsNotFoundError(err error) bool {
	return err == mongo.ErrNoDocuments
}

// IsDuplicateError checks if an error is a duplacte index error
func IsDuplicateError(err error) bool {
	if we, ok := err.(mongo.WriteException); ok {
		for _, e := range we.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}
	return false
}
