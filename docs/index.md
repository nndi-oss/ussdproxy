Architecture of UDCP Server
===========================

The UDCP server provides a means for a GSM connected device to
interact with web applications over the GSM USSD protocol.
UDCP server acts as a proxy or gateway to compliant services that
offer low data web services. 

The server is an implementation of the USSD Dialogue Control Protocol 
- a protocol that allows bidirectional communication over USSD protocol 
to allow HTTP connections over GSM USSD. Essentially, this enables 
development of specialized applications that can provide data services over
USSD which results in lower costs.

From the spec

> The USSD dialogue provides a two-way-alternate interactive service to the user. This means that only
> the entity (mobile phone or network node) with the turn may send and its correspondent is permitted
> only to receive. To be able to use the USSD dialogue as a full duplex service a special protocol has to
> be specified that deals with the management of the dialogue. The protocol has to hide the two-way-
> alternate characteristics of the USSD dialogue to the upper layer, and allow the upper layer to use
> USSD as a full duplex service onto which datagrams can be sent and received.

## Architecture

UDCP Server is essentially a webserver that provides a custom protocol to enable
GSM clients to interact with USSD applications

### Client

A client is a device/program that sends USSD requests to the Server for processing 
by either the server or an application

### Server

The core server responds to all requests and may forward responses to *applications*
it is configured with

### Application

An application is a program that provides some service to clients. The application 
receives data from the Server after it has processed the complete request from a client.
An Application can be either `passive` or `active`

#### Active Applications 

Active applications are bi-directional programs, that is to say - they process a request
from a Client and typically return some result of that processing (other than an `ErrorPDU`).
Active applications process requests and may send to an `external system` for processing, once they
have a response they send the received response from the `external system` to the requesting
Client. 
Active applications, include:

* API proxy service 


#### Passive Applications

Passive applications are *uni-directional* applications in the sense that they
mostly wait for requests from clients and then forward the request data to an
`external system`. Passive applications only return a response if the request to
the `external system` system returned an error.
 Otherwise, passive applications are mostly just dumb waiters. They only receive data
and deliver it to the configured external system.
Passive applications are best used for data collection, particularly from sensors and
other IOT devices.
Passive applications include:

* Sensor data collection 
* Location tracking (e.g. Equipment/Vehicle)
* Alerting (e.g. on errors in sensors/device/program)
* Command initiating applications (e.g. Event based task scheduling:- )

## Server Configuration

The Server will be configured via YAML and the following keys are supported

```yaml
host: localhost
port: 8327
tls:
  key: /path/to/server.key
  ca_store: /path/to/server.pem
udcp:
  keepAlive: true # Whether to wait for data 
  receiveReadyLimit: 5 # Number of RR pdus to send to the server 
  maxBufferSize: 8096 # Maximum size of the buffer on the server and client side
  # Apps or Services are applications running on the UDCP server 
  apps:
  - echo # Echos messages from Clients
  - time # Provides the time 
  - weather # Retrieves weather updates from a Weather API
  # Other apps can be custom in that the server can route requests to them
  - weatherApi: 
    host: http://api.example.com:4000/udcp/
    apiKey: ...

  commands:
  - querySessionID: true
  - queryKeepAlive: true
  - queryReceiveReadyLimit: true
    ## Buffer ops
  - queryMaxBufferSize: true
  - clearBuffer: true # Clear the server's buffer for the session
  - growBuffer: true # Ability for a client to grow session buffer to some value less than maxBufferSize
  - shrinkBuffer: true # Ability for client to shrink the session buffer (recommended for IOT apps)
    ## Session
  - cacheSession: true # Keep the session with a specified TTL (time-to-live)
  - closeSession: true # Close the session/end the connection
# Configuration used for Client Application information
# also used for session storage if session.database == postgresql
database:
  host: localhost
  port: 5432
  name: udcp
  username: udcp
  password: changeit

session:
  database: boltdb | redis | postgresql
  ## For BoltDB
  path: /path/to/boltdb/instance
  ## For Redis
  host: localhost
  port: 6172
  
alerts:
  email:
    host: ...
    port: 
    username:
    password:
    useTls: true
    recipients:
    - user@example.com
  sms:
    africastalkingApiKey: SOME_KEY_HERE
    recipients:
    - 08888888

## Logging
logging:
  enabled: true
  level: INFO
```

## Server Operations

These are commands that the client can send to the server to query or command the server to
perform some supported function. 

See [UDCP extensions](./udcp-extensions.md) for supported commands and queries


## REFERENCES

- [Plain Text Protocols, Blain Smith](https://blainsmith.com/articles/plain-text-protocols/)
