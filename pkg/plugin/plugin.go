package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/client"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/datasource"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/storage/bolt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

type TezosDatasource struct {
	storage *bolt.BoltStorage
}

func NewTezosDatasource(_ backend.DataSourceInstanceSettings, storage *bolt.BoltStorage) (instancemgmt.Instance, error) {
	return &TezosDatasource{storage: storage}, nil
}

type datasourceConfig struct {
	Chain string `json:"chain"`
}

func (d *TezosDatasource) newDatasource(is *backend.DataSourceInstanceSettings) (*datasource.Datasource, error) {
	var conf datasourceConfig
	if err := json.Unmarshal(is.JSONData, &conf); err != nil {
		return nil, err
	}
	return &datasource.Datasource{
		DB: d.storage,
		Client: &client.Client{
			URL:   is.URL,
			Chain: conf.Chain,
		},
	}, nil
}

func (d *TezosDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ds, err := d.newDatasource(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return nil, err
	}
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		res := d.doQuery(ctx, ds, req.PluginContext, q)
		response.Responses[q.RefID] = res
	}
	return response, nil
}

type queryModel struct {
	Streaming bool     `json:"streaming"`
	Fields    []string `json:"fields"`
}

func (d *TezosDatasource) doQuery(ctx context.Context, ds *datasource.Datasource, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	var q queryModel

	response.Error = json.Unmarshal(query.JSON, &q)
	if response.Error != nil {
		return response
	}

	var blockInfo []*datasource.BlockInfo
	blockInfo, response.Error = ds.GetBlocksInfo(ctx, query.TimeRange.From, query.TimeRange.To)
	if response.Error != nil {
		return response
	}

	timestamps := make([]time.Time, len(blockInfo))
	for i, bi := range blockInfo {
		timestamps[i] = bi.Header.Timestamp
	}
	frame := data.NewFrame("Block Info", data.NewField("time", nil, timestamps))

	validFields := make([]string, 0, len(q.Fields))
	(func() {
		// NewField may panic if data type is not supported because types were expected to be hard coded. Return panic value it as an error instead
		defer (func() {
			if r := recover(); r != nil {
				response.Error = fmt.Errorf("%w", r)
			}
		})()
		for _, field := range q.Fields {
			values := pickFieldByName(blockInfo, field)
			if values != nil {
				validFields = append(validFields, field)
				frame.Fields = append(frame.Fields, data.NewField(field, nil, values))
			}
		}
	})()
	if response.Error != nil {
		return response
	}

	if q.Streaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      strings.Join(validFields, ","),
		}
		frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	}

	response.Frames = append(response.Frames, frame)
	return response
}

func (d *TezosDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	ds, err := d.newDatasource(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return nil, err
	}
	var status *backend.CheckHealthResult
	_, err = ds.Client.GetBlockHeader(ctx, "head")
	if err != nil {
		status = &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}
	}
	status = &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}
	return status, nil
}

func (d *TezosDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
	}, nil
}

func (d *TezosDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	ds, err := d.newDatasource(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return err
	}

	fields := strings.Split(req.Path, ",")

	blockinfoCh, errCh, err := ds.MonitorBlockInfo(ctx)
	if err != nil {
		return err
	}
	for bi := range blockinfoCh {
		frame := data.NewFrame("Block Info", data.NewField("time", nil, []time.Time{bi.Header.Timestamp}))
		for _, field := range fields {
			values := pickFieldByName([]*datasource.BlockInfo{bi}, field)
			if values != nil {
				frame.Fields = append(frame.Fields, data.NewField(field, nil, values))
			}
		}
		err := sender.SendFrame(frame, data.IncludeAll)
		if err != nil {
			log.DefaultLogger.Error("Error sending frame", "error", err)
			continue
		}
	}
	if err, ok := <-errCh; ok && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (d *TezosDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

var (
	_ backend.QueryDataHandler   = (*TezosDatasource)(nil)
	_ backend.CheckHealthHandler = (*TezosDatasource)(nil)
	_ backend.StreamHandler      = (*TezosDatasource)(nil)
)
