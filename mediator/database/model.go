package database

import (
	"time"

	"github.com/gocql/gocql"
)

type MediatorDid struct {
	Id    int
	Did   string
	Added time.Time
}

type MediateeBase struct {
	RemoteDid  string            `json:"remoteDid" example:"did:peer:2.Ez6LSjrTnGzLHVRhrpAkubSd5Fs9355B454sfJAimQedtgJ.Vz6MkqLpAm2EKufwbMxXqXNZwSVxtTh3LdYB8Vp7MCoTTkSUq.SeyJ0IjoiZG0iLCJzIjp7InVyaSI6Imh0dHA6Ly9sb2NhbGhvc3Q6OTA5MC9tZXNzYWdlL3JlY2VpdmUiLCJhIjpbImRpZGNvbW0vdjIiXSwiciI6W119fQ"`
	Protocol   string            `json:"protocol" example:"nats"`
	Topic      string            `json:"topic" example:"topic-example"`
	Properties map[string]string `json:"properties" example:"did:remotedid:example:67890fghijk#key-2"`
	EventType  string            `json:"eventType"`
	Group      string            `json:"group"`
}

type Mediatee struct {
	RemoteDid     string            `json:"remoteDid" example:"did:peer:2.Ez6LSjrTnGzLHVRhrpAkubSd5Fs9355B454sfJAimQedtgJ.Vz6MkqLpAm2EKufwbMxXqXNZwSVxtTh3LdYB8Vp7MCoTTkSUq.SeyJ0IjoiZG0iLCJzIjp7InVyaSI6Imh0dHA6Ly9sb2NhbGhvc3Q6OTA5MC9tZXNzYWdlL3JlY2VpdmUiLCJhIjpbImRpZGNvbW0vdjIiXSwiciI6W119fQ"`
	RoutingKey    string            `json:"routingKey" example:"did::routingkey:example:12345abcde#key-1"`
	Protocol      string            `json:"protocol" example:"nats"`
	Topic         string            `json:"topic" example:"topic-example"`
	EventType     string            `json:"eventType"`
	Properties    map[string]string `json:"properties" example:"did:remotedid:example:67890fghijk#key-2"`
	RecipientDids []string          `json:"recipientDids" example:"did:recipientdid:example:12345abcde#key-1,did:recipientdid:example:12345abcde#key-2"`
	Added         time.Time         `json:"added" example:"2024-01-16 12:23:34.952000+0000"`
	Group         string            `json:"group"`
}

type Message struct {
	Id             gocql.UUID
	AttachmentId   string
	RecipientDid   string
	Description    string
	Filename       string
	MediaType      string
	Format         string
	LastmodTime    uint64
	ByteCount      uint64
	AttachmentData string
	Added          time.Time
}
