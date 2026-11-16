
'use strict';
const router = require('express').Router();
const _ = require('lodash');
const fs = require('fs');
const esClientService = require('./esquery');
const esService = new esClientService();
const { Client } = require('@elastic/elasticsearch');

const esServerConfig = require('../settings').esServerConfig;
const traceIndex = esServerConfig.trace_index;
const onoffCpuIndex = esServerConfig.onoffcpu_index;
const esClient = new Client({ node: `http://${esServerConfig.host}:${esServerConfig.port}` });

router.get('/getTraceList', (req, res, next) => {
    let { pid, startTimestamp, endTimestamp, protocol, isServer } = req.query;
    console.log('/esserver/getTraceList:', pid, startTimestamp, endTimestamp, protocol, isServer);
    let query = [
        {
            term: {'labels.is_server': 'true'}
        }
    ];
    if (protocol && protocol !== '') {
        query.push({
            match: {'labels.protocol': protocol}
        });
    }
    if (pid && pid !== 0) {
        query.push({
            match: {'labels.pid': pid}
        });
    }
    startTimestamp = startTimestamp * 1000000;
    endTimestamp = endTimestamp * 1000000;
    query.push({
        range: {'timestamp': {
            'gte': startTimestamp,
            'lte': endTimestamp
        }}
    });
    esClient.search({
        index: traceIndex,
        body: {
            query: {
                bool: {
                    must: query
                }
            },
            size: 100,
            sort: [
                {
                    'timestamp': {"order" : "desc"}
                }
            ]
        },
    }, {
        headers: {
            'content-type': 'application/json'
        }
    }, (err, result) => {
        if (err) {
            console.log(err);
            res.status(504).json({
                success: false,
                data: err
            });
        } else {
            let hits = result.body.hits.hits;
            let fresult = _.map(hits, '_source');

            res.status(200).json(fresult);
        }        
    });
});

router.get('/onoffcpu', (req, res) => {
    const { pid, startTimestamp, endTimestamp } = req.query;
    esClient.search({
        index: onoffCpuIndex,
        body: {
            query: {
                bool: {
                    must: [
                        {
                            match: {'labels.pid': pid}
                        }, {
                            range: {
                                'labels.startTime': {
                                    'lte': endTimestamp * 1000000
                                }
                            }
                        }, {
                            range: {
                                'labels.endTime': {
                                    'gte': startTimestamp * 1000000
                                }
                            }
                        }

                    ]
                }
            },
            size: 10000,
            sort: [
                {
                    'labels.tid': {"order" : "desc"}
                }, {
                    'labels.startTime': {"order" : "asc"}
                }
            ]
        }
    }, {
        headers: {
            'content-type': 'application/json'
        }
    }, (err, result) => {
        if (err) {
            console.log(err);
            res.status(504).json({
                success: false,
                data: err
            });
        } else {
            let hits = result.body.hits.hits;
            let fresult = _.map(hits, '_source');

            res.status(200).json(fresult);
        }        
    });
})


function handleData(data) {
    let nodes = [], edges = [];
    _.forEach(data, item => {
        let metric = _.find(item.metrics, { Name: 'request_total_time' });
        let timestamp = parseInt(item.timestamp / 1000000);
        if (_.findIndex(nodes, {apm_span_ids: item.labels.apm_span_ids}) === -1) {
            let node = {
                content_key: item.labels.content_key,
                id: item.labels.apm_span_ids,
                apm_span_ids: item.labels.apm_span_ids,
                dst_container: item.labels.dst_container,
                dst_pod: item.labels.dst_pod,
                dst_workload_name: item.labels.dst_workload_name,
                pid: item.labels.pid,
                protocol: item.labels.protocol,
                list: [
                    {
                        endTime: timestamp,
                        totalTime: metric ? metric.Data.Value : 0,
                        p90: item.labels.p90,
                        is_profiled: item.labels.is_profiled,
                    }
                ]
            }
            nodes.push(node);
        } else {
            let node = _.find(nodes, {apm_span_ids: item.labels.apm_span_ids});
            if (_.findIndex(node.list, {endTime: timestamp}) === -1) {
                node.list.push({
                    endTime: timestamp,
                    totalTime: metric ? metric.Data.Value : 0,
                    p90: item.labels.p90,
                    is_profiled: item.labels.is_profiled,
                });
            }
        }
    });
    _.forEach(data, item => {
        if (item.labels.apm_parent_id !== '0') {
            let sourceNode = _.find(nodes, node => node.apm_span_ids.split(',').indexOf(item.labels.apm_parent_id) > -1);
            if (sourceNode) {
                let sourceId = sourceNode.apm_span_ids;
                if (_.findIndex(edges, {source: sourceId, target: item.labels.apm_span_ids}) === -1) {
                    let edge = {
                        source: sourceId,
                        target: item.labels.apm_span_ids
                    }
                    edges.push(edge);
                }
            }
        }
    });
    return { nodes, edges };
}

router.get('/getTraceData', async(req, res, next) => {
    const { traceId } = req.query;
    console.log(traceId);
    // try {
    //     let result = await esService.getEsData('span_trace_group_dev', 'labels.trace_id', traceId, 1, 100);
    //     console.log('result', result);
    //     let hits = result.body.hits.hits;
    //     console.log(JSON.stringify(result.body.hits));
    //     let data = _.map(hits, '_source');
    //     let finalResult = handleData(data);

    //     res.status(200).json({
    //         success: true,
    //         data: finalResult
    //     });
    // } catch (error) {
    //     console.log('我报错了 哈哈哈哈哈哈');
    //     console.log('error', error)
    // }
    
    esClient.search({
        index: traceIndex,
        // 如果您使用的是Elasticsearch≤6，取消对这一行的注释
        // type: '_doc',
        body: {
            query: {
                match: { 'labels.trace_id.keyword': traceId }
            }
        },
    }, {
        headers: {
            'content-type': 'application/json'
        }
    }, (err, result) => {
        if (err) {
            console.log(err);
            res.status(504).json({
                success: false,
                data: err
            });
        } else {
            let hits = result.body.hits.hits;
            let data = _.map(hits, '_source');
            let finalResult = handleData(data);

            res.status(200).json({
                success: true,
                data: finalResult
            });
        }        
    });
}); 

module.exports = router;