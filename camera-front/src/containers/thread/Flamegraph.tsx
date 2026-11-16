import { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import moment from 'moment';
import { Divider } from 'antd';
import { flamegraph } from 'd3-flame-graph';
import 'd3-flame-graph/dist/d3-flamegraph.css';

import DescriptionList, { Specification } from '../components/DescriptionList';
import './index.less';

const DateFormat = 'YYYY-MM-DD HH:mm:ss.SSS';
const descList: Specification[] = [
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
      render: (value, allObj) => value ? parseFloat(value).toFixed(2) + 'ms' : '--'
    },
    {
      name: '线程名称',
      key: 'threadName',
      space: 'quater'
    }
];
interface TreeData {
  children?: TreeData[],
  depth: number;
  index: number;
  name: string;
  width: number;
  color: number;
  value: number;
  nextNodeIndex: number;
}

export const sleep = (n = 500) => {
  return new Promise((resolve) => setTimeout(resolve, n));
};


const sortData = (a, b) => {
  if (a.depth < b.depth) return -1;
  else if (a.depth > b.depth) return 1;
  else return a.index - b.index;
};
function transferTree(stackList: TreeData[]) {
  let treeData: any = {};
  if (!Array.isArray(stackList)) return treeData;
  stackList.sort(sortData);
  let tmp: TreeData;
  let nodeList: TreeData[] = [];
  stackList.forEach((item, index) => {
    if (item.depth === 0) {
      const obj: TreeData = {
        ...item,
        value: item.width,
        children: [],
        nextNodeIndex: index + 1
      };
      treeData = obj;
      nodeList.push(obj);
      tmp = treeData;
    } else {
      let flag = true;
      do {
        if ((item.depth === tmp.depth + 1) && item.index >= tmp.index && item.index < tmp.index + tmp.width) {
          flag = false;
          const obj: TreeData = {
            ...item,
            value: item.width,
            children: [],
            nextNodeIndex: index + 1
          };
          tmp.children?.push(obj);
          nodeList.push(obj);
        } else {
          tmp = nodeList[tmp.nextNodeIndex]
        }
      } while (flag && tmp);
    }
  })
  return treeData;
}
export default function Flamegraph(props) {

  const { data } = props;

  const ref = useRef<HTMLDivElement>(null);
  const [hoverCtx, setHoverCtx] = useState<any>(null)
  let chart, renderFlag, init = true;

  // 插件自带的tooltip在边界位置无法自适应，故手动实现
  const renderTooltip = (d) => {
    renderFlag = true;
    const { data } = d;
    const mouseEvent = window.event as any;
    const { layerX, layerY } = mouseEvent;
    const dom: any = ref.current;
    const alignL = layerX + 250 > dom.offsetWidth;
    const ctx = <div className="fg_tooltip" style={{ left: alignL ? layerX - 250 : layerX + 10, top: layerY - 120 }}>
      <div className="fg_tooltip_title">{data.name ?? '--'}</div>
      <Divider style={{ margin: "8px 0px" }} />
      <div>share of CPU：{data.cpu ?? '--'}</div>
      {/* <div>CPU Time：6.31 minutes</div> */}
      {/* <div>Samples：4327</div> */}
    </div>;
    if (hoverCtx !== ctx) {
      setHoverCtx(ctx);
    }
  }

  const drawFlamegraph = (container, drawData) => {
    const width = container.offsetWidth - 10;
    d3.select('.flamegraph').selectAll('svg').remove()
    chart = flamegraph()
      .width(width)
      .cellHeight(18)
      .transitionDuration(750)
      .minFrameSize(5)
      // .transitionEase('0.5')
      // .sort(true)
      // .setColorHue('pastelgreen')
      // .onClick(onClick)
      //@ts-ignore  
      // 先注释掉不显示hover
      // .onHover(renderTooltip)
      .inverted(true)
      .selfValue(false);
    console.log(drawData);
    d3.select(`.flamegraph`)
      .datum(drawData)
      .call(chart)
    d3.select('.flamegraph').selectAll('g').on('mouseleave', async () => {
      renderFlag = false;
      await sleep(100);
      if (!renderFlag) {
        setHoverCtx(null);
      }
    });
  }

  const onResize = () => {
    const dom = ref.current;
    if (dom) {
      drawFlamegraph(dom, data.stackList);
    }
  }

  useEffect(() => {
    // window.addEventListener("resize", onResize);
    // console.log(data);
    if (data.stackList) {
      if (init) {
        init = false;
        const dom = ref.current;
        if (dom) {
          drawFlamegraph(dom, transferTree(data.stackList));
        }
      }
      else if (chart) {
        chart.update(transferTree(data.stackList))
      }
    }
    return () => {
      d3.select('.flamegraph').selectAll('svg').remove();
    }
  }, [data])

  return (
    <div className="event_detail f-padding10">
      <DescriptionList
        title={`${data.eventType || '--'}`}
        data={data}
        specifications={descList}
      />
      {data?.stackList?.length ? <>
        <div
          className="flamegraph"
          ref={ref}
        />
        {hoverCtx}</>
        : <div>--</div>}
    </div>
  );
}
