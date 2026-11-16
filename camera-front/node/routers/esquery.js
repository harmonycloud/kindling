
'use strict';
const { Client } = require('@elastic/elasticsearch')


const esServerConfig = require('../settings').esServerConfig;
const esClient = new Client({ node: `http://${esServerConfig.host}:${esServerConfig.port}` });


function esClientService() {
    /**
     * 获取ES查询出来的数据，这里使用的是对name进行类似于sql中的like "%aaa%" 的查询方式，其他的查询方式，后面再补上
     * @param index 索引
     * @param keyword 关键词
     * @param value
     * @param pagenum 页码
     * @param size 单页大小
     */
    this.getEsData = async (index, keyword, value, pagenum, size) => {
        // 获取数据
        const result = await esClient.search({
            index,
            body: {
                track_total_hits: true,
                query: {
                    bool: {
                        must: [{
                            "query_string": {
                                "default_field": keyword,
                                "query": value
                            }
                        }],
                        must_not: [],
                        should: []
                    }
                },
                search_after: [(pagenum - 1) * size],
                size: size,
                sort: [
                    {"_id": "asc"}
                ]
            }
        }, {
            headers: {
                'content-type': 'application/json'
            }
        });
        result.size = size;
        result.pages = Math.ceil(result.body.hits.total.value / size);
        return Promise.resolve(result);
        // return result;
    }


}

module.exports = esClientService;