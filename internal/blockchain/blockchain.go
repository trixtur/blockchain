package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Block models a single entry in the chain.
type Block struct {
	Index        int       `json:"index"`
	Timestamp    time.Time `json:"timestamp"`
	Data         string    `json:"data"`
	PreviousHash string    `json:"previousHash"`
	Hash         string    `json:"hash"`
}

// Chain is a threadsafe slice of blocks.
type Chain struct {
	mu     sync.RWMutex
	blocks []Block
}

// NewChain allocates a blockchain seeded with the genesis block.
func NewChain() *Chain {
	genesis := Block{
		Index:        0,
		Timestamp:    time.Now().UTC(),
		Data:         "Genesis Block",
		PreviousHash: "0",
	}
	genesis.Hash = calculateHash(genesis)

	return &Chain{
		blocks: []Block{genesis},
	}
}

// Blocks returns a copy of the block slice so callers cannot mutate the chain directly.
func (c *Chain) Blocks() []Block {
	c.mu.RLock()
	defer c.mu.RUnlock()

	copySlice := make([]Block, len(c.blocks))
	copy(copySlice, c.blocks)
	return copySlice
}

// Latest returns the tip of the chain.
func (c *Chain) Latest() Block {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.blocks[len(c.blocks)-1]
}

// AddBlock appends a block that carries the provided payload.
func (c *Chain) AddBlock(data string) (Block, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	newBlock := generateBlock(c.blocks[len(c.blocks)-1], data)
	if err := validateBlock(newBlock, c.blocks[len(c.blocks)-1]); err != nil {
		return Block{}, err
	}

	c.blocks = append(c.blocks, newBlock)
	return newBlock, nil
}

// Replace swaps in a longer valid chain if the provided slice satisfies the rules.
func (c *Chain) Replace(newBlocks []Block) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(newBlocks) <= len(c.blocks) {
		return false
	}

	if err := validateChain(newBlocks); err != nil {
		return false
	}

	c.blocks = append([]Block(nil), newBlocks...)
	return true
}

func generateBlock(oldBlock Block, data string) Block {
	block := Block{
		Index:        oldBlock.Index + 1,
		Timestamp:    time.Now().UTC(),
		Data:         data,
		PreviousHash: oldBlock.Hash,
	}
	block.Hash = calculateHash(block)
	return block
}

func validateBlock(newBlock, oldBlock Block) error {
	switch {
	case oldBlock.Index+1 != newBlock.Index:
		return fmt.Errorf("invalid index: got %d expected %d", newBlock.Index, oldBlock.Index+1)
	case oldBlock.Hash != newBlock.PreviousHash:
		return errors.New("previous hash mismatch")
	case calculateHash(newBlock) != newBlock.Hash:
		return errors.New("hash mismatch")
	default:
		return nil
	}
}

func validateChain(blocks []Block) error {
	if len(blocks) == 0 {
		return errors.New("chain is empty")
	}

	for i := 1; i < len(blocks); i++ {
		if err := validateBlock(blocks[i], blocks[i-1]); err != nil {
			return err
		}
	}
	return nil
}

func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%s%s", block.Index, block.Timestamp.UTC().Format(time.RFC3339Nano), block.Data, block.PreviousHash)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}
