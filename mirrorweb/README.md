# Reflect Demo

Reflect sends video data from the browser to the go server, which then returns it, all via WebRTC. 

To use:

1\) Start the server 

```bash
$ go run main.go
now serving on localhost:8000
```

2\) Click the send session button on the browser. This will send 
the browser's WebRTC session data over to the server via request and start a session.

3\) Then click start video to start and close video to close. 

