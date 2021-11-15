package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/client"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

const (
	defaultNode = "https://hangzhounet.api.tez.ie"
	blockNum    = 10
)

type block struct {
	Header header `json:"header"`
}

type header struct {
	Level                     int64         `json:"level"`
	Proto                     uint64        `json:"proto"`
	Predecessor               model.Base58  `json:"predecessor"`
	Timestamp                 time.Time     `json:"timestamp"`
	ValidationPass            uint64        `json:"validation_pass"`
	OperationsHash            model.Base58  `json:"operations_hash"`
	Fitness                   []model.Bytes `json:"fitness"`
	Context                   model.Base58  `json:"context"`
	Priority                  uint64        `json:"priority"`
	ProofOfWorkNonce          model.Bytes   `json:"proof_of_work_nonce"`
	SeedNonceHash             model.Base58  `json:"seed_nonce_hash,omitempty"`
	LiquidityBakingEscapeVote bool          `json:"liquidity_baking_escape_vote"`
	Signature                 model.Base58  `json:"signature"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	node := flag.String("a", defaultNode, "Node address")
	num := flag.Int("n", blockNum, "Block number")
	flag.Parse()

	client := &client.Client{
		URL: *node,
	}

	var b block
	req, err := client.NewGetBlockRequest(context.Background(), "head")
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		log.Fatal(res.StatusCode)
	}

	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&b); err != nil {
		log.Fatal(err)
	}

	head := b.Header.Level
	res.Body.Close()

	for i := 0; i < *num; i++ {
		id := rand.Int31n(int32(head) + 1)
		req, err := client.NewGetBlockRequest(context.Background(), strconv.FormatInt(int64(id), 10))
		if err != nil {
			log.Fatal(err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if res.StatusCode != 200 {
			log.Fatal(res.StatusCode)
		}
		out, err := os.Create(fmt.Sprintf("%d.json", id))
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(out, res.Body); err != nil {
			log.Fatal(err)
		}
		res.Body.Close()
		out.Close()
	}
}
