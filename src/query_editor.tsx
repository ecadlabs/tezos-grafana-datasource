import React, { PureComponent, ChangeEvent } from 'react';
import { InlineField, InlineSwitch, Input, MultiSelect } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { DataSourceOptions, FieldType, Query } from './types';

type Props = QueryEditorProps<DataSource, Query, DataSourceOptions>;

interface State {
  fields?: FieldType[];
}

export class QueryEditor extends PureComponent<Props, State> {
  private onFieldsChange = (values: Array<SelectableValue<string>>) => {
    const { onChange, query } = this.props;
    const fields = values.filter((v) => v.value !== undefined).map((v) => v.value || '');
    onChange({ ...query, fields: fields, expr: DataSource.buildExpression(fields) });
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

  private async loadOptions() {
    const { datasource } = this.props;
    this.setState({ ...this.state, fields: await datasource.getFieldsQuery('block_info_fields') });
  }

  constructor(props: Readonly<Props> | Props) {
    super(props);
    this.state = {};
  }

  componentDidMount() {
    this.loadOptions();
  }

  render() {
    const { query } = this.props;
    if (query.fields === undefined) {
      query.fields = ['header.timestamp'];
      query.expr = DataSource.buildExpression(query.fields);
    }

    const fieldsVal = (v: string[]) => v.map<SelectableValue<string>>((v) => ({ label: v, value: v }));

    return (
      <div className="gf-form">
        <InlineField label="Expression">
          <InlineSwitch checked={query.useExpr || false} onChange={this.onUseExprChange} />
        </InlineField>
        {query.useExpr ? (
          <InlineField grow>
            <Input value={query.expr} type="text" onChange={this.onExprChange}></Input>
          </InlineField>
        ) : (
          <InlineField label="Select fields" grow>
            <MultiSelect
              menuShouldPortal
              isSearchable
              options={this.state.fields?.map<SelectableValue<string>>((v) => ({
                label: v.selector,
                value: v.selector,
                description: v.type,
              }))}
              isLoading={this.state.fields === undefined}
              value={fieldsVal(query.fields)}
              onChange={this.onFieldsChange}
            ></MultiSelect>
          </InlineField>
        )}
        <InlineField label="Enable streaming">
          <InlineSwitch checked={query.streaming || false} onChange={this.onWithStreamingChange} />
        </InlineField>
      </div>
    );
  }
}
