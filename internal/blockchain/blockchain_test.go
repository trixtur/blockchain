package blockchain

import (
	"fmt"
	"testing"
)

func TestAddBlock(t *testing.T) {
	tests := []struct {
		name string
		data []string
	}{
		{name: "single append", data: []string{"hello"}},
		{name: "multiple appends", data: []string{"block-1", "block-2", "block-3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewChain()
			for i, payload := range tt.data {
				block, err := chain.AddBlock(payload)
				if err != nil {
					t.Fatalf("AddBlock error: %v", err)
				}
				if block.Index != i+1 {
					t.Fatalf("unexpected index %d for iteration %d", block.Index, i)
				}
				if block.Data != payload {
					t.Fatalf("block contains %q want %q", block.Data, payload)
				}
				if block.PreviousHash != chain.Blocks()[i].Hash {
					t.Fatalf("previous hash mismatch for block %d", block.Index)
				}
			}
		})
	}
}

func TestReplaceChain(t *testing.T) {
	tests := []struct {
		name      string
		modifier  func([]Block) []Block
		wantRepl  bool
		finalSize int
	}{
		{
			name: "accepts longer valid chain",
			modifier: func(blocks []Block) []Block {
				return extend(blocks, 2)
			},
			wantRepl:  true,
			finalSize: 3,
		},
		{
			name: "rejects shorter chain",
			modifier: func(blocks []Block) []Block {
				return blocks[:1]
			},
			wantRepl:  false,
			finalSize: 1,
		},
		{
			name: "rejects invalid chain",
			modifier: func(blocks []Block) []Block {
				replacement := extend(blocks, 1)
				replacement[1].PreviousHash = "tampered"
				return replacement
			},
			wantRepl:  false,
			finalSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewChain()
			replacement := tt.modifier(chain.Blocks())
			if got := chain.Replace(replacement); got != tt.wantRepl {
				t.Fatalf("Replace returned %v want %v", got, tt.wantRepl)
			}
			if gotSize := len(chain.Blocks()); gotSize != tt.finalSize {
				t.Fatalf("chain length %d want %d", gotSize, tt.finalSize)
			}
		})
	}
}

func TestBranchingAndMerge(t *testing.T) {
	tests := []struct {
		name       string
		run        func(t *testing.T) *Chain
		wantTip    string
		wantLength int
	}{
		{
			name: "branch overrides shorter trunk",
			run: func(t *testing.T) *Chain {
				chain := NewChain()
				if _, err := chain.AddBlock("trunk-1"); err != nil {
					t.Fatalf("add trunk block: %v", err)
				}
				branch := extendWithData(chain.Blocks()[:1], "branch-1", "branch-2")
				if ok := chain.Replace(branch); !ok {
					t.Fatalf("expected branch replacement to succeed")
				}
				return chain
			},
			wantTip:    "branch-2",
			wantLength: 3,
		},
		{
			name: "trunk reasserts when longer",
			run: func(t *testing.T) *Chain {
				chain := NewChain()
				if _, err := chain.AddBlock("trunk-1"); err != nil {
					t.Fatalf("add trunk block: %v", err)
				}
				branch := extendWithData(chain.Blocks()[:1], "branch-1", "branch-2")
				if ok := chain.Replace(branch); !ok {
					t.Fatalf("expected branch replacement to succeed")
				}
				trunkCandidate := extendWithData(chain.Blocks()[:1], "trunk-1", "trunk-2", "trunk-3")
				if ok := chain.Replace(trunkCandidate); !ok {
					t.Fatalf("expected trunk to reassert when longer")
				}
				return chain
			},
			wantTip:    "trunk-3",
			wantLength: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := tt.run(t)
			if got := chain.Latest().Data; got != tt.wantTip {
				t.Fatalf("tip data %q want %q", got, tt.wantTip)
			}
			if gotLen := len(chain.Blocks()); gotLen != tt.wantLength {
				t.Fatalf("chain length %d want %d", gotLen, tt.wantLength)
			}
		})
	}
}

func TestAttackSimulationTamperedReplacement(t *testing.T) {
	chain := NewChain()
	_, _ = chain.AddBlock("legit-1")
	_, _ = chain.AddBlock("legit-2")
	original := chain.Blocks()

	attackerChain := extendWithData(chain.Blocks(), "attacker-1", "attacker-2")
	attackerChain[2].PreviousHash = "forged"

	if ok := chain.Replace(attackerChain); ok {
		t.Fatalf("expected tampered chain to be rejected")
	}
	if gotLen := len(chain.Blocks()); gotLen != len(original) {
		t.Fatalf("chain length changed after rejected attack: %d", gotLen)
	}
	if chain.Latest().Data != "legit-2" {
		t.Fatalf("chain tip mutated after rejected attack")
	}
}

func extend(blocks []Block, count int) []Block {
	copySlice := append([]Block(nil), blocks...)
	prev := copySlice[len(copySlice)-1]
	for i := 0; i < count; i++ {
		next := generateBlock(prev, fmt.Sprintf("data-%d", i))
		copySlice = append(copySlice, next)
		prev = next
	}
	return copySlice
}

func extendWithData(blocks []Block, payloads ...string) []Block {
	copySlice := append([]Block(nil), blocks...)
	prev := copySlice[len(copySlice)-1]
	for _, payload := range payloads {
		next := generateBlock(prev, payload)
		copySlice = append(copySlice, next)
		prev = next
	}
	return copySlice
}
