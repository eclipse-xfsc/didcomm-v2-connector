# Create Connection
- Status: [RELEASED](/README.md#released)
- Specification: None

## Summary

Establish a connection between a user of the cloud (using his DID) and the DIDCommConnector.

## Motivation

The DIDCommConnector needs to know with which protocol the he has to forward messages to the User. If NATS is the value of the protocol, a topic is required as well. In addition the user can add custom properties.

## Tutorial

The feature can be used over the REST API endpoint `/admin/connections` with a `POST` request. To create a connection a user has to send a request containing the following:

``` json 
{
    "protocol": "NATS", // NATS or HTTP
    "remoteDid": "did:example:123456",
    "topic": "nats-topic",
    "properties": {
        "key": "value",
        "greeting": "hello-world"
    }
}
```

The DIDCommConnector will reply with an OOB invitation if the DIDCommConnector allows the connection. If the DIDCommConnector denies the connection it will reply with a DIDComm problem report.

Each user can only connect once with a DID.

## Implementation

See file: [connection.go](/cmd/api/connection.go)
