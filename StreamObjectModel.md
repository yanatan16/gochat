# Stream Object Model

## Intro

This document specifies the notation and structure of messages sent over a text chat stream object to websocket clients. Note that streams are chat-area-specific, so each message will only be in reference to the chat area that opened the stream.

There are two directions of communication; both are detailed below. Each message (in both directions) will be formatted as JSON objects.

## Server To Client

Each message will be formatted as JSON objects with one field always present: "type". The "type" field will always be a string, which will contain one of the following values:

- "chat"
- "user"

### Chat Messages

Chat Messages denote new chat messages are being sent into a chat area. They are formatted with the basic form {"type":"msgs"}. Each messages object also contains a field: "msgs", which is an array of objects containing the fields: "User" and "Msg" (note the capitalization, inherent from Go).

Example:
    {
    	"type":	"chat",
    	"msgs":	[
    		{
    			"User": "frank",
    			"Msg": "hopscotch! I love hopscotch!"
    		},
    		{
    			"User": "joe",
    			"Msg": "Dude, calm down."
    		}
    	]
    }

### User Messages

User messages denote new users have joined, or that users have left. Each user object comes with a field "op", which takes the values "add" and "rem" (to add an user or remove an user respectively), as well as a "name" field. The "users" field contains an array of these objects.

Example:
    {
    	"type": 	"user",
    	"users":	[
    		{
    			"op": 	"add",
    			"name": "billy-bob"
    		},
    		{
    			"op": 	"rem",
    			"name":	"russha11235"
    		}
    	]
	}

## Client to Server

The primary use of the client to server connection is to send new chat messages to the chat area. To do so, simply wrap a JSON object with a single field: "msg" with the text of that message, and forward down the websocket connection.

Example:
    {
    	"msg": "hey, this is tammy; long time, so see!"
	}