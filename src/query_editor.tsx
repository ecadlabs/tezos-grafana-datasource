import React, { PureComponent, ChangeEvent } from 'react';
import { AsyncMultiSelect, InlineField, InlineSwitch, Input } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { DataSourceOptions, Query } from './types';

type Props = QueryEditorProps<DataSource, Query, DataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  private onFieldsChange = (values: Array<SelectableValue<string>>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, fields: values.map<string>((v) => v.value || '') });
  };

  private onWithStreamingChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, streaming: event.currentTarget.checked });
    onRunQuery();
  };

  private onExprChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, expr: event.currentTarget.value });
  };

  private onUseExprChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, useExpr: event.currentTarget.checked });
  };

  private loadOptions = async (): Promise<Array<SelectableValue<string>>> => {
    const { datasource } = this.props;
    const res = await datasource.getFieldsQuery('block_info_fields');
    return res.map<SelectableValue<string>>((v) => ({ label: `${v.selector}: ${v.type}`, value: v.selector }));
  };

  render() {
    const { query } = this.props;
    query.fields = query.fields || [];

    const fieldsVal = (v: string[]) => v.map<SelectableValue<string>>((v) => ({ label: v, value: v }));

    return (
      <div className="gf-form">
        <InlineField label="Extended">
          <InlineSwitch checked={query.useExpr || false} onChange={this.onUseExprChange} />
        </InlineField>
        {query.useExpr ? (
          <InlineField label="Expression" grow>
            <Input value={query.expr} type="text" onChange={this.onExprChange}></Input>
          </InlineField>
        ) : (
          <InlineField label="Select fields" grow>
            <AsyncMultiSelect
              menuShouldPortal
              defaultOptions
              loadOptions={this.loadOptions}
              value={fieldsVal(query.fields)}
              onChange={this.onFieldsChange}
            ></AsyncMultiSelect>
          </InlineField>
        )}
        <InlineField label="Enable streaming">
          <InlineSwitch checked={query.streaming || false} onChange={this.onWithStreamingChange} />
        </InlineField>
      </div>
    );
  }
}
