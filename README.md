# DIDComm Connector

# Status

The DIDComm Connector is an mediator like tool which enables DIDComm V2 Protocol between a Cloud Entity, Devices and Other DIDComm Connectors. Please note that this connector should be considered as Mediator in an Standalone Running way for a third party, because it implements currently no encryption/signing in this version. Ensure therefore that all connections between the device/party and this connector is TLS protected to avoid privacy problems. Multiple Router Hops of messages are also not recommended so long as the party is not trusted to process the data within the messages. 


## Protocol

1. Create Invitation from Actor to Mediator. Result is a URL with an token
2. Call the mediate request to register your actor by using the token in bearer auth
3. Follow the message flow. The message flow for delivering messages can be found [here](https://didcomm.org/messagepickup/3.0/) Check status route if message are arriving in your inbox. If yes, call delivery request. If received, call message received with the ids of messages which can be removed from mediator.

If a message should be forwarded to anyone use the forwarding message.

## Description

The DIDCommConnector can be used as a Mediator and Connection Management Service by parties who want to set up trust with another party. The DIDCommConnector uses DIDComm v2 and provides a message layer and a management component for the following two use cases:

- Pairing a cloud solution with a smartphone / app solution
- DIDComm v2 based message protocols

## Installation

Perquisite installations:
- go
- cargo

How to install:

- Execute `go get` in the main directory.
- To install all needed dependencies use the `makefile` and run `make`.

For Development (Linux):

- Install air for `make install-air`
- Add `export $PATH=PATH:your/path/to/go/bin` to path (`~/.bashrc`) or `your/path/to/go/bin` to `/etc/paths`
- Build the libs executing `make build-rust`


## Usage

Before you can start the application, an instance of the [Universal Resolver](https://github.com/decentralized-identity/universal-resolver) should be available and configured. Note: The version of [uport/uni-resolver-driver-did-uport](https://hub.docker.com/r/uport/uni-resolver-driver-did-uport/) must be `4.3.0`, select a compatible resolver version.

How to run the application:

- To start the application execute `make dev`

Alternatively the application can be run using following chaing of commands:

```
go build -o dcc ./cmd/api/ 
export LD_LIBRARY_PATH=${PWD}/didcomm/lib
./dcc
```

The application may also be executed inside docker. For that scenarion following steps must be executed:

- build the docker container using `make docker-build`
- run the docker container on the local docker host using `make docker-run`
- start the docker container using `make docker-start` (or use docker desktop)
- optional: before re-building the container, do not forget to clean using `make docker-clean`


### Operation Modes

For the DIDComm message based communication to the Cloud there are three different modes. The modes are configured in the [config.yaml](/config.yaml). The [CloudEventProvider](https://github.com/eclipse-xfsc/cloud-event-provider) is used to handle the communication.

- NATS: receive and send messages with the NATS protocol
- HTTP: receive and send messages with the HTTP protocol
- HYBRID: receive and send messages with the NATS and HTTP protocols

⚠️ Hybrid mode can not be used at the moment. The used protocol of the CloudEventProvider can only be set at the start of an application and can not be changed while an application is running.

### DIDComm

The following DID methods are supported:
- [did:peer](https://identity.foundation/peer-did-method-spec/) (For creating a peer DID [Method 2](https://identity.foundation/peer-did-method-spec/#method-2-multiple-inception-key-without-doc) is used)

The used encryption algorithms are the following:
- `Ed25519VerificationKey2020` for signing (Base58 encoded)
- `X25519KeyAgreementKey2020` for encryption (Base58 encoded)

Example of an encrypted DIDComm message:

``` json
{
    "ciphertext": "vxqMZabuz1CMfFBO1HGR6Sj6fvibjL91fQvtIiVQ-fO25T9zRiZnbaFRGysmzxhjbWAErD4cDAzcyzROxAqMKy3f49bF1KZ7QfLuqe_EcYDjn1ifNrBLvPaqz56xvWLVdkbEEWc0XgG-19esejnIHd0h2lmkxzJcesV6dyiqsVkNd55gGg-Gyk1e4YpS0YVKrSnI-2DTI1ZOOh1t6feo5-z8ozPCucDbkFDBz1l5qDhU2wpmITYeICKRdC2Rr2Whh-Kyg3ODTj-_31lSBx3cDMlzNbLtgytXuC12NWJJam8vZSFMUm97x1BlYlz75qqrd1awECQtWY0HsyXjYw4ahxcs8bsq2C6I-v295pxW48LeNnTtMhJQL1lcYMJTJSDH2cy_GcVcu3mOXc-ZS7SrRAoXarYb7ZIWwTkaTSk319R8sck0D8llx--he1jSjolOmWD1QoXclCiep2NPxhACIJFSPJx0CHGTtyR7ebMDobs_OxXadbVZYbdIBzL0yhZ9BF69u3-ovY1ck7mY_SsWFXCL5k_U73iBLWj_eDaU6nDva7zDv1Z3XYlRy8AoagoRfA18OZyCJmfmzmvntTEqk60v5wy-dsCZMPJS20ybzA-XzAzMRVAeYYfAvwvja49lfhy3kZFVVRy8qjjRa2_wYVcJb8a2228EPotkfSLP5O_mAOqNZ1_VUeRuhwXJgUWK6W-oRJ7vPJrswXZeSCnURO-c7bVgB1kYBJljRH5QdrZyxFnP1y2ZtcbLH9N1-m06OSSw58BR5GqxJKcfExanaBqmhJtOX8e6kc3FkrGOJTxi1gt8B1u97gFlmIi0UYrj66QbqC8g56mcEOuc8x2q3TTLVtrMn9YT4vI2O0d75A7v_fYFUeNk9iaJnwHZ1iFw7OzIQ_jPJacC0GGkHCHtMHPug3rz3oyn68tibxQKvbnP2yVeuBu9wnDEquchJ1UQWBTrVg0BaBNZLUtEM9k0Yza-YWKk7lII1d7GVbS1C4DpyHV3R9fMVZKa7PgYCdc_aqU5YyPdzwREy_l3vNbOy8Upc9S4eUipDuF_7yOE9YOtkyjLDFfZsbi4O84ZhtHpvCpdYPtZWYopFqP1-fmRXpXC-67Fkq_qOalBl9kXvdymDC68D02zfG3hnbS8Y22wjKPcwzU1I3bJYjqRbKo8Wy8ePRgKUR0cLCU-yFAqbY7vR4FAPOd66u9qWSZv38wnHvz6eBkxqdzTovst4tZcsoLiUuKPe5vnJbgcP99v-aXu-ipU2fZzcQS4dEYGWCdkq1Nipti1LKeL_pqTyyQr4tI2WZ_0yV1Jv2Rk_XHyxzOO0Oje70r8YDa5XLVJSOnC8FjFqfNtMVBDNOaM8UY4flVNv4bYGafaYQOKTWSKJ-prDMeer-okvdLOfZm5T2HzgcC79Jz-uf4U__Q81hOFuwdsZUx4SlWrvN3GAhOdMwDDzsgFlIDpqlPywb4C55BTXgykAxIi_3Jjbspk_ksDbFa3rXopgN4IJvlR9wJ13XHfAloIhXSiDmNs2JK3FFsl1xyD65_ZGMWS46rSBy1WcHOGFTnet2yiC50enkCI5XPCZLA_bcMJWDXJyE2c1NvjcIZbQaFw8wV9cPd-CgidEFOnpN_Nw3tO0lAAR8Im0rInzWEfaOkMYNc1FX4zdHV87xQ6EZB9GetRN8fQrY3Sl9mV2zxYerxw6JakHA38hssIT6EFBe-GCbRbh7KnfJ8plVECSaGwsdiF_09GEHQP51sOhRlIsYUbDrWtjzU1ltNYmPUfAtnAjQAHYW7f7JmURwMk464zFtZ1vibC-vZ8xDiclKEC3s39qaCndTj8JQ39TUqV2SdVsfpzGkRr7WZB2Rgy5ol5-w-Y5w6aYroThj8jsiOJjKNEENRNsHsnuXE",
    "iv": "51Ivppe4PbYhdjOeSz83JQ",
    "protected": "eyJ0eXAiOiJhcHBsaWNhdGlvbi9kaWRjb21tLWVuY3J5cHRlZCtqc29uIiwiYWxnIjoiRUNESC0xUFUrQTI1NktXIiwiZW5jIjoiQTI1NkNCQy1IUzUxMiIsInNraWQiOiJkaWQ6cGVlcjoyLkV6NkxTaUdLalVnN2NRUUtNTTVkV1lBWGN1YlJRRG5rclV2NmIzb0piWkFBbkV3cWMuVno2TWtrY0xrQ3JrRjU4WXdkaUF6eVFhbTZEYUYyR1o4NjlKaEFZSzdHM2UxQlRvYS5TZXlKMElqb2laRzBpTENKeklqcDdJblZ5YVNJNkltaDBkSEE2THk5c2IyTmhiR2h2YzNRNk9UQTVNQzl0WlhOellXZGxMM0psWTJWcGRtVWlMQ0poSWpwYkltUnBaR052YlcwdmRqSWlYU3dpY2lJNlcxMTlmUSM2TFNpR0tqVWc3Y1FRS01NNWRXWUFYY3ViUlFEbmtyVXY2YjNvSmJaQUFuRXdxYyIsImFwdSI6IlpHbGtPbkJsWlhJNk1pNUZlalpNVTJsSFMycFZaemRqVVZGTFRVMDFaRmRaUVZoamRXSlNVVVJ1YTNKVmRqWmlNMjlLWWxwQlFXNUZkM0ZqTGxaNk5rMXJhMk5NYTBOeWEwWTFPRmwzWkdsQmVubFJZVzAyUkdGR01rZGFPRFk1U21oQldVczNSek5sTVVKVWIyRXVVMlY1U2pCSmFtOXBXa2N3YVV4RFNucEphbkEzU1c1V2VXRlRTVFpKYldnd1pFaEJOa3g1T1hOaU1rNW9Za2RvZG1NelVUWlBWRUUxVFVNNWRGcFlUbnBaVjJSc1RETktiRmt5Vm5Ca2JWVnBURU5LYUVscWNHSkpiVkp3V2tkT2RtSlhNSFprYWtscFdGTjNhV05wU1RaWE1URTVabEVqTmt4VGFVZExhbFZuTjJOUlVVdE5UVFZrVjFsQldHTjFZbEpSUkc1cmNsVjJObUl6YjBwaVdrRkJia1YzY1dNIiwiYXB2IjoiS0Y3V0kxMmwwdC01c2FOamt0b25vZGdqb0RMUFJWdzlTa3RkRlpfRlFmOCIsImVwayI6eyJjcnYiOiJYMjU1MTkiLCJrdHkiOiJPS1AiLCJ4IjoiWl91d212RTR4b1ZQWTdsV2p3TmU3VnE5aHBYbmFaa1U5alFWWGMzdnNqayJ9fQ",
    "recipients": [
        {
            "encrypted_key": "xHqGRY5syCSrRcJEuYKNmC2_c8a6X3T46pM2lotqcV6XqmeP6vVpBxDazDrvjm-DcZP89UwAp5cQEdekwZPImCIs5fy5bzhn",
            "header": {
                "kid": "did:peer:2.Ez6LSmg6eZ5FdMcE8PPSMMBWXDvDPQ2weFhbTWjabmgeo3hQh.Vz6Mknr5Zt1YeLF6XpCchBSCrepSoaXpFV93TR5YyhnU3nu8A.SeyJ0IjoiZG0iLCJzIjp7InVyaSI6Imh0dHA6Ly9sb2NhbGhvc3Q6OTA5MC9tZXNzYWdlL3JlY2VpdmUiLCJhIjpbImRpZGNvbW0vdjIiXSwiciI6W119fQ#6LSmg6eZ5FdMcE8PPSMMBWXDvDPQ2weFhbTWjabmgeo3hQh"
            }
        }
    ],
    "tag": "vU2QDMvobdzBmgyfqHlnV9Gkw0ScK7bbm5c2-7eaG34"
}
```

### Flow

#### DIDComm V2

The diagram shows the basic didcomm v2 flow between to peers (Alice and Bob).

![didcomm-v2-flow](/docs/features/images/didcomm-v2.drawio.png)

#### DIDCommConnector

The standard flow between a user and the DIDCommConnector are described in following diagram.

![didcommconnector-flow](/docs/features/images/didcommcoordinator-flow.drawio.png)

The flow only represents the happy path and does not show any details. The details of each step are described in the section [Features](#features). The dots at the end should indicate that at the end most of the messages are forward or receiving messages.

## Docker

To build a docker image, run `cd deployment/docker && docker build . --tag didcommconnector --build-context files=../..`

To run the image (create a container from it and launch it), run `docker run -p 9090:9090 -d --name dcc didcommconnector`

To launch the container, run `docker start dcc`

To remove the container, run `docker rm -f dcc`

To remove the image, run `docker image rm didcommconnector:latest`

## Configuration
configuration file: [config.yaml](/config.yaml)

#### application:
- **env**: `DEV` or `PROD` environment
- **logLevel**: `info`, `debug`, `warning` or `error` 
- **port**: *(example: 9090) needs to be 9090 if db.inMemory:true*
- **url**: *(example:"http://localhost:9090") if changed, the mediator DID in the DB needs to be deleted*

#### didcomm:
- **resolverUrl**: the url of the DID resolver *(example: "http://localhost:8081")*
- **messageEncrypted**: set the messages encryption - `true` or `false`

#### database:

**db**:
- **inMemory**: store data in application's memory - `true` or `false` (true only for demo purposes, set to *false* to use the database)
- **host**: database's connection url *(development example: "localhost")*
- **port**: database connection port *(example: 9042)*
- **user**: database user
- **password**: database password
- **keyspace**: database keyspace
- **dbName**: database name

#### cloudEventProvider

See https://github.com/eclipse-xfsc/cloud-event-provider for more info.

- messaging:
  - **protocol**: messaging's protocol -  `nats` or `http`
  - **nats**:
    - **url**: url to send cloud event *(example: "http://localhost:4222")*
    - **topic**: the topic to receive didcomm messages
    - **queueGroup**: *optional (example: logger)*
    - **timeoutInSec**: *optional (example: 10)*

  - **http**:
    - **url**: url to send cloud event *(example: "http://localhost:1111")*
    - **port**: port to send cloud event *(example: 1111)*
    - **path**: path to receive cloud event *(example: "xyz")* 

## Database

[gocql](https://github.com/gocql/gocql) is used to access the database.
 
Connection settings:
To connect the needed adjustments need to be set in the configuration (see [config.yaml](/config.yaml)).
 
Adapter interface:
The application contains a database adapter interface. An implementation of it is done for cassandra (see [mediator/database/cassandra.go](/mediator/database/cassandra.go) and [mediator/secretsResolver/cassandra.go](/mediator/secretsResolver/cassandra.go)). To use another database, add a new implementation of that interface and consider the potential adapting of the table structure.
 
Migration:
The database-initialization script, stored in `database/migrations`, will be executed once the application is being run for the first time and the db.inMemory in `config.yaml` is set to `false`. To make changes to the database, add another script(s) and run the application again.

Retrieve connections:

```bash
cqlsh <cassandra host> <cassandra port> -u <cassandra user> -p <cassandra password> -e "SELECT * FROM dcc.mediatees;"

```

## Test

To run all  unit tests execute `make test`.

In addition there are some end to end tests in the `tests` folder. Before executing these tests the application needs to be up and running. The extension [Rest-Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) in VS-Code is used to execute these tests. Select the environment `local` for the Rest-Client extension (Ctrl + Alt + E).

Alternative the postman collections can be used (see [docs/postman-collections](/docs/postman-collections/)).

### Swagger (Management API)

To test the management API it is possible to make rest requests. A UI for the management API is provided by swagger. The swagger for the management API is available at the URL: `localhost:9090/swagger/index.html`

## Features

All features are documentes in `docs/features`.

### Status of a feature

Possible status of a feature. Features are documented in `/docs/features`. For the description of the featrues the [Aries-RFC-Template](https://github.com/hyperledger/aries-rfcs/blob/main/0000-template.md) is used.

#### PLANNED
The feature is planned to be implemented in the future.

#### WIP
The feature is partly implemented and can be tested. It should not be used in production.

#### RELEASED
The feature is implemented and tested. It can be used in production.

## FAQ

- I get the error `/usr/bin/ld: cannot find -ldidcomm_uniffi` when starting the application
  - Check if the path of the library is exported
  - Execute `make build-rust` and try again to start the application
- I get the error `Error resolving peer DID: ...`
  - Check if the universal resolver is running and if the url correct in the config file
- The application imidiatly exits after start with an error `Resolver not available`
  - Check if the url in the config is correct and if the DID resolver is online and available for the DIDCommConnector