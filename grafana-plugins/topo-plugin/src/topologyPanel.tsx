// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import React, { useRef, useState, useEffect } from 'react';
import * as G6 from '@antv/g6';
import _ from 'lodash';
import './topology/node';
import './topology/edge';
import { nodeTooltip } from './topology/tooltip';
import TopoLegend from './topology/legend';
import { metricList, metricDataName, layoutOptions, directionOptions, viewRadioOptions, showServiceOptions, NodeDataProps, EdgeDataProps, 
  transformData, transformWorkload, getNodeTypes, nsRelationHandle, connFailNSRelationHandle, workloadRelationHandle, connFailWorkloadRelationHandle, TopologyProps, topoMerge,
  updateEdge } from './topology/services'; 
import { buildLayout, serviceLineUpdate, updateLinesAndNodes } from './topology/topology'
import { PanelProps, SelectableValue } from '@grafana/data';
import { MetricType, SimpleOptions } from 'types';
import { css, cx } from 'emotion';
import { stylesFactory, Select, RadioButtonGroup, Icon, Tooltip, Spinner } from '@grafana/ui';
import { metricQuery } from './dataSource'; 

interface VolumeProps {
  maxSentVolume: number; 
  maxReceiveVolume: number;
  minSentVolume: number; 
  minReceiveVolume: number;
}

let SGraph: any;
// let connFailTopoData: any;
let connFailTopo: TopologyProps;
const initActiveMetricList: MetricType[] = ['latency', 'calls', 'errorRate'];

interface Props extends PanelProps<SimpleOptions> {}

export const TopologyPanel: React.FC<Props> = ({ options, data, width, height, timeRange, replaceVariables }) => {
  let graphRef: any = useRef();
  const styles = getStyles();
  const namespace = replaceVariables('$namespace');
  const workload = replaceVariables('$workload');

  const topoData = useRef<any>();
  const nodeData = useRef<NodeDataProps>();
  const edgeData = useRef<EdgeDataProps>();

  const [graphData, setGraphData] = useState<any>({}); 
  const [layout, setLayout] = useState<'dagre' | 'force'>('dagre'); 
  const [loading, setLoading] = useState<boolean>(false); 
  const [showCheckbox] = useState<boolean>(true); // 是否显示show services的切换选项
  const [showService, setShowService] = useState<boolean>(false); // show services
  const [showView, setShowView] = useState<boolean>(false); // 单个workload拓扑视图下  workload与pod切换的选项
  const [firstChangeDir, setFirstChangeDir] = useState<boolean>(false);
  const [direction, setDirection] = useState<string>('LR');
  const [view, setView] = useState<string>('workload_view');
  const [lineMetric, setLineMetric] = useState<any>('latency');
  const [volumes, setVolumes] = useState<VolumeProps>({maxSentVolume: 0, maxReceiveVolume: 0, minSentVolume: 0, minReceiveVolume: 0});
  const [nodeTypesList, setNodeTypesList] = useState<any[]>([]);
  const [activeMetricList, setActiveMetricList] = useState<MetricType[]>(initActiveMetricList);

  // console.log(options, data, namespace, workload, width, height, timeRange);
  // draw topology
  const draw = (gdata: any, metric = lineMetric, serviceLine = showService) => {
    const inner: any = document.getElementById('kindling_topo');
    inner.innerHTML = '';
    let data = _.cloneDeep(gdata);
    const graph = new G6.Graph({
      // renderer: 'svg',
      container: 'kindling_topo',
      width: width - 240,
      height: height,
      fitView: true,
      fitViewPadding: 10,
      maxZoom: 1.5,
      minZoom: 0.25,
      fitCenter: true,
      autoPaint: false,
      plugins: [nodeTooltip],
      modes: {
        default: [
          {
            type: 'drag-canvas',
            enableOptimize: true,
          }, {
            type: 'zoom-canvas',
            maxZoom: 1.5,
            minZoom: 0.25
          }, 
          'drag-node'
        ]
      },
      layout: buildLayout(layout, direction),
      defaultNode: {
        type: 'custom-node'
      },
      defaultEdge: {
        type: 'service-edge',
        labelCfg: {
          refY: serviceLine ? 15 : 10,
          autoRotate: true,
          style: {
            fontWeight: 400,
            fill: '#C2C8D5',
          }
        },
        style: {
          radius: 10,
          offset: 5,
          endArrow: true,
          lineWidth: 1,
          stroke: '#C2C8D5',
        }
      }
    });
    graph.data(data);
    graph.render();
    /**
     * Close local rendering to solve the lingering problem of some browser drag nodes
     * 关闭局部渲染，解决部分浏览器拖拽节点的残影问题
     */
    graph.get('canvas').set('localRefresh', false);

    SGraph = graph;
    serviceLineUpdate(SGraph, direction);
    updateLinesAndNodes(SGraph, options, volumes, metric, serviceLine);
  };
 
  /**
   * When you go back to the topology drawing data, update the type array of the corresponding node and the value of Max and min of the flow on the side
   * 重新回去拓扑绘制数据时，更新对应的节点的类型数组和边上流量max、min的数值
   */
  const calculateVolume = (gdata: any) => {
    let volumeData: VolumeProps = {
      maxSentVolume: _.max(_.map(gdata.edges, 'sentVolume')),
      maxReceiveVolume: _.max(_.map(gdata.edges, 'receiveVolume')),
      minSentVolume: _.min(_.map(gdata.edges, 'sentVolume')),
      minReceiveVolume: _.min(_.map(gdata.edges, 'receiveVolume'))
    }
    setVolumes(volumeData);
  }
  const buildtopoData = () => {
    let nodes: any[] = [], edges: any[] = [];
    // namespace select all
    if (namespace.indexOf(',') > -1) {
      let result: any = nsRelationHandle(topoData.current, nodeData.current!, edgeData.current!);
      nodes = [].concat(result.nodes);
      edges = [].concat(result.edges);
    } else {
      let showPod = workload.indexOf(',') === -1;
      setView(showPod ? 'pod_view' : 'workload_view');
      let result: any = workloadRelationHandle(transformWorkload(workload), namespace, topoData.current, nodeData.current!, edgeData.current!, showPod, showService);
      nodes = [].concat(result.nodes);
      edges = [].concat(result.edges);
    }
    
    let gdata = {
      nodes: nodes,
      edges: edges
    }
    console.log(gdata);
    setGraphData(gdata);
    draw(gdata, lineMetric);
    let nodeTypesList = getNodeTypes(gdata.nodes);
    setNodeTypesList(nodeTypesList);
  }

  const initData = () => {
    topoData.current = transformData(_.filter(data.series, (item: any) => item.refId === 'A'));
    let edgeTimeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'B'));
    edgeData.current = {
      edgeCallData: topoData.current,
      edgeTimeData,
      edgeSendVolumeData: [],
      edgeReceiveVolumeData: [],
      edgeRetransmitData: [],
      edgeRTTData: [],
      edgePackageLostData: []
    };
    
    let nodeCallsData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'C'));
    let nodeTimeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'D'));
    let nodeSendVolumeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'E'));
    let nodeReceiveVolumeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'F'));
    nodeData.current = {
      nodeCallsData,
      nodeTimeData,
      nodeSendVolumeData,
      nodeReceiveVolumeData
    };
    // console.log('edgeData', edgeData.current);
    // console.log('nodeData', nodeData.current);

    buildtopoData();
  }

  useEffect(() => {
    if (SGraph) {
      updateLinesAndNodes(SGraph, options, volumes, lineMetric, showService);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [volumes, lineMetric]);

  useEffect(() => {
    setLineMetric('latency');
    setActiveMetricList(initActiveMetricList);
    if (namespace.indexOf(',') > -1) {
      setShowView(workload.indexOf(',') === -1);
    } else {
      setShowView(workload.indexOf(',') === -1);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [namespace, workload]);
  
  useEffect(() => {
    if (data.state === 'Done') {
      setLoading(false);
      initData();
    }
	// eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data]);

  // 改变拓扑布局方式，重新绘制拓扑图
  useEffect(() => {
    if (SGraph && !_.isEmpty(graphData)) {
      if (lineMetric === 'connFail') {
        const data = topoMerge(graphData, connFailTopo);
        draw(data);
      } else {
        console.log(graphData);
        draw(graphData);
      }  
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [layout]);

  // When the segment indicator is switched, the corresponding segment style is updated
  const lineMetricChange = async (opt: SelectableValue<MetricType>) => {
    let metric: MetricType = opt.value!;
    if (activeMetricList.indexOf(metric) === -1) {
      setLoading(true);
      setActiveMetricList([...activeMetricList, metric]);
      let metricResult = await metricQuery(metric, namespace, workload, timeRange);
      const metricData = transformData(metricResult.data);
      console.log('metricData', metricData);
      setLoading(false);
      if (metric === 'connFail') {
        // connFailTopoData = metricData;
        if (namespace.indexOf(',') > -1) {
          connFailTopo = connFailNSRelationHandle(metricData);
        } else {
          connFailTopo = connFailWorkloadRelationHandle(transformWorkload(workload), namespace, metricData, showView, showService);
        }
        console.log('connFailTopo', connFailTopo);
        const data = topoMerge(graphData, connFailTopo);
        draw(data, metric);
      } else {
        // @ts-ignore
        edgeData.current[metricDataName[metric]] = metricData;
        updateEdge(graphData, metric, metricData, SGraph, namespace.indexOf(',') > -1, showService);
        // updateLinesAndNodes(SGraph, options, volumes, metric, showService);
        if (metric === 'sentVolume' || metric === 'receiveVolume') {
          console.log('graphData', graphData);
          calculateVolume(graphData);
        }
      }
    } else {
      if (metric === 'connFail') {
        const data = topoMerge(graphData, connFailTopo);
        draw(data, metric);
      } else {
        // 若当前选中指标时connFail时，需要重新单独绘制没有connFailTopo的拓扑图
        if (lineMetric === 'connFail') {
          draw(graphData, metric);
        }
      }
    } 
    setLineMetric(metric);
  }
  // change layout
  const changeLayout = (opt: any) => {
    if (opt.value === layout) {
      return;
    }
    setLayout(opt.value);
  } 
  // change direction whrn layout is dagre
  const changeDirection = (value: any) => {
    setDirection(value);
    SGraph.updateLayout({
      rankdir: value
    });
    SGraph.fitView(10);
    if (!firstChangeDir) {
      serviceLineUpdate(SGraph, value);
      setFirstChangeDir(true);
    }
  }
  // Whether to display service calls on invocation relationships
  const changeShowService = () => {
    let show = !showService ? true : false;
    setShowService(show);
    let { nodes, edges } = workloadRelationHandle(workload, namespace, topoData.current, nodeData.current!, edgeData.current!, view === 'pod_view', show);
    let gdata = {
      nodes: nodes,
      edges: edges
    }
    draw(gdata, lineMetric, show);
    setGraphData(gdata);
    calculateVolume(gdata);
  }
  // toggle View Mode。Switch between the workload view and pod view
  const changeView = (value: any) => {
    setView(value);
    let { nodes, edges } = workloadRelationHandle(workload, namespace, topoData.current, nodeData.current!, edgeData.current!, value === 'pod_view', showService);
    let gdata = {
      nodes: nodes,
      edges: edges
    }
    draw(gdata);
    setGraphData(gdata);
    calculateVolume(gdata);
  }

  return (
    <div
      className={cx(
        styles.wrapper,
        css`
          width: ${width}px;
          height: ${height}px;
        `
      )}
    >
      <div className={styles.topLineMetric}>
        <div className={styles.metricSelect}>
          <span style={{ width: '180px' }}>Call Line Metric</span>
          <Select value={lineMetric} options={metricList} onChange={lineMetricChange}/>
        </div>
      </div>
      <div className={styles.topRightWarp}>
        <div className={styles.viewRadioMode}>
          <div>
            <span>Layout</span>
            <Tooltip content="change topology layout。">
              <Icon name="question-circle" />
            </Tooltip>
          </div>
          <div style={{ width: 150 }}>
            <Select value={layout} options={layoutOptions} onChange={changeLayout}/>
          </div>
        </div>
        {
          layout === 'dagre' ? <div className={styles.viewRadioMode}>
            <div>
              <span>Layout Direction</span>
              <Tooltip content="change Dargre topology layout direction mean top to bottom，LR mean left to right。">
                <Icon name="question-circle" />
              </Tooltip>
            </div>
            <RadioButtonGroup options={directionOptions} value={direction} onChange={changeDirection}/>
          </div> : null
        }
        {
          showView ? <div className={styles.viewRadioMode}>
            <span>View Mode</span>
            <RadioButtonGroup options={viewRadioOptions} value={view} onChange={changeView}/>
          </div> : null
        }
        {
          showCheckbox ? <div className={styles.viewRadioMode}>
            <div>
              <span>Show Services</span>
              <Tooltip content="if the network communicate by Kubernetes service, the service name will be shown。">
                <Icon name="question-circle" />
              </Tooltip>
            </div>
            <RadioButtonGroup options={showServiceOptions} value={showService} onChange={changeShowService}/>
          </div> : null
        }
        <TopoLegend typeList={nodeTypesList} metric={lineMetric} volumes={volumes} options={options}/>
      </div>
      <div id="kindling_topo" style={{ height: '100%' }} ref={graphRef}></div>
      {
        loading ? <div className={styles.spinner_warp}>
          <Spinner className={styles.spinner_icon}/>
        </div> : null
      }
    </div>
  );
};

const getStyles = stylesFactory(() => {
  return {
    wrapper: css`
      position: relative;
    `,
    topLineMetric: css`
      position: absolute;
      top: 0;
      left: 0;
      z-index: 10;
      display: flex;
      flex-direction: column;
    `,
    filterWarp: css`
      display: flex;
      align-items: center;
      margin-bottom: 10px;
    `,
    metricSelect: css`
      display: flex;
      align-items: center;
      margin-bottom: 10px;
    `,
    topRightWarp: css`
      position: absolute;
      top: 0;
      right: 0;
      z-index: 10;
      display: flex;
      flex-direction: column;
      width: 245px;
    `,
    viewRadioMode: css`
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 10px;
    `,
    svg: css`
      position: absolute;
      top: 0;
      left: 0;
    `,
    textBox: css`
      position: absolute;
      bottom: 0;
      left: 0;
      padding: 10px;
    `,
    spinner_warp: css`
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-color: #181b1fc2;
      z-index: 20;
    `,
    spinner_icon: css`
      position: absolute;
      font-size: xx-large;
      top: 48%;
      left: 49%;
    `,
  };
});
