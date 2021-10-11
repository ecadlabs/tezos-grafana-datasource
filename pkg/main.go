package main

import (
	"os"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/plugin"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/storage/bolt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	log.DefaultLogger.Debug("Running Tezos datasource")

	storage, err := bolt.NewBoltStorage()
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}

	newInstanceFunc := func(is backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
		return plugin.NewTezosDatasource(is, storage)
	}

	if err := datasource.Manage("tezos-datasource", newInstanceFunc, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}

	if err := storage.Close(); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
