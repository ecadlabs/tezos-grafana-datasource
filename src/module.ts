import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './config_editor';
import { QueryEditor } from './query_editor';
import { Query, DataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, Query, DataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
