import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { DataSourceOptions, Query } from './types';

export class DataSource extends DataSourceWithBackend<Query, DataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<DataSourceOptions>) {
    super(instanceSettings);
  }

  getFields(): string[] {
    // TODO
    return ['hash', 'level', 'priority', 'timestamp', 'minimal_delay', 'delay'];
  }
}
