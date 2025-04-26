# Coordinate Mediation
- Status: [RELEASED](/README.md#released)
- Specification: [Coordinate Mediation 3.0](https://didcomm.org/coordinate-mediation/3.0/)

## Summary

Configuration to coordinate the mediation between a mediating agent and the recipient.

## Motivation

To forward messages in the routing, it is required to exchange some information of the recipient. The recipient must know which endpoint and routing key(s) to share, and the Mediator needs to know which keys should be routed via this relationship.

## Prerequisite

- Create Connection (see feature [Create Connection](/docs/features/0010-create-connection.md))

## Tutorial

The feature consists of three different message requests from the recipient that the mediator replies to:

- [Mediate Request](#mediation-request) -> [Mediate Grant](#mediate-grant) or [Mediate Deny](#mediate-deny)
- [Recipient Update](#recipient-update) -> [Recipient Update Response](#recipient-update-response)
- [Recipient Query](#recipient-query) -> [Recipient](#recipient)

The feature can be used over the REST API endpoint `/message/receive` with a `POST` request.
To test this feature use the provided file [didcomm-coordinate-mediation.http](/tests/didcomm-coordinate-mediation.http).

### Roles:
- mediator: will receive messages on behalf of the recipient
- recipient: intended receiver of the payload of a message

### Mediation Request

The recipient asks the mediator for permission (and routing information) to publish the endpoint as a mediator.

Message of the mediation request:
``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/coordinate-mediation/2.0/mediate-request",
    "return_route": "all"
}
```
### Mediate Deny

After the mediation request the mediator can deny the request. The recipient cloud get a deny response if he is already mediated or if he is blocked.

Message if a mediation is denied:
``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/coordinate-mediation/3.0/mediate-deny",
}
```

### Mediate Grant

The mediator grants to be the mediator of the recipient and replies with the corresponding message. The recipient then can use the included information as an inbound route.

- routing_did: DID of the mediator where forwarded messages should be sent. The recipient may use this DID as an endpoint (see [Using a DID as an endpoint](https://identity.foundation/didcomm-messaging/spec/#using-a-did-as-an-endpoint)).

Message if a mediation is granted:
``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/coordinate-mediation/3.0/mediate-grant",
    "body": {
        "routing_did": ["did:peer:0z6Mkfriq1MqLBoPWecGoDLjguo1sB9brj6wT3qZ5BxkKpuP6"]
    }
}
```

Important: After the granted mediation the recipient should add DIDs so that the mediator can accept and forward messages for those DIDs.

### Recipient Update

The recipient notifies the mediator of DIDs he uses. The following actions are allowed:

- add
- remove

Example message of a recipient update:
``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/coordinate-mediation/3.0/recipient-update",
    "body": {
        "updates": [
            {
                "recipient_did": "did:peer:0z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
                "action": "add"
            }
        ]
    },
    "return_route": "all"
}
```

### Recipient Update Response

The mediator confirms the received DID updates of the recipient. The following results are possible:

- client_error
- server_error
- no_changes
- success

Example message of a recipient update response:
``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/coordinate-mediation/3.0/recipient-update-response",
    "body": {
        "updated": [
            {
                "recipient_did": "did:peer:0z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH",
                "action": "" // "add" or "remove"
                "result": "" // [client_error | server_error | no_change | success]
            }
        ]
    }
}
```

### Recipient Query

Get a list of registered recipient DIDs for a connection. The pagination is optional but must include limit and offset if it is used.

Example message of a recipient query:
``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/coordinate-mediation/3.0/recipient-query",
    "body": {
        "paginate": {
            "limit": 30,
            "offset": 0
        }
    }
}
```

### Recipient

The mediator responds with a list of registered DIDs for a recipient. Pagination here is optional too but must include count, offset and remaining if it is used.

Example message of a recipient query response:
``` json
{
    "id": "123456780",
    "type": "https://didcomm.org/coordinate-mediation/3.0/recipient",
    "body": {
        "dids": [
            {
                "recipient_did": "did:peer:0z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH"
            }
        ],
        "pagination": {
            "count": 30,
            "offset": 30,
            "remaining": 100
        }
    }
}
```

### Flow

![image](/docs/features/images/didcomm-coordinate-mediation.drawio.png)

   
## Implementations

See file: [coordinateMediation.go](/protocol/coordinateMediation.go)
