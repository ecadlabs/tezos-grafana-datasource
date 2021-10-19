import React, { ChangeEvent, PureComponent } from 'react';
import { InlineField, Input, Legend } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { DataSourceOptions } from './types';

export class ConfigEditor extends PureComponent<DataSourcePluginOptionsEditorProps<DataSourceOptions>> {
  render() {
    const { options, onOptionsChange } = this.props;
    const { jsonData } = options;
    const isValidUrl = (u: string) =>
      /^(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?$/.test(u);

    return (
      <div className="gf-form-group">
        <Legend>RPC</Legend>
        <div className="gf-form">
          <InlineField required invalid={!isValidUrl(options.url)} label="URL" labelWidth={15}>
            <Input
              width={40}
              type="text"
              value={options.url}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                onOptionsChange({ ...options, url: event.currentTarget.value })
              }
            />
          </InlineField>
        </div>
        <div className="gf-form">
          <InlineField label="Chain" labelWidth={15}>
            <Input
              width={40}
              type="text"
              value={jsonData.chain}
              onChange={(event: ChangeEvent<HTMLInputElement>) =>
                onOptionsChange({ ...options, jsonData: { ...jsonData, chain: event.currentTarget.value } })
              }
            />
          </InlineField>
        </div>
      </div>
    );
  }
}
