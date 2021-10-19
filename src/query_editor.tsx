import { defaults } from 'lodash';
import React, { PureComponent, ChangeEvent } from 'react';
import { InlineField, MultiSelect, Switch } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { DataSourceOptions, defaultQuery, Query } from './types';

type Props = QueryEditorProps<DataSource, Query, DataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onFieldsChange = (values: Array<SelectableValue<string>>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, fields: values.map<string>((v) => v.value || '') });
  };

  onWithStreamingChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, streaming: event.currentTarget.checked });
    onRunQuery();
  };

  render() {
    const { datasource } = this.props;
    const query = defaults(this.props.query, defaultQuery);

    const fieldsVal = (v: string[]) => v.map<SelectableValue<string>>((v) => ({ label: v, value: v }));

    return (
      <div className="gf-form">
        <InlineField label="Select fields" grow>
          <MultiSelect
            options={fieldsVal(datasource.getFields())}
            value={fieldsVal(query.fields)}
            onChange={this.onFieldsChange}
          ></MultiSelect>
        </InlineField>
        <InlineField label="Enable streaming">
          <Switch checked={query.streaming || false} onChange={this.onWithStreamingChange} />
        </InlineField>
      </div>
    );
  }
}
