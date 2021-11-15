package plugin

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/client"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/datasource"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/storage/bolt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

const (
	queryBlockInfo       = "block_info"
	queryBlockInfoFields = "block_info_fields"
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
	if err := jtree.Unmarshal(is.JSONData, &conf); err != nil {
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
		res := d.doQuery(ctx, ds, req.PluginContext, &q)
		response.Responses[q.RefID] = res
	}
	return response, nil
}

type queryModel struct {
	Streaming bool   `json:"streaming"`
	Expr      string `json:"expr"`
}
type blockScope struct {
	Block *datasource.BlockInfo `json:"block"`
}

func makeFrame(info []*datasource.BlockInfo, expr string) (*data.Frame, error) {
	var fields []fieldConverter
	fieldIdx := make(map[string]int)
	ctx := cuecontext.New()

	for i, bi := range info {
		scope := blockScope{
			Block: bi,
		}
		val := ctx.CompileString(expr, cue.Scope(ctx.Encode(&scope)))
		if val.Err() != nil {
			return nil, val.Err()
		}

		if val.Kind() == cue.StructKind {
			f, err := val.Fields(cue.All())
			if err != nil {
				return nil, err
			}
			for f.Next() {
				if f.Value().Err() != nil {
					return nil, f.Value().Err()
				}
				name := f.Selector().String()
				if fi, ok := fieldIdx[name]; !ok {
					if converter, err := newFieldConverter(name, f.Value(), len(info)); err != nil {
						return nil, err
					} else {
						fieldIdx[name] = len(fields)
						fields = append(fields, converter)
						if err := converter.Set(i, f.Value()); err != nil {
							return nil, err
						}
					}
				} else if err := fields[fi].Set(i, f.Value()); err != nil {
					return nil, err
				}
			}
		} else if val.Kind() == cue.ListKind {
			v, err := val.List()
			if err != nil {
				return nil, err
			}
			var ii int
			for v.Next() && ii <= len(fields) {
				if v.Value().Err() != nil {
					return nil, v.Value().Err()
				}

				if ii == len(fields) {
					if converter, err := newFieldConverter("", v.Value(), len(info)); err != nil {
						return nil, err
					} else {
						fields = append(fields, converter)
						if err := converter.Set(i, v.Value()); err != nil {
							return nil, err
						}
					}
				} else if err := fields[ii].Set(i, v.Value()); err != nil {
					return nil, err
				}
				ii++
			}
		} else {
			return nil, fmt.Errorf("list or struct type expected: %v", val.Kind())
		}
	}
	frame := data.NewFrame("")
	for _, f := range fields {
		frame.Fields = append(frame.Fields, f.Field())
	}
	return frame, nil
}

func (d *TezosDatasource) doQuery(ctx context.Context, ds *datasource.Datasource, pCtx backend.PluginContext, query *backend.DataQuery) backend.DataResponse {
	queryType := query.QueryType
	if queryType == "" {
		queryType = queryBlockInfo
	}
	response := backend.DataResponse{}
	var q queryModel

	if response.Error = jtree.Unmarshal(query.JSON, &q); response.Error != nil {
		return response
	}

	switch queryType {
	case queryBlockInfo:
		var blockInfo []*datasource.BlockInfo
		if blockInfo, response.Error = ds.GetBlocksInfo(ctx, query.TimeRange.From, query.TimeRange.To); response.Error != nil {
			return response
		}

		var frame *data.Frame
		if frame, response.Error = makeFrame(blockInfo, q.Expr); response.Error != nil {
			return response
		}

		if q.Streaming {
			channel := live.Channel{
				Scope:     live.ScopeDatasource,
				Namespace: pCtx.DataSourceInstanceSettings.UID,
				Path:      base64.RawStdEncoding.EncodeToString([]byte(q.Expr)),
			}
			frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
		}

		response.Frames = append(response.Frames, frame)
		return response

	case queryBlockInfoFields:
		fields := getStructFields((*datasource.BlockInfo)(nil))
		selectors := make([]string, len(fields))
		types := make([]string, len(fields))
		for i, f := range fields {
			selectors[i] = strings.Join(f.Selector, ".")
			types[i] = f.Type.Name()
		}
		frame := data.NewFrame("", data.NewField("selector", nil, selectors), data.NewField("type", nil, types))
		response.Frames = append(response.Frames, frame)
		return response

	default:
		response.Error = fmt.Errorf("unknown query type: %v", queryType)
		return response
	}
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

	expr, err := base64.RawStdEncoding.DecodeString(req.Path)
	if err != nil {
		return err
	}

	blockinfoCh, errCh, err := ds.MonitorBlockInfo(ctx)
	if err != nil {
		return err
	}
	for bi := range blockinfoCh {
		frame, err := makeFrame([]*datasource.BlockInfo{bi}, string(expr))
		if err != nil {
			return err
		}
		if err = sender.SendFrame(frame, data.IncludeAll); err != nil {
			log.DefaultLogger.Error("Error sending frame", "error", err)
			continue
		}
	}
	if err, ok := <-errCh; ok && !errors.Is(err, context.Canceled) {
		return err
	}
	log.DefaultLogger.Info("Stream stopped")
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
