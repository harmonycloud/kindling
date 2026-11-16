import axios from 'axios';

console.log('env.DEV', import.meta.env.DEV);
const basicUrl = import.meta.env.DEV ? '/api' : '';
interface IParams1 {
    pid?: any;
    startTimestamp: number;
    endTimestamp: number;
    isServer: boolean;
    podName?: string;
}
// export const getTraceList = (params: IParams1) => {
//     return axios.get(basicUrl + '/camera/trace', {params});
// }
export const getTraceList = (params) => {
    return axios.get(basicUrl + '/esserver/getTraceList', {params});
}

interface IParams2 {
    pid: number;
    startTimestamp: number;
    endTimestamp: number;
}
// export const getTraceData = (params: IParams2) => {
//     return axios.get(basicUrl + '/camera/onoffcpu', {params});
// }
export const getTraceData = (params: IParams2) => {
    return axios.get(basicUrl + '/esserver/onoffcpu', {params});
}

interface IParams3 {
    pid: number;
    type: string;
    startTimestamp: number;
    endTimestamp: number;
    srcIp: string;
    srcPort: string;
    dstIp: string;
    dstPort: string;
}
export const getTracePayload = (params: IParams3) => {
    return axios.get(basicUrl + '/camera/tracePayload', {params});
}



//  node 层接口
export const getFolderList = () => {
    return axios.get(basicUrl + '/file/getFolders', {});
}
export const getFileList = (params: any) => {
    return axios.get(basicUrl + '/file/getAllTraceFileList', {params});
}
export const getFileDetail = (params: any) => {
    return axios.get(basicUrl + '/file/getTraceFile', {params});
}


// profile接口
export const toggleProfile = (params: any) => {
    return axios.post(basicUrl + '/profile', params);
}