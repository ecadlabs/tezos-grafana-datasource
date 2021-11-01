import { DataQueryRequest, DataSourceInstanceSettings, toDataFrame } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { DataSourceOptions, FieldType, Query, QueryType } from './types';
import { lastValueFrom } from 'rxjs';

export class DataSource extends DataSourceWithBackend<Query, DataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<DataSourceOptions>) {
    super(instanceSettings);
  }

  static buildExpression(fields: string[]): string {
    return '{ ' + fields.map((f) => f.slice(f.lastIndexOf('.') + 1) + ': block.' + f).join(', ') + ' }';
  }

  async getFieldsQuery(request: QueryType): Promise<FieldType[]> {
    const targets: Query[] = [{ refId: 'fields', queryType: request }];
    const response = await lastValueFrom(this.query({ targets } as DataQueryRequest<Query>));
    if (response.data.length) {
      const df = toDataFrame(response.data[0]);
      if (df.fields.length > 1 && df.fields[0].type === 'string' && df.fields[1].type === 'string') {
        return df.fields[0].values.toArray().map((v, i) => ({ selector: v, type: df.fields[1].values.get(i) || '' }));
      }
    }
    return [];
  }
}
