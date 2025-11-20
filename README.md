# Go Blockchain Proof-of-Concept

Minimal blockchain implementation in Go, inspired by [naivechain](https://github.com/lhartikk/naivechain). Provides an HTTP node plus a CLI client to list the ledger, mine new blocks, and test longest-chain replacement.

## Layout
- `internal/blockchain`: core chain logic (genesis, hashing, validation, replacement).
- `server`: HTTP node exposing `/blocks`, `/mine`, `/replace`.
- `client`: CLI for interacting with a running node.

## Quickstart
1) Start the node:
```bash
cd server
go run .
```
2) In another terminal, interact with the client:
```bash
cd client
go run . -action list
go run . -action mine -data "hello chain"
```
Use `-host` to point at a remote node (default `http://localhost:8080`).

## Testing
From the repo root:
```bash
GOCACHE=$(pwd)/.gocache go test ./...
```

## Notes
- Longest-chain replacement logic is covered by table-driven tests, including branch/merge scenarios.
