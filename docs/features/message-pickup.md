# Message Pickup
- Status: [RELEASED](/README.md#released)
- Specification: [Message Pickup 3.0](https://didcomm.org/messagepickup/3.0/)

## Summary

Is used to get messages for a user that are stored at the mediator.

## Motivation

Wallets that are not always connected can pick up their messages from the mediator.

## Tutorial

This feature has multiple request that the recipient can send to the mediator:
- Status Request -> Status
- Delivery Request -> Message Delivery
- Message Received -> Status
- Live Mode -> Status or Problem Report

The feature can be used over the REST API endpoint `/message/receive` that handles DIDComm messages.

To test this feature use the provided file [didcomm-message-pickup.http](/tests/didcomm-message-pickup.http).

### Roles:
- mediator: stores the messages waiting for pickup by the recipient.
- recipient: is picking up messages from the mediator.

### Status Request

To check if there a messages for a recipient, the recipient can sent the following status request:

``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/messagepickup/3.0/status-request",
    "body" : {
        "recipient_did": "<did for messages>"
    },
    "return_route": "all"
}
```

The mediator will then reply with a status.

Example of a status reply of the mediator.
``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/messagepickup/3.0/status",
    "body": {
            "recipient_did": "<did for messages>",
            "message_count": 7,
            "longest_waited_seconds": 3600,
            "newest_received_time": 1658085169,
            "oldest_received_time": 1658084293,
            "total_bytes": 8096,
            "live_delivery": false
    }
}
```

### Delivery Request

To receive messages the recipient has to make the following delivery request:

``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/messagepickup/3.0/delivery-request",
    "body": {
        "limit": 10,
        "recipient_did": "<did for messages>"
    },
    "return_route": "all"
}
```

The mediator will then deliver messages to the recipient.

Example of a delivery message:
``` json
{
    "id": "123456780",
    "thid": "<message id of delivery-request message>",
    "type": "https://didcomm.org/messagepickup/3.0/delivery",
    "body": {
        "recipient_did": "<did for messages>"
    },
    "attachments": [{
        "id": "<id of message>",
        "data": {
            "base64": "<message>"
        }
    }]
}
```

### Message Received

When a recipient has received one or multiple messages he has to inform the mediator. To inform the mediator about the received messages, the recipients sends the following message:

``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/messagepickup/3.0/messages-received",
    "body": {
        "message_id_list": ["123","456"]
    },
    "return_route": "all"
}
```

The mediator then can delete the the messages that the recipient has received. The mediator then replies with a status (see status reply in chapter [Status Request](#status-request)).

### Live Mode

This mode is not supported. If a recipient still makes a request for this mode he will get a problem report as reply.

## Flow

The following diagram shows the flow how the recipient picks up his messages form the mediator:
![Message-Pickup](/docs/features/images/didcomm-message-pickup.drawio.png)

## Implementation

See file: [messagePickup.go](/protocol/messagePickup.go)
