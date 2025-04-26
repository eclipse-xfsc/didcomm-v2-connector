package database

import "github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"

type Adapter interface {
	// Mediator Did
	GetMediatorDid() (mediatorDid string, err error)
	StoreMediatorDid(mediatorDid string) (err error)
	// Connections (Mediatees)
	GetMediatees(group *string) ([]Mediatee, error)
	GetMediatee(remoteDid string) (*Mediatee, error)
	UpdateMediatee(mediatee Mediatee) error
	AddMediatee(mediatee Mediatee) (err error)
	DeleteMediatee(remoteDid string) (err error)
	IsMediated(remoteDid string) (isMediated bool, err error)
	// Block Connections (Mediatees)
	BlockMediatee(remoteDid string) error
	UnblockMediatee(remoteDid string) error
	IsBlocked(remoteDid string) (bool, error)
	// Mediatees / RecipientDids
	IsRecipientDidRegistered(recipientDid string) (isRecDidRegistered bool, err error)
	GetRecipientDids(remoteDid string) (recipientDids []string, err error)
	AddRecipientDid(remoteDid string, recipientDid string) error
	DeleteRecipientDid(remoteDid string, recipientDid string) error
	GetMediateeByRecipientDid(recipientDid string) (mediatee *Mediatee, err error)

	RecipientAndRemoteDidBelongTogether(recipientDid string, remoteDid string) (bool, error)

	SetRoutingKey(remoteDid string, routingKey string) (err error)
	GetRoutingKey(remoteDid string) (routingKey string, err error)

	// Messages / Attachments
	GetMessage(id string) (*Message, error)
	GetMessagesForRecipient(recipientDid string, limit int) ([]didcomm.Attachment, error)
	GetMessagesCountForRecipient(recipientDid string) (count int, err error)
	AddMessage(recipientDid string, message didcomm.Attachment) error
	DeleteMessagesByIds(messageIds []string) (int, error)
	RemoteDidBelongsToMessage(remoteDid string, messageId string) (bool, error)

	Close() error
}
