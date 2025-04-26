package mediator

import "github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"

type DatabaseAdapter interface {
	AddMediatee(remoteDid string, routingKey string) error
	IsMediated(remoteDid string) (bool, error)
	IsRecipientDidRegistered(recipientDid string) (bool, error)
	GetRecipientDids(remoteDid string) (recipientDids []string, err error)
	RecipientAndRemoteDidBelongTogether(recipientDid string, remoteDid string) (bool, error)
	AddRecipientDid(remoteDid string, recipientDid string) error
	DeleteRecipientDid(remoteDid string, recipientDid string) error
	RemoteDidBelongsToMessage(remoteDid string, messageId string) (bool, error)
	GetMessageCountForRecipient(recipientDid string) (count int, err error)
	GetMessagesForRecipient(recipientDid string, limit int) ([]didcomm.Attachment, error)
	DeleteMessagesByIds(messageIds []string) error
	AddMessage(recipientDid string, message didcomm.Attachment) error
}
