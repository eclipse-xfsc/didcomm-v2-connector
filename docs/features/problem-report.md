# Problem Report
- Status: [RELEASED](/README.md#released)
- Specification: [Problem Report 2.0](https://identity.foundation/didcomm-messaging/spec/#problem-reports)

## Summary

Standard to report problems if a request can not be handled as expected.

## Motivation

The User is informed that a Problem appeared during his request and can handle it.

## Tutorial

This feature is is only for problem reporting.

Example of a problem report message:
``` json
{
  "type": "https://didcomm.org/report-problem/2.0/problem-report",
  "id": "7c9de639-c51c-4d60-ab95-103fa613c805",
  "pthid": "1e513ad4-48c9-444e-9e7e-5b8b45c5e325",
  "ack": ["1e513ad4-48c9-444e-9e7e-5b8b45c5e325"],
  "body": {
    "code": "e.p.xfer.cant-use-endpoint",
    "comment": "Unable to use the {1} endpoint for {2}.",
    "args": [
      "https://agents.r.us/inbox",
      "did:sov:C805sNYhMrjHiqZDTUASHg"
    ],
    "escalate_to": "mailto:admin@foo.org"
  }
}
```

For more information about the problem code see [Problem Codes](https://identity.foundation/didcomm-messaging/spec/#problem-codes)

## Implementation

See files:
- [problemReport.go](/protocol/problemReport.go)
- [problemReports.go](/protocol/problemReports.go)

