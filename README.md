# Overview

This is a toy proof of concept of a UDP load balancer which duplicates requests to all backends and responds to the client with the first response it receives (therefore the fastest).

- Configure listening port with `listen` flag, for example `-listen 127.0.0.1:999`
- Configure backends with the `-backends` flag, providing a comma separated list, such as `-backends 127.0.0.1:123,127.0.0.1:567`

This is essentially the first iteration, and may contain bugs and performance issues. There are further optimisations, as each request creates a new listening socket for the backend response - this causes overhead (and syscalls), if the UDP messages could contain the IP and port, or other identifier which could be used to lookup where to send the responses to, a listening socket could be created at startup instead of at each request/response.

```
$ go get github.com/bradleyfalzon/udp-proxy
$ udp-quickest-responder -listen :1234 -backends 10.0.0.1:1234,10.0.0.2:1234
```
