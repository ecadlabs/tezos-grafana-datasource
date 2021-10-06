package datasource

import (
	"context"
	"testing"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/client"
	"github.com/ecadlabs/tezos-grafana-datasource/storage/bolt"
	"github.com/stretchr/testify/require"
)

func TestDatasource(t *testing.T) {
	client := client.Client{
		URL: "https://rpc.tzstats.com",
	}

	db, err := bolt.NewBoltStorage()
	require.NoError(t, err)

	defer db.Close()

	ds := Datasource{
		DB:     db,
		Client: &client,
	}

	now := time.Now()
	info, err := ds.GetBlockTimes(context.Background(), now.Add(-time.Hour), now)
	require.NoError(t, err)

	for _, bi := range info {
		t.Logf("%s: p=%d s=%d delay=%d dif=%d rel=%f\n", bi.Header.Hash, bi.Header.Priority, bi.Stat.Slots, bi.Delay, bi.Delay-bi.MinDelay, float64(bi.Delay)/float64(bi.MinDelay))
	}
}
