// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import ReactDOM from 'react-dom';
import G6 from '@antv/g6';
import { css } from 'emotion';
import { stylesFactory } from '@grafana/ui';

export const formatTime = (time: number) => {
    if (time === undefined) {
        return 0;
    } else {
        time = Math.abs(time);
        if (time > 3600000) {
            return (time / 3600000).toFixed(2) + 'hour';
        } else if (time >= 60000) {
            return (time / 60000).toFixed(2) + 'min';
        } else if (time >= 1000) {
            return (time / 1000).toFixed(2) + 's';
        } else if (time >= 10) {
            return time.toFixed(2) + 'ms';
        } else if (time > 0) {
            return time.toFixed(2) + 'ms';
        } else {
            return time.toFixed(2);
        }
    }
};
export const formatCount = (y: number) => {
    let yy: number = Math.abs(y);
    if (yy >= 1000000) {
        return (yy / 1000000).toFixed(2) + 'M';
    } else if (yy >= 1000) {
        return (yy / 1000).toFixed(2) + 'K';
    } else if (yy >= 1) {
        return yy.toFixed(2);
    } else if (yy > 0) {
        return yy.toFixed(4);
    } else {
        return 0;
    }
};
export const formatKMBT = (y: number) => {
    let yy: number = Math.abs(y);
    if (yy >= Math.pow(1024, 4)) {
        return (yy / Math.pow(1024, 4)).toFixed(2) + 'TiB';
    } else if (yy >= Math.pow(1024, 3)) {
        return (yy / Math.pow(1024, 3)).toFixed(2) + 'GiB';
    } else if (yy >= Math.pow(1024, 2)) {
        return (yy / Math.pow(1024, 2)).toFixed(2) + 'MiB';
    } else if (yy >= 1024) {
        return (yy / 1024).toFixed(2) + 'KiB';
    } else if (yy < 1024 && yy >= 1) {
        return yy.toFixed(2) + 'B';
    } else if (yy < 1 && yy > 0) {
        return yy.toFixed(2) + 'B';
    } else if (yy === 0) {
        return 0;
    } else {
        return yy;
    }
};
//小数转化百分比
export const formatPercent = (number: number, size = 2, hasSymbol = true) => {
    if (number !== 0 && !number) {
		return '-';
	}
    if (hasSymbol) {
        return `${Number(number).toFixed(size)}%`;
    }
    return `${Number(number).toFixed(size)}`;
};


export const nameTooltip = new G6.Tooltip({
    offestX: 10,
    offestY: 10,
    itemTypes: ['node'],
    getContent: (e: any) => {
        let node: any = e.item.getModel();
        return node.name
    },
    shouldBegin: (e: any) => {
        const target = e.target;
        const model = e.item.getModel();
        if (target.get('name') === 'node-name' && model) {
			return true;
		} else {
			return false;
		}
    }
});

export const nodeTooltip = new G6.Tooltip({
    offestX: 10,
    offestY: 10,
    trigger: 'click',
    fixToNode: [0.5, 0.5],
    itemTypes: ['node'],
    getContent: (e: any) => {
        // let type = e.item.getType();
        let model: any = e.item.getModel();
        const styles = getStyles();
        // console.log(type, model);
        const tooltipDom = (
            <div className={styles.tooltip_warp}>
                <div className={styles.tooltip_name}>{model.name}</div>
                <ul className={styles.tooltip_filed}>
                    <li>
                        <span className={styles.field_value}>Latency: {formatTime(model.latency || 0)}</span>
                    </li>
                    <li>
                        <span className={styles.field_value}>Calls: {formatCount(model.calls || 0)}</span>
                    </li>
                    <li>
                        <span className={styles.field_value}>Error Rate: {formatPercent(model.errorRate || 0)}</span>
                    </li>
                </ul>
                <ul className={styles.tooltip_filed}>
                    <li>
                        <span className={styles.field_value}>Sent Volume: {formatKMBT(model.sentVolume || 0)}</span>
                    </li>
                    <li>
                        <span className={styles.field_value}>Receive Volume: {formatKMBT(model.receiveVolume || 0)}</span>
                    </li>
                </ul>
            </div>
        );
        let elem = document.createElement("div");
        elem.setAttribute('id', 'nodeTooltip');
        ReactDOM.render(tooltipDom, elem);
        return ['external', 'dns', 'default', 'unknow'].indexOf(model.nodeType) > -1 ? model.name : elem;
    }
});

// export const edgeTooltip = new G6.Tooltip({
//     offestX: 10,
//     offestY: 10,
//     trigger: 'click',
//     itemTypes: ['edge'],
//     getContent: (e: any) => {
//         let item: any = e.item;
//         let model: any = item.getModel();
//         let sourceNode = item.getSource().getModel();
//         let targetNode = item.getTarget().getModel();

//         // console.log(item, model, sourceNode, targetNode);
//         const tooltipDom = (
//             <div className="tooltip_warp">
//                 <div className="tooltip_edge_top">
//                     <span className="source">{sourceNode.name}</span>
//                     <span className="triangle"></span>
//                     <span className="target">{targetNode.name || model.target}</span>
//                     {/* <span>{sourceNode.name} - {targetNode.name}</span> */}
//                 </div>
//                 <ul className="tooltip_filed">
//                     <li>
//                         <div className="field_title">响应时间</div>
//                         <div className="field_value">{model.responseTimeAvg !== null ? formatTime(model.responseTimeAvg) : '--'}</div>
//                     </li>
//                     <li>
//                         <div className="field_title">错误率</div>
//                         <div className="field_value">{model.errorRate !== null ? `${Number(model.errorRate).toFixed(2)}%` : '--'}</div>
//                     </li>
//                     <li>
//                         <div className="field_title">请求量</div>
//                         <div className="field_value">{model.requestTotal !== null ? formatCount(parseFloat(model.requestTotal || 0)) : '--'}</div>
//                     </li>
//                 </ul>
//             </div>
//         );
//         let elem = document.createElement("div");
//         elem.setAttribute('id', 'edgeTooltip');
//         ReactDOM.render(tooltipDom, elem);
//         return elem;
//     },
//     shouldBegin: (e: any) => {
//         let item: any = e.item;
//         let sourceNode = item.getSource().getModel();
//         let targetNode = item.getTarget().getModel();
//         // console.log(sourceNode, targetNode);
//         if (sourceNode.collapsed || targetNode.collapsed) return false;
//         return true;
//     }
// });

const getStyles = stylesFactory(() => {
    return {
        tooltip_warp: css`
            min-width: 350px;
        `,
        tooltip_name: css`
            font-size: 14px;
            font-weight: 600;
        `,
        tooltip_filed: css`
            display: flex;
            list-style: none;
            margin: 0;
            padding: 0;
            > li {
                flex: 1;
                padding-left: 4px;
            }
        `,
        field_value: css`
            font-size: 12px;
            line-height: 16px;
            color: rgba(0, 0, 0, 0.8)
        `,
    };
});
