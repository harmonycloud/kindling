// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import { getDataSourceSrv, getBackendSrv, toDataQueryResponse } from '@grafana/runtime';
import { TimeRange } from '@grafana/data';
import { MetricType } from './types';
import _ from 'lodash';

// // 获取对应数据库的UID用于后续api代理转发跟参数构造
// export const fetchDataSources = () => {
//     return getBackendSrv().get('/api/datasources')
// };
// // 通过DataSource proxy 代理请求数据库获取数据，需要拼接完整的promql跟对应Prometheus的api
// export const queryPrometheusData1 = async (prometheusId: string, timeRange: any) => {
//     try {
//         const promql2 = 'increase(kindling_topology_request_total{src_namespace="kindling"}[30m])';

//         const response2 = await getBackendSrv().datasourceRequest({
//             url: `/api/datasources/proxy/uid/${prometheusId}/api/v1/query`, // 使用 prometheusId 构建数据查询请求 URL
//             method: 'GET',
//             params: { 
//                 query: promql2,
//                 to: timeRange.to.toISOString()
//             },
//         });
//         console.log(response2);
//     } catch (error) {
//       // 处理错误
//       console.error('Error:', error);
//     }
// };
// // 通过grafana的api query data source，需要构造对应的请求参数
// export const queryPrometheusData2 = async (prometheusId: string, timeRange: any) => {
//     try {
//         const promql = 'increase(kindling_topology_request_total{src_namespace="kindling"}[$__range])'; // 替换为你的 PromQL 查询语句
    
//         const response = await getBackendSrv().datasourceRequest({
//             url: `/api/ds/query`, // 使用 prometheusId 构建数据查询请求 URL
//             method: 'POST',
//             data: { 
//                 queries: [
//                     {
//                         refId: 'K',
//                         expr: promql,
//                         datasource: { uid: prometheusId },
//                         format: 'time_series',
//                         maxDataPoints: 1,
//                         intervalMs: 1800000
//                     }
//                 ], 
//                 range: timeRange,
//                 from: timeRange.from.valueOf().toString(), 
//                 to: timeRange.to.valueOf().toString() 
//             },
//         });
  
//         const data = response.data;
//         // 处理获取的数据
//         console.log('Data:', data);
//     } catch (error) {
//         // 处理错误
//         console.error('Error:', error);
//     }
// };

const excuteQuery = async(queryText: string, timeRange: TimeRange) => {
    const kindlingDataSourceName = 'Prometheus';
    // const prometheusDataSource = await getDataSourceSrv().get(kindlingDataSourceName);
    const prometheusDataSource = getDataSourceSrv().getInstanceSettings(kindlingDataSourceName);
    let prometheusUid = prometheusDataSource?.uid;
    
    const from =  timeRange.from.valueOf().toString();
    const to = timeRange.to.valueOf().toString();
    const intervalMs = parseInt(to, 10) - parseInt(from, 10);
    try {
        const response = await getBackendSrv().datasourceRequest({
            url: `/api/ds/query`, // 使用 prometheusId 构建数据查询请求 URL
            method: 'POST',
            data: { 
                queries: [
                    {
                        refId: 'metric',
                        expr: queryText,
                        datasource: { uid: prometheusUid },
                        format: 'time_series',
                        maxDataPoints: 1,
                        intervalMs: intervalMs
                    }
                ], 
                range: timeRange,
                from: from, 
                to: to 
            },
        });
        const data = toDataQueryResponse(response);
        // 处理获取的数据
        console.log('Data:', data);
        return Promise.resolve(data);
    } catch (error) {
        // 处理错误
        console.error('Error:', error);
        return Promise.resolve([]);
    }
    // const response: any = await request(api, `query=${queryText}&time=${nowTime}`);
    // if (response.data.status) {
    //     const result = transformResponseData(response.data.data.result);
    //     return Promise.resolve(result);
    // } else {
    //     console.error(response.statusText);
    //     return Promise.resolve([]);
    // }
}

const getQueryText = (metric: string, namespace: string, workload: string) => {
    console.log(metric, namespace, workload);
    let queryPQL = '';
    let functionName = 'increase';
    if (metric === 'kindling_tcp_srtt_microseconds') {
        functionName = 'avg_over_time';
    }

    if (workload.indexOf(',') > -1) {
        queryPQL = `${functionName}(${metric}{src_namespace="${namespace}"}[$__range]) or ${functionName}(${metric}{dst_namespace="${namespace}"}[$__range])`;
    } else {
        queryPQL = `${functionName}(${metric}{src_namespace="${namespace}", src_workload_name="${workload}"}[$__range]) or ${functionName}(${metric}{dst_namespace="${namespace}", dst_workload_name="${workload}"}[$__range])`;
    }
    return queryPQL;
}

/**
 * 切换指标时再查询对应的指标详细信息数据
 * @param metric 
 * @param timeRange 
 */
export const metricQuery = async (metric: MetricType, namespace: string, workload: string, timeRange: TimeRange) => {
    // console.log(metric, timeRange);
    let data: any;
    if (metric === 'sentVolume') {
        let promQL = getQueryText('kindling_topology_request_request_bytes_total', namespace, workload);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'receiveVolume') {
        let promQL = getQueryText('kindling_topology_request_response_bytes_total', namespace, workload);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'retransmit') {
        let promQL = getQueryText('kindling_tcp_retransmit_total', namespace, workload);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'rtt') {
        let promQL = getQueryText('kindling_tcp_srtt_microseconds', namespace, workload);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'packageLost') {
        let promQL = getQueryText('kindling_tcp_packet_loss_total', namespace, workload);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'connFail') {
        let promQL = getQueryText('kindling_tcp_connect_total', namespace, workload);
        data = await excuteQuery(promQL, timeRange);
    }
    return data;
}
