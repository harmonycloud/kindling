// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import { formatTime, formatCount, formatKMBT, formatPercent } from './tooltip';

// Create a topology layout based on the Layout field in the configuration
export const buildLayout = (type: string, direction: string) => {
    const layout: any = {
        type: 'dagre',
        rankdir: direction,
        align: 'DL',
        ranksep: 60
    }
    if (type === 'dagre') {
        return layout;
    }
    if (type === 'force') {
        return {
            type: 'fruchterman',
            center: [0, 0],
            gravity: 2,
            gpuEnabled: true
            // type: 'gForce', // 布局名称
            // center: [0, 0],
            // preventOverlap: true, // 布局参数，是否允许重叠
            // nodeSize: 40, // 布局参数，节点大小，用于判断节点是否重叠
            // linkDistance: 150, // 布局参数，边长  
        };
    }
    return layout;
}

// text overFlow handle
export const nodeTextHandle = (text: string, num = 20) => {
    if (text && text.length > num) {
        return text.substring(0, num) + '...';
    } else {
        return text;
    }
};

// 当勾选View Service Call时，显示service的调用边，两个节点之间存在多条调用关系，使用弧线绘制对应的调用关系
export const serviceLineUpdate = (SGraph: any, dir: string) => {
    let activeList: any[] = [];
    const edges = SGraph.getEdges();
    const offest = 5;
    edges.forEach((edge: any) => {
        let edgeModel = edge.getModel();
        let active = activeList.findIndex((item: any) => (item.source === edgeModel.source && item.target === edgeModel.target) || (item.source === edgeModel.target && item.target === edgeModel.source));
        if (active === -1) {
            activeList.push({
                source: edgeModel.source,
                target: edgeModel.target
            });
            let lines = edges.filter((itemEdge: any) => {
                let item = itemEdge.getModel();
                return (item.source === edgeModel.source && item.target === edgeModel.target) || (item.source === edgeModel.target && item.target === edgeModel.source)
            });
            if (lines.length > 1) {
                let oddNum = 0, evenNum = 0;
                lines.forEach((item: any, idx: number) => {
                    let line: any = item.getContainer();
                    // line.type = 'service-edge2';
                    let curveOffset = 0;
                    // if (idx % 2 === 0) {
                    //   curveOffset = arc * (1 + (1 * evenNum));
                    //   evenNum ++;
                    // } else {
                    //   curveOffset = -arc * (1 + (1 * oddNum));
                    //   oddNum ++;
                    // }
                    // console.log(item, curveOffset)
                    // line.curveOffset = curveOffset;
                    // SGraph.updateItem(item, line);
                    if (idx % 2 === 0) {
                        curveOffset = -offest * (1 + (1 * evenNum));
                        evenNum ++;
                    } else {
                        curveOffset = offest * (1 + (1 * oddNum));
                        oddNum ++;
                    }
                    if (dir === 'TB') {
                        line.translate(curveOffset, 0);
                    } else {
                        line.translate(0, curveOffset);
                    }
                });
            }
        }
    });
}
/**
 * update edge style based on the current metric
 * 根据当前指标选择更新边的样式
 */
export const updateLinesAndNodes = (SGraph: any, options: any, volumes: any, metric: string, serviceLine: boolean) => {
    const nodes = SGraph.getNodes();
    const edges = SGraph.getEdges();
    if (metric === 'latency' || metric === 'rtt' || metric === 'errorRate' || metric === 'connFail') {
        edges.forEach((edge: any, idx: number) => {
            let edgeModel = edge.getModel();
            let color: string;
            
            if (metric === 'latency') {
                color = edgeModel.latency > options.abnormalLatency ? '#ff4c4c' : (edgeModel.latency > options.normalLatency ? '#f3ff69' : '#C2C8D5');
                edgeModel.label = formatTime(edgeModel.latency);
            } else if (metric === 'rtt') {
                color = edgeModel.rtt > options.abnormalRtt ? '#ff4c4c' : (edgeModel.rtt > options.normalRtt ? '#f3ff69' : '#C2C8D5');
                edgeModel.label = formatTime(edgeModel.rtt);
            } else if (metric === 'errorRate') {
                color = edgeModel.errorRate > 0 ? '#ff4c4c' : '#C2C8D5';
                edgeModel.label = formatPercent(edgeModel.errorRate);
            } else {
                color = edgeModel.connFail > 0 ? '#ff4c4c' : '#C2C8D5';
                edgeModel.label = formatPercent(edgeModel.connFail);
            }
            edgeModel.opposite && (edgeModel.labelCfg.refY = -10);
            edgeModel.style.stroke = color;
            if (serviceLine) {
                edgeModel.rectColor = color;
            }
            edgeModel.style.lineWidth = 1;
            SGraph.updateItem(edge, edgeModel);
        });
        nodes.forEach((node: any) => {
            let nodeModel = node.getModel();
            if (metric === 'latency') {
                nodeModel.status = nodeModel.latency > options.abnormalLatency ? 'red' : (nodeModel.latency > options.normalLatency ? 'yellow' : 'green');
            } else if (metric === 'rtt') {
                nodeModel.status = 'green';
            } else if (metric === 'errorRate') {
                nodeModel.status = nodeModel.errorRate > 0 ? 'red' : 'green';
            } else {
                nodeModel.status = 'green';
            }
            SGraph.updateItem(node, nodeModel);
        });
    } else if (metric === 'sentVolume' || metric === 'receiveVolume'){
        if (metric === 'sentVolume') {
            let volumeStep = volumes.maxSentVolume / 5;
            if (edges.length === 1) {
                let edge = edges[0];
                let edgeModel = edge.getModel();
                edgeModel.style.lineWidth = 1;
                edgeModel.style.stroke = '#C2C8D5';
                edgeModel.label = formatKMBT(edgeModel.sentVolume);
                SGraph.updateItem(edge, edgeModel);
            } else {
                edges.forEach((edge: any, idx: number) => {
                    let edgeModel = edge.getModel();
                    let step = Math.floor(edgeModel.sentVolume / volumeStep);
                    edgeModel.style.lineWidth = step === 0 ? 1 : 1.5 * step;
                    edgeModel.style.stroke = '#C2C8D5';
                    edgeModel.label = formatKMBT(edgeModel.sentVolume);
                    edgeModel.opposite && (edgeModel.labelCfg.refY = -10);
                    SGraph.updateItem(edge, edgeModel);
                });
            }
        } else {
            let volumeStep = volumes.maxReceiveVolume / 5;
            if (edges.length === 1) {
                let edge = edges[0];
                let edgeModel = edge.getModel();
                edgeModel.style.lineWidth = 1;
                edgeModel.style.stroke = '#C2C8D5';
                edgeModel.label = formatKMBT(edgeModel.receiveVolume);
                SGraph.updateItem(edge, edgeModel);
            } else {
                edges.forEach((edge: any, idx: number) => {
                    let edgeModel = edge.getModel();
                    let step = Math.floor(edgeModel.receiveVolume / volumeStep);
                    edgeModel.style.lineWidth = step === 0 ? 1 : 1.5 * step;
                    edgeModel.style.stroke = '#C2C8D5';
                    edgeModel.label = formatKMBT(edgeModel.receiveVolume);
                    edgeModel.opposite && (edgeModel.labelCfg.refY = -10);
                    SGraph.updateItem(edge, edgeModel);
                });
            }
        }
        nodes.forEach((node: any) => {
            let nodeModel = node.getModel();
            nodeModel.status = 'green';
            SGraph.updateItem(node, nodeModel);
        });
    } else {
        edges.forEach((edge: any) => {
            let edgeModel = edge.getModel();
            edgeModel.style.stroke = '#C2C8D5';
            edgeModel.style.lineWidth = 1;
            edgeModel.label = formatCount(edgeModel[metric]);
            SGraph.updateItem(edge, edgeModel);
        });
        nodes.forEach((node: any) => {
            let nodeModel = node.getModel();
            nodeModel.status = 'green';
            SGraph.updateItem(node, nodeModel);
        });
    }
}
