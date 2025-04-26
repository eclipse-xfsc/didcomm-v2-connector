package database

import (
	"errors"
	"strings"
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"

	"github.com/gocql/gocql"
)

type Cassandra struct {
	session  *gocql.Session
	keyspace string
}

func NewCassandra() *Cassandra {
	session, err := newCassandraSession()
	if err != nil {
		config.Logger.Error("NewCassandra", "Error creating session:", err)
		panic("Error creating cassandra session")
	}
	return &Cassandra{
		session:  session,
		keyspace: config.CurrentConfiguration.Database.Keyspace,
	}
}

// Mediator Did
func (db *Cassandra) GetMediatorDid() (string, error) {
	logTag := "GetMediatorDid"
	config.Logger.Info(logTag, "Start", true)

	var did string = ""
	query := "SELECT did FROM mediator_did LIMIT 1;"
	iter := db.session.Query(query).Iter()
	for iter.Scan(&did) {
	}

	if err := iter.Close(); err != nil {
		config.Logger.Error(logTag, "Error while closing iter", err)
		err = errors.New(logTag + ": Error while closing iter:" + err.Error())
		return "", err
	}
	config.Logger.Info(logTag, "End", true)
	return did, nil
}

func (db *Cassandra) StoreMediatorDid(mediatorDid string) (err error) {
	// The function inserts a dataset with id = 1, in order to avoid multiple datasets. Only one is needed
	logTag := "StoreMediatorDid"
	config.Logger.Info(logTag, "Start", true, "Did", mediatorDid)

	query := "INSERT INTO mediator_did (id, did, added) VALUES (1, ?, ?) ; "
	if err := db.session.Query(query, mediatorDid, time.Now()).Exec(); err != nil {
		config.Logger.Error(logTag, "Error while executing the query", err)
		return errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	return nil
}

// Connections (Mediatees)

func (db *Cassandra) GetMediatees(group *string) (datasets []Mediatee, err error) {
	logTag := "GetMediatees"
	config.Logger.Info(logTag, "Start", true)
	query := "SELECT * FROM mediatees ;"
	dbQuery := db.session.Query(query)
	if group != nil {
		query = "SELECT * FROM mediatees where group = ?"
		dbQuery = db.session.Query(query, group)
	}

	iter := dbQuery.Iter()
	datasets, err = readMediateeRows(iter)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return make([]Mediatee, 0), errors.New(logTag + ". Error: " + err.Error())
	}
	config.Logger.Info(logTag, "End", true)
	return datasets, nil
}

func (db *Cassandra) GetMediatee(remoteDid string) (*Mediatee, error) {
	logTag := "GetMediatee"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)

	dataset, err := db.getMediatee(remoteDid)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return nil, errors.New(logTag + ". Error: " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	return dataset, nil
}

func (db *Cassandra) UpdateMediatee(mediatee Mediatee) error {
	logTag := "UpdateMediatee"
	config.Logger.Info(logTag, "Start", true, "mediatee", mediatee)
	query := "UPDATE mediatees SET "
	var values []interface{} = make([]interface{}, 0)
	if mediatee.RoutingKey != "" {
		query = query + "routing_key=?,"
		values = append(values, mediatee.RoutingKey)
	}

	if mediatee.Protocol != "" {
		query = query + "protocol=?,"
		values = append(values, mediatee.Protocol)
	}

	if mediatee.RecipientDids != nil {
		query = query + "recipient_dids=?,"
		values = append(values, mediatee.RecipientDids)
	}

	if mediatee.Topic != "" {
		query = query + "topic=?,"
		values = append(values, mediatee.Topic)
	}

	if mediatee.Properties != nil {
		query = query + "properties=?,"
		values = append(values, mediatee.Properties)
	}

	if mediatee.EventType != "" {
		query = query + "eventtype=?,"
		values = append(values, mediatee.EventType)
	}

	if mediatee.Group != "" {
		query = query + "group=?,"
		values = append(values, mediatee.Group)
	}

	values = append(values, mediatee.RemoteDid)
	query = strings.Trim(query, ",") + " WHERE remote_did=?"
	if err := db.session.Query(query, values...).Exec(); err != nil {
		config.Logger.Error(logTag, "Error while executing the query", err)
		return errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	return nil

}

func (db *Cassandra) AddMediatee(mediatee Mediatee) error {
	logTag := "AddMediatee"
	config.Logger.Info(logTag, "Start", true, "mediatee", mediatee)

	query := "INSERT INTO mediatees (remote_did, routing_key, protocol, recipient_dids, topic, properties, added, eventtype,group) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ; "
	if err := db.session.Query(query, mediatee.RemoteDid, mediatee.RoutingKey, mediatee.Protocol, mediatee.RecipientDids, mediatee.Topic, mediatee.Properties, time.Now(), mediatee.EventType, mediatee.Group).Exec(); err != nil {
		config.Logger.Error(logTag, "Error while executing the query", err)
		return errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	return nil
}

func (db *Cassandra) DeleteMediatee(remoteDid string) error {
	logTag := "DeleteMediatee"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)

	query := "DELETE FROM mediatees WHERE remote_did = ? ; "
	if err := db.session.Query(query, remoteDid).Exec(); err != nil {
		config.Logger.Error(logTag, "Error", err)
		return errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	return nil
}

func (db *Cassandra) IsMediated(remoteDid string) (bool, error) {
	logTag := "IsMediated"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)

	m, err := db.getMediatee(remoteDid)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return false, errors.New(logTag + ". Error: " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	return m != nil, nil
}

// Block Connections (Mediatees)

func (db *Cassandra) BlockMediatee(remoteDid string) error {
	logTag := "BlockMediatee"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)

	query := "INSERT INTO blocked_dids (remote_did, added) VALUES (?, ?) ;"
	if err := db.session.Query(query, remoteDid, time.Now()).Exec(); err != nil {
		config.Logger.Error(logTag, "Error while executing the query", err)
		return errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
	}
	config.Logger.Info(logTag, "End", true)
	return nil
}

func (db *Cassandra) UnblockMediatee(remoteDid string) error {
	logTag := "UnblockMediatee"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)

	query := "DELETE FROM blocked_dids WHERE remote_did = ? ;"
	if err := db.session.Query(query, remoteDid).Exec(); err != nil {
		config.Logger.Error(logTag, "Error while executing the query", err)
		return errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
	}
	config.Logger.Info(logTag, "End", true)
	return nil
}

func (db *Cassandra) IsBlocked(remoteDid string) (bool, error) {
	logTag := "IsBlocked"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)

	query := "SELECT remote_did FROM blocked_dids WHERE remote_did = ? ;"
	iter := db.session.Query(query, remoteDid).Iter()
	var rd *string
	for iter.Scan(&rd) {
	}
	if err := iter.Close(); err != nil {
		config.Logger.Error(logTag, "Error closing iter:", err)
		return false, err
	}
	config.Logger.Info(logTag, "End", true)
	return rd != nil, nil
}

// Mediatees / RecipientDids

func (db *Cassandra) IsRecipientDidRegistered(recipientDid string) (bool, error) {
	logTag := "IsRecipientDidRegistered"
	config.Logger.Info(logTag, "Start", true, "recipientDid", recipientDid)

	rows := 0
	query := "SELECT COUNT(*) FROM mediatees WHERE recipient_dids contains ?; "
	iter := db.session.Query(query, recipientDid).Iter()
	iter.Scan(&rows)
	return rows > 0, nil
}

func (db *Cassandra) GetRecipientDids(remoteDid string) (recipientDids []string, err error) {
	logTag := "GetRecipientDids"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)

	mediatee, err := db.getMediatee(remoteDid)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return nil, errors.New(logTag + ". Error: " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	if mediatee == nil {
		config.Logger.Info(logTag, "No mediatee found in the database", true)
		return make([]string, 0), nil
	} else {
		return mediatee.RecipientDids, nil
	}
}

func (db *Cassandra) AddRecipientDid(remoteDid string, recipientDid string) (err error) {
	logTag := "AddRecipientDid"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid, "recipientDid", recipientDid)

	dataset, err := db.getMediatee(remoteDid)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return errors.New(logTag + ". Error: " + err.Error())
	}

	if dataset == nil {
		config.Logger.Warn(logTag, "Datasets count found with that remoteDid", 0)
	} else {
		query := "UPDATE mediatees SET recipient_dids = recipient_dids + { '" + recipientDid + "' } WHERE remote_did = ? ;"
		if err = db.session.Query(query, remoteDid).Exec(); err != nil {
			config.Logger.Error(logTag, "Error while executing the query", err)
			return errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
		}
	}
	config.Logger.Info(logTag, "End", true)
	return nil
}

func (db *Cassandra) DeleteRecipientDid(remoteDid string, recipientDid string) (err error) {
	logTag := "DeleteRecipientDid"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid, "recipientDid", recipientDid)

	dataset, err := db.getMediatee(remoteDid)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return errors.New(logTag + ". Error: " + err.Error())
	}

	if dataset == nil {
		config.Logger.Warn(logTag, "No datasets found with that remoteDid.", true)
	} else {
		query := "UPDATE mediatees SET recipient_dids = recipient_dids - { '" + recipientDid + "' } WHERE remote_did = ? ;"

		if err = db.session.Query(query, dataset.RemoteDid).Exec(); err != nil {
			config.Logger.Error(logTag, "Error while executing the query", err)
			err = errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
			return
		}
	}
	config.Logger.Info(logTag, "End", true)
	return nil
}

func (db *Cassandra) GetMediateeByRecipientDid(recipientDid string) (mediatee *Mediatee, err error) {
	logTag := "GetMediateeByRecipientDid"
	config.Logger.Info(logTag, "Start", true, "recipientDid", recipientDid)

	query := "SELECT * FROM mediatees WHERE recipient_dids contains ?; "
	iter := db.session.Query(query, recipientDid).Iter()
	datasets, err := readMediateeRows(iter)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return nil, errors.New(logTag + ". Error: " + err.Error())
	}

	if len(datasets) == 0 {
		config.Logger.Info(logTag, "End", "No mediatee found with that recipientDid")
		return nil, nil
	}
	config.Logger.Info(logTag, "End", true)
	return &datasets[0], nil

}

func (db *Cassandra) RecipientAndRemoteDidBelongTogether(recipientDid string, remoteDid string) (bool, error) {
	logTag := "RecipientAndRemoteDidBelongTogether"
	config.Logger.Info(logTag, "Start", true, "recipientDid", recipientDid, "remoteDid", remoteDid)

	query := "SELECT * FROM mediatees WHERE remote_did = ? AND recipient_dids contains ?; "
	iter := db.session.Query(query, remoteDid, recipientDid).Iter()
	datasets, err := readMediateeRows(iter)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return false, errors.New(logTag + ". Error: " + err.Error())
	}
	config.Logger.Info(logTag, "End", true)
	return len(datasets) > 0, nil
}

func (db *Cassandra) GetRoutingKey(remoteDid string) (routingKey string, err error) {
	logTag := "GetRoutingKey"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)

	mediatee, err := db.getMediatee(remoteDid)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return "", errors.New(logTag + ". Error: " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	if mediatee == nil {
		config.Logger.Info(logTag, "No mediatee found in the database", true)
		return "", nil
	} else {
		return mediatee.RoutingKey, nil
	}
}

func (db *Cassandra) SetRoutingKey(remoteDid string, routingKey string) (err error) {
	logTag := "SetRoutingKey"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid, "routingKey", routingKey)

	dataset, err := db.getMediatee(remoteDid)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return errors.New(logTag + ". Error: " + err.Error())
	}

	if dataset == nil {
		config.Logger.Warn(logTag, "Datasets count found with that remoteDid", 0)
	} else {
		query := "UPDATE mediatees SET routing_key = ? WHERE remote_did = ? ;"
		if err = db.session.Query(query, routingKey, remoteDid).Exec(); err != nil {
			config.Logger.Error(logTag, "Error while executing the query", err)
			return errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
		}
	}
	config.Logger.Info(logTag, "End", true)
	return nil
}

// Messages / Attachments

func (db *Cassandra) GetMessage(id string) (*Message, error) {
	logTag := "GetMessage"
	config.Logger.Info(logTag, "Start", true, "messageId", id)

	message, err := db.getMessage(id)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return nil, errors.New(logTag + ". Error: " + err.Error())
	}
	config.Logger.Info(logTag, "End", true)
	return message, nil
}

func (db *Cassandra) GetMessagesForRecipient(recipientDid string, limit int) (messages []didcomm.Attachment, err error) {
	logTag := "GetMessagesForRecipient"
	config.Logger.Info(logTag, "Start", true, "recipientDid", recipientDid, "limit", limit)

	var rDid string
	var message didcomm.Attachment
	var data didcomm.AttachmentDataBase64
	query := "SELECT recipient_did, id, description, filename, media_type, format, lastmod_time, byte_count, attachment_data " +
		"FROM messages WHERE recipient_did = ? LIMIT ? ; "
	iter := db.session.Query(query, recipientDid, limit).Iter()
	for iter.Scan(
		&rDid,
		&message.Id, &message.Description, &message.Filename, &message.MediaType, &message.Format, &message.LastmodTime, &message.ByteCount,
		&data.Value.Base64) {
		message.Data = data
		messages = append(messages, message)
	}

	if err := iter.Close(); err != nil {
		config.Logger.Error(logTag, "Error while closing iter", err)
		err = errors.New(logTag + ": Error while closing iter:" + err.Error())
		return make([]didcomm.Attachment, 0), err
	}
	config.Logger.Info(logTag, "End", true)
	return messages, nil
}

func (db *Cassandra) GetMessagesCountForRecipient(recipientDid string) (count int, err error) {
	logTag := "GetMessageCountForRecipient"
	config.Logger.Info(logTag, "Start", true, "recipientDid", recipientDid)

	rows := 0
	query := "SELECT COUNT(*) FROM messages WHERE recipient_did = ? ;"
	iter := db.session.Query(query, recipientDid).Iter()
	iter.Scan(&rows)
	config.Logger.Info(logTag, "End", true)
	return rows, nil
}

func (db *Cassandra) AddMessage(recipientDid string, message didcomm.Attachment) (err error) {
	logTag := "AddMessage"
	config.Logger.Info(logTag, "Start", message)

	query := "INSERT INTO messages " +
		"(id, recipient_did, description, filename, media_type, format, lastmod_time, byte_count, attachment_data, added) VALUES (now(), ?, ?, ?, ?, ?, ?, ?, ?, ?) ;"
	if err := db.session.Query(query, recipientDid, message.Description, message.Filename, message.MediaType,
		message.Format, message.LastmodTime, message.ByteCount, message.Data.(didcomm.AttachmentDataBase64).Value.Base64, time.Now()).Exec(); err != nil {
		config.Logger.Error(logTag, "Error while executing the query", err)
		return errors.New(logTag + ". Error while executing the query: '" + query + "'. " + err.Error())
	}
	config.Logger.Info(logTag, "End", true)
	return nil
}

func (db *Cassandra) DeleteMessagesByIds(messageIds []string) (deletedCount int, err error) {
	logTag := "DeleteMessagesByIds"
	config.Logger.Info(logTag, "Start", messageIds)

	foundDatasetsCount := 0
	query := "SELECT COUNT(*) FROM messages WHERE id IN ? ;"
	iter := db.session.Query(query, messageIds).Iter()
	iter.Scan(&foundDatasetsCount)
	if foundDatasetsCount == 0 {
		config.Logger.Info(logTag, "No datasets found. Deleted datasets' count", foundDatasetsCount)
	} else {
		config.Logger.Info(logTag, "Count of the datasets to be deleted", foundDatasetsCount)
		query = "DELETE FROM messages WHERE id in ? ;"
		if err = db.session.Query(query, messageIds).Exec(); err != nil {
			config.Logger.Error(logTag, "Error while executing the query", err)
			return 0, errors.New(logTag + ". Error while executing the query: " + query + ". " + err.Error())
		}
	}
	config.Logger.Info(logTag, "End", true)
	return foundDatasetsCount, nil
}

func (db *Cassandra) RemoteDidBelongsToMessage(remoteDid string, messageId string) (b bool, err error) {
	logTag := "RemoteDidBelongsToMessage"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid, "messageId", messageId)

	att, err := db.getMessage(messageId)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return false, errors.New(logTag + ". Error: " + err.Error())
	}

	if att == nil {
		return false, nil
	}

	query := "SELECT * FROM mediatees WHERE remote_did = ? AND recipient_dids CONTAINS ? ; "
	iter := db.session.Query(query, remoteDid, att.RecipientDid).Iter()
	datasets, err := readMediateeRows(iter)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return false, errors.New(logTag + ". Error: " + err.Error())
	}

	config.Logger.Info(logTag, "End", true)
	return len(datasets) > 0, nil
}

// Help Functions

func (db *Cassandra) getMediateeGroup(group string) (*Mediatee, error) {
	logTag := "getAttachmentById"
	config.Logger.Info(logTag, "Start", true, "group", group)
	query := "SELECT * FROM mediatees WHERE group = ? ; "
	iter := db.session.Query(query, group).Iter()
	datasets, err := readMediateeRows(iter)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return &Mediatee{}, err
	}
	switch len(datasets) {
	case 0:
		config.Logger.Info(logTag, "End", "No datasets found with that id")
		return nil, nil
	case 1:
		config.Logger.Info(logTag, "End", true)
		return &datasets[0], nil
	default:
		return nil, errors.New(logTag + ". End - Multiple datasets found with that id")
	}
}

func (db *Cassandra) getMediatee(remoteDid string) (*Mediatee, error) {
	logTag := "getAttachmentById"
	config.Logger.Info(logTag, "Start", true, "remoteDid", remoteDid)
	query := "SELECT * FROM mediatees WHERE remote_did = ? ; "
	iter := db.session.Query(query, remoteDid).Iter()
	datasets, err := readMediateeRows(iter)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return &Mediatee{}, err
	}
	switch len(datasets) {
	case 0:
		config.Logger.Info(logTag, "End", "No datasets found with that id")
		return nil, nil
	case 1:
		config.Logger.Info(logTag, "End", true)
		return &datasets[0], nil
	default:
		return nil, errors.New(logTag + ". End - Multiple datasets found with that id")
	}
}

func (db *Cassandra) Close() error {
	logTag := "Database Closing"
	config.Logger.Info(logTag, "Start", true)
	db.session.Close()
	config.Logger.Info(logTag, "End", true)
	return nil
}

func readMediateeRows(iter *gocql.Iter) (datasets []Mediatee, err error) {
	m := map[string]interface{}{}
	for iter.MapScan(m) {
		datasets = append(datasets, Mediatee{
			RemoteDid:     m["remote_did"].(string),
			RoutingKey:    m["routing_key"].(string),
			Protocol:      m["protocol"].(string),
			RecipientDids: m["recipient_dids"].([]string),
			Properties:    m["properties"].(map[string]string),
			Added:         m["added"].(time.Time),
			Topic:         m["topic"].(string),
			EventType:     m["eventtype"].(string),
			Group: 		   m["group"].(string),
		})
		m = map[string]interface{}{}
	}
	if err := iter.Close(); err != nil {
		err = errors.New("readMediateeRows: Error while closing iter:" + err.Error())
		return make([]Mediatee, 0), err
	}

	if len(datasets) == 0 {
		datasets = []Mediatee{}
	}
	return datasets, nil
}

func (db *Cassandra) getMessage(messageId string) (message *Message, err error) {
	logTag := "getAttachmentById"
	config.Logger.Info(logTag, "Start", true, "messageId", messageId)

	query := "SELECT recipient_did, id, description, filename, media_type, format, lastmod_time, byte_count, attachment_data " +
		"FROM messages WHERE id = ?; "

	iter := db.session.Query(query, messageId).Iter()
	datasets, err := readMessageRows(iter)

	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		return &Message{}, err
	}
	switch len(datasets) {
	case 0:
		config.Logger.Info(logTag, "End", "No datasets found with that id")
		return nil, nil
	case 1:
		config.Logger.Info(logTag, "End", true)
		return &datasets[0], nil
	default:
		return nil, errors.New(logTag + " - End - Multiple datasets found with that id")
	}
}

func readMessageRows(iter *gocql.Iter) (messages []Message, err error) {
	var message Message
	for iter.Scan(&message.RecipientDid,
		&message.Id, &message.Description, &message.Filename, &message.MediaType, &message.Format, &message.LastmodTime, &message.ByteCount,
		&message.AttachmentData) {
		messages = append(messages, message)
	}
	if err := iter.Close(); err != nil {
		err = errors.New("readMessageRows: Error while closing iter:" + err.Error())
		return make([]Message, 0), err
	}
	return messages, nil
}

func newCassandraSession() (*gocql.Session, error) {
	logTag := "Database session"

	dbConfig := config.CurrentConfiguration.Database

	cluster := gocql.NewCluster(dbConfig.Host)
	cluster.Port = dbConfig.Port
	cluster.Keyspace = dbConfig.Keyspace

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: dbConfig.User,
		Password: dbConfig.Password,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		config.Logger.Error(logTag, "Error while creating session", err)
		return nil, err
	}
	config.Logger.Info(logTag, "Initiated", true)
	return session, nil
}
