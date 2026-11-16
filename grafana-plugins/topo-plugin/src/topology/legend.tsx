// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import { css } from 'emotion';
import { stylesFactory } from '@grafana/ui';
import { formatKMBT } from './tooltip';
import dnsPng from '../img/dns.png';
import externalPng from '../img/externel.png';
import namespacePng from '../img/namespace.png';
import workloadPng from '../img/workload.png';
import deploymentPng from '../img/deployment.png';
import daemonsetPng from '../img/daemonset.png';
import statefulsetPng from '../img/statefulset.png';
import podPng from '../img/pod.png';
import unkonwPng from '../img/unknown.png';

const imgs: any = {
    dns: dnsPng,
    external: externalPng,
    namespace: namespacePng,
    workload: workloadPng,
    node: workloadPng,
    deployment: deploymentPng,
    daemonset: daemonsetPng,
    statefulset: statefulsetPng,
    pod: podPng,
    unknow: unkonwPng
}

interface CProps {
    color?: string;
    title: string;
}
function ColorLegend(props: CProps) {
    const { color, title } = props;
    const styles = getStyles();
    return (<div className={styles.color_item}>
        {
            color ? <span className={styles.color_pointer} style={{ backgroundColor: color }}></span> : null
        }
        <span>{title}</span>
    </div>)
}
interface LProps {
    typeList: string[];
    metric: string;
    volumes: any;
    options: any;
}
function TopoLegend(props: LProps) {
    const { typeList, metric, volumes, options } = props;
    const styles = getStyles();

    const metricRender = () => {
        let domNode: any = null;
        if (metric === 'latency') {
            domNode = (
                <div>
                    <ColorLegend color="#C2C8D5" title={`Normal(<${options.normalLatency}ms)`}/>
                    <ColorLegend color="#f3ff69" title={`Warning(${options.normalLatency}ms~${options.abnormalLatency}ms)`}/>
                    <ColorLegend color="#ff4c4c" title={`AbNormal(>${options.abnormalLatency}ms)`}/>
                </div>
            ); 
        } else if (metric === 'rtt') {
            domNode = (<div>
                <ColorLegend color="#C2C8D5" title={`Normal(<${options.normalRtt}ms)`}/>
                <ColorLegend color="#f3ff69" title={`Warning(${options.normalRtt}ms~${options.abnormalRtt}ms)`}/>
                <ColorLegend color="#ff4c4c" title={`AbNormal(>${options.abnormalRtt}ms)`}/>
            </div>);
        } else if (metric === 'errorRate') {
            domNode = (<div>
                <ColorLegend color="#C2C8D5" title="Normal(0%)"/>
                <ColorLegend color="#ff4c4c" title="AbNormal(>0%)"/>
            </div>);
        } else if (metric === 'sentVolume') {
            domNode = (<div>
                <ColorLegend title={`min Volume ${formatKMBT(volumes.minSentVolume)}`}/>
                <ColorLegend title={`max Volume ${formatKMBT(volumes.maxSentVolume)}`}/>
            </div>);
        } else if (metric === 'receiveVolume') {
            domNode = (<div>
                <ColorLegend title={`min Volume ${formatKMBT(volumes.minReceiveVolume)}`}/>
                <ColorLegend title={`max Volume ${formatKMBT(volumes.maxReceiveVolume)}`}/>
            </div>);
        }
        return <>
            {
                domNode !== null ? <span className={styles.status_label}>node and call line status</span> : null
            }
            {domNode}
        </>;
    }
    return (
        <div className={styles.legend_warp}>
            {
                typeList.map((type, idx) => <div key={idx} className={styles.legend_item}>
                    <img className={styles.legend_icon} alt="icon" src={imgs[type]}/>
                    <span>{type}</span>
                </div>)
            }
            {
                metricRender()
            }
        </div>
    )
}

const getStyles = stylesFactory(() => {
    return {
        legend_warp: css`
            z-index: 10;
            display: flex;
            flex-direction: column;
            width: 245px;
        `,
        legend_item: css`
            height: 28px;
            line-height: 28px;
        `,
        legend_icon: css`
            width: 18px;
            margin-right: 10px;
        `,
        status_label: css`
            margin: 8px 0;
        `,
        color_item: css`
            height: 28px;
            line-height: 28px;
        `,
        color_pointer: css`
            border-radius: 6px;
            width: 12px;
            height: 12px;
            display: inline-block;
            margin-right: 8px;
        `,
    };
});
  
export default TopoLegend;
