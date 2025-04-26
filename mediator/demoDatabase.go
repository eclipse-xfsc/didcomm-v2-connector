package mediator

import (
	"errors"
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
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

type DemoDatabaseElement struct {
	message      didcomm.Attachment
	recipientDid string
}

type Mediatee struct {
	RecipientDids []string
	RemoteDid     string
	RoutingKey    string
	Created       time.Time
}

type DemoDatabase struct {
	attachments []DemoDatabaseElement
	mediatees   []Mediatee
}

func NewDemoDatabase() *DemoDatabase {
	return &DemoDatabase{
		attachments: []DemoDatabaseElement{},
		mediatees:   []Mediatee{},
	}
}

func (d *DemoDatabase) AddMediatee(remoteDid string, routingKey string) error {
	d.mediatees = append(d.mediatees, Mediatee{
		RecipientDids: []string{},
		RemoteDid:     remoteDid,
		RoutingKey:    routingKey,
		Created:       time.Now(),
	})
	return nil
}

func (d *DemoDatabase) IsMediated(remoteDid string) (bool, error) {
	for _, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			return true, nil
		}
	}
	return false, nil
}

func (d *DemoDatabase) IsRecipientDidRegistered(recipientDid string) (bool, error) {
	for _, mediatee := range d.mediatees {
		for _, did := range mediatee.RecipientDids {
			if did == recipientDid {
				return true, nil
			}
		}
	}
	return false, nil
}

func (d *DemoDatabase) RecipientAndRemoteDidBelongTogether(recipientDid string, remoteDid string) (bool, error) {
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

func (d *DemoDatabase) GetRecipientDids(remoteDid string) (recipientDids []string, err error) {

	for _, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			return mediatee.RecipientDids, nil
		}
	}

	return []string{}, errors.New("could not find mediatee")
}

func (d *DemoDatabase) AddRecipientDid(remoteDid string, recipientDid string) error {
	for i, mediatee := range d.mediatees {
		if mediatee.RemoteDid == remoteDid {
			d.mediatees[i].RecipientDids = append(d.mediatees[i].RecipientDids, recipientDid)
			return nil
		}
	}
	return errors.New("could not find mediatee")
}

func (d *DemoDatabase) DeleteRecipientDid(remoteDid string, recipientDid string) error {
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

func (d *DemoDatabase) RemoteDidBelongsToMessage(remoteDid string, messageId string) (bool, error) {
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

func (m *DemoDatabase) GetMessageCountForRecipient(recipientDid string) (count int, err error) {
	count = 0
	for _, message := range m.attachments {
		if message.recipientDid == recipientDid {
			count++
		}
	}
	return count, nil
}

func (m *DemoDatabase) GetMessagesForRecipient(recipientDid string, limit int) ([]didcomm.Attachment, error) {
	messages := []didcomm.Attachment{}
	for _, e := range m.attachments {
		if e.recipientDid == recipientDid {
			messages = append(messages, e.message)
		}
	}
	return messages, nil
}

func (m *DemoDatabase) DeleteMessagesByIds(messageIds []string) error {
	for _, id := range messageIds {
		for i, e := range m.attachments {
			if *e.message.Id == id {
				m.attachments = DeleteIdFromSlice(m.attachments, i)
			}
		}
	}

	return nil
}

func (m *DemoDatabase) AddMessage(recipientDid string, message didcomm.Attachment) error {
	m.attachments = append(m.attachments, DemoDatabaseElement{
		message:      message,
		recipientDid: recipientDid,
	})

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
