package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/blockchain/internal/blockchain"
)

func TestNodeEndpoints(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, n *node)
	}{
		{
			name: "lists genesis block",
			run: func(t *testing.T, n *node) {
				blocks := requestBlocks(t, n)
				if len(blocks) != 1 {
					t.Fatalf("expected 1 block got %d", len(blocks))
				}
				if blocks[0].Index != 0 {
					t.Fatalf("unexpected genesis index %d", blocks[0].Index)
				}
			},
		},
		{
			name: "mines block and lists it",
			run: func(t *testing.T, n *node) {
				block := requestMine(t, n, "payload")
				if block.Index != 1 {
					t.Fatalf("expected mined index 1 got %d", block.Index)
				}
				blocks := requestBlocks(t, n)
				if len(blocks) != 2 {
					t.Fatalf("want 2 blocks got %d", len(blocks))
				}
				if blocks[1].Data != "payload" {
					t.Fatalf("last block data %q want payload", blocks[1].Data)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newNode()
			tt.run(t, n)
		})
	}
}

func TestNodeReplaceEndToEnd(t *testing.T) {
	nodeA := newNode()
	nodeB := newNode()

	requestMine(t, nodeA, "a-1")
	requestMine(t, nodeA, "a-2")
	longChain := requestBlocks(t, nodeA)
	if len(longChain) != 3 {
		t.Fatalf("expected node A chain to have 3 blocks")
	}

	resp := requestReplace(t, nodeB, longChain)
	if resp.Code != http.StatusOK {
		t.Fatalf("replace returned %d", resp.Code)
	}

	updated := requestBlocks(t, nodeB)
	if len(updated) != len(longChain) {
		t.Fatalf("node B blocks %d want %d", len(updated), len(longChain))
	}
	if updated[2].Data != "a-2" {
		t.Fatalf("node B did not copy last block data")
	}

	longChain[1].PreviousHash = "bad"
	resp = requestReplace(t, nodeB, longChain)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 got %d", resp.Code)
	}
}

func requestBlocks(t *testing.T, n *node) []blockchain.Block {
	t.Helper()
	rr := doRequest(t, n, http.MethodGet, "/blocks", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("/blocks status %d", rr.Code)
	}

	var blocks []blockchain.Block
	if err := json.Unmarshal(rr.Body.Bytes(), &blocks); err != nil {
		t.Fatalf("decode blocks: %v", err)
	}
	return blocks
}

func requestMine(t *testing.T, n *node, data string) blockchain.Block {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"data": data})
	if err != nil {
		t.Fatalf("marshal mine payload: %v", err)
	}
	rr := doRequest(t, n, http.MethodPost, "/mine", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("mine status %d", rr.Code)
	}

	var block blockchain.Block
	if err := json.Unmarshal(rr.Body.Bytes(), &block); err != nil {
		t.Fatalf("decode mined block: %v", err)
	}
	return block
}

func requestReplace(t *testing.T, n *node, blocks []blockchain.Block) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(blocks)
	if err != nil {
		t.Fatalf("marshal chain: %v", err)
	}
	return doRequest(t, n, http.MethodPost, "/replace", payload)
}

func doRequest(t *testing.T, n *node, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, "http://example.com"+path, reader)
	rr := httptest.NewRecorder()
	n.handler().ServeHTTP(rr, req)
	return rr
}
