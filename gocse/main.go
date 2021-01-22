package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mongodb-university/csfle-guides/gocse/csfle"
	"github.com/mongodb-university/csfle-guides/gocse/kms"
	"github.com/mongodb-university/csfle-guides/gocse/patient"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	keyVaultNamespace = "encryption.__keyVault"
	uri               = "mongodb+srv://testmongo:testmongo@cluster0.nre0n.mongodb.net/abc?retryWrites=true&w=majority"
	dbName            = "medicalRecords"
	collName          = "patients"
)

func localMasterKey() []byte {
	if _, err := os.Stat("master-key.txt"); err == nil {
		return getMasterKey()
	}
	key := make([]byte, 96)
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("Unable to create a random 96 byte data key: %v", err)
	}
	if err := ioutil.WriteFile("master-key.txt", key, 0644); err != nil {
		log.Fatalf("Unable to write key to file: %v", err)
	}
	return key
}

func getMasterKey() []byte {
	key, err := ioutil.ReadFile("master-key.txt")
	if err != nil {
		log.Fatalf("Could not read the key from master-key.txt: %v", err)
	}
	return key
}

func main() {
	err := godotenv.Load()

	preferredProvider := kms.AWSProvider()
	// preferredProvider := kms.AzureProvider()
	// preferredProvider := kms.GCPProvider()
	// preferredProvider := kms.LocalProvider(localMasterKey())

	// getting the base64 representation of a new data key
	clientEncryption, err := csfle.GetClientEncryption(keyVaultNamespace, uri, preferredProvider)
	//dataKey, clientEncryption, err := csfle.GetDataKey(keyVaultNamespace, uri, preferredProvider)
	defer func() {
		_ = clientEncryption.Close(context.TODO())
	}()
	if err != nil {
		log.Fatalf("problem during data key creation: %v", err)
	}

	// configuring our jsonSchema for automatic helpers
	// the driver only uses this for encryption information,
	// not to enforce schema constraints

	if err != nil {
		log.Panic(err)
	}

	// get a client with auto encryption using the new schema
	eclient, err := csfle.EncryptedClient(keyVaultNamespace, uri, preferredProvider)
	if err != nil {
		log.Panic(err)
	}
	// the auto encrypting client is ready for work
	defer func() {
		_ = eclient.Disconnect(context.TODO())
	}()
	encryptionOpts := options.Encrypt().
		SetAlgorithm("AEAD_AES_256_CBC_HMAC_SHA_512-Deterministic").
		SetKeyAltName("ongev")
	rawValueType, rawValueData, err := bson.MarshalValue("123456789")
	fmt.Println(rawValueType)
	rawValue := bson.RawValue{Type: rawValueType, Value: rawValueData}
	fmt.Println(rawValue)
	encryptedField, err := clientEncryption.Encrypt(context.TODO(), rawValue, encryptionOpts)
	fmt.Println(encryptedField)
	doc := patient.GetExamplePatient()
	doc.SSN = encryptedField
	// drop the collection and insert a new document with the encrypted client
	if err := csfle.InsertTestData(eclient, doc, dbName, collName); err != nil {
		log.Panic(err)
	}

	// creating an unencrypted client for a read operation to verify encryption
	uclient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Panicf("unencrypted client connect error %v", err)
	}
	defer func() {
		_ = uclient.Disconnect(context.TODO())
	}()

	FindDocument(eclient, dbName, collName)
	FindDocument(uclient, dbName, collName)

}

// FindDocument is just a wrapper around FindOne that accepts a client to perform
// the find operation with to illustrate CSFLE
func FindDocument(client *mongo.Client, dbName, collName string) {
	collection := client.Database(dbName).Collection(collName)
	cur, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var ele patient.Patient
		err := cur.Decode(&ele)
		if err != nil {
			log.Fatal(err)
		}

		b, err := json.Marshal(ele)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(b))
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	//Close the cursor once finished
	cur.Close(context.TODO())
}
