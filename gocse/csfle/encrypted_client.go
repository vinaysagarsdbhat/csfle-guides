package csfle

import (
	"context"
	"fmt"

	"github.com/mongodb-university/csfle-guides/gocse/kms"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EncryptedClient(keyVaultNamespace, uri string, provider kms.Provider) (*mongo.Client, error) {

	autoEncryptionOpts := options.AutoEncryption().
		SetKmsProviders(provider.Credentials()).
		SetKeyVaultNamespace(keyVaultNamespace).
		SetBypassAutoEncryption(true)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri).SetAutoEncryptionOptions(autoEncryptionOpts))
	if err != nil {
		return nil, fmt.Errorf("Connect error for encrypted client: %v", err)
	}
	return client, nil
}

func InsertTestData(client *mongo.Client, doc interface{}, dbName, collName string) error {
	collection := client.Database(dbName).Collection(collName)

	// if err := collection.Drop(context.TODO()); err != nil {
	// 	return fmt.Errorf("Drop error: %v", err)
	// }

	if _, err := collection.InsertOne(context.TODO(), doc); err != nil {
		return fmt.Errorf("InsertOne error: %v", err)
	}
	return nil
}
