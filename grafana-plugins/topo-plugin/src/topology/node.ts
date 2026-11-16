// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import * as G6 from '@antv/g6';
import dnsPng from '../img/dns.png';
import externalPng from '../img/externel.png';
import namespacePng from '../img/namespace.png';
import namespaceYPng from '../img/namespace-yellow.png';
import namespaceRPng from '../img/namespace-red.png';
import workloadPng from '../img/workload.png';
import workloadYPng from '../img/workload-yellow.png';
import workloadRPng from '../img/workload-red.png';
import deploymentPng from '../img/deployment.png';
import deploymentYPng from '../img/deployment-yellow.png';
import deploymentRPng from '../img/deployment-red.png';
import daemonsetPng from '../img/daemonset.png';
import daemonsetYPng from '../img/daemonset-yellow.png';
import daemonsetRPng from '../img/daemonset-red.png';
import statefulsetPng from '../img/statefulset.png';
import statefulsetYPng from '../img/statefulset-yellow.png';
import statefulsetRPng from '../img/statefulset-red.png';
import podPng from '../img/pod.png';
import podYPng from '../img/pod-yellow.png';
import podRPng from '../img/pod-red.png';
import unkonwPng from '../img/unknown.png';

import { nodeTextHandle } from './topology';

const ImgW = 40, ImgH = 40;
const nodeImgHandle = (node: any) => {
    switch (node.nodeType) {
        case 'dns':
            return dnsPng;
        case 'external':
            return externalPng;
        case 'namespace':
            switch (node.status) {
                case 'green':
                    return namespacePng;
                case 'yellow':
                    return namespaceYPng;
                case 'red':
                    return namespaceRPng;
                default:
                    return namespacePng;
            }
        case 'workload':
        case 'node':
            switch (node.status) {
                case 'green':
                    return workloadPng;
                case 'yellow':
                    return workloadYPng;
                case 'red':
                    return workloadRPng;
                default:
                    return workloadPng;
            }
        case 'deployment':
            switch (node.status) {
                case 'green':
                    return deploymentPng;
                case 'yellow':
                    return deploymentYPng;
                case 'red':
                    return deploymentRPng;
                default:
                    return deploymentPng;
            }
        case 'daemonset':
            switch (node.status) {
                case 'green':
                    return daemonsetPng;
                case 'yellow':
                    return daemonsetYPng;
                case 'red':
                    return daemonsetRPng;
                default:
                    return daemonsetPng;
            }
        case 'statefulset':
            switch (node.status) {
                case 'green':
                    return statefulsetPng;
                case 'yellow':
                    return statefulsetYPng;
                case 'red':
                    return statefulsetRPng;
                default:
                    return statefulsetPng;
            }
        case 'pod':
            switch (node.status) {
                case 'green':
                    return podPng;
                case 'yellow':
                    return podYPng;
                case 'red':
                    return podRPng;
                default:
                    return podPng;
            }
        case 'unknow': 
            return unkonwPng;
        default:
            return unkonwPng;
    }
};

// register custom node
G6.registerNode('custom-node', {
    // getAnchorPoints() {
    //     return [
    //         [0.5, 0],
    //         [0, 0.5], // 左侧中间
    //         [1, 0.5], // 右侧中间
    //         [0.5, 1]
    //     ];
    // },
    draw: (node: any, group: any) => {

        let shape = group.addShape('image', {
            attrs: {
                x: -ImgW / 2,
                y: -ImgW / 2,
                width: ImgW,
                height: ImgH,
                cursor: 'pointer',
                img: nodeImgHandle(node)
            },
            name: 'node-image',
            id: 'node-image',
            draggable: true
        });
        group.addShape('text', {
            attrs: {
                x: 0,
                y: ImgH / 2 + 6,
                textAlign: 'center',
                textBaseline: 'top',
                class: 'node_text',
                fontSize: 10,
                lineHeight: 10,
                text: nodeTextHandle(node.name),
                fill: '#C2C8D5',
            },
            id: 'node-name',
            name: 'node-name',
        }); 
        if (node.showNamespace) {
            group.addShape('rect', {
                attrs: {
                    x: -ImgW * 2 / 2,
                    y: ImgH / 2 + 19,
                    radius: 5,
                    width: ImgW * 2,
                    height: 10,
                    class: 'node_ns_rect',
                    fill: '#F3FF69',
                    stroke: '#F3FF69'
                },
                id: 'node-ns-rect',
                name: 'node-ns-rect',
            });    
            group.addShape('text', {
                attrs: {
                    x: 0,
                    y: ImgH / 2 + 18,
                    textAlign: 'center',
                    textBaseline: 'top',
                    class: 'node_text',
                    fontSize: 10,
                    lineHeight: 10,
                    text: nodeTextHandle(`ns:${node.namespace}`, 15),
                    fill: '#595959',
                },
                id: 'node-ns-name',
                name: 'node-ns-name',
            });    
        }    
        return shape;
    },
    update: (cfg: any, node: any) => {
        
    },
    afterUpdate(cfg, item: any) {
        const group = item.get('group');
        const model: any = item.getModel();
        const imgShape = group.findById('node-image');
        imgShape.attr({
            img: nodeImgHandle(model)
        });
    }
}, 'single-node');

