# Trust Ping
- Status: [RELEASED](/README.md#released)
- Specification: [Trust Ping Protocol 2.0](https://identity.foundation/didcomm-messaging/spec/#trust-ping-protocol-20)

## Summary

A standard way to test connectivity, responsiveness, and security of a DIDComm channel. It is similar to the network `ping`. 

## Motivation

It is not guaranteed that other parties are connected or running all the time. Therefore it is vital to have insights into privacy and security of a DIDComm channel that a regular ping cannot.

## Tutorial

The feature can be used over the REST API endpoint `/message/receive` with a `POST` request.

Example request:
``` json
{
  "type": "https://didcomm.org/trust-ping/2.0/ping",
  "id": "518be002-de8e-456e-b3d5-8fe472477a86",
  "from": "did:example:123456",
  "body": {
      "response_requested": true
  }
}
```

If `response_requested` is `false` the sender is not requesting a response but is able to verify if the recipient is available by the HTTP status code `200 OK`.
If the sender is requests a response, the `response_requested` value is `true` and the response looks like the following example. 

Example response:
``` json
{
  "type": "https://didcomm.org/trust-ping/2.0/ping-response",
  "id": "e002518b-456e-b3d5-de8e-7a86fe472847",
  "thid": "518be002-de8e-456e-b3d5-8fe472477a86"
}
```
To test this feature use the provided file [didcomm-trust-ping.http](/tests/didcomm-trust-ping.http).
   
## Implementation

See file: [trustPing.go](/protocol/trustPing.go).
