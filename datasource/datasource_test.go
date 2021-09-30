package datasource

import (
	"context"
	"testing"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/client"
	"github.com/ecadlabs/tezos-grafana-datasource/storage"
	"github.com/stretchr/testify/require"
)

func TestDatasource(t *testing.T) {
	client := client.Client{
		URL: "https://mainnet.api.tez.ie/",
	}

	db, err := storage.NewBoltStorage()
	require.NoError(t, err)

	defer db.Close()

	ds := Datasource{
		DB:     db,
		Client: &client,
	}

	now := time.Now()
	info, err := ds.GetBlockTimes(context.Background(), now.Add(-time.Minute*10), now)
	require.NoError(t, err)

	for _, bi := range info {
		delay := bi.Timestamp.Sub(bi.PredecessorTimestamp)
		ideal := bi.MinimalValidTime.Sub(bi.PredecessorTimestamp)
		t.Logf("%s: p=%d s=%d delay=%d dif=%d rel=%f\n", bi.Hash, bi.Priority, bi.EndorsementSlots, delay, delay-ideal, float64(delay)/float64(ideal))
	}
}
