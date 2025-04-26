# Routing
- Status: [RELEASED](/README.md#released)
- Specification: [Routing Protocol 2.0](https://identity.foundation/didcomm-messaging/spec/#routing-protocol-20)

## Summary

The goal is to forward DIDComm messages for the communication between two parties.

## Motivation

Create a message route from the sender to the recipient.

## Tutorial

There a multiple routes that can be used to forward a message:
- user -> mediator -> outgoing message
- incoming message -> mediator -> user (cloud/inbox)
- incoming message -> mediator -> outgoing message

The feature can be used over the REST API endpoint `/message/receive` that handles DIDComm messages.

To test this feature use the provided file [didcomm-routing.http](/tests/didcomm-routing.http).

An example of a message that will be forwarded looks like this:
``` json
{
    "type": "https://didcomm.org/routing/2.0/forward",
    "id": "abc123xyz456",
    "to": ["did:example:mediator"],
    "expires_time": 1516385931,
    "body":{
        "next": "did:foo:1234abcd"
    },
    "attachments": [
        // The payload(s) to be forwarded
    ]
}
```

## Flow

Example of all possibles routing are described in the following flow diagram:

![forward from user](/docs/features/images/didcomm-routing-forward-from-user.drawio.png)

## Implementation

See file: [routing.go](/protocol/routing.go)
