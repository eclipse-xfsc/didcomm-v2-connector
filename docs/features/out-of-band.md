# Out Of Band Messages
- Status: [RELEASED](/README.md#released)
- Specification: [Out Of Band 2.0](https://identity.foundation/didcomm-messaging/spec/#out-of-band-messages)

## Summary

The OOB is used to create an invitation in the form of an URL that can also be turned into an QR-Code.

## Motivation

The invitation can be presented in various forms. Therefor it is also possible present the invitation to a smartphone camera as a QR-Code.

## Tutorial

The feature can be used over the REST API endpoint `/admin/invitation` with a `GET` request.

Example of a request response
```
https://example.com/path?_oob=eyJ0eXBlIjoiaHR0cHM6Ly9kaWRjb21tLm9yZy9vdXQtb2YtYmFuZC8yLjAvaW52aXRhdGlvbiIsImlkIjoiNjkyMTJhM2EtZDA2OC00ZjlkLWEyZGQtNDc0MWJjYTg5YWYzIiwiZnJvbSI6ImRpZDpleGFtcGxlOmFsaWNlIiwiYm9keSI6eyJnb2FsX2NvZGUiOiIiLCJnb2FsIjoiIn0sImF0dGFjaG1lbnRzIjpbeyJpZCI6InJlcXVlc3QtMCIsIm1lZGlhX3R5cGUiOiJhcHBsaWNhdGlvbi9qc29uIiwiZGF0YSI6eyJqc29uIjoiPGpzb24gb2YgcHJvdG9jb2wgbWVzc2FnZT4ifX1dfQ
```

Example message that is base64 URL encoded in the above URL
``` json
{
  "type": "https://didcomm.org/out-of-band/2.0/invitation",
  "id": "69212a3a-d068-4f9d-a2dd-4741bca89af3",
  "from": "did:example:alice",
  "body": {
      "goal_code": "",
      "goal": ""
  },
  "attachments": [
      {
          "id": "request-0",
          "media_type": "application/json",
          "data": {
              "json": "<json of protocol message>"
          }
      }
  ]
}
```

To test this feature use the provided file [didcomm-invitation.http](/tests/didcomm-invitation.http).

## Implementation

See file: [outOfBand.go](/protocol/outOfBand.go)
