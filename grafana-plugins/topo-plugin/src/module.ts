// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import { PanelPlugin } from '@grafana/data';
import { SimpleOptions } from './types';
import { TopologyPanel } from './topologyPanel';

export const plugin = new PanelPlugin<SimpleOptions>(TopologyPanel).setPanelOptions(builder => {
  return builder
    // .addSelect({
    //   path: 'layout', 
    //   name: 'Layout',
    //   defaultValue: 'dagre',
    //   description: 'change topology layout。',
    //   settings: {
    //     options: [
    //       {
    //         value: 'dagre',
    //         label: 'Dagre Layout',
    //       },
    //       {
    //         value: 'force',
    //         label: 'Force Layout',
    //       }
    //     ],
    //   },
    // })
    .addBooleanSwitch({
      path: 'showLatency',
      name: 'Latency Config',
      description: 'The value change latency threshold config',
      defaultValue: false,
    })
    .addNumberInput({
      path: 'normalLatency',
      defaultValue: 20,
      name: 'Normal',
      showIf: config => config.showLatency,
    })
    .addNumberInput({
      path: 'abnormalLatency',
      defaultValue: 1000,
      name: 'Abnormal',
      showIf: config => config.showLatency,
    })
    .addBooleanSwitch({
      path: 'showRtt',
      name: 'RTT Config',
      description: 'The value change rtt threshold config',
      defaultValue: false,
    })
    .addNumberInput({
      path: 'normalRtt',
      defaultValue: 100,
      name: 'Normal',
      showIf: config => config.showRtt,
    })
    .addNumberInput({
      path: 'abnormalRtt',
      defaultValue: 200,
      name: 'Abnormal',
      showIf: config => config.showRtt,
    })
});
