package main

import (
	"encoding/json"
	"net/http"

	"example.com/blockchain/internal/blockchain"
)

type node struct {
	chain *blockchain.Chain
}

func newNode() *node {
	return &node{
		chain: blockchain.NewChain(),
	}
}

func (n *node) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /blocks", n.handleBlocks)
	mux.HandleFunc("POST /mine", n.handleMine)
	mux.HandleFunc("POST /replace", n.handleReplace)
	return mux
}

func (n *node) handleBlocks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, n.chain.Blocks())
}

func (n *node) handleMine(w http.ResponseWriter, r *http.Request) {
	var req mineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if req.Data == "" {
		http.Error(w, "data is required", http.StatusBadRequest)
		return
	}
	block, err := n.chain.AddBlock(req.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, block)
}

func (n *node) handleReplace(w http.ResponseWriter, r *http.Request) {
	var blocks []blockchain.Block
	if err := json.NewDecoder(r.Body).Decode(&blocks); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if ok := n.chain.Replace(blocks); !ok {
		http.Error(w, "replacement failed", http.StatusBadRequest)
		return
	}
	writeJSON(w, n.chain.Blocks())
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
