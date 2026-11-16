// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

import _ from 'lodash';
export interface SelectOption {
    label: string, 
    value: string,
}

class FilterList {
    namespaceList: SelectOption[];
    workloadList: SelectOption[];
    workloadListByNamespace: any;

    constructor(data: any, namespace: string) {
        this.namespaceList = [];
        this.workloadList = [];
        this.workloadListByNamespace = {};

        this.init(data, namespace);
    }

    init(data: any[], namespace: string) {
        let namespaces = _.uniq(_.concat(_.map(data, 'src_namespace'), _.map(data, 'dst_namespace')));
        // this.namespaceList.push({
        //     label: 'All', 
        //     value: 'all'
        // });
        this.workloadList.push({
            label: 'All', 
            value: 'all'
        });
        
        _.forEach(namespaces, namespace => {
            this.namespaceList.push({
                label: namespace, 
                value: namespace
            });

            let workloadlist = _.filter(data, item => item.src_namespace === namespace || item.dst_namespace === namespace);
            let srcWorkloads = _.map(_.filter(workloadlist, item => item.src_namespace === namespace), 'src_workload_name');
            let dstWorkloads = _.map(_.filter(workloadlist, item => item.dst_namespace === namespace), 'dst_workload_name');

            let workloads = _.uniq(_.concat(srcWorkloads, dstWorkloads));
            this.workloadListByNamespace[namespace] = [];
            _.forEach(workloads, workload => {
                if (workload) {
                    this.workloadListByNamespace[namespace].push({
                        label: workload, 
                        value: workload
                    });
                }
            });
        });
        this.namespaceList.sort((a, b) => a.value.localeCompare(b.value));
        this.namespaceList.push({
            label: 'All', 
            value: 'all'
        });
        if (namespace !== 'all') {
            this.changeNamespace(namespace);
        }
    }

    changeNamespace(namespace: string) {
        let workloads = this.workloadListByNamespace[namespace];
        this.workloadList = _.concat([{
            label: 'All', 
            value: 'all'
        }], workloads);
    }
}
    
export default FilterList;
