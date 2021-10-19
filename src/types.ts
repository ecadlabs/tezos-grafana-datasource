import { DataQuery, DataSourceJsonData } from '@grafana/data';

export const defaultQuery: Partial<Query> = {
  fields: [],
};

export interface Query extends DataQuery {
  streaming?: boolean;
  fields: string[];
}

export interface DataSourceOptions extends DataSourceJsonData {
  chain?: string;
}
