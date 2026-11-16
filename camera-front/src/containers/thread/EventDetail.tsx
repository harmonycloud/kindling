import _ from 'lodash';
import moment from 'moment';
import { useEffect, useState } from 'react';
import DescriptionList, { Specification } from '../components/DescriptionList';
import { formatTimeNs } from '@/containers/thread/camera/util';
import LogTable from '../components/logTable';
import './index.less';

const DateFormat = 'YYYY-MM-DD HH:mm:ss.SSS';
interface Props {
  data: any;
}

const netTraceList: Specification[] = [
  {
    name: '发生时间',
    key: 'startTime',
    space: 'quater',
    render: (value) => value ? moment(value).format(DateFormat) : '--'
  }, {
    name: '响应时间',
    key: 'time',
    space: 'quater',
    render: (value) => value ? formatTimeNs(value) : '--'
  }, {
    name: '协议',
    key: 'protocol',
    space: 'quater'
  }, {
    name: '返回码',
    key: 'statuscode',
    space: 'quater',
    // render: (value) => value ? <span style={{color: parseInt(value) >= 400 ? '' : ''}}>{value}</span> : '--'
  }, {
    name: '请求报文',
    key: 'requestMessage',
    space: 'full'
  }, {
    name: '响应报文',
    key: 'responseMessage',
    space: 'full'
  }
];
function getDescList(data): Specification[] {
  return [
    {
      name: '发生时间',
      key: 'startTime',
      space: 'quater',
      render: (value, allObj) => value ? moment(value).format(DateFormat) : '--'
    },
    {
      name: '操作时间',
      key: 'time',
      space: 'quater',
      render: (value, allObj) => value ? (data.eventType === 'runqLatency' ? (parseFloat(value) / 1000).toFixed(3) : parseFloat(value).toFixed(2)) + 'ms' : '--'
    },
    {
      name: '线程名称',
      key: 'threadName',
      space: 'quater'
    }, 
    ...(data.eventType !== "on" && data.runqLatency ? [{
      name: '调度等待时间',
      key: 'runqLatency',
      space: 'quater',
      render: (value, allObj) => value ? (parseInt(value) / 1000 < 0 ? (parseInt(value) / 1000).toFixed(3) + 'ms' : (parseInt(value) / 1000).toFixed(2) + 'ms') : '--'
    }] as Specification[] : []),
    ...(data.eventType === "trace" ? [{
      name: 'Trace ID',
      key: 'traceId',
      space: 'quater'
    }] as Specification[] : []),
    ...(data.eventType === "futex" ? [{
      name: '锁地址',
      key: 'address',
      space: 'quater'
    }] as Specification[] : []),
    ...(data.type === "lock" ? [{
      name: '锁地址',
      key: 'lockAddress',
      space: 'quater'
    }, {
      name: '等待线程',
      key: 'waitThread',
      space: 'quater'
    }] as Specification[] : []),
    ...(data.eventType === "epoll" ? [{
      name: '操作文件类型',
      key: 'operate',
      space: 'quater'
    }, {
      name: '大小',
      key: 'size',
      space: 'quater'
    }, {
      name: '时间戳',
      key: 'timestamp',
      space: 'quater'
    }, {
      name: '连接信息',
      key: 'file',
      space: 'half'
    }] as Specification[] : []),
    ...(data.eventType === "net" ? [{
      name: '操作文件类型',
      key: 'operate',
      space: 'quater'
    }, {
      name: '操作大小',
      key: 'size',
      space: 'quater',
      render: (value) => `${value}byte`
    }, {
      name: '连接信息',
      key: 'file',
      space: 'half'
    }, {
      name: '报文信息',
      key: 'message',
      space: 'full'
    }] as Specification[] : []),
    ...(data.eventType === "file" ? [{
      name: '操作文件',
      key: 'file',
      space: 'quater'
    }, {
      name: '操作文件类型',
      key: 'operate',
      space: 'quater'
    }] as Specification[] : []),
    ...(data.type === "lock" ? [{
      name: '当前堆栈',
      key: 'stack',
      space: 'full',
      render: value => <LogTable list={_.compact(value.split(';'))}/>
    }] as Specification[] : []),
    ...(data.eventType === "log" ? [{
      name: '日志',
      key: 'log',
      space: 'full',
      // render: list => <LogTable list={list}/>
    }] as Specification[] : [])
  ];
}
export default function EventDetail(props: Props) {

  const { data } = props;

  const [newData, setNewData] = useState<any>({});
  const [traceData, setTraceData] = useState<any>({});
  const [specifications, setSpecifications] = useState<Specification[]>([]);
  const [netColumnList, setNetColumnList] = useState<Specification[]>([]);

  const init = (data) => {
    let info: any = {};
    if (data.type === 'net' && !_.isEmpty(data.traceInfo)) {
      let netColumns = _.cloneDeep(netTraceList);
      let totalTime = _.find(data.traceInfo?.metrics, {Name: 'request_total_time'});
      let traceData: any = {
        startTime: Math.floor(data.traceInfo?.timestamp / 1000000),
        // time: data.traceInfo?.metrics?.request_total_time,
        time: totalTime.Data.Value,
        statuscode: data.traceInfo?.labels?.http_status_code,
        protocol: data.traceInfo?.labels?.protocol,
        requestMessage: data.traceInfo.labels?.request_payload,
        responseMessage: data.traceInfo.labels?.response_payload
        // message: data.requestType === 'request' ? data.traceInfo.labels?.request_payload : data.traceInfo.labels?.response_payload
      };
      if (traceData.protocol === 'http') {
        traceData.url = data.traceInfo?.labels?.http_url;
        netColumns.push({
          name: '请求URL',
          key: 'url',
          space: 'full'
        });
      } else if (traceData.protocol === 'mysql') {
        traceData.url = data.traceInfo?.labels?.sql;
        netColumns.push({
          name: 'SQL语句',
          key: 'url',
          space: 'full'
        });
      }
      setNetColumnList(netColumns);
      setTraceData(traceData);
    } else {
      setNetColumnList(netTraceList);
      setTraceData({});
    }
    if (data.info) {
      info = {
        ...data.info,
        eventType: data.info.type
      }
      delete info.type;
      delete info.startTime;
    }
    const newdata = {
      ...data,
      ...info
    };
    setNewData(newdata);
    setSpecifications(getDescList(newdata));
  }

  useEffect(() => {
    if (data) {
      init(data);
    } else {
      setSpecifications([]);
      setNetColumnList([]);
    }
  }, [data]);

  return (
    <div className="event_detail f-padding10">
      <DescriptionList
        title={`${newData.eventType ?? '--'}`}
        data={newData}
        specifications={specifications}
      />
      {
        data.type === 'net' && _.keys(traceData).length > 0 ? <div style={{ marginTop: '10px' }}>
          <DescriptionList
            title="相关请求信息"
            data={traceData}
            specifications={netColumnList}
          />
        </div> : null
      }
    </div>
  );
}
