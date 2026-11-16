// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import _ from 'lodash';
import { MetricType } from '../types';

export const metricList: Array<{label: string; value: MetricType; description?: string}> = [
    { label: 'Latency', value: 'latency' },
    { label: 'Calls', value: 'calls' },
    { label: 'Error Rate', value: 'errorRate' },
    { label: 'Sent Volume', value: 'sentVolume' },
    { label: 'Receive Volume', value: 'receiveVolume' },
    { label: 'SRTT', value: 'rtt' },
    { label: 'Retransmit', value: 'retransmit' },
    { label: 'Package Lost', value: 'packageLost' },
    { label: 'Connection Failure', value: 'connFail' }
];
export const metricDataName = {
    sentVolume: 'edgeSendVolumeData',
    receiveVolume: 'edgeReceiveVolumeData',
    retransmit: 'edgeRetransmitData',
    rtt: 'edgeRTTData',
    packageLost: 'edgePackageLostData'
}

export const layoutOptions = [
    { label: 'Hierarchy Chart', value: 'dagre' },
    { label: 'Mesh Chart', value: 'force' }
];
export const directionOptions = [
    { label: 'TB', value: 'TB' },
    { label: 'LR', value: 'LR' }
];
export const viewRadioOptions = [
    { label: 'Workload', value: 'workload_view' },
    { label: 'Pod', value: 'pod_view' }
];
export const showServiceOptions = [
    { label: 'ON', value: true },
    { label: 'OFF', value: false }
];


// externalTypes：The namespace enumeration value of the current external call
const externalTypes: string[] = ['NOT_FOUND_EXTERNAL', 'NOT_FOUND_INTERNAL', 'EXTERNAL', 'external'];
// workloadTypes
const workloadTypes: string[] = ['workload', 'deployment', 'daemonset', 'statefulset', 'node'];

type NodeField = 'calls' | 'latency' | 'errorRate' | 'sentVolume' | 'receiveVolume';
export interface TopologyProps {
    nodes: NodeProps[];
    edges: EdgeProps[];
}

export interface NodeProps {
    id: string;
    name: string;
    namespace?: string;
    nodeType: string;
    showNamespace: Boolean;
    calls?: number;
    latency?: number;
    errorRate?: number;
    sentVolume?: number;
    receiveVolume?: number;
    status?: string;
}
export interface EdgeProps {
    source: string;
    target: string;
    service?: string;
    opposite?: boolean;
    dnsEdge?: boolean;
    calls?: number;
    latency?: number;
    errorRate?: number;
    sentVolume?: number;
    receiveVolume?: number;
    rtt?: number;
    retransmit?: number;
    packageLost?: number;
    connFail?: number;
}
export interface NodeDataProps {
    nodeCallsData: any[];
    nodeTimeData: any[];
    nodeSendVolumeData: any[];
    nodeReceiveVolumeData: any[];
}
export interface EdgeDataProps {
    edgeCallData: any[];
    edgeTimeData: any[];
    edgeSendVolumeData?: any[];
    edgeReceiveVolumeData?: any[];
    edgeRetransmitData?: any[];
    edgeRTTData?: any[];
    edgePackageLostData?: any[];
}

export const transformWorkload = (workload: string) => {
    if (workload.indexOf(',') > -1) {
        let list = workload.match(/[^{},\s]+/g);
        return list!.toString();
    } else {
        return workload;
    }
}
/**
 * Grafana data format conversion to facilitate subsequent debugging
 * 高版本grafana返回数据时不在有buffer这个字段，直接取values
 * Grafana的数据格式转化提取，方便后续调试
 */
export const transformData = (data: any[]) => {
    let result: any[] = [];
    _.forEach(data, item => {
        let tdata: any = {
            ...item.fields[1].labels,
            values: item.fields[1].values.buffer ? item.fields[1].values.buffer : item.fields[1].values
        }
        result.push(tdata);
    });
    return result;
}

/**
 * Gets an array of node types in the current topology for legend drawing on the right
 * 获取当前拓扑图下节点的类型数组，用于右侧的legend绘制
 */
export const getNodeTypes = (nodes: any[]) => {
    let nodeByType = _.groupBy(nodes, 'nodeType');
    let types: string[] = _.keys(nodeByType);
    _.remove(types, opt => opt === 'undefined');
    return types;
}


const edgeFilter = (item: any, edge: EdgeProps) => {
    if (edge.dnsEdge) {
        return item.src_namespace === edge.source && item.dst_ip === edge.target;
    } else {
        return item.src_namespace === edge.source && item.dst_namespace === edge.target;
    }
};
const nsNodeInfoHandle = (node: NodeProps, nodeData: NodeDataProps) => {
    let callsList = _.filter(nodeData.nodeCallsData, item => item.namespace === node.id);
    let timeList = _.filter(nodeData.nodeTimeData, item => item.namespace === node.id);
    let sendVolumeList = _.filter(nodeData.nodeSendVolumeData, item => item.namespace === node.id);
    let receiveVolumeList = _.filter(nodeData.nodeReceiveVolumeData, item => item.namespace === node.id);
    let errorList = _.filter(callsList, item => {
        if (item.protocol === 'http') {
            return parseInt(item.response_content, 10) >= 400;
        } else if (item.protocol === 'dns') {
            return parseInt(item.response_content, 10) > 0;
        } else {
            return false
        }
    });

    let calls = callsList.length > 0 ? _.chain(callsList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    let timeValue = timeList.length > 0 ? _.chain(timeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    let latency = calls ? timeValue / calls / 1000000 : 0;
    let errorValue = errorList.length > 0 ? _.chain(errorList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    let errorRate = calls ? errorValue / calls * 100 : 0;
    let sentVolume = sendVolumeList.length > 0 ? _.chain(sendVolumeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    let receiveVolume = receiveVolumeList.length > 0 ? _.chain(receiveVolumeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    
    return { calls, latency, errorRate, sentVolume, receiveVolume }
}   
// namespace relational data handle
export const nsRelationHandle = (topoData: any, nodeData: NodeDataProps, edgeData: EdgeDataProps) => {
    let nodes: NodeProps[] = [], edges: EdgeProps[] = [];
    _.forEach(topoData, tdata => {
        let target: string;
        if (tdata.src_namespace && tdata.dst_namespace) {
            if (tdata.protocol === 'dns') {
                target = tdata.dst_ip;
                if (_.findIndex(nodes, {id: tdata.dst_ip}) === -1) {
                    nodes.push({
                        id: tdata.dst_ip,
                        name: tdata.dst_ip,
                        nodeType: 'dns',
                        showNamespace: false
                    });
                }
            } else {
                target = tdata.dst_namespace;
                if (_.findIndex(nodes, {id: tdata.dst_namespace}) === -1) {
                    nodes.push({
                        id: tdata.dst_namespace,
                        name: tdata.dst_namespace,
                        nodeType: externalTypes.indexOf(tdata.dst_namespace) > -1 ? 'external' : 'namespace',
                        showNamespace: false
                    });
                }
            }
            if (_.findIndex(nodes, {id: tdata.src_namespace}) === -1) {
                nodes.push({
                    id: tdata.src_namespace,
                    name: tdata.src_namespace,
                    nodeType: externalTypes.indexOf(tdata.src_namespace) > -1 ? 'external' : 'namespace',
                    showNamespace: false
                });
            }
            if (_.findIndex(edges, {source: tdata.src_namespace, target: target}) === -1) {
                let opposite: boolean = _.findIndex(edges, {source: target, target: tdata.src_namespace}) > -1 ? true : false;
                edges.push({
                    source: tdata.src_namespace,
                    target: target,
                    opposite,
                    dnsEdge: tdata.protocol === 'dns'
                });
            }
        }
    });
    _.forEach(nodes, (node: NodeProps) => {
        node.status = 'green';
        if (externalTypes.indexOf(node.nodeType) === -1) {
            let info: Record<NodeField, any> = nsNodeInfoHandle(node, nodeData);
            (_.keys(info) as NodeField[]).forEach((field: NodeField) => {
                node[field] = info[field];
            });
        }
    });

    _.remove(edges, edge => edge.source === edge.target);
    edges.forEach((edge: EdgeProps) => {
        let callsList = _.filter(topoData, item => edgeFilter(item, edge));
        let errorList = _.filter(callsList, item => {
            if (item.protocol === 'http') {
                return parseInt(item.status_code, 10) >= 400;
            } else if (item.protocol === 'dns') {
                return parseInt(item.status_code, 10) > 0;
            } else {
                return false
            }
        });
        let timeList = _.filter(edgeData.edgeTimeData, item => edgeFilter(item, edge));
        let sendVolumeList = _.filter(edgeData.edgeSendVolumeData, item => edgeFilter(item, edge));
        let receiveVolumeList = _.filter(edgeData.edgeReceiveVolumeData, item => edgeFilter(item, edge));
        let retransmitList = _.filter(edgeData.edgeRetransmitData, item => edgeFilter(item, edge));
        let rttList = _.filter(edgeData.edgeRTTData, item => edgeFilter(item, edge));
        let packageLostList = _.filter(edgeData.edgePackageLostData, item => edgeFilter(item, edge));
        
        edge.calls = callsList.length > 0 ? _.chain(callsList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        let timeValue = timeList.length > 0 ? _.chain(timeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.latency = edge.calls ? timeValue / edge.calls / 1000000 : 0;
        let errorValue = errorList.length > 0 ? _.chain(errorList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.errorRate = edge.calls ? errorValue / edge.calls * 100 : 0;
        edge.sentVolume = sendVolumeList.length > 0 ? _.chain(sendVolumeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.receiveVolume = receiveVolumeList.length > 0 ? _.chain(receiveVolumeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.rtt = rttList.length > 0 ? _.chain(rttList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() / 1000 : 0;
        edge.retransmit = retransmitList.length > 0 ? _.chain(retransmitList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.packageLost = packageLostList.length > 0 ? _.chain(packageLostList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    });
    return { nodes, edges };
}
export const connFailNSRelationHandle = (topoData: any) => {
    let nodes: NodeProps[] = [], edges: EdgeProps[] = [];
    _.forEach(topoData, tdata => {
        let target: string;
        if (tdata.src_namespace && tdata.dst_namespace) {
            if (tdata.protocol === 'dns') {
                target = tdata.dst_ip;
                if (_.findIndex(nodes, {id: tdata.dst_ip}) === -1) {
                    nodes.push({
                        id: tdata.dst_ip,
                        name: tdata.dst_ip,
                        nodeType: 'dns',
                        showNamespace: false
                    });
                }
            } else {
                target = tdata.dst_namespace;
                if (_.findIndex(nodes, {id: tdata.dst_namespace}) === -1) {
                    nodes.push({
                        id: tdata.dst_namespace,
                        name: tdata.dst_namespace,
                        nodeType: externalTypes.indexOf(tdata.dst_namespace) > -1 ? 'external' : 'namespace',
                        showNamespace: false
                    });
                }
            }
            if (_.findIndex(nodes, {id: tdata.src_namespace}) === -1) {
                nodes.push({
                    id: tdata.src_namespace,
                    name: tdata.src_namespace,
                    nodeType: externalTypes.indexOf(tdata.src_namespace) > -1 ? 'external' : 'namespace',
                    showNamespace: false
                });
            }
            if (_.findIndex(edges, {source: tdata.src_namespace, target: target}) === -1) {
                let opposite: boolean = _.findIndex(edges, {source: target, target: tdata.src_namespace}) > -1 ? true : false;
                edges.push({
                    source: tdata.src_namespace,
                    target: target,
                    opposite,
                    dnsEdge: tdata.protocol === 'dns'
                });
            }
        }
    });
    _.remove(edges, edge => edge.source === edge.target);
    edges.forEach((edge: EdgeProps) => {
        let connectData = _.filter(topoData, item => edgeFilter(item, edge));
        let connectFailData = _.filter(connectData, item => item.success === 'false');
        edge.connFail = connectFailData.length > 0 ? _.chain(connectFailData).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() / _.chain(connectData).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() * 100 : 0;
    });
    return { nodes, edges };
}
export const connFailWorkloadRelationHandle = (workload: string, namespace: string, topoData: any, showPod: boolean, serviceLine: boolean) => {
    let nodes: any[] = [], edges: any[] = [];
    let result: any[] = [];
    if (workload.indexOf(',') > -1) {
        // 当workload为all的时候，筛选对应namespace下所有workload的调用关系
        result = _.filter(topoData, (item: any) => item.dst_namespace === namespace || item.src_namespace === namespace);
    } else {
        // filter Topology data when namespace and workload is specific
        result = _.filter(topoData, (item: any) => (item.dst_namespace === namespace && item.dst_workload_name === workload) || (item.src_namespace === namespace && item.src_workload_name === workload));
    }
    // console.log('topology data', result);
    _.forEach(result, tdata => {
        let { node: targetNode, target, service } = detailRelationHandle(nodes, edges, namespace, tdata, 'dst', showPod, serviceLine);
        let { node: sourceNode, source } = detailRelationHandle(nodes, edges, namespace, tdata, 'src', showPod, serviceLine);
        sourceNode && nodes.push(sourceNode);
        targetNode && nodes.push(targetNode);
        let edgeId = `edge_${source}_${target}${service ? '_' + service : ''}`
        if (_.findIndex(edges, {id: edgeId}) === -1) {
            let opposite: boolean = _.findIndex(edges, {source: target, target: source}) > -1 ? true : false;
            edges.push({
                id: edgeId,
                source: source,
                target: target,
                service: service || '',
                opposite
            });
        }
    });
    edges = connFailDetailEdgesHandle(nodes, edges, topoData, serviceLine);
    return { nodes, edges };
}

export const topoMerge = (topo: TopologyProps, connTopo: TopologyProps) => {
    let result = _.cloneDeep(topo);
    let mergeNodes: NodeProps[] = [], mergeEdges: EdgeProps[] = [];
    _.forEach(connTopo.edges, edge => {
        if (_.findIndex(result.edges, {source: edge.source, target: edge.target}) === -1) {
            if (_.findIndex(mergeEdges, {source: edge.source, target: edge.target}) === -1) {
                mergeEdges.push(edge);
            }
            if (_.findIndex(result.nodes, {id: edge.source}) === -1) {
                let node = _.find(connTopo.nodes, {id: edge.source}) as NodeProps;
                mergeNodes.push(node);
            }
            if (_.findIndex(result.nodes, {id: edge.target}) === -1) {
                let node = _.find(connTopo.nodes, {id: edge.target}) as NodeProps;
                mergeNodes.push(node);
            }
        } else {
            let edgeItem: EdgeProps = _.find(result.edges, {source: edge.source, target: edge.target}) as EdgeProps;
            edgeItem.connFail = edge.connFail;
        }
    });
    result.nodes = result.nodes.concat(mergeNodes);
    result.edges = result.edges.concat(mergeEdges);
    console.log(result);
    return result;
}

/**
 * value data handle when select only one namespace that workload is all or workload is a single 
 * 只勾选一个namespace是workload为all或者workload为单个值的调用关系处理
 */
export const workloadRelationHandle = (workload: string, namespace: string, topoData: any, nodeData: NodeDataProps, edgeData: EdgeDataProps, showPod: boolean, serviceLine: boolean) => {
    let nodes: any[] = [], edges: any[] = [];
    let result: any[] = topoData;
    if (workload.indexOf(',') > -1) {
        // 当workload为all的时候，筛选对应namespace下所有workload的调用关系
        result = _.filter(topoData, (item: any) => item.dst_namespace === namespace || item.src_namespace === namespace);
    } else {
        // filter Topology data when namespace and workload is specific
        result = _.filter(topoData, (item: any) => (item.dst_namespace === namespace && item.dst_workload_name === workload) || (item.src_namespace === namespace && item.src_workload_name === workload));
    }
    // console.log('topology data', result);
    _.forEach(result, tdata => {
        let { node: targetNode, target, service } = detailRelationHandle(nodes, edges, namespace, tdata, 'dst', showPod, serviceLine);
        let { node: sourceNode, source } = detailRelationHandle(nodes, edges, namespace, tdata, 'src', showPod, serviceLine);
        (sourceNode && _.findIndex(nodes, {id: sourceNode.id}) === -1) && nodes.push(sourceNode);
        (targetNode && _.findIndex(nodes, {id: targetNode.id}) === -1) && nodes.push(targetNode);
        //  TODO 先去掉节点自调用的数据
        if (source !== target) {
            let edgeId = `edge_${source}_${target}${service ? '_' + service : ''}`
            if (_.findIndex(edges, {id: edgeId}) === -1) {
                let opposite: boolean = _.findIndex(edges, {source: target, target: source}) > -1 ? true : false;
                    edges.push({
                        id: edgeId,
                        source: source,
                        target: target,
                        service: service || '',
                        opposite
                    });
            }
        }
    });
    nodes = detailNodesHandle(nodes, nodeData);
    edges = detailEdgesHandle(nodes, edges, edgeData, serviceLine);
    return { nodes, edges };
}
/**
 * Construct the node data and edge data in the Workload view
 * 构造workload视图下的对应节点数据和对应调用source、target
 * @param nodes node list | 当前节点数组
 * @param namespace current namespace | 当前查询的namespace
 * @param tdata current call data | 当前遍历的调用数据
 * @param pre dst|src   pre-type field | 判断当前节点使用字段的前置类型
 * @param showPod pod view | 是否为单个workload下显示pod视图
 * @returns node：当前构造的节点数据、source：tdata的调用方、target：tdata的被调用方
 */
export const detailRelationHandle = (nodes: any[], edges: any[], namespace: string, tdata: any, pre: string, showPod: boolean, showService: boolean) => {
    let source, target, service;
    let node: any = undefined;
    if (externalTypes.indexOf(tdata[`${pre}_namespace`]) > -1) {
        if (_.findIndex(nodes, node => node.namespace === tdata[`${pre}_namespace`] && node.id.indexOf(tdata[`${pre}_ip`]) > -1) > -1) {
            let tnode = _.find(nodes, node => node.namespace === tdata[`${pre}_namespace`] && node.id.indexOf(tdata[`${pre}_ip`]) > -1);
            let prots = tnode.id.indexOf(':') > -1 ? tnode.id.substring(tnode.id.indexOf(':') + 1).split('%') : [];
            if (tdata[`${pre}_port`] && tdata[`${pre}_port`] > 0 && prots.indexOf(tdata[`${pre}_port`]) === -1) {
                let needRenameEdges = _.filter(edges, edge => edge.source === tnode.id || edge.target === tnode.id);
                needRenameEdges.forEach(edge => {
                    if (edge.source === tnode.id) {
                        edge.source = `${tnode.id}${prots.length > 0 ? '%' : ':'}${tdata[`${pre}_port`]}`;
                    } else {
                        edge.target = `${tnode.id}${prots.length > 0 ? '%' : ':'}${tdata[`${pre}_port`]}`;
                    }   
                    edge.id = `edge_${edge.source}_${edge.target}${edge.service ? '_' + edge.service : ''}`
                });
                tnode.id = `${tnode.id}${prots.length > 0 ? '%' : ':'}${tdata[`${pre}_port`]}`;
                tnode.name = `${tnode.name}${prots.length > 0 ? '%' : ':'}${tdata[`${pre}_port`]}`;
            }
            if (pre === 'dst') {
                target = tnode.id;
            } else {
                source = tnode.id;
            }
        } else {
            let IPPort = `${tdata[`${pre}_ip`]}` + (tdata[`${pre}_port`] && tdata[`${pre}_port`] > 0 ? `:${tdata[`${pre}_port`]}` : ''); //:${tdata[`${pre}_port`]}
            let nodeId = `${tdata[`${pre}_namespace`]}_${IPPort}`;
            node = {
                id: nodeId,
                name: IPPort,
                namespace: tdata[`${pre}_namespace`],
                nodeType: 'external',
                showNamespace: false
            };
            if (pre === 'dst') {
                target = nodeId;
            } else {
                source = nodeId;
            }
        }
    } else {
        if (pre === 'dst') {
            target = `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_workload_name`]}`;
            if (showService) {
                service = tdata[`${pre}_service`];
            }
        } else {
            source = `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_workload_name`]}`;
        }
        if (showPod) {
            if (tdata[`${pre}_pod`]) {
                if (pre === 'dst') {
                    target = `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_pod`]}`;
                } else {
                    source = `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_pod`]}`;
                }
                if (_.findIndex(nodes, {id: `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_pod`]}`}) === -1) {
                    node = {
                        id: `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_pod`]}`,
                        name: tdata[`${pre}_pod`],
                        namespace: tdata[`${pre}_namespace`],
                        nodeType: 'pod',
                        showNamespace: tdata[`${pre}_namespace`] !== namespace
                    };
                }
            } else {
                if (_.findIndex(nodes, {id: `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_workload_name`]}`}) === -1) {
                    node = {
                        id: `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_workload_name`]}`,
                        name: tdata[`${pre}_workload_name`],
                        namespace: tdata[`${pre}_namespace`],
                        nodeType: tdata[`${pre}_workload_kind`] || 'unknow',
                        showNamespace: tdata[`${pre}_namespace`] !== namespace
                    };
                }
            }
        } else {
            if (_.findIndex(nodes, {id: `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_workload_name`]}`}) === -1) {
                node = {
                    id: `${tdata[`${pre}_namespace`]}_${tdata[`${pre}_workload_name`]}`,
                    name: tdata[`${pre}_workload_name`],
                    namespace: tdata[`${pre}_namespace`],
                    nodeType: tdata[`${pre}_workload_kind`],
                    showNamespace: tdata[`${pre}_namespace`] !== namespace
                };
            }
        }
    }
    return { node, source, target, service }
}

const detailNodeInfoHandle = (node: NodeProps, nodeData: NodeDataProps) => {
    let callsList = [], timeList = [], sendVolumeList = [], receiveVolumeList = []; 
    if (workloadTypes.indexOf(node.nodeType) > -1) {
        callsList = _.filter(nodeData.nodeCallsData, item => item.namespace === node.namespace && node.name === item.workload_name);
        timeList = _.filter(nodeData.nodeTimeData, item => item.namespace === node.namespace && node.name === item.workload_name);
        sendVolumeList = _.filter(nodeData.nodeSendVolumeData, item => item.namespace === node.namespace && node.name === item.workload_name);
        receiveVolumeList = _.filter(nodeData.nodeReceiveVolumeData, item => item.namespace === node.namespace && node.name === item.workload_name);
    } else if (node.nodeType === 'pod') {
        callsList = _.filter(nodeData.nodeCallsData, item => item.namespace === node.namespace && node.name === item.pod);
        timeList = _.filter(nodeData.nodeTimeData, item => item.namespace === node.namespace && node.name === item.pod);
        sendVolumeList = _.filter(nodeData.nodeSendVolumeData, item => item.namespace === node.namespace && node.name === item.pod);
        receiveVolumeList = _.filter(nodeData.nodeReceiveVolumeData, item => item.namespace === node.namespace && node.name === item.pod);
    } else {
        callsList = _.filter(nodeData.nodeCallsData, item => item.namespace === node.namespace && node.name === item.workload_name && !item.pod);
        timeList = _.filter(nodeData.nodeTimeData, item => item.namespace === node.namespace && node.name === item.workload_name && !item.pod);
        sendVolumeList = _.filter(nodeData.nodeSendVolumeData, item => item.namespace === node.namespace && node.name === item.workload_name && !item.pod);
        receiveVolumeList = _.filter(nodeData.nodeReceiveVolumeData, item => item.namespace === node.namespace && node.name === item.workload_name && !item.pod);
    }
    let errorList = _.filter(callsList, item => {
        if (item.protocol === 'http') {
            return parseInt(item.response_content, 10) >= 400;
        } else if (item.protocol === 'dns') {
            return parseInt(item.response_content, 10) > 0;
        } else {
            return false
        }
    });

    let calls = callsList.length > 0 ? _.chain(callsList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    let timeValue = timeList.length > 0 ? _.chain(timeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    let latency = calls ? timeValue / calls / 1000000 : 0;
    let errorValue = errorList.length > 0 ? _.chain(errorList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    let errorRate = calls ? errorValue / calls * 100 : 0;
    let sentVolume = sendVolumeList.length > 0 ? _.chain(sendVolumeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    let receiveVolume = receiveVolumeList.length > 0 ? _.chain(receiveVolumeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
    
    return { calls, latency, errorRate, sentVolume, receiveVolume }
}
/**
 * handle node indicator data
 * node节点数据处理（时延、调用次数、错误率、流量）
 * @param nodes nodes list | 节点数组
 * @param nodeData Indicator data of node | 节点的指标数据（调用次数、延时、错误率、进出流量）
 */
export const detailNodesHandle = (nodes: any[], nodeData: any) => {
    let nodelist = _.cloneDeep(nodes);
    nodelist.forEach(node => {
        node.status = 'green';
        if (externalTypes.indexOf(node.nodeType) === -1) {
            let info: Record<NodeField, any> = detailNodeInfoHandle(node, nodeData);
            (_.keys(info) as NodeField[]).forEach((field: NodeField) => {
                node[field] = info[field];
            });
        }
    });
    return nodelist;
}

export const detailEdgesHandle = (nodes: any[], edges: any[], edgeData: any, showService: boolean) => {
    let edgelist = _.cloneDeep(edges);
    edgelist.forEach((edge: EdgeProps) => {
        let sourceNode = _.find(nodes, {id: edge.source});
        let targteNode = _.find(nodes, {id: edge.target});

        let callsList = [], timeList = [], sendVolumeList = [], receiveVolumeList = [], retransmitList = [], rttList = [], packageLostList = [];
        if (externalTypes.indexOf(sourceNode.nodeType) > -1) {
            let ip = sourceNode.id.substring(sourceNode.id.lastIndexOf('_') + 1, sourceNode.id.indexOf(':') > -1 ? sourceNode.id.indexOf(':') : sourceNode.id.length);
            callsList = _.filter(edgeData.edgeCallData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
            timeList = _.filter(edgeData.edgeTimeData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
            sendVolumeList = _.filter(edgeData.edgeSendVolumeData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
            receiveVolumeList = _.filter(edgeData.edgeReceiveVolumeData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
            retransmitList = _.filter(edgeData.edgeRetransmitData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
            rttList = _.filter(edgeData.edgeRTTData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
            packageLostList = _.filter(edgeData.edgePackageLostData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
        } else if (sourceNode.nodeType === 'pod') {
            callsList = _.filter(edgeData.edgeCallData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
            timeList = _.filter(edgeData.edgeTimeData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
            sendVolumeList = _.filter(edgeData.edgeSendVolumeData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
            receiveVolumeList = _.filter(edgeData.edgeReceiveVolumeData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
            retransmitList = _.filter(edgeData.edgeRetransmitData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
            rttList = _.filter(edgeData.edgeRTTData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
            packageLostList = _.filter(edgeData.edgePackageLostData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
        } else if (sourceNode.nodeType === 'unknow') {
            callsList = _.filter(edgeData.edgeCallData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
            timeList = _.filter(edgeData.edgeTimeData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
            sendVolumeList = _.filter(edgeData.edgeSendVolumeData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
            receiveVolumeList = _.filter(edgeData.edgeReceiveVolumeData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
            retransmitList = _.filter(edgeData.edgeRetransmitData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
            rttList = _.filter(edgeData.edgeRTTData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
            packageLostList = _.filter(edgeData.edgePackageLostData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
        } else if (workloadTypes.indexOf(sourceNode.nodeType) > -1) {
            callsList = _.filter(edgeData.edgeCallData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
            timeList = _.filter(edgeData.edgeTimeData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
            sendVolumeList = _.filter(edgeData.edgeSendVolumeData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
            receiveVolumeList = _.filter(edgeData.edgeReceiveVolumeData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
            retransmitList = _.filter(edgeData.edgeRetransmitData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
            rttList = _.filter(edgeData.edgeRTTData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
            packageLostList = _.filter(edgeData.edgePackageLostData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
        }

        if (externalTypes.indexOf(targteNode.nodeType) > -1) {
            let ip = targteNode.id.substring(targteNode.id.lastIndexOf('_') + 1, targteNode.id.indexOf(':') > -1 ? targteNode.id.indexOf(':') : targteNode.id.length);
            callsList = _.filter(callsList, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
            timeList = _.filter(timeList, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
            sendVolumeList = _.filter(sendVolumeList, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
            receiveVolumeList = _.filter(receiveVolumeList, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
            retransmitList = _.filter(retransmitList, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
            rttList = _.filter(rttList, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
            packageLostList = _.filter(packageLostList, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
        } else if (targteNode.nodeType === 'pod') {
            callsList = _.filter(callsList, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
            timeList = _.filter(timeList, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
            sendVolumeList = _.filter(sendVolumeList, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
            receiveVolumeList = _.filter(receiveVolumeList, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
            retransmitList = _.filter(retransmitList, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
            rttList = _.filter(rttList, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
            packageLostList = _.filter(packageLostList, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
        } else if (targteNode.nodeType === 'unknow') {
            callsList = _.filter(callsList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
            timeList = _.filter(timeList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
            sendVolumeList = _.filter(sendVolumeList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
            receiveVolumeList = _.filter(receiveVolumeList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
            retransmitList = _.filter(retransmitList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
            rttList = _.filter(rttList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
            packageLostList = _.filter(packageLostList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
        } else if (workloadTypes.indexOf(targteNode.nodeType) > -1)  {
            callsList = _.filter(callsList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
            timeList = _.filter(timeList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
            sendVolumeList = _.filter(sendVolumeList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
            receiveVolumeList = _.filter(receiveVolumeList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
            retransmitList = _.filter(retransmitList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
            rttList = _.filter(rttList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
            packageLostList = _.filter(packageLostList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
        }

        if (showService && edge.service) {
            callsList = _.filter(callsList, item => item.dst_service === edge.service);
            timeList = _.filter(timeList, item => item.dst_service === edge.service);
            sendVolumeList = _.filter(sendVolumeList, item => item.dst_service === edge.service);
            receiveVolumeList = _.filter(receiveVolumeList, item => item.dst_service === edge.service);
            retransmitList = _.filter(retransmitList, item => item.dst_service === edge.service);
            rttList = _.filter(rttList, item => item.dst_service === edge.service);
            packageLostList = _.filter(packageLostList, item => item.dst_service === edge.service);
        }
        let errorList = _.filter(callsList, item => {
            if (item.protocol === 'http') {
                return parseInt(item.status_code, 10) >= 400;
            } else if (item.protocol === 'dns') {
                return parseInt(item.status_code, 10) > 0;
            } else {
                return false
            }
        });
        
        edge.calls = callsList.length > 0 ? _.chain(callsList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        let timeValue = timeList.length > 0 ? _.chain(timeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.latency = edge.calls ? timeValue / edge.calls / 1000000 : 0;
        let errorValue = errorList.length > 0 ? _.chain(errorList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.errorRate = edge.calls ? errorValue / edge.calls * 100 : 0;
        edge.sentVolume = sendVolumeList.length > 0 ? _.chain(sendVolumeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.receiveVolume = receiveVolumeList.length > 0 ? _.chain(receiveVolumeList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.rtt = rttList.length > 0 ? _.chain(rttList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() / 1000 : 0;
        edge.retransmit = retransmitList.length > 0 ? _.chain(retransmitList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        edge.packageLost = packageLostList.length > 0 ? _.chain(packageLostList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
        // console.log(edge);
    });
    return edgelist;
}

export const connFailDetailEdgesHandle = (nodes: any[], edges: any[], edgeData: any, showService: boolean) => {
    let edgelist = _.cloneDeep(edges);
    edgelist.forEach((edge: EdgeProps) => {
        let sourceNode = _.find(nodes, {id: edge.source});
        let targteNode = _.find(nodes, {id: edge.target});

        let connectData = [];
        if (externalTypes.indexOf(sourceNode.nodeType) > -1) {
            let ip = sourceNode.id.substring(sourceNode.id.lastIndexOf('_') + 1, sourceNode.id.indexOf(':') > -1 ? sourceNode.id.indexOf(':') : sourceNode.id.length);
            connectData = _.filter(edgeData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
        } else if (sourceNode.nodeType === 'pod') {
            connectData = _.filter(edgeData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
        } else if (sourceNode.nodeType === 'unknow') {
            connectData = _.filter(edgeData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
        } else if (workloadTypes.indexOf(sourceNode.nodeType) > -1) {
            connectData = _.filter(edgeData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
        }

        if (externalTypes.indexOf(targteNode.nodeType) > -1) {
            let ip = targteNode.id.substring(targteNode.id.lastIndexOf('_') + 1, targteNode.id.indexOf(':') > -1 ? targteNode.id.indexOf(':') : targteNode.id.length);
            connectData = _.filter(connectData, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
        } else if (targteNode.nodeType === 'pod') {
            connectData = _.filter(connectData, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
        } else if (targteNode.nodeType === 'unknow') {
            connectData = _.filter(connectData, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
        } else if (workloadTypes.indexOf(targteNode.nodeType) > -1)  {
            connectData = _.filter(connectData, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
        }

        if (showService && edge.service) {
            connectData = _.filter(connectData, item => item.dst_service === edge.service);
        }
        let connectFailData = _.filter(connectData, item => item.success === 'false');
        edge.connFail = connectFailData.length > 0 ? _.chain(connectFailData).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() / _.chain(connectData).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() * 100 : 0;
    });
    return edgelist;
}

export const updateNode = (graphData: TopologyProps, nodeData: NodeDataProps, SGraph: any, namespaceAll: boolean) => {
    if (namespaceAll) {
        SGraph.getNodes().forEach((dNode: any) => {
            let node = dNode.getModel();
            let info = nsNodeInfoHandle(node, nodeData);
            node = {...node, ...info};
            dNode.update(node);
        });
    } else {
        SGraph.getNodes().forEach((dNode: any) => {
            let node = dNode.getModel();
            let info = detailNodeInfoHandle(node, nodeData);
            node = {...node, ...info};
            dNode.update(node);
        });
    }
}

export const updateEdge = (graphData: TopologyProps, metric: MetricType, metricData: any, SGraph: any, namespaceAll: boolean, showService: boolean) => {
    let {nodes, edges} = graphData;
    if (namespaceAll) {
        SGraph.getEdges().forEach((dEdge: any) => {
            let edgeModel = dEdge.getModel();
            let metricList = _.filter(metricData, item => edgeFilter(item, edgeModel));
            let metricValue = metricList.length > 0 ? _.chain(metricList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
            if (metric === 'rtt') {
                metricValue = metricValue / 1000;
            }
            edgeModel[metric] = metricValue;
            edges.splice(_.findIndex(edges, {source: edgeModel.source, target: edgeModel.target}), 1, edgeModel);
            dEdge.update(edgeModel);
        });
    } else {
        SGraph.getEdges().forEach((dEdge: any) => {
            let edge = dEdge.getModel();
            let sourceNode = _.find(nodes, {id: edge.source})!;
            let targteNode = _.find(nodes, {id: edge.target})!;

            let metricList = [];

            if (externalTypes.indexOf(sourceNode.nodeType) > -1) {
                let ip = sourceNode.id.substring(sourceNode.id.lastIndexOf('_') + 1, sourceNode.id.indexOf(':') > -1 ? sourceNode.id.indexOf(':') : sourceNode.id.length);
                metricList = _.filter(metricData, item => item.src_namespace === sourceNode.namespace && item.src_ip === ip);
            } else if (sourceNode.nodeType === 'pod') {
                metricList = _.filter(metricData, item => item.src_namespace === sourceNode.namespace && item.src_pod === sourceNode.name);
            } else if (sourceNode.nodeType === 'unknow') {
                metricList = _.filter(metricData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name && !item.src_pod);
            } else if (workloadTypes.indexOf(sourceNode.nodeType) > -1) {
                metricList = _.filter(metricData, item => item.src_namespace === sourceNode.namespace && item.src_workload_name === sourceNode.name);
            }

            if (externalTypes.indexOf(targteNode.nodeType) > -1) {
                let ip = targteNode.id.substring(targteNode.id.lastIndexOf('_') + 1, targteNode.id.indexOf(':') > -1 ? targteNode.id.indexOf(':') : targteNode.id.length);
                metricList = _.filter(metricList, item => item.dst_namespace === targteNode.namespace && item.dst_ip === ip);
            } else if (targteNode.nodeType === 'pod') {
                metricList = _.filter(metricList, item => item.dst_namespace === targteNode.namespace && item.dst_pod === targteNode.name);
            } else if (targteNode.nodeType === 'unknow') {
                metricList = _.filter(metricList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name && !item.dst_pod);
            } else if (workloadTypes.indexOf(targteNode.nodeType) > -1)  {
                metricList = _.filter(metricList, item => item.dst_namespace === targteNode.namespace && item.dst_workload_name === targteNode.name);
            }

            if (showService && edge.service) {
                metricList = _.filter(metricList, item => item.dst_service === edge.service);
            }

        
            let metricValue = metricList.length > 0 ? _.chain(metricList).map(item => _.isNumber(_.last(item.values)) ? _.last(item.values) : 0).sum().value() : 0;
            if (metric === 'rtt') {
                metricValue = metricValue / 1000;
            }
            edge[metric] = metricValue;
            edges.splice(_.findIndex(edges, {source: edge.source, target: edge.target}), 1, edge);
            dEdge.update(edge);
        });
    }
}
