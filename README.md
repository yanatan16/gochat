gochat
======

A backend chat server in Go.

Gochat is a backend chat server which uses Redis as a database and messaging backend ([Godis](http://github.com/simonz05/godis) as Go-based client) and websockets as the primary form of send/receive message routing.

Structure
=========

There are two main parts of Gochat:

- Server: This interface provides the management functions of Gochat such as creating areas, joining areas, listing users and areas, etc.
- Stream Service: This interface provides the methods to start websockets and connect them to Redis pub/sub channels for sending and receiving messages for a particular user and chat area.

To embed gochat into your server, simply make the calls to Server that you require. For joining areas, make sure to call Server.JoinArea() AND StreamService.InitiateStream() to create the websocket stream.

Streams
=======

When a stream is opened up as a websocket to a client, the messages that travel from the server to the client (as well as visa versa) are specified in the [Stream Object Model](https://github.com/yanatan16/gochat/blob/master/StreamObjectModel.md).
