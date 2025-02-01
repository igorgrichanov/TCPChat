# TCPChat
Simple TCP Chat. 

TCP Server listens on port localhost:8080 and sent notifications to all connected peers when events happen. Messages are shown in stdout.

## Description

You can connect to server using tools like "netcat".

Server and Client code has been covered by unit-tests. To run tests, enter "go test"

## Features

- When Client connects to the Server, Server sends "You are [client_ip]" to every connected peer, including just connected Client
- When any of the peers send a message (via input from keyboard), every connected peer receives it and prints to stdout in the following format: "[client_ip]: [message]"
- When any of the peers disconnects, every connected peers receives "[leaving_peer_ip] has left"




