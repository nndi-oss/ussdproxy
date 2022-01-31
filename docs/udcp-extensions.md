UDCP Extensions
===============

> **DISCLAIMER**: These are just ideas, they do not actually work yet.

## Summary

This document defines extensions to the UDCP protocol defined in [1].
The extensions have been adapted primarily to enable UDCP to support multiple
applications/services, commands and querying capabilities.

## Introduction

> TODO: Write introduction

## Extension PDUs

## PDU ASCII 

In order to be interoperable with most telcos, the following reserved character
combinations will be used to represent the PDUs described below. 

`U;[{PARAMETERS}]` (`0;`) - Initialize the protocol e.g. `U;v:1.0,b:256,`

PARAMETERS
- `v` - *version* *required* `String`
- `b` - *bufferLength* `int`

`A;` (`0;`) - Initialize an application
`E;` (`- Error Response
`C;`   - Command Operation
`Q;`  - Query Operation
`F;`  - Query for feature availability
`R;`  - Receive Ready 
`D;`  - Data PDU Without more to send

`CX;`   - Command Operation with more to send
`QX;`  - Query Operation with more to send
`DX;`  - Data PDU With more to send

### Application PDU

An ApplicatonPDU selects an application on the server. The server MUST have
support for applications in order to process an ApplicationPDU. Servers 
that do not support multiple applications must return an `ErrorPDU` with the
Error Code set to Protocol Error (0x02) (`ERR;`).

A server that supports applications MUST implement the *Available Applications*(`q:apps`) 
query inorder to advertise which applications it provides.

An `ApplicationPDU` must be sent before any application specific DataPDU. A server
supporting multiple applications MUST also provide the *Current Application*(`q:app`)
query to identify the application that will process the subsequent DataPDUs.

### Command PDU

A CommandPDU specifies a command to execute on the server. The server MUST have
support for executable commands in order to process a CommandPDU. Servers that do not support
client executed commands must return an `ErrorPDU` with the Error Code set to
Protocol Error (0x02).

In order to construct a CommandPDU the Data content of the DataPDU is expected to
contain data with the following content.

Command Header:

+------+-------------+-----------------------------------------------------------------------+   
| c    | commandName | The command to execute on the server                                  |   
+------+-------------+-----------------------------------------------------------------------+   
| args | arguments   | `key:value` arguments for the command, separated by space             |   
+------+-------------+-----------------------------------------------------------------------+   

e.g. `c:sum x:9 y:10` - this executes a sum operation on the server and returns `19`

#### Empty Command Results

A server MAY return DataPDU after executing a command, depending on whether the 
command returns a result or not. In the case that a command does not return a 
result a server may elect to wait for more data from the client and send `ReceiveReadyPDU`.

#### Excessive Arguments

Ideally, server commands should be simple and require few arguments - but 
in the case that a command requires an excessive number of arguments or the
value of the arguments exceed the `DATA_LIMIT` then the CommandPDU MUST be
sent with the `MoreToSend` flag set. A server processing a CommandPDU with 
`MoreToSend` flag set MUST buffer the command data/arguments until the client
has sent a CommandPDU with the `MoreToSend` flag unset.

### Query PDU

A QueryPDU specifies a query to execute on the server. The server MUST have
support for the `query` commands in order to process a QueryPDU. Servers that do not support
client executed commands must return an `ErrorPDU` with the Error Code set to
Protocol Error (0x02).

In order to construct a QueryPDU the CommandPDU specification is followed with one 
difference, the parameter `c` is replaced with `q` to indicate that the command is a 
query.

e.g. `q:sum x:9 y:10` - this executes a sum operation on the server and returns `19`

A server that has received a QueryPDU MUST always return `DataPDU`.

## Basic Applications

A server supporting the Extended UDCP protocol defined in this document 
MUST support the following basic applications. A server that does
not implement the following applications is NOT a compliant server.

### Echo Application (`echo`)

The `echo` application is used to debug the requests and responses exchanged
between the server and the client. The `echo` application echoes all the client
input when it has received a `DataPDU` with the `MoreToSend` flag unset.
When the server receives `DataPDU` with `MoreToSend` flag set, it MUST buffer
the data until it receives a pdu with the flag unset - at which point it must
begin to send the data back to the client in chunks (not necessarily the same
as it received); i.e. the server must send the data in it's `receive buffer`
back with `MoreToSend` flag if the server's send buffer has `size() > DATA_LIMIT`.


## Basic Commands and Queries

A server supporting the Extended UDCP protocol defined in this document 
MUST support the following basic commands and queries. A server that does
not implement all of the commands and queries below is NOT a compliant
server.

### Server Information

Query Name: Server Information (`q:info`)
Return: `name` `version` `buildId[optional]` `author`

### Application

Command Name: Select application (`c:app id:APPLICATION_ID`)

Query Name: Available Applications (`q:apps`)   

Query Name: Current Application(`q:app [verbose:(bool)]`)

### Session

Query Name: Session ID (`q:sessID`)   
Return: Unique ID of the session


Query Name: Keep Alive (`q:keepAlive`)   
Return: (bool)

Query Name: Receive Ready Limit (`q:rrLimit`)   
Return: integer

Query Name: Receive Ready's Sent (`q:rrSent`)   
Return: integer   

Command Name: CacheSession (`c:sessCache ttl:(integer)`)   
Description: Keep the session with a specified TTL (time-to-live)   

Command Name: XloseSession (`c:sessClose [id:SESSION_ID]`)
Description: Close the session/end the connection

### Buffer

Query Name: Min Buffer Size (`q:bufMinSize`)   

Query Name: Max Buffer Size (`q:bufMaxSize`)   

Query Name: Current (Read) Buffer Size (`q:bufCurSize which:read`)   
Query Name: Current (Send) Buffer Size (`q:bufCurSize which:send`)   

  - clearBuffer: true # Clear the server's buffer for the session
  - growBuffer: true # Ability for a client to grow session buffer to some value less than maxBufferSize
  - shrinkBuffer: true # Ability for client to shrink the session buffer (recommended for IOT apps)
    ## Session


# References

[1]: WAP over GSM USSD, Open Mobile Alliance. 2003

---

Copyright (c) 2018, NNDI