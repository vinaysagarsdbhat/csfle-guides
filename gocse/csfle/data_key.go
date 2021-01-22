package csfle

import (
	"context"
	"fmt"

	"github.com/mongodb-university/csfle-guides/gocse/kms"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetDataKey creates a new data key and returns the base64 encoding to be used
// in schema configuration for automatic encryption
func GetDataKey(keyVaultNamespace, uri string, provider kms.Provider) (primitive.Binary, *mongo.ClientEncryption, error) {

	// configuring encryption options by setting the keyVault namespace and the kms providers information
	// we configure this client to fetch the master key so that we can
	// create a data key in the next step
	clientEncryptionOpts := options.ClientEncryption().SetKeyVaultNamespace(keyVaultNamespace).SetKmsProviders(provider.Credentials())
	keyVaultClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return primitive.Binary{}, nil, fmt.Errorf("Client encryption connect error %v", err)
	}
	clientEnc, err := mongo.NewClientEncryption(keyVaultClient, clientEncryptionOpts)
	if err != nil {
		return primitive.Binary{}, nil, fmt.Errorf("NewClientEncryption error %v", err)
	}

	// specify the master key information that will be used to
	// encrypt the data key(s) that will in turn be used to encrypt
	// fields, and create the data key
	dataKeyOpts := options.DataKey().SetMasterKey(provider.DataKeyOpts()).SetKeyAltNames([]string{"ongev"})
	dataKeyID, err := clientEnc.CreateDataKey(context.TODO(), provider.Name(), dataKeyOpts)
	if err != nil {
		return primitive.Binary{}, nil, fmt.Errorf("create data key error %v", err)
	}

	return dataKeyID, clientEnc, nil
}

func GetClientEncryption(keyVaultNamespace, uri string, provider kms.Provider) (*mongo.ClientEncryption, error) {

	// configuring encryption options by setting the keyVault namespace and the kms providers information
	// we configure this client to fetch the master key so that we can
	// create a data key in the next step
	clientEncryptionOpts := options.ClientEncryption().SetKeyVaultNamespace(keyVaultNamespace).SetKmsProviders(provider.Credentials())
	keyVaultClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("Client encryption connect error %v", err)
	}
	clientEnc, err := mongo.NewClientEncryption(keyVaultClient, clientEncryptionOpts)
	if err != nil {
		return nil, fmt.Errorf("NewClientEncryption error %v", err)
	}

	// specify the master key information that will be used to
	// encrypt the data key(s) that will in turn be used to encrypt
	// fields, and create the data key

	return clientEnc, nil
}
