import * as d3 from 'd3';
import { D3BrushEvent } from 'd3';
import { message, Modal } from 'antd';
import _ from 'lodash';
import { ILegend, IEvent, IOption, IThread, IEventTime, IJavaLock, ILogEvent, IFilterParams, ILineTime } from './types';
import { eventList as LEvent, darkEventList as DEvent, textHandle, timeFormat, containTime } from './util';
import { getStore } from '@/services/util';


type ILineType = 'requestTime' | 'cruxTime' | 'trace' | 'rt';
// export interface IOption {
//     svgId?: string;
//     nameWidth?: number;
//     barWidth?: number;
//     barPadding?: number;
//     padding?: number;
// }
const eventList = (getStore('theme') || 'light') === 'light' ? LEvent : DEvent;
class Camera {
    parentRef: any;
    theme: 'light' | 'dark' = 'light';
    data: IThread[] = [];
    lineTimeList: ILineTime[] = [];
    timeThreshold: number = 0.0025;
    trace: any;
    traceId: string;
    svgId: string;
    svg: any;
    svgWidth: number = 500;     // 绘制svg区域的宽度
    nameWidth: number;          // 线程名称区域的宽度
    barWidth: number;           // 柱状的高度 (完整绘制线程图的高度 = barWidth + barPadding * 2)
    barPadding: number;         // 柱状上下padding距离 
    padding: number;            // svg绘制区域的padding
    barWarpWidth: number = 500;
    topTimeheight: number = 30;
    ifShowTooltipLine: boolean = false;
    threadHeight: number = 50;
    chartWarpHeight: number = 50;    // 主题绘制进程柱状进度图的高度
    bottomxAxisHeight: number = 28;
    lockEventTypeList: string[] = _.chain(eventList).filter(item => item.type === 'lock').map('value').value();
    
    requestTraceXDiff: number = 2; // request线跟trace线的x坐标合并间隔
    filters: IFilterParams;
    timeRange: Date[];          // traceId发生的时间前后推五秒，组成的时间区间
    timeRangeDiff: number;      // timeRange的时间区间差值
    requestTimes: Date[] = [];  // trace请求开始跟结束时间
    cruxTimes: number[] = [];     // 用户自定义的关键时刻数组
    showRunQFlag: boolean = true;
    showLogFlag: boolean = true;       // 是否显示日志的标志位
    showJavaLockFlag: boolean = true;  // 是否显示lock事件的标志位
    showTraceFlag: boolean = true;         // 是否开启trace分析标志位
    activeEventDomId: string = '';      // 当前选中查看详情的event id => id_idx1_idx2

    xScale: any;                // 上方线程图的时间序数比例尺对象
    xAxis: any;                 // 上方线程图生成X轴对象
    xScale2: any;               // 底部时间筛选的时间序数比例尺对象
    xAxisWarp: any;
    chartBrush: any;
    chartBrushTimeRange: any[] = [];
    xAxisBrush: any;
    isPrint: boolean = false;

    nameClick: (tid: string) => void;
    eventClick: (evt: any) => void;

    constructor (option: IOption) {
        this.parentRef = option.parentRef;
        this.theme = (getStore('theme') as any) || 'light';
        this.data = option.data;
        this.lineTimeList = option.lineTimeList;
        this.trace = option.trace;
        this.traceId = option.traceId;
        this.svgId = option?.svgId || 'camera_svg';
        this.nameWidth = option?.nameWidth || 150; // 侧边轴上的名称宽度包括跟柱图的间隔20
        this.barWidth = option?.barWidth || 26; // 柱图的高度
        this.barPadding = option?.barPadding || 12; // 柱图的上下padding
        this.padding = option?.padding || 20;

        
        this.timeRange = option.timeRange;
        this.timeRangeDiff = new Date(option.timeRange[1]).getTime() - new Date(option.timeRange[0]).getTime();
        this.filters = {
            threadList: [], 
            logList: [],
            fileList: [],
            eventList: []
        };

        this.nameClick = option.nameClick;
        this.eventClick = option.eventClick.bind(this);
    }

    draw() {
        let _this = this;

        const inner: any = document.getElementById('camera_svg');
        inner.innerHTML = '';

        // 柱图高度加上下间隔组成对应rect的高度，撑开thread_warp
        const threadHeight = this.barWidth + this.barPadding * 2; 
        this.threadHeight = threadHeight;
        this.chartWarpHeight = threadHeight * this.data.length;

        this.svg = d3.select('#camera_svg').attr('height', this.chartWarpHeight);
        const cameraSize = document.getElementById('camera')?.getBoundingClientRect();
        this.svgWidth = cameraSize?.width as number;
        
        const chartWarp = this.svg.append('g').attr('id', 'chart_warp').attr('transform', `translate(0, 0)` );

        this.xScale = d3.scaleLinear()
            .domain(this.timeRange)
            .range([this.nameWidth, this.svgWidth]);
        this.xAxis = d3.axisBottom(this.xScale)
            // .ticks(d3.timeMonth.every(2))
            .tickFormat(d => timeFormat(d as Date));

        // 在线程图上监听zoom事件，实现zoom跟brush联动
        const zoomed = (e: any) => {
            console.log('zoom', e);
            let t = e.transform;
            const timeRange = t.rescaleX(this.xScale).domain();
            console.log(timeRange);
            if (timeRange[0] === this.timeRange[0] && timeRange[1] === this.timeRange[1]) {
                return;
            }
            if (timeRange[1] - timeRange[0] <= 100) {
                return;
            }
            this.xScale.domain(timeRange);
            this.updateChart(timeRange);


            let s = this.xScale.domain();
            //把当前的domain通过x2Scale转化为range的数字数组(即为brush的位置信息)
            let d = s.map(item => {
                return this.xScale2(item)
            })
            //通过brush.move方法动态修改brush的大小与位置
            this.xAxisBrush.move(this.xAxisWarp, d)
        }
        //定义缩放zoom
        const zoom = d3
            .zoom() // 设置zoom
            //设置缩放范围
            .scaleExtent([1, Infinity])
            //设置transform的范围
            .translateExtent([[this.nameWidth, 0], [this.svgWidth, this.chartWarpHeight]])
            //设置缩放的视口的大小; 注:此时视口大小与transform范围一样说明无法拖动只可滚轮缩放
            .extent([[this.nameWidth, 0], [this.svgWidth, this.chartWarpHeight]])
            .on('zoom end', zoomed)

        // 图表上增加brush操作 - 定义brush
        const brushEnd = function(e) {
            const { selection } = e;
            if (selection) {
                let tiemRange = selection.map(_this.xScale.invert, _this.xScale);
                console.log('chart brush brushEnd', tiemRange);
                if (tiemRange[1] - tiemRange[0] >= 10) {
                    _this.chartBrushTimeRange.push(tiemRange);
                    _this.updateChart(tiemRange);   
                    _this.setxAxisBrushByChartBrush();
                } else {
                    message.warning('只支持到十毫秒的时间区间');
                }
                // 每次筛选完之后直接移除brush（隐藏brush操作区域）
                _this.removeChartBrush();
                d3.select('.chart_brush .selection').attr('width', 0);
            }
        }
        this.chartBrush = d3.brushX()
            .extent([[this.nameWidth, 0], [this.svgWidth, this.chartWarpHeight]])
            .on('end', brushEnd);
        
        const barWarpWidth = this.svgWidth - this.nameWidth - this.padding;
        this.barWarpWidth = barWarpWidth;

        // 绘制主体柱图区域
        _.forEach(this.data, (item, idx: number) => {
            let threadWarp = chartWarp.append('g')
                .attr('id', `thread_warp_${item.tid}`)
                .attr('data-tid', item.tid)
                .attr('class', 'thread_warp')
                .attr('height', threadHeight)
                .attr('transform', `translate(0, ${threadHeight * idx})` );
            threadWarp.append('rect')
                .attr('class', 'thread_rect_warp')
                .attr('fill', () => this.theme === 'light' ? (idx % 2 === 0 ? '#FFFFFF' : '#F9F9F9') : (idx % 2 === 0 ? '#212226' : '#333333'))
                .attr('width', this.svgWidth)
                .attr('height', threadHeight);

            if (item.name.length > 18) {
                let nameLine1 = item.name.substring(0, 18);
                let nameLine2 = item.name.substring(18);
                const textWarp = threadWarp.append('g').attr('class', 'left_name_warp');
                this.drawThreadName(textWarp, nameLine1, item, this.barPadding + this.barWidth / 2 - 5);
                this.drawThreadName(textWarp, textHandle(nameLine2, 18), item, this.barPadding + this.barWidth / 2 + 10);
                textWarp.append('title').text(item.name);
                // textWarp.append('text')
                //     .text(nameLine1)
                //     .attr('data-tid', item.tid)
                //     .attr('class', `left_name ${item.active && 'active'}`)
                //     .attr('x', function (this: any) {
                //         let textlen: number = this.getComputedTextLength();
                //         return _this.nameWidth - 20 - textlen;
                //     })
                //     .attr('y', this.barPadding + this.barWidth / 2 + 5)
                //     .on('click', function(this: any) {
                //         let tid = d3.select(this).attr('data-tid');
                //         _this.nameClick(tid);
                //     });
            } else {
                this.drawThreadName(threadWarp, item.name, item, this.barPadding + this.barWidth / 2 + 5);
            }

            let barWarp = threadWarp.append('g').attr('class', 'bar_warp');

            _.forEach(item.eventList, (event, idx2: number) => {
                event.idx = idx2;
                event.timeRate = event.time / this.timeRangeDiff;
                event.left = this.xScale(new Date(event.startTime)) as number;
            });
            _.forEach(item.eventList, (event, idx2: number) => {
                if (event.timeRate && event.timeRate > this.timeThreshold) {
                    // let timeWidth = event.timeRate * (this.svgWidth - this.nameWidth - this.padding);
                    let right = this.xScale(new Date(event.endTime));
                    let timeWidth = right - event.left!;
                    let sevent: IEvent = _.find(eventList, {type: event.type}) as IEvent;
                    const eventWarp = barWarp.append('g')
                        .attr('id', `id_${idx}_${idx2}`)
                        .attr('data-type', event.eventType)
                        .attr('class', `event_warp ${event.eventType === 'runqLatency' ? 'event_runq' : ''}`);
                    eventWarp.append('rect')
                        .attr('class', 'event_rect')
                        .attr('width', timeWidth)
                        .attr('height', this.barWidth)
                        .attr('fill', sevent.fillColor)
                        .attr('x', event.left)
                        .attr('y', this.barPadding);
                    if (event.eventType === 'runqLatency') {
                        if (timeWidth! > 75) {
                            eventWarp.append('text').text(sevent.type)
                                .attr('class', 'event_rect_left_text')
                                .attr('stroke', sevent.color)
                                .attr('x', event.left && event.left + 5)
                                .attr('y', this.barPadding + this.barWidth / 2 + 5)
                        } 
                    } else {
                        if (event.timeRate > 0.05) {
                            eventWarp.append('text').text(sevent.type)
                                .attr('class', 'event_rect_left_text')
                                .attr('stroke', sevent.color)
                                .attr('x', event.left && event.left + 5)
                                .attr('y', this.barPadding + this.barWidth / 2 + 5)
                        }
                    }
                }
            });
            // 绘制虚线事件
            let lineData: IEventTime[] = _.filter(item.eventList, event => event.timeRate && event.timeRate <= this.timeThreshold) as IEventTime[];
            this.drawThreadEventDashLine(idx, lineData, barWarp);
            
            // 绘制trace的线段
            this.drawTraceLine(threadWarp, item.traceList, item.name);

            // 日志Icon绘制，默认初始不显示
            var symbol = d3.symbol()
                .size(60)
                .type(d3.symbolTriangle);
            threadWarp.selectAll('.log_icon')
                .data(item.logList).enter()
                .append('path')
                .attr('class', 'log_icon')
                .attr('d', symbol)
                .attr('data-tid', item.tid)
                .attr('fill', (log) => log.traceId === this.traceId ? '#1291A2' : '#E6F4F5')
                .attr('cursor', 'pointer')
                .attr('transform', (log: ILogEvent) => {
                    let position = this.xScale(new Date(log.startTime));
                    return `translate(${position}, ${this.barPadding + this.barWidth + 6})`
                })
                .style('visibility', 'hidden');

            // draw java lock => 绘制java lock锁的区域
            // const javalockWarp = threadWarp.selectAll('.javalock_warp')       
            //         .data(item.javaLockList)
            //         .enter()
            //         .append('g')
            //         .attr('class', 'javalock_warp')
            //         .attr('style', 'display: none')
            //     javalockWarp.append('rect')
            //         .attr('class', 'javalock_rect')
            //         .attr('data-thread', item.name)
            //         .attr('width', (lock: IJavaLock) => barWarpWidth * (lock.time / this.timeRangeDiff))
            //         .attr('height', this.barWidth + 8)
            //         .attr('cursor', 'pointer')
            //         .attr('x', (lock: IJavaLock) => this.xScale(new Date(lock.startTime)))
            //         .attr('y', this.barPadding - 4);
                
            //     javalockWarp.append('text')
            //         .text(d => d.eventType)
            //         .attr('x', (lock: IJavaLock) => this.xScale(new Date(lock.startTime)) + 5 )
            //         .attr('y', this.barPadding + this.barWidth / 2)
            threadWarp.selectAll('.javalock_rect')
                .data(item.javaLockList)
                .enter()
                .append('rect')
                .attr('class', 'javalock_rect')
                .attr('data-thread', item.name)
                .attr('data-type', d => d.eventType)
                .attr('width', (lock: IJavaLock) => this.xScale(new Date(lock.endTime)) - this.xScale(new Date(lock.startTime)))
                .attr('height', this.barWidth + 8)
                .attr('cursor', 'pointer')
                // .attr('style', 'display: none')
                .style('visibility', 'hidden')
                .attr('x', (lock: IJavaLock) => this.xScale(new Date(lock.startTime)))
                .attr('y', this.barPadding - 4);


            item.traceStartTimestamp && this.drawTraceStar(threadWarp, item.traceStartTimestamp, item.tid);
            item.traceEndTimestamp && this.drawTraceStar(threadWarp, item.traceEndTimestamp, item.tid);

        });
        // 绘制图标上的X轴时间轴用于时间映射，但是页面不显示
        chartWarp.append('g')
            .attr('class', 'xaxis_line')
            .attr('opacity', 0)
            .attr('transform', `translate(0, ${this.chartWarpHeight})`)
            .call(this.xAxis);

        // 绘制tooltip的虚线跟时间text
        this.drawTooltipLine(chartWarp);

        // chartWarp.call(zoom);
        chartWarp.append('g').attr('class', 'chart_brush')
            .style('display', 'none')
            .call(this.chartBrush)
            .call(this.chartBrush.move);

        // 绘制顶部的时间tooltip显示区域
        this.drawTopTimeWarp();
        // 绘制底部时间条形选择框
        this.drawxAxisBrush();
        // 绘制trace请求开始结束和trace处理开始结束的两条竖直线（若时间接近的情况下将两个竖线进行合并）
        this.drawLineTime(this.lineTimeList);
        
        this.isPrint = true;

        if (this.filters.threadList.length > 0 || this.filters.eventList.length > 0) {
            this.reprintByFilter(this.filters);
        }

        // event warp click 监听事件点击事件
        this.addEventClickListener();
        
        // log click 日志ICON的点击事件监听
        d3.selectAll('.log_icon').on('click', function(e, log: any) {
            const tid = d3.select(this).attr('data-tid');
            let threadItem = _.find(_this.data, {tid: parseInt(tid)});
            let evt = _.find(threadItem?.eventList, {startTime: log.startTime});
            const data = {
                threadName: threadItem?.name,
                // ...evt,
                ...log
            }
            _this.eventClick(data);
        });
        // Java lock 点击事件监听
        d3.selectAll('.javalock_rect').on('click', function(e, lock: any) {
            const thread = d3.select(this).attr('data-thread');
            const data = {
                threadName: thread,
                ...lock
            }
            _this.eventClick(data);
        });

        !this.showRunQFlag && this.hideRunQ();
        // 默认开启trace 分析，默认打开javalock 跟 log绘制
        this.showTraceFlag && this.startTrace(true);
        this.showJavaLockFlag && this.showJavaLock();
        this.showLogFlag && this.showLog();
        
        // 单个事件event的Tooltip事件监听
        // d3.selectAll('.event_warp').on('mouseenter', (e) => {
        //     console.log('event_warp mouseenter', e);
        //     let { layerX, layerY } = e;
        //     this.showEventTooltip(layerX, layerY);
        // });
        // // d3.selectAll('.event_warp').on('mousemove',  (e) => {
        // //     let { offsetX, offsetY } = e;
        // //     this.showEventTooltip(offsetX, offsetY, true);
        // // });
        // d3.selectAll('.event_warp').on('mouseleave', () => {
        //     console.log('event_warp mouseleave');
        //     this.hideEventTooltip();
        // });
    }
    drawThreadName(textWarp, text, item, y) {
        let _this = this;
        textWarp.append('text')
            .text(text)
            .attr('data-tid', item.tid)
            .attr('class', `left_name ${item.active && 'active'}`)
            .attr('x', function (this: any) {
                let textlen: number = this.getComputedTextLength();
                return _this.nameWidth - 20 - textlen;
            })
            .attr('y', y)
            .on('click', function(this: any) {
                let tid = d3.select(this).attr('data-tid');
                _this.nameClick(tid);
            });
    }
    // 绘制关键IO线程的read/write事件的 ☆
    drawTraceStar(threadWarp, timestamp, tid) {
        const starSymbol = d3.symbol().size(50).type(d3.symbolStar);
        let position = this.xScale(new Date(timestamp));
        threadWarp.append('path')
            .attr('class', 'start_icon')
            .attr('d', starSymbol)
            .attr('data-tid', tid)
            .attr('data-timestamp', timestamp)
            // .attr('id', `star_${idx}_${event.idx}`)
            .attr('cursor', 'pointer')
            .attr('transform', `translate(${position}, ${this.barPadding - 8})`);
    }
    // 点击 ☆ 的事件处理
    handleStarClick(tid, timestamp) {
        let {src_ip, src_port, dst_ip, dst_port} = this.trace.labels;
        let operateFile = `${src_ip}:${src_port}->${dst_ip}:${dst_port}`;
        let threadObj: IThread = _.find(this.data, {tid: parseInt(tid)})!;
        timestamp = parseInt(timestamp);
        let netEvent: any = null;
        if (threadObj) {
            _.forEach(threadObj.eventList, event => {
                if (event.type === 'net') {
                    if (event.info && event.info.file === operateFile) {
                        if (event.startTime > timestamp - 2 && event.startTime < timestamp + 2 ) {
                            netEvent = _.cloneDeep(event);
                        }
                    }
                }
            });
            if (netEvent) {
                // @ts-ignore
                netEvent.message = netEvent.eventType === "netread" ? this.trace.labels.request_payload : this.trace.labels.response_payload;
            }
            this.eventClick({...netEvent, traceInfo: this.trace});
        }
    }
    /**
     * 根据合并的事件判断当前绘制的虚线的颜色
     * 1.若存在net类型事件，返回net的颜色；2.若存在file类型事件，返回file的颜色；3.否则直接范返回黑色代码合并事件
     * @param list 当前合并的事件list
     * @returns color
     */
    getEventLineColor(list) {
        if (list.length > 1) {
            if (_.some(list, item => item.type === 'net')) {
                let sevent: IEvent = _.find(eventList, {type: 'net'}) as IEvent;
                return sevent.color;
            } else if (_.some(list, item => item.type === 'file')) {
                let sevent: IEvent = _.find(eventList, {type: 'file'}) as IEvent;
                return sevent.color;
            } else {
                return this.theme === 'dark' ? '#dcdcdc' : '#3d3d3d';
            }
        } else {
            let sevent: IEvent = _.find(eventList, {type: list[0].type}) as IEvent;
            return sevent.color;
        }
    }
    // 每次渲染或者筛选brush时，对时间占比很小绘制虚线的事件进行group聚合并绘制对应的虚线
    drawThreadEventDashLine(idx: number, lineData: IEventTime[], barWarp: any) {
        const _this = this;
        const groupLineData = _.groupBy(lineData, (item: IEventTime) => Math.ceil(item.left as number));
        let leftList = _.keys(groupLineData).map(v => parseInt(v));
        let group2: any = {};
        while(leftList.length > 0) {
            let p1 = leftList[0];
            let list = _.remove(leftList, p => p >= p1 && p <= p1 + 5);
            group2[p1] = list;
        }
        let group2LineData: any = {};
        _.forEach(group2, (lefts, k) => {
            group2LineData[k] = _.reduce(lefts, (result: any, left) => {
                return _.concat(result, groupLineData[left])
            }, []);
        });
        // console.log('groupLineData', groupLineData);
        // console.log('group2', group2);
        // console.log('group2LineData', group2LineData);
        _.forEach(group2LineData, (list, left) => {
            // let startT = list[0].startTime;
            // let endT = list[list.length - 1].endTime;
            // let time = endT - startT;
            // let timer = time / this.timeRangeDiff;
            // const width = timer * (this.svgWidth - this.nameWidth - this.padding);
            // console.log('timer', timer, timer * (this.svgWidth - this.nameWidth - this.padding));
            const lineWarp = barWarp.append('g').attr('class', 'event_dash_warp');
            lineWarp.append('rect')
                .attr('class', 'event_dash_rect')
                // .attr('id', `line_rect_${idx}_${left}`)
                // .attr('data-idxs', _.join(_.map(list, 'idx'), '_'))
                // .attr('data-types', _.join(_.map(list, 'eventType'), '_'))
                .attr('x', parseInt(left) - 1)
                .attr('y', this.barPadding - 2)
                .attr('width', 3)
                .attr('height', this.barWidth + 4);
            lineWarp.append('line')
                .attr('class', `event_dash ${list.length > 1 ? 'dash_line' : ''}`)
                .attr('id', `line_id_${idx}_${left}`)
                .attr('data-idxs', _.join(_.map(list, 'idx'), '_'))
                .attr('data-types', _.join(_.map(list, 'eventType'), '_'))
                .attr('stroke', this.getEventLineColor(list))
                .attr('x1', left)
                .attr('x2', left)
                .attr('y1', this.barPadding)
                .attr('y2', this.barPadding + this.barWidth);
            lineWarp.on('mouseenter', function(this: any) {
                    d3.select(this).select('.event_dash_rect')
                        .attr('width', 8)
                        .attr('x', parseInt(left) - 7)
                })
                .on('mouseleave', function(this: any) {
                    d3.select(this).select('.event_dash_rect')
                        .attr('width', 4)
                        .attr('x', parseInt(left) - 2)
                });
        });
    }
    // 绘制各线程对应trace的处理时间段
    drawTraceLine(threadWarp, traceList, threadName) {
        const _this = this;
        const traceWarp = threadWarp.selectAll('.trace_line_warp').data(traceList).enter().append('g').attr('class', 'trace_line_warp');
        traceWarp.append('line')
            .attr('class', 'trace_line')
            .attr('x1', d => this.xScale(new Date(d.startTime)))
            .attr('x2', d => this.xScale(new Date(d.endTime)))
            .attr('y1', this.barWidth + this.barPadding * 2 - 4)
            .attr('y2', this.barWidth + this.barPadding * 2 - 4);

        traceWarp.append('rect')
            .attr('class', 'trace_line_rect')
            .attr('data-name', threadName)
            .attr('x', d => this.xScale(new Date(d.startTime)))
            .attr('y', this.barWidth + this.barPadding * 2 - 6)
            .attr('width', d => this.xScale(new Date(d.endTime)) - this.xScale(new Date(d.startTime)))
            .attr('height', 4)
            .on('click', function(this: any, e, d) {
                const threadName = d3.select(this).attr('data-name');
                const traceEvent = {
                    type: 'trace',
                    eventType: 'trace',
                    threadName: threadName,
                    ...d
                }
                _this.eventClick(traceEvent);
            });
    }
    // 支持在图上使用brush筛选
    addChartBrush() {
        d3.select('.chart_brush').style('display', 'block');
    }
    // 隐藏图上的brush筛选
    removeChartBrush() {
        d3.select('.chart_brush').style('display', 'none');
    }
    // 图上进行brush筛选之后，将对应的时间范围更新到底部的xAxisBrush上
    setxAxisBrushByChartBrush() {
        let s = this.xScale.domain();
        //把当前的domain通过x2Scale转化为range的数字数组(即为brush的位置信息)
        let d = s.map(item => {
            return this.xScale2(item)
        })
        //通过brush.move方法动态修改brush的大小与位置
        this.xAxisBrush.move(this.xAxisWarp, d);
    }
    // 图上进行多次brush筛选之后，通过reset操作可以将上一次筛选区间还远，直至筛选交互还原到初始状态
    resetChartBrush() {
        if (this.chartBrushTimeRange.length > 1) {
            let tiemRange = this.chartBrushTimeRange[this.chartBrushTimeRange.length - 2];
            this.chartBrushTimeRange.pop();
            this.updateChart(tiemRange);
        } else if (this.chartBrushTimeRange.length === 1) {
            this.chartBrushTimeRange = [];
            this.updateChart(this.timeRange);
        }
        this.setxAxisBrushByChartBrush();
    }
    // 监听事件的click事件
    private changeEventRectColor(id, active) {
        let eventType = d3.select(`#${id}`).attr('data-type');
        let sevent = _.find(eventList, {value: eventType});
        sevent && d3.select(`#${id}`).select('.event_rect').attr('fill', active ? sevent.activeColor : sevent.fillColor);
    }
    addEventClickListener() {
        const _this = this;
        d3.selectAll('.event_warp').on('click', (e) => {
            this.handleEventRectClick(e, true);
        });
        d3.selectAll('.event_dash_warp').on('click', function(e) {
            let id = d3.select(this).select('.event_dash').attr('id');
            let idxs = d3.select(this).select('.event_dash').attr('data-idxs').split('_');
            let firstIdx = id.split('_')[2];
            if (idxs.length > 1) {
                _this.showDashEvent(firstIdx, idxs);
                let { layerX, layerY } = e;
                _this.showEventTooltip(layerX, layerY);
            } else {
                const evt: any = _this.data[firstIdx].eventList[idxs[0]];
                evt.threadName = _this.data[firstIdx].name;
                if (_this.activeEventDomId) {
                    _this.changeEventRectColor(_this.activeEventDomId, false);
                }
                _this.eventClick(evt);
            }
        });
        d3.selectAll('.start_icon').on('click', function(e) {
            let tid = d3.select(this).attr('data-tid');
            let timestamp = d3.select(this).attr('data-timestamp');
            _this.handleStarClick(tid, timestamp);
        });
    }
    // 点击虚线时，显示tooltip，显示对应聚合的事件列表。点击对应事件展示事件详情
    showDashEvent(firstIdx, idxs) {
        let dom = '<ul class="tooltip_events">';
        _.forEach(this.data[firstIdx].eventList, (event, idx) => {
            if (idxs.indexOf(idx + '') > -1) {
                let sevent = _.find(eventList, {value: event.eventType});
                dom += `<li id="${`id_${firstIdx}_${idx}`}"><span style="color: ${sevent?.color}">${event.eventType}</span></li>`
            }
        });
        dom += '</ul>';
        d3.select('#tooltip_warp').html(dom);

        document.addEventListener('click', this.documentClickToHideTooltip, true);
        // 增加事件监听
        const _this = this;
        d3.selectAll('.tooltip_events li').on('click', function(e) {
            _this.handleEventRectClick(e);
        });
    }
    handleEventRectClick(e: any, showActive = false) {
        let id = e.currentTarget.id;
        let temp = id.split('_');
        const evt: any = this.data[temp[1]].eventList[temp[2]];
        evt.threadName = this.data[temp[1]].name;
        evt.tid = this.data[temp[1]].tid;
        this.eventClick(evt);
        if (showActive) {
            if (this.activeEventDomId) {
                this.changeEventRectColor(this.activeEventDomId, false);
            }
            this.activeEventDomId = id;
            this.changeEventRectColor(this.activeEventDomId, true);
        } else {
            if (this.activeEventDomId) {
                this.changeEventRectColor(this.activeEventDomId, false);
            }
            this.activeEventDomId = '';
        }
    }
    showEventTooltip(x: number, y: number) {
        d3.select('#tooltip_warp')
            .attr('class', 'camera_tooltip show_toolip')
            .attr('style', `left: ${x + 10}px;top: ${y - 10}px`);
    }
    hideEventTooltip() {
        d3.select('#tooltip_warp').attr('class', 'camera_tooltip');
    }
    documentClickToHideTooltip(e) {
        let targetArea: any = document.getElementById("tooltip_warp");   // 设置目标区域
        if (targetArea && !targetArea.contains(e.target)) {
            targetArea.setAttribute('style', 'display: none');
            document.removeEventListener('click', this.documentClickToHideTooltip);
        }
    }
    // 添加对应请求开始跟结束的时间
    addRequestLine(time: Date, type: ILineType = 'requestTime', timeType: 0|1) {
        let timeStamp = new Date(time).getTime();
        this.requestTimes.push(time);
        let position1 = this.xScale(time);
        this.drawChartWarpLine(position1, timeStamp, type, timeType);
        let position2 = this.xScale2(time);
        this.drawBottomxAxisLine(position2, timeStamp, type);
    }
    drawLineTime(timeList: ILineTime[]) {
        if (timeList.length > 2) {
            let requestSX = this.xScale(timeList[0].time);
            let traceSX = this.xScale(timeList[2].time);
            let requestEX = this.xScale(timeList[1].time);
            let traceEX = this.xScale(timeList[3].time);
            // console.log(requestSX, traceSX, requestEX, traceEX);

            if (Math.abs(requestSX - traceSX) < this.requestTraceXDiff) {
                this.addRequestLine(timeList[0].time, 'rt', 0);
            } else {
                this.addRequestLine(timeList[0].time, 'requestTime', 0);
                this.addRequestLine(timeList[2].time, 'trace', 0);
            }
            if (Math.abs(requestEX - traceEX) < this.requestTraceXDiff) {
                this.addRequestLine(timeList[1].time, 'rt', 1);
            } else {
                this.addRequestLine(timeList[1].time, 'requestTime', 1);
                this.addRequestLine(timeList[3].time, 'trace', 1);
            }
        } else {
            this.addRequestLine(timeList[0].time, 'requestTime', 0);
            this.addRequestLine(timeList[1].time, 'requestTime', 1);
        }
    }
    // 点击总览的上的事件，图上对应请求区间内的事件rect增加动画效果shining
    shiningEvent(type) {
        let requestTimeRange = [this.lineTimeList[0].time, this.lineTimeList[1].time].map(value => new Date(value).getTime());
        let traceTids = _.chain(this.data).filter(item => item.traceList.length > 0).map('tid').value();
        // 没有开启trace分析时，将trace的线程滚动到指定视窗内
        (document.getElementById(`thread_warp_${traceTids[0]}`) as any).scrollIntoView({ behavior: 'smooth' });
        if (type === 'lock') {
            // let traceTids = _.chain(this.data).filter(item => item.traceList.length > 0).map('tid').value();
            d3.selectAll('.javalock_rect').attr('class', (d: IJavaLock | any) => {
                if (traceTids.indexOf(d.threadTid) > -1) {
                    if (containTime(requestTimeRange, d.startTime, d.endTime)) {
                        return `javalock_rect lock_shining`; 
                    } else {
                        return 'javalock_rect';
                    }
                } else {
                    return 'javalock_rect';
                }
            });
            setTimeout(() => {
                d3.selectAll('.lock_shining').attr('class', 'javalock_rect');
            }, 3000);
        } else {
            _.forEach(this.data, (item, idx) => {
                if (item.traceList.length > 0) {
                    _.forEach(item.eventList, (evt, idx2) => {
                        if (containTime(requestTimeRange, evt.startTime, evt.endTime)) {
                            if (evt.type === type) {
                                // console.log(evt);
                                // let sevent = _.find(eventList, {type: type}) as IEvent;
                                // d3.select(`#id_${idx}_${idx2}`).select('.event_rect').attr('fill', sevent.activeColor);
                                d3.select(`#id_${idx}_${idx2}`).select('.event_rect').attr('class', `event_rect ${type}_shining`);
                                // d3.select(`#line_id_${idx}_${idx2}`).attr('fill', sevent.activeColor);
                                setTimeout(() => {
                                    d3.select(`#id_${idx}_${idx2}`).select('.event_rect').attr('class', 'event_rect');
                                }, 3000);
                            }
                        }
                    })
                }
            });
        }
    }

    // 判断当前绘制的时间线是 关键时刻（cruxTime） | 请求时间（requestTime）| trace | rt (requestTime、trace)
    private getLineColorByType(type: ILineType) {
        switch(type) {
            case 'cruxTime':
                return '#e10700';
            case 'requestTime':
                return '#f4bd80';
            case 'trace':
                return '#3d3d3d';
            case 'rt':
                return '#b96100';
            default:
                return '#dcdcdc';
        }
    }
    // 删除关键时刻
    deleteCruxTime(time) {
        console.log(time);
        Modal.confirm({
            title: '删除关键时刻',
            icon: null,
            content: '确定删除该关键时刻？',
            onOk() {
                d3.select(`#top_line_${time}`).remove();
                d3.select(`#line_warp_${time}`).remove();
                d3.select(`#xAxis_line_${time}`).remove();
            }
        });
    }
    // 绘制主体线程图上的请求Line和关键时刻Line
    drawChartWarpLine(x: number, timeStamp: number, type: ILineType, timeType = 0) {
        let _this = this;
        const topLineWarp = d3.select('.top_time_warp').append('g')
            .attr('id', `top_line_${timeStamp}`)
            .attr('class', 'chart_warp_line_warp')
            .attr('data-time', timeStamp);
        const lineWarp = d3.select('#chart_warp').append('g')
            .attr('id', `line_warp_${timeStamp}`)
            .attr('class', 'chart_warp_line_warp')
            .attr('data-time', timeStamp);
        topLineWarp.append('line')
            .attr('class', type === 'cruxTime' ? `chart_warp_line dash_line` : 'chart_warp_line')
            .attr('x1', x)
            .attr('x2', x)
            .attr('y1', 0)
            .attr('y2', this.topTimeheight)
            .attr('stroke', () => this.getLineColorByType(type));
        lineWarp.append('line')
            .attr('class', type === 'cruxTime' ? `chart_warp_line dash_line` : 'chart_warp_line')
            .attr('x1', x)
            .attr('x2', x)
            .attr('y1', 0)
            .attr('y2', this.chartWarpHeight)
            .attr('stroke', () => this.getLineColorByType(type));
        // TODO 如果绘制关键时刻line，额外绘制一个rect用于双击删除事件
        if (type === 'cruxTime') {
            topLineWarp.append('rect')
                .attr('class', 'chart_warp_line_rect')
                .attr('x', x - 2)
                .attr('y', 0)
                .attr('width', 4)
                .attr('height', this.topTimeheight)
                .attr('data-time', timeStamp)
                .on('click', function(e) {
                    e.stopPropagation();
                    let time = d3.select(this).attr('data-time');
                    _this.deleteCruxTime(time);
                });
            lineWarp.append('rect')
                .attr('class', 'chart_warp_line_rect')
                .attr('x', x - 2)
                .attr('y', 0)
                .attr('width', 4)
                .attr('height', this.chartWarpHeight)
                .attr('data-time', timeStamp)
                .on('click', function(e) {
                    e.stopPropagation();
                    let time = d3.select(this).attr('data-time');
                    _this.deleteCruxTime(time);
                });
        } else {
            topLineWarp.append('text')
                .attr('class', 'chart_warp_line_text')
                .attr('data-type', timeType)
                .text(() => {
                    let typeText = type === 'requestTime' ? 'IO' : (type === 'trace' ? 'Trace' : 'IO/Trace');
                    let timeTypeText = timeType === 0 ? 'Start' : 'End';
                    return `${typeText} ${timeTypeText}`;
                })
                .attr('fill', () => this.getLineColorByType(type))
                .attr('x', function(this: any) {
                    if (timeType === 0) {
                        let textlen: number = this.getComputedTextLength();
                        return x - 3 - textlen;
                    } else {
                        return x + 3;
                    }
                })
                .attr('y', 10);
            topLineWarp.append('text')
                .attr('class', 'chart_warp_line_text')
                .attr('data-type', timeType)
                .text(timeFormat(new Date(timeStamp)))
                .attr('fill', () => this.getLineColorByType(type))
                .attr('x', function(this: any) {
                    if (timeType === 0) {
                        let textlen: number = this.getComputedTextLength();
                        return x - 3 - textlen;
                    } else {
                        return x + 3;
                    }
                })
                .attr('y', 20);
        }
    }

    // 添加关键时间事件
    supportAddLine() {
        d3.select('.top_time_warp')
            .attr('cursor', 'pointer')
            .on('click', (e) => {
                let { offsetX } = e;
                console.log(this);
                let time = this.xScale.invert(offsetX);
                let timeStamp = new Date(time).getTime();
                this.drawChartWarpLine(offsetX, timeStamp, 'cruxTime');
                let x2 = this.xScale2(new Date(time));
                this.drawBottomxAxisLine(x2, timeStamp, 'cruxTime');
                // if (this.cruxTimes.indexOf(timeStamp) === -1) {
                //     this.cruxTimes.push(timeStamp);
                //     this.drawChartWarpLine(offsetX, timeStamp, 'cruxTime');
                //     let x2 = this.xScale2(new Date(time));
                //     this.drawBottomxAxisLine(x2, 'cruxTime');
                // } else {
                //     message.warn('当前关键时刻已添加')
                // }
            });
    }
    // 移除添加关键时间事件
    removeAddLine() {
        d3.select('.top_time_warp').attr('cursor', 'default').on('click', null);
    }

    showRunQ() {
        this.showRunQFlag = true;
        d3.selectAll('.event_runq').style('visibility', 'visible');
    }
    hideRunQ() {
        this.showRunQFlag = false;
        d3.selectAll('.event_runq').style('visibility', 'hidden');
    }
    // 显示Java lock事件
    showJavaLock() {
        this.showJavaLockFlag = true;
        d3.selectAll('.javalock_rect').style('visibility', 'visible');
        if (this.filters.eventList.length > 0) {
            this.reprintJavaLockByFilter(this.filters.eventList);
        }
    }
    // 隐藏Java lock事件
    hideJavaLock() {
        this.showJavaLockFlag = false;
        d3.selectAll('.javalock_rect').style('visibility', 'hidden');
    }
    // 绘制日志输出所在的标志位
    // 显示日志标记
    showLog() {
        this.showLogFlag = true;
        d3.selectAll('.log_icon').style('visibility', 'visible');
    }
    // 隐藏日志标记
    hideLog() { 
        this.showLogFlag = false;
        d3.selectAll('.log_icon').style('visibility', 'hidden');
    }

    containTime = (timeRange, startTime, endTime) => {
        return endTime > timeRange[0] && startTime < timeRange[1];
    }
    timeHandle = (startTime, endTime, timeNum, timeRange) => {
        let time = timeNum;
        // 若事件的开始事件小于当前请求的时间区间，需要根据当前时间的开始值进行截取，同时重新计算事件耗时
        let stime = startTime;
        if (stime < timeRange[0]) {
            time = time - (timeRange[0] - stime);
            stime = timeRange[0];
        }
        // 若事件的结束时间大于当前请求的时间区间，需要根据当前时间的结束值进行截取，同时重新计算事件耗时
        let etime = endTime;
        if (etime > timeRange[1]) {
            time = time - (etime - timeRange[1]);
            etime = timeRange[1];
        }
        return {stime, etime, time}
    }
    updateThread(timeDateRange: Date[]) {
        let _this = this;
        let startTimeStamp = new Date(timeDateRange[0]).getTime();
        let endTimeStamp = new Date(timeDateRange[1]).getTime();
        let timeRangeDiff = endTimeStamp - startTimeStamp;
        let timeRange = [startTimeStamp, endTimeStamp];
        _.forEach(this.data, (item: IThread, idx: number) => {
            const barWarp = d3.select(`#thread_warp_${item.tid} .bar_warp`);
            barWarp.html(null);
            
            let lineData: IEventTime[] = [];
            _.forEach(item.eventList, (event: IEventTime, idx2: number) => {
                let { startTime, endTime, time } = event;
                if (this.containTime(timeRange, startTime, endTime)) {
                    // console.log(event);
                    let { stime, etime, time: ftime } = this.timeHandle(startTime, endTime, time, timeRange);
                    const timeRate = ftime / timeRangeDiff;
                    event.left = this.xScale(new Date(stime)) as number;
                    if (timeRate && timeRate > this.timeThreshold) {
                        // let timeWidth = timeRate * (this.svgWidth - this.nameWidth - this.padding);
                        let right = this.xScale(new Date(event.endTime));
                        let timeWidth = right - event.left!;
                        let sevent: IEvent = _.find(eventList, {type: event.type}) as IEvent;
                        const eventWarp = barWarp.append('g')
                            .attr('id', `id_${idx}_${idx2}`)
                            .attr('data-type', event.eventType as string)
                            .attr('class', `event_warp ${event.eventType === 'runqLatency' ? 'event_runq' : ''}`);
                        eventWarp.append('rect')
                            .attr('class', 'event_rect')
                            .attr('width', timeWidth)
                            .attr('height', this.barWidth)
                            .attr('fill', sevent.fillColor)
                            .attr('x', event.left)
                            .attr('y', this.barPadding);
                        if (event.eventType === 'runqLatency') {
                            if (timeWidth! > 75) {
                                eventWarp.append('text').text(sevent.type)
                                    .attr('class', 'event_rect_left_text')
                                    .attr('stroke', sevent.color)
                                    .attr('x', event.left && event.left + 5)
                                    .attr('y', this.barPadding + this.barWidth / 2 + 5)
                            } 
                        } else {
                            if (timeRate > 0.05) {
                                eventWarp.append('text').text(sevent.type)
                                    .attr('class', 'event_rect_left_text')
                                    .attr('stroke', sevent.color)
                                    .attr('x', event.left && event.left + 5)
                                    .attr('y', this.barPadding + this.barWidth / 2 + 5)
                            }
                        }
                    } else {
                        lineData.push(event);
                    }
                }
            });
            // 绘制虚线事件
            this.drawThreadEventDashLine(idx, lineData, barWarp);
        });
        if (this.filters.threadList.length > 0 || this.filters.eventList.length > 0) {
            this.reprintByFilter(this.filters);
        }

        // event warp click 监听事件点击事件
        this.addEventClickListener();

        !this.showRunQFlag && this.hideRunQ();
        /**
         * 更新javaLock事件占比
         * 时间筛选后会出现event startTime 小于筛选的起始时间，endTime 大于筛选的终止时间，需要截取event的对应时间段 重新计算
         */
        d3.selectAll('.javalock_rect').each(function(d: any) {
            let {startTime, endTime, time} = d;
            if (_this.containTime(timeRange, startTime, endTime)) {
                let { stime, etime, time: ftime } = _this.timeHandle(startTime, endTime, time, timeRange);
                // let timediff = endTime - startTime;
                if (ftime > 0) {
                    let x = _this.xScale(new Date(stime));
                    let left = x > _this.nameWidth ? x : _this.nameWidth;
                    let right = _this.xScale(etime);
                    d3.select(this)
                        .style('display', 'block')
                        .style('visibility', _this.showJavaLockFlag ? 'visible' : 'hidden')
                        .attr('width', right - left)
                        .attr('x', x > _this.nameWidth ? x : _this.nameWidth); 
                } else {
                    d3.select(this).style('display', 'none');
                }
            } else {
                d3.select(this).style('display', 'none');
            }
        });
    }
    // 进行brush筛选时图表相关的更新操作
    updateChart(timeRange: Date[]) {
        let _this = this;
        this.xScale.domain(timeRange);
        d3.select('.xaxis_line').call(this.xAxis);
        if (this.isPrint) {
            // TODO 当requestTime跟traceTime合并时，筛选后两个x坐标相对分离时要重新绘制成两条线？
            // 请求线跟关键时刻的update
            d3.selectAll('.chart_warp_line_warp').each(function(d,i) {
                let timeStamp = d3.select(this).attr('data-time');
                let x = _this.xScale(new Date(parseInt(timeStamp)));
                d3.select(this).attr('opacity', x > _this.nameWidth ? 1 : 0);
                d3.select(this).select('.chart_warp_line').attr('x1', x).attr('x2', x);
                d3.select(this).selectAll('.chart_warp_line_text').attr('x', function(this: any) {
                    let type = d3.select(this).attr('data-type');
                    if (parseInt(type) === 0) {
                        let textlen: number = this.getComputedTextLength();
                        return x - 3 - textlen;
                    } else {
                        return x + 3;
                    }
                });
            });
            // 更新log事件坐标
            d3.selectAll('.log_icon').each(function(log: any) {
                let position = _this.xScale(new Date(log.startTime));
                d3.select(this)
                    .attr('transform', `translate(${position}, ${_this.barPadding + _this.barWidth})`)
                    .attr('style', () => {
                        if (position > _this.nameWidth) {
                            return _this.showLogFlag ? 'visibility: visible' : 'visibility: hidden';
                        } else {
                            return 'display: none';
                        }
                    });
            }); 
            d3.selectAll('.start_icon').each(function() {
                let timestamp = d3.select(this).attr('data-timestamp');
                let position = _this.xScale(new Date(parseInt(timestamp)));
                d3.select(this)
                    .attr('transform', `translate(${position}, ${_this.barPadding - 8})`)
                    .attr('style', () => {
                        if (position > _this.nameWidth && position < _this.svgWidth) {
                            return 'display: block';
                        } else {
                            return 'display: none';
                        }
                    });
            }); 
            // 更新trace
            d3.selectAll('.trace_line_warp').each(function(d: any) {
                let sx = _this.xScale(new Date(d.startTime)), ex = _this.xScale(new Date(d.endTime));
                if (sx > _this.nameWidth && ex < _this.svgWidth) {
                    d3.select(this).style('display', 'block');
                    d3.select(this).select('.trace_line').attr('x1', sx).attr('x2', ex);
                    d3.select(this).select('.trace_line_rect').attr('x', sx).attr('width', ex - sx);
                } else {
                    d3.select(this).style('display', 'none');
                }
            });
            // 更新各线程事件图
            this.updateThread(timeRange);
        }
    }
    // 根据勾选的java lock事件进行筛选
    reprintJavaLockByFilter(selectEvents) {
        if (_.some(this.lockEventTypeList, type => selectEvents.indexOf(type) > -1)) {
            d3.selectAll('.javalock_rect').each(function() {
                let eventType = d3.select(this).attr('data-type');
                d3.select(this).attr('style', `visibility: ${selectEvents.indexOf(eventType) > -1 ? 'visible' : 'hidden'}`);
            });        
        }
    }
    // rect跟line的筛选事件处理
    reprintEventRect(filters, eventD3Dom) {
        let eventType = eventD3Dom.attr('data-type');
        if (eventType.indexOf('file') > -1 && filters.eventList.indexOf(eventType) > -1) {
            if (filters.fileList.length > 0) {
                let ids = eventD3Dom.attr('id').split('_');
                let event = this.data[ids[1]].eventList[ids[2]];
                eventD3Dom.attr('style', `visibility: ${_.some(filters.fileList, file => event.info.address.indexOf(file) > -1) ? 'visible' : 'hidden'}`);
            } else {
                eventD3Dom.attr('style', `visibility: visible`);
            }
        } else {
            eventD3Dom.attr('style', `visibility: ${filters.eventList.indexOf(eventType) > -1 ? 'visible' : 'hidden'}`);
        }
    }
    reprintEventLine(filters, eventD3Dom) {
        let eventTypes = eventD3Dom.attr('data-types');
        let eventTypeList = eventTypes.split('_');
        let typeExist = _.some(eventTypeList, type => filters.eventList.indexOf(type) > -1);
        if (typeExist && _.some(eventTypeList, type => type.indexOf('file') > -1)) {
            if (filters.fileList.length > 0) {
                // TODO file场景待验证
                let ids = eventD3Dom.attr('id').split('_');
                let idxs = eventD3Dom.attr('data-idxs').split('_');
                let visible = 'hidden';
                _.forEach(eventTypeList, (type, idx) => {
                    if (type.indexOf('file') > -1) {
                        let event = this.data[ids[2]].eventList[idxs[idx]];
                        if (_.some(filters.fileList, file => event.info.address.indexOf(file) > -1)) {
                            visible = 'visible';
                        }
                    } 
                });
                eventD3Dom.attr('style', `visibility: ${visible}`);
            } else {
                eventD3Dom.attr('style', `visibility: visible`);
            }
        } else {
            eventD3Dom.attr('style', `visibility: ${typeExist ? 'visible' : 'hidden'}`);
        }
    }
    /**
     * 根据线程筛选里面的筛选条件重新进行图表绘制，显示或隐藏相关筛选条件筛选后的线程事件
     */
    reprintByFilter(filter = this.filters) {
        let _this = this;
        this.filters = filter;

        let data: IThread[] = [];
        if (this.filters.threadList.length === 0) {
            data = this.data;
        } else {
            data = _.filter(this.data, (item: IThread) => this.filters.threadList.indexOf(item.tid) > -1);
        }
        this.chartWarpHeight = this.threadHeight * data.length;
        this.svg.attr('height', this.chartWarpHeight);
        // d3.select(`#${this.svgId}`).attr('height', this.chartWarpHeight);
        // 请求线跟关键时刻的高度重新update
        d3.selectAll('.chart_warp_line').each(function(d,i) {
            d3.select(this).attr('y2', _this.chartWarpHeight);
        });

        // 过滤线程名称
        const threadTids = _.map(data, 'tid');
        d3.selectAll('#chart_warp .thread_warp').each(function() {
            let tid = d3.select(this).attr('data-tid');
            d3.select(this).attr('style', `display: ${threadTids.indexOf(parseInt(tid)) > -1 ? 'block' : 'none'}`);
        });
        _.forEach(threadTids, (tid, idx) => {
            // 重新计算线程的Y轴坐标
            d3.select(`#thread_warp_${tid}`).attr('transform', `translate(0, ${this.threadHeight * idx})` );
            if (this.filters.eventList.length > 0) {
                // 根据勾选的时间类型过滤显示的事件
                d3.selectAll(`#thread_warp_${tid} .bar_warp .event_warp`).each(function() {
                    _this.reprintEventRect(_this.filters, d3.select(this));
                });
                d3.selectAll(`#thread_warp_${tid} .bar_warp .event_dash`).each(function() {
                    _this.reprintEventLine(_this.filters, d3.select(this));
                });
                if (this.showJavaLockFlag) {
                    this.reprintJavaLockByFilter(this.filters.eventList);
                }
            }
            // 日志搜索筛选
            if (this.filters.logList.length > 0 && this.showLogFlag) {
                d3.selectAll('.log_icon').each(function(log: any) {
                    d3.select(this).attr('style', `visibility: ${_.some(_this.filters.logList, v => log.log.indexOf(v) > -1) ? 'visible' : 'hidden'}`);
                }); 
            }
        });
    }

    /**
     * 是否开启trace分析，点击trace分析后，只查看持有相关trace的线程跟部分gc和vm线程的白名单  
     */
    startTrace(showTrace) {
        let _this = this;
        const whiteThread: string[] = [];
        let data: any = [];
        this.showTraceFlag = showTrace;
        if (showTrace) {
            // data = _.filter(this.data, item => item.traceList.length > 0 || whiteThread.indexOf(item.name) > -1 || _.some(item.eventList, opt => opt.active));
            data = _.filter(this.data, item => item.traceList.length > 0 || whiteThread.indexOf(item.name) > -1 || item.active || (item.traceStartTimestamp && item.traceEndTimestamp));
            if (this.filters.threadList.length > 0) {
                data = data.concat(_.filter(this.data, (item: IThread) => this.filters.threadList.indexOf(item.tid) > -1));
            }
            if (data.length === 0) {
                data = _.concat([], this.data);
                this.parentRef && this.parentRef.closeTraceAnaliysis();
            }
        } else {
            data = _.concat([], this.data);
            if (this.filters.threadList.length > 0) {
                data = _.filter(this.data, (item: IThread) => this.filters.threadList.indexOf(item.tid) > -1);
            }
        }

        this.chartWarpHeight = this.threadHeight * data.length;
        this.svg.attr('height', this.chartWarpHeight);
        // 请求线跟关键时刻的高度重新update
        d3.selectAll('.chart_warp_line').each(function(d,i) {
            d3.select(this).attr('y2', _this.chartWarpHeight);
        });
        // 过滤线程名称
        const threadTids = _.map(data, 'tid');
        d3.selectAll('#chart_warp .thread_warp').each(function() {
            let tid = d3.select(this).attr('data-tid');
            d3.select(this).attr('style', `display: ${threadTids.indexOf(parseInt(tid)) > -1 ? 'block' : 'none'}`);
        });
        _.forEach(threadTids, (tid, idx) => {
            // 重新计算线程的Y轴坐标
            d3.select(`#thread_warp_${tid}`).attr('transform', `translate(0, ${this.threadHeight * idx})` );
        });
    }
    closeShowTrace() {
        this.showTraceFlag = false;
    }
    
    /**
     * 底部brush 时间轴相关事件、绘制方法
     */
    // 绘制底部时间筛选区域的标志位 - Line 包括(请求开始结束时间、关键时刻)
    drawBottomxAxisLine(x: number, time: number, type: ILineType) {
        d3.select('#bottom_xaxis_svg #xAxis_warp').append('line')
            .attr('id', `xAxis_line_${time}`)
            .attr('class', 'xAxis_warp_line')
            .attr('x1', x)
            .attr('x2', x)
            .attr('y1', 1)
            .attr('y2', this.bottomxAxisHeight)
            .attr('stroke', () => this.getLineColorByType(type))
    }
    // 绘制时间筛选两端的时间text-tspan
    private drawXAxisWEText(xAxisWarp: any, type: 'start' | 'end', time: string, x: number) {
        let textWarp: any;
        if (type === 'start') {
            textWarp = xAxisWarp.append('text').attr('class', 'handle-text handle-text-w');
        } else {
            textWarp = xAxisWarp.append('text').attr('class', 'handle-text handle-text-e');
        }
        textWarp.text(time)
            .attr('opacity', 0)
            .attr('x', function(this: any) {
                if (type === 'start') {
                    let textlen: number = this.getComputedTextLength();
                    return x - 3 - textlen;
                } else {
                    return x + 3;
                }
            })
            .attr('y', 18);
    }
    // 时间筛选框筛选时更新text-tspan中的时间值
    private updateXAxisWEText(textWarp: any, type: 'start' | 'end', time: string, x: number) {
        textWarp.text(time)
            .attr('x', function(this: any) {
                if (type === 'start') {
                    let textlen: number = this.getComputedTextLength();
                    return x - 3 - textlen;
                } else {
                    return x + 3;
                }
            });
    }
    // 绘制时间筛选的区域
    drawxAxisBrush() {
        const _this = this;
        const bottomXAxisWarp = d3.select('#bottom_xaxis_svg');
        bottomXAxisWarp.html(null);
        // 返回起始和终止两个时间构造线性时间比例尺
        // 定义xAxis坐标轴的比例尺
        let xAxisWarp = bottomXAxisWarp.append('g').attr('id', 'xAxis_warp').attr('transform', `translate(0, 0)`);
        this.xAxisWarp = xAxisWarp;
        const brushxAxis = d3.scaleTime().domain(this.xScale.domain()).range([this.nameWidth, this.svgWidth]);
        this.xScale2 = brushxAxis;
        const xAxis = d3.axisBottom(brushxAxis)
            // .ticks(d3.timeMonth.every(2))
            .tickFormat(d => timeFormat(d as Date));

        // brush事件监听
        // brush开始事件 - 显示左右两边的具体文案
        const xAxisBrushStart = () => {
            // console.log('start');
            d3.selectAll('.handle-text').attr('opacity', 1);
        }
        const xAxisBrushing = function(e: any) {
            const selection = e.selection;
            // x.domain(selection.map(x2.invert, x2));
            // 获取brush筛选的时间区间
            let tiemRange = selection.map(brushxAxis.invert, brushxAxis);
            _this.updateXAxisWEText(d3.select('.handle-text-w'), 'start', timeFormat(tiemRange[0]), selection[0]);
            _this.updateXAxisWEText(d3.select('.handle-text-e'), 'end',timeFormat(tiemRange[1]), selection[1]);
        }
        const xAxisBrushed = (e: any) => {
            // console.log('end');
            const { selection } = e;
            d3.selectAll('.handle-text').attr('opacity', 0);
            let tiemRange = selection.map(brushxAxis.invert, brushxAxis);
            this.updateChart(tiemRange);
        }
        // 构造xAxis上的brush对象
        this.xAxisBrush = d3.brushX()
            .extent([[this.nameWidth, 0], [this.svgWidth - this.nameWidth, this.bottomxAxisHeight]])
            .on('start', xAxisBrushStart)
            .on('brush', xAxisBrushing)
            .on('end', xAxisBrushed);

        xAxisWarp.append('rect')
            .attr('class', 'xAxis_rect')
            .attr('width', this.svgWidth - this.nameWidth)
            .attr('height', this.bottomxAxisHeight)
            .attr('x', this.nameWidth)
            .attr('y', 0)
        xAxisWarp.call(this.xAxisBrush).call(this.xAxisBrush.move, brushxAxis.range());

        xAxisWarp.append('g')
            .attr('class', 'xaxis-x')
            .attr('transform', `translate(0, ${this.bottomxAxisHeight})`)
            .call(xAxis);
        let startTime: string = timeFormat(this.timeRange[0]);
        this.drawXAxisWEText(xAxisWarp, 'start', startTime, this.nameWidth);
        let endTime: string = timeFormat(this.timeRange[1]);
        this.drawXAxisWEText(xAxisWarp, 'end', endTime, this.svgWidth);
    }


    // 绘制全局移动的tooltip line
    drawTooltipLine(chartWarp: any) {
        let lineWarp = chartWarp.append('g').attr('class', 'tooltip_line_warp').attr('opacity', 0);
        lineWarp.append('line')
            .attr('class', 'tooltip_line')
            .attr('x1', 0)
            .attr('x2', 0)
            .attr('y1', 0)
            .attr('y2', this.chartWarpHeight);
    }   
    drawTopTooltip(chartWarp: any) {
        let lineWarp = chartWarp.append('g').attr('class', 'tooltip_line_warp').attr('opacity', 0);
        const FirTime = timeFormat(new Date(this.xScale.invert(this.nameWidth)));
        lineWarp.append('rect')
            .attr('class', 'tooltip_line_rect')
            .attr('width', 12)
            .attr('height', this.topTimeheight)
            .attr('x', -6)
            .attr('y', 0);
        lineWarp.append('text')
            .text(FirTime)
            .attr('class', 'tooltip_line_text')
            .attr('x', function(this: any) {
                let textlen: number = this.getComputedTextLength();
                return - textlen / 2;
            })
            .attr('y', 8);
        lineWarp.append('line')
            .attr('class', 'tooltip_line')
            .attr('x1', 0)
            .attr('x2', 0)
            .attr('y1', 0)
            .attr('y2', this.topTimeheight);
    }
    // tooltip - 鼠标mouseenter和mousemove的公共事件
    private tooltipLineAction(x: number) {
        d3.selectAll('.tooltip_line_warp')
            .attr('opacity', 1)
            .attr('transform', `translate(${x})` );
        let time = timeFormat(new Date(this.xScale.invert(x)));
        d3.select('.tooltip_line_text').text(time);
    }
    addTooltipMouse(timeWarp: any) {
        let _this = this;
        timeWarp
            .on('mouseenter', function(this: any, e: any) {
                let [x] = d3.pointer(e, this);
                _this.ifShowTooltipLine = true;
                _this.tooltipLineAction(x);
            })
            .on('mousemove', function(this: any, e: any) {
                let [x] = d3.pointer(e, this);
                _this.tooltipLineAction(x);
            })
            .on('mouseleave', function() {
                _this.ifShowTooltipLine = false;
                d3.selectAll('.tooltip_line_warp').attr('opacity', 0);
            });
    }
    drawTopTimeWarp() {
        const _this = this;
        const topTimeSvg = d3.select('#top_time_svg').style('height', this.topTimeheight);
        topTimeSvg.html(null);
        
        // 绘制顶部时间tooltip筛选区域
        const topTimeWarp = topTimeSvg.append('g').attr('class', 'top_time_warp').attr('transform', `translate(0, 0)`);
        topTimeWarp.append('rect')
            .attr('id', 'top_time_rect')
            .attr('width', this.svgWidth - this.nameWidth)
            .attr('height', this.topTimeheight)
            .attr('fill', '#dcdcdc51')
            .attr('x', this.nameWidth)
            .attr('y', 0);
        // 监听tooltip的mouse鼠标移动事件
        this.addTooltipMouse(topTimeWarp);
        // 绘制tooltip的虚线跟时间text
        this.drawTopTooltip(topTimeWarp);
    }

    
    // 改变图表大小之后重新绘制
    fullScreen() {
        console.log('changeSize now width: ',  this.svgWidth);
    }
    changeSize() {
        // TODO 改变布局大小时因为宽度变化，所有的事件节点都需要重新计算宽度跟相应的起始坐标，brush等也需要更新操作范围
        console.log('changeSize now width: ',  this.svgWidth);
        // const cameraSize = document.getElementById('camera')?.getBoundingClientRect();
        // this.svgWidth = cameraSize?.width as number;

        // d3.select('#top_time_rect').attr('width', this.svgWidth - this.nameWidth);
        // d3.selectAll('.thread_rect_warp').attr('width', this.svgWidth);

        // // 更新底部时间筛选宽度
        // d3.select('.xAxis_rect').attr('width', this.svgWidth - 40);
        // this.xScale2.range([20, this.svgWidth - 20]);
        // this.xAxisBrush.extent([[20, 0], [this.svgWidth - 20, this.bottomxAxisHeight]]);

        // 是重新绘制整个图还是一个个更新布局？
        this.draw();
    }
}
export default Camera;