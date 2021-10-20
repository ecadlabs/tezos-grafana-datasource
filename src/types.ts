import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface Query extends DataQuery {
  queryType?: QueryType;
  streaming?: boolean;
  fields?: string[];
  expr?: string;
  useExpr?: boolean;
}

export interface DataSourceOptions extends DataSourceJsonData {
  chain?: string;
}

export interface FieldType {
  selector: string;
  type: string;
}

export type QueryType = 'block_info' | 'block_info_fields';
