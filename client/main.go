package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"example.com/blockchain/internal/blockchain"
)

func main() {
	host := flag.String("host", "http://localhost:8080", "blockchain server base URL")
	action := flag.String("action", "list", "action to perform: list or mine")
	data := flag.String("data", "", "payload for the mine action")
	flag.Parse()

	switch strings.ToLower(*action) {
	case "list":
		if err := listBlocks(*host); err != nil {
			log.Fatal(err)
		}
	case "mine":
		if *data == "" {
			log.Fatal("data flag is required for mine action")
		}
		if err := mineBlock(*host, *data); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown action %q", *action)
	}
}

func listBlocks(host string) error {
	resp, err := http.Get(host + "/blocks")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}

	var blocks []blockchain.Block
	if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
		return err
	}

	if len(blocks) == 0 {
		fmt.Println("No blocks yet.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "INDEX\tTIMESTAMP\tDATA\tHASH")
	for _, block := range blocks {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", block.Index, block.Timestamp.Format("2006-01-02T15:04:05Z07:00"), block.Data, block.Hash[:16])
	}
	return w.Flush()
}

func mineBlock(host, data string) error {
	payload, err := json.Marshal(map[string]string{"data": data})
	if err != nil {
		return err
	}

	resp, err := http.Post(host+"/mine", "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}

	var block blockchain.Block
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return err
	}

	fmt.Printf("Mined block #%d with hash %s\n", block.Index, block.Hash)
	return nil
}
