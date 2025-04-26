package protocol

import (
	"log/slog"
	"testing"

	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"

	"github.com/stretchr/testify/assert"
)

func init() {
	config.Logger = slog.Default()
}

func TestUpdate_RemoveSuccess(t *testing.T) {

	// Sample data
	did := "sampleDID"
	dbKeys := []string{"key1"}
	newKeys := []Update{
		{RecipientDid: "key1", Action: "remove"},
	}

	updatedKeys, keysToAdd, keysToDelete, err := update(did, dbKeys, newKeys)

	// Check if the result matches the expected outcome
	expectedUpdatedKeys := []Update{
		{RecipientDid: "key1", Action: "remove", Result: "success"},
	}

	expectedKeysToAdd := []string{}
	expectedKeysToDelete := []string{"key1"}

	assert.Equal(t, expectedUpdatedKeys, updatedKeys)
	assert.Equal(t, expectedKeysToAdd, keysToAdd)
	assert.Equal(t, expectedKeysToDelete, keysToDelete)

	assert.Equal(t, err, nil)
}

func TestUpdate_AddSuccess(t *testing.T) {

	// Sample data
	did := "sampleDID"
	dbKeys := []string{"key1", "key2", "key3"}
	newKeys := []Update{
		{RecipientDid: "key5", Action: "add"},
	}

	updatedKeys, keysToAdd, keysToDelete, err := update(did, dbKeys, newKeys)

	// Check if the result matches the expected outcome
	expectedUpdatedKeys := []Update{
		{RecipientDid: "key5", Action: "add", Result: "success"},
	}

	expectedKeysToAdd := []string{"key5"}
	expectedKeysToDelete := []string{}

	assert.Equal(t, expectedUpdatedKeys, updatedKeys)
	assert.Equal(t, expectedKeysToAdd, keysToAdd)
	assert.Equal(t, expectedKeysToDelete, keysToDelete)

	assert.Equal(t, err, nil)
}

func TestUpdate_RemoveClientError(t *testing.T) {

	// Sample data
	did := "sampleDID"
	dbKeys := []string{"key1"}
	newKeys := []Update{
		{RecipientDid: "key2", Action: "remove"},
	}

	updatedKeys, keysToAdd, keysToDelete, err := update(did, dbKeys, newKeys)

	// Check if the result matches the expected outcome
	expectedUpdatedKeys := []Update{
		{RecipientDid: "key2", Action: "remove", Result: "client_error"},
	}

	expectedKeysToAdd := []string{}
	expectedKeysToDelete := []string{}

	assert.Equal(t, expectedUpdatedKeys, updatedKeys)
	assert.Equal(t, expectedKeysToAdd, keysToAdd)
	assert.Equal(t, expectedKeysToDelete, keysToDelete)

	assert.Equal(t, err, nil)
}

func TestUpdate_AddClientError(t *testing.T) {

	// Sample data
	did := "sampleDID"
	dbKeys := []string{"key1", "key2", "key3"}
	newKeys := []Update{
		{RecipientDid: "key1", Action: "add"},
	}

	updatedKeys, keysToAdd, keysToDelete, err := update(did, dbKeys, newKeys)

	// Check if the result matches the expected outcome
	expectedUpdatedKeys := []Update{
		{RecipientDid: "key1", Action: "add", Result: "no_changes"},
	}

	expectedKeysToAdd := []string{}
	expectedKeysToDelete := []string{}

	assert.Equal(t, expectedUpdatedKeys, updatedKeys)
	assert.Equal(t, expectedKeysToAdd, keysToAdd)
	assert.Equal(t, expectedKeysToDelete, keysToDelete)

	assert.Equal(t, err, nil)
}

func TestUpdate_UnknwonAction(t *testing.T) {

	// Sample data
	did := "sampleDID"
	dbKeys := []string{"key1", "key2", "key3"}
	newKeys := []Update{
		{RecipientDid: "key1", Action: "unknwonAction"},
	}

	updatedKeys, keysToAdd, keysToDelete, err := update(did, dbKeys, newKeys)

	// Check if the result matches the expected outcome
	expectedUpdatedKeys := []Update{}

	expectedKeysToAdd := []string{}
	expectedKeysToDelete := []string{}

	assert.Equal(t, expectedUpdatedKeys, updatedKeys)
	assert.Equal(t, expectedKeysToAdd, keysToAdd)
	assert.Equal(t, expectedKeysToDelete, keysToDelete)

	assert.Equal(t, err, nil)
}
