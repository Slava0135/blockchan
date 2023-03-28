# Blockchan

[![Run Tests](https://github.com/Slava0135/blockchan/actions/workflows/test.yml/badge.svg)](https://github.com/Slava0135/blockchan/actions/workflows/test.yml)

## About

Worthless service implementing PoW blockchain with at most 3 nodes using UDP sockets.

## Building

To build docker image:

```sh
docker build -t blockchan .
```

To run docker compose (3 nodes) in interactive mode:

```sh
docker compose -f "docker-compose.yml" up
```

To run single node in interactive mode:

```sh
docker run -it blockchan:latest /app -genesis
```

Available flags:

```text
Usage of ./blockchan:
  -addr string
        address node will be running (default "localhost:8888")
  -debug
        enable debug level logging
  -genesis
        generate genesis block on start?
  -name string
        node name for logging (default "NONAME")
```

Node connections are configured through arguments:

```sh
docker run -it blockchan:latest /app localhost:8001 localhost:8002 ...
```

## Tests

To run tests:

```sh
go test -timeout 10s -count 1 ./...
```

Add `-v` flag for verbose output

## Algorithm

* One node should be configured to generate first (genesis) block.
* Nodes compete with each other trying to generate a hash ending with 0x0000 for the next block.
* When succeded node sends mined block to others and they verify it
* If two nodes mined blocks at the same time then third node verifies first block it got and tells "loser" to drop their block.
* If node missed any blocks it asks other nodes to resend them. Longest chain is accepted.

## Structure

package `blockgen` - block generation and validation

package `encode` - block json encoding  

package `mesh` - virtual "mesh" connecting "forks" which may represent local node or remote nodes connected through "links"

package `messages` - messages that nodes send to communicate with each other

package `node` - node initialisation and blockchain generation

package `protocol` - nodes communication protocol

package `validate` - blockchain validation

`./tests` - integration tests

`network.go` - networking code

## Output Files

Every node writes verified blocks (to docker volume `blockchan`) to file `/blockchan/${name}.txt`

You can then access them directly on host mountpoint or using docker:

```sh
docker run -it --rm -v /path/on/host:/vol busybox ls -l /vol
```

And read and take diff (host path may be different):

```sh
docker run -it --rm -v /var/lib/docker/volumes/blockchan_blockchan/_data:/blockchan busybox cat /blockchan/ZERO.txt

docker run -it --rm -v /var/lib/docker/volumes/blockchan_blockchan/_data:/blockchan busybox diff /blockchan/ZERO.txt /blockchan/FIRST.txt
```
