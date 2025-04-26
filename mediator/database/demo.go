package database

import (
	"errors"
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	secretsResolver "github.com/eclipse-xfsc/didcomm-v2-connector/mediator/secretsResolver"
)

// Table Structure
// Mediatee / Connection
// id				@id
// recipientDids	@did[] peer did created by client using routing key from mediator
// remoteDid		@did string peer did of the client and not created by the mediator
// routingKey		@string routing key created by the mediator
// created			@datetime

// Deny List
// For deviced outside of the network, the mediator will deny the connection
// id			@id
// did			@did string remote did of the client

// Secret
// Id             string
// Type           uint 		enums?
// SecretMaterial string

type DemoElement struct {
	message      didcomm.Attachment
	recipientDid string
}

type Demo struct {
	attachments []DemoElement
	mediatees   []Mediatee
	blockedDids []string
}

func NewDemo() *Demo {
	return &Demo{
		attachments: []DemoElement{},
		mediatees:   []Mediatee{},
		blockedDids: []string{},
	}
}

// Mediator Did
func (d *Demo) GetMediatorDid() (string, error) {
	// Did does not have meaningful service endpoint
	return secretsResolver.DID, nil
}

func (d *Demo) StoreMediatorDid(mediatorDid string) (err error) {
	// Do nothing, because DID is hardcoded in the demo database
	return nil
}

// Connections (Mediatees)
func (d *Demo) GetMediatees(group *string) ([]Mediatee, error) {
	return d.mediatees, nil
}

func (d *Demo) GetMediatee(remoteDid string) (*Mediatee, error) {
	var m Mediatee
	for _, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			m = mediatee
			return &m, nil
		}
	}
	return nil, errors.New("could not find mediatee")
}

func (d *Demo) AddMediatee(mediatee Mediatee) error {
	mediatee.Added = time.Now()
	d.mediatees = append(d.mediatees, mediatee)
	return nil
}

func (d *Demo) UpdateMediatee(mediatee Mediatee) error {
	for i, m := range d.mediatees {
		if m.RemoteDid == mediatee.RemoteDid {
			d.mediatees[i] = mediatee
			return nil
		}
	}
	return errors.New("could not find mediatee")
}

func (d *Demo) DeleteMediatee(remoteDid string) (err error) {
	for i, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			d.mediatees = DeleteIdFromSlice(d.mediatees, i)
			return nil
		}
	}
	return errors.New("could not find mediatee")
}

func (d *Demo) IsMediated(remoteDid string) (bool, error) {
	for _, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			return true, nil
		}
	}
	return false, nil
}

// Block Connections (Mediatees)
func (d *Demo) BlockMediatee(remoteDid string) error {
	d.blockedDids = append(d.blockedDids, remoteDid)
	return nil
}

func (d *Demo) UnblockMediatee(remoteDid string) error {
	for i, blockedDid := range d.blockedDids {
		if blockedDid == remoteDid {
			d.blockedDids = DeleteIdFromSlice(d.blockedDids, i)
			return nil
		}
	}
	return nil
}

func (d *Demo) IsBlocked(remoteDid string) (bool, error) {
	for _, blockedDid := range d.blockedDids {
		if blockedDid == remoteDid {
			return true, nil
		}
	}
	return false, nil
}

// Mediatees / RecipientDids
func (d *Demo) IsRecipientDidRegistered(recipientDid string) (bool, error) {
	for _, mediatee := range d.mediatees {
		for _, did := range mediatee.RecipientDids {
			if did == recipientDid {
				return true, nil
			}
		}
	}
	return false, nil
}

func (d *Demo) GetRecipientDids(remoteDid string) (recipientDids []string, err error) {

	for _, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			return mediatee.RecipientDids, nil
		}
	}

	return nil, nil
}

func (d *Demo) AddRecipientDid(remoteDid string, recipientDid string) error {
	for i, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			d.mediatees[i].RecipientDids = append(d.mediatees[i].RecipientDids, recipientDid)
			return nil
		}
	}
	return errors.New("could not find mediatee")
}

func (d *Demo) DeleteRecipientDid(remoteDid string, recipientDid string) error {
	for i, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			for j, recipient := range mediatee.RecipientDids {
				if recipient == recipientDid {
					d.mediatees[i].RecipientDids = DeleteIdFromSlice(mediatee.RecipientDids, j)
					return nil
				}
			}
		}
	}
	return errors.New("could not find mediatee")
}

func (d *Demo) GetMediateeByRecipientDid(recipientDid string) (*Mediatee, error) {
	for _, mediatee := range d.mediatees {
		for _, did := range mediatee.RecipientDids {
			if did == recipientDid {
				return &mediatee, nil
			}
		}
	}
	return nil, nil
}

func (d *Demo) RecipientAndRemoteDidBelongTogether(recipientDid string, remoteDid string) (bool, error) {
	for _, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			for _, did := range mediatee.RecipientDids {
				if did == recipientDid {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (d *Demo) SetRoutingKey(remoteDid string, routingKey string) error {
	for i, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			d.mediatees[i].RoutingKey = routingKey
			return nil
		}
	}
	return nil
}

func (d *Demo) GetRoutingKey(remoteDid string) (string, error) {
	for _, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			return mediatee.RoutingKey, nil
		}
	}
	return "", nil
}

// Messages / Attachments

func (d *Demo) GetMessage(id string) (*Message, error) {
	for _, message := range d.attachments {
		if *message.message.Id == id {
			return &Message{AttachmentId: *message.message.Id, RecipientDid: message.recipientDid, Description: *message.message.Description,
				Filename: *message.message.Filename, MediaType: *message.message.MediaType, Format: *message.message.Format, LastmodTime: *message.message.LastmodTime,
				ByteCount: *message.message.ByteCount, AttachmentData: message.message.Data.(didcomm.AttachmentDataBase64).Value.Base64}, nil
		}
	}
	return nil, errors.New("cannot find the message")
}

func (m *Demo) GetMessagesForRecipient(recipientDid string, limit int) ([]didcomm.Attachment, error) {
	messages := []didcomm.Attachment{}
	for _, e := range m.attachments {
		if e.recipientDid == recipientDid {
			messages = append(messages, e.message)
		}
	}
	return messages, nil
}

func (m *Demo) GetMessagesCountForRecipient(recipientDid string) (count int, err error) {
	count = 0
	for _, message := range m.attachments {
		if message.recipientDid == recipientDid {
			count++
		}
	}
	return count, nil
}

func (m *Demo) AddMessage(recipientDid string, message didcomm.Attachment) error {
	m.attachments = append(m.attachments, DemoElement{
		message:      message,
		recipientDid: recipientDid,
	})

	return nil
}

func (m *Demo) DeleteMessagesByIds(messageIds []string) (deletedCount int, err error) {
	count := 0
	for _, id := range messageIds {
		for i, e := range m.attachments {
			if *e.message.Id == id {
				m.attachments = DeleteIdFromSlice(m.attachments, i)
				count++
			}
		}
	}

	return count, nil
}

func (d *Demo) RemoteDidBelongsToMessage(remoteDid string, messageId string) (bool, error) {
	for _, e := range d.attachments {
		if *e.message.Id == messageId {
			recipientDid := e.recipientDid
			for _, mediatee := range d.mediatees {
				for _, r := range mediatee.RecipientDids {
					if r == recipientDid {
						if mediatee.RemoteDid == remoteDid {
							return true, nil
						} else {
							return false, nil
						}
					}
				}
			}
		}
	}
	return false, nil
}

func (d *Demo) Close() error {
	logTag := "Database Closing"
	config.Logger.Info(logTag, "Start", true)
	config.Logger.Info(logTag, "End", true)
	return nil
}

func DeleteIdFromSlice[T any](slice []T, id int) []T {
	if len := len(slice); len == 1 {
		return []T{}
	} else if len > 1 && id == len-1 {
		return slice[:id]
	} else if len > 1 && id == 0 {
		return slice[1:]
	} else {
		return append(slice[:id], slice[id+1:]...)
	}
}
