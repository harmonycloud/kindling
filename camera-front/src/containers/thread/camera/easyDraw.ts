import * as d3 from 'd3';
import _ from 'lodash';
import { IEvent, IOption, IThread, IEventTime, IJavaLock, ILogEvent, ILineTime } from './types';
import { eventList as LEvent, darkEventList as DEvent, textHandle, timeFormat } from './util';
import { getStore } from '@/services/util';
import netReadPng from '@/assets/images/netread.png';
import netWritePng from '@/assets/images/netwrite.png';
import logPng from '@/assets/images/log.png';


const eventList = (getStore('theme') || 'light') === 'light' ? LEvent : DEvent;
class EasyCamera {
    parentRef: any;
    theme: 'light' | 'dark' = 'light';
    spanList: any[] = [];
    data: IThread[] = [];
    trace: any;
    lineTimeList: ILineTime[] = [];
    timeThreshold: number = 0.001;
    traceId: string;
    svgId: string;
    svg: any;
    spanSvg: any;
    svgWidth: number = 500;     // 绘制svg区域的宽度
    svgHeight: number = 300;
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
    subEventNumWidth: number = 20;
    subEventWidth: number = 40;
    subEvents: string[] = ['on', 'net', 'file'];
    lockEventTypeList: string[] = _.chain(eventList).filter(item => item.type === 'lock').map('value').value();
    groupLineData = {};
    
    timeRange: number[];
    timeRangeDiff: number;
    showRunQFlag: boolean = true;
    showLogFlag: boolean = true;       // 是否显示日志的标志位
    showJavaLockFlag: boolean = true;  // 是否显示lock事件的标志位
    activeEventDomId: string = '';      // 当前选中查看详情的event id => id_idx1_idx2
    timeDiffStep = 10; // 时间戳前后各扩大毫秒数间隔

    xScale: any;                
    xAxis: any;    
    xScale2: any;            

    isPrint: boolean = false;
    eventClick: (evt: any) => void;

    constructor (option: IOption) {
        this.parentRef = option.parentRef;
        this.theme = (getStore('theme') as any) || 'light';
        this.spanList = option.spanList || [];
        this.data = this.startTrace(option.data);
        this.trace = option.trace;
        this.lineTimeList = option.lineTimeList;
        this.traceId = option.traceId;
        this.svgId = option?.svgId || 'camera_svg';
        this.nameWidth = option?.nameWidth || 150; // 侧边轴上的名称宽度包括跟柱图的间隔20
        this.barWidth = 40; // 柱图的高度
        this.barPadding = option?.barPadding || 12; // 柱图的上下padding
        this.padding = option?.padding || 20;

        let lineTimes: any[] = _.chain(this.lineTimeList).map(item => new Date(item.time).getTime()).value();
        // this.timeRange = traceTimes.map((opt, idx) => idx === 0 ? new Date(opt).getTime() - this.timeDiffStep : new Date(opt).getTime() + this.timeDiffStep);
        this.timeRange = [_.min(lineTimes) - this.timeDiffStep, _.max(lineTimes) + this.timeDiffStep];
        this.timeRangeDiff = this.timeRange[1] - this.timeRange[0] + this.timeDiffStep * 2;

        this.eventClick = option.eventClick.bind(this);
    }

    /**
     * 简易视图默认开启trace分析，只查看持有相关trace的线程跟部分gc和vm线程的白名单  
     */
    startTrace(data: any) {
        const whiteThread: string[] = [];
        let finalData: IThread[] = [];
        finalData = _.filter(data, item => item.traceList.length > 0 || whiteThread.indexOf(item.name) > -1 || item.traceStartTimestamp || item.traceEndTimestamp);
        if (finalData.length === 0) {
            finalData = _.concat([], data);
        }
        // console.log('finalData', finalData);
        return finalData;
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
        this.svgHeight = cameraSize?.height as number - 30;
        
        const chartWarp = this.svg.append('g').attr('id', 'chart_warp').attr('transform', `translate(0, 0)` );

        this.xScale = d3.scaleLinear()
            .domain(this.timeRange)
            .range([this.nameWidth, this.svgWidth]);
        this.xAxis = d3.axisBottom(this.xScale)
            .tickFormat(d => timeFormat(d as Date));

        
        const barWarpWidth = this.svgWidth - this.nameWidth;
        this.barWarpWidth = barWarpWidth;

        this.drawSpanTree();

        // 绘制主体柱图区域
        let translateY = 0;
        _.forEach(this.data, (item, idx: number) => {
            let threadWarp = chartWarp.append('g')
                .attr('id', `thread_warp_${item.tid}`)
                .attr('data-tid', item.tid)
                .attr('class', 'thread_warp')
                .attr('height', threadHeight)
                .attr('transform', `translate(0, ${translateY})` );
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
            } else {
                this.drawThreadName(threadWarp, item.name, item, this.barPadding + this.barWidth / 2 + 5);
            }

            let barWarp = threadWarp.append('g').attr('class', 'bar_warp');
            _.forEach(item.eventList, (event, idx2: number) => {
                event.idx = idx2;
            });
            let lineData: IEventTime[] = [];
            _.forEach(item.eventList, (event, idx2: number) => {
                let { startTime, endTime, time } = event;
                if (this.containTime(this.timeRange, startTime, endTime)) {
                    let { stime, etime, time: ftime } = this.timeHandle(startTime, endTime, time, this.timeRange);
                    const timeRate = ftime / this.timeRangeDiff;
                    event.left = this.xScale(stime) as number;
                    event.startTime = stime;
                    event.endTime = etime;
                    event.time = ftime;
                    let right = this.xScale(etime) as number;
                    if (timeRate && timeRate > this.timeThreshold) {
                        // 发现使用rate计算的event宽度存在缺陷，会出现绘制的事件不连续，通过定位startTime和endTime的坐标相减获得的宽度更符合实际场景
                        // let timeWidth = timeRate * (this.svgWidth - this.nameWidth - this.padding);
                        let timeWidth = right - event.left;
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
                        if (timeWidth > 50) {
                            eventWarp.append('text').text(sevent.alias)
                                .attr('class', 'event_rect_left_text')
                                .attr('stroke', sevent.color)
                                .attr('x', event.left && event.left + 5)
                                .attr('y', this.barPadding + this.barWidth / 2 + 5)
                        }
                    } else {
                        // lineData.push(event);
                        if (['file', 'net', 'on'].indexOf(event.type) > -1) {
                            lineData.push(event);
                        } else {
                            // file，net，on以外的事件类型只保留时间大于一毫秒的事件
                            if (event.time > 0.1) {
                                lineData.push(event);
                            }
                        }
                    }
                }
            });

            if (lineData.length > 0) {
                this.drawThreadEventDashLine(idx, lineData, barWarp);
                translateY += threadHeight * 3;
            } else {
                translateY += threadHeight;
            }

            // 日志Icon绘制
            threadWarp.selectAll('.log_icon')
                .data(item.logList)
                .enter()
                .append('image')
                .attr('class', 'log_icon')
                .attr('xlink:href', logPng)
                .attr('data-tid', item.tid)
                // .attr('fill', (log) => log.traceId === this.traceId ? '#1291A2' : '#E6F4F5')
                .attr('cursor', 'pointer')
                .attr('x', (log: ILogEvent) => this.xScale(log.startTime))
                .attr('y', 0)
                .attr('width', 15)
                .attr('height', 15)
                .style('visibility', 'hidden');
            // 更新log事件坐标
            d3.selectAll('.log_icon').each(function(log: any) {
                let position = _this.xScale(log.startTime);
                d3.select(this)
                    // .attr('transform', `translate(${position}, ${_this.barPadding + _this.barWidth})`)
                    .attr('x', position)
                    .attr('style', () => {
                        if (position > _this.nameWidth) {
                            return _this.showLogFlag ? 'visibility: visible' : 'visibility: hidden';
                        } else {
                            return 'display: none';
                        }
                    });
            }); 

            threadWarp.selectAll('.javalock_rect')
                .data(item.javaLockList)
                .enter()
                .append('rect')
                .attr('class', 'javalock_rect')
                .attr('data-thread', item.name)
                .attr('data-type', d => d.eventType)
                .attr('width', (lock: IJavaLock) => this.xScale(lock.endTime) - this.xScale(lock.startTime))
                .attr('height', this.barWidth + 8)
                .attr('cursor', 'pointer')
                // .attr('style', 'display: none')
                .style('visibility', 'hidden')
                .attr('x', (lock: IJavaLock) => this.xScale(lock.startTime))
                .attr('y', this.barPadding - 4);
            // 根据时间区间更新javalock的绘制与位置
            d3.selectAll('.javalock_rect').each(function(d: any) {
                let {startTime, endTime, time} = d;
                if (_this.containTime(_this.timeRange, startTime, endTime)) {
                    let { stime, etime, time: ftime } = _this.timeHandle(startTime, endTime, time, _this.timeRange);
                    if (ftime > 0) {
                        let left = _this.xScale(stime);
                        left = left > _this.nameWidth ? left : _this.nameWidth;
                        let right = _this.xScale(etime);
                        d3.select(this)
                            .style('display', 'block')
                            .style('visibility', _this.showJavaLockFlag ? 'visible' : 'hidden')
                            .attr('width', right - left)
                            .attr('x', left); 
                    } else {
                        d3.select(this).style('display', 'none');
                    }
                } else {
                    d3.select(this).style('display', 'none');
                }
            });

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
        this.drawTooltipLine(chartWarp, this.chartWarpHeight);

        // 绘制顶部的时间tooltip显示区域
        this.drawTopTimeWarp();
        // // 绘制底部时间条形选择框
        this.drawxAxisBrush();
        
        this.isPrint = true;

        // // event warp click 监听事件点击事件
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

        // 默认开启trace 分析，默认打开javalock 跟 log绘制
        this.showJavaLockFlag && this.showJavaLock();
        this.showLogFlag && this.showLog();

        d3.select('#span_chart').attr('style', 'display: block');
        d3.select('#event_chart').attr('style', 'display: none');
        d3.select('.bottom_xaxis_warp').attr('style', 'display: none');
    }
    hideEventChart = () => {
        d3.select('#span_chart').attr('style', 'display: block');
        d3.select('#event_chart').attr('style', 'display: none');
        d3.select('.bottom_xaxis_warp').attr('style', 'display: none');
        d3.selectAll('.event_span_shining').remove();
    }
    // 绘制线程名称
    drawThreadName(textWarp, text, item, y) {
        let _this = this;
        textWarp.append('text')
            .text(text)
            .attr('data-tid', item.tid)
            .attr('class', `left_name ${item.active && 'active'}`)
            .attr('x', 15)
            .attr('y', y);
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
                            netEvent = event;
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
    // 绘制虚线与下级详情的框线
    drawSubEventWarp(subRectWarp, subEventList: IEventTime[], idx, left) {
        const _this = this;
        const middleWidth = this.svgWidth / 2 + this.nameWidth / 2;
        const lineColor = this.getEventLineColor(subEventList);

        let subLeft;
        left = parseFloat(left)
        let startLeft = 0, endLeft = 0;
        if (left < middleWidth) {
            startLeft = subLeft = left;
        } else {
            startLeft = subLeft = left - subEventList.length * this.subEventWidth;
        }
        _.forEach(subEventList, (event: IEventTime) => {
            let sevent: IEvent = _.find(eventList, {value: event.eventType}) as IEvent;
            const eventWarp = subRectWarp.append('g')
                .attr('id', `id_${idx}_${event.idx}`)
                .attr('data-type', event.eventType)
                .attr('class', `event_warp ${event.eventType === 'runqLatency' ? 'event_runq' : ''}`);
            eventWarp.append('rect')
                .attr('class', 'event_rect')
                .attr('width', this.subEventWidth)
                .attr('height', this.barWidth)
                .attr('fill', sevent.fillColor || '#1890ff')
                .attr('x', subLeft)
                .attr('y', this.barPadding * 5 + this.barWidth * 2);
            
            this.drawSubRectText(eventWarp, sevent, event, subLeft + 2, this.barPadding * 5 + this.barWidth * 2.5 - 8);
            if(event.type === 'net') {
                this.drawSubImage(eventWarp, event, subLeft);
            }
            subLeft += this.subEventWidth;
        });
        endLeft = subLeft;

        this.drawSubPath(subRectWarp, left + this.subEventNumWidth / 2, this.barPadding * 3 + this.barWidth * 2, startLeft, endLeft, this.barPadding * 5 + this.barWidth * 2, lineColor);
    
        d3.selectAll('.event_warp').on('click', (e) => {
            _this.handleEventRectClick(e, true);
        });
    }
    drawSubPath(barWarp, sx, sy, ex1, ex2, ey, color) {
        barWarp.append('line')
            .attr('class', 'event_dash dash_line')
            .attr('stroke', color)
            .attr('x1', sx)
            .attr('x2', ex1)
            .attr('y1', sy)
            .attr('y2', ey);
        barWarp.append('line')
            .attr('class', 'event_dash dash_line')
            .attr('stroke', color)
            .attr('x1', sx)
            .attr('x2', ex2)
            .attr('y1', sy)
            .attr('y2', ey);
    }
    drawSubRectText(eventWarp, sevent: IEvent, event: IEventTime, x, y) {
        let name = sevent.alias;
        let top = y;
        // if (sevent.type === 'net' && sevent.alias) {
        //     let namelist: string[] = [];
        //     if (event.active) {
        //         if (event.eventType === 'netread') {
        //             namelist = ['server', 'request'];
        //         } else {
        //             namelist = ['server', 'response'];
        //         }
        //     } else {
        //         namelist = sevent.alias?.split(' ');
        //     }
        //     eventWarp.append('text').text(namelist[0])
        //         .attr('class', 'event_rect_small_text')
        //         .attr('stroke', sevent.color)
        //         .attr('x', x)
        //         .attr('y', top)
        //     top += 10;
        //     eventWarp.append('text').text(namelist[1])
        //         .attr('class', 'event_rect_small_text')
        //         .attr('stroke', sevent.color)
        //         .attr('x', x)
        //         .attr('y', top)
        // } else {
            eventWarp.append('text').text(name)
            .attr('class', 'event_rect_small_text')
            .attr('stroke', sevent.color)
            .attr('x', x)
            .attr('y', y + 2)
        // }
        top += 15;
        eventWarp.append('text').text((event.time).toFixed(2))
            .attr('class', 'event_rect_small_text')
            .attr('stroke', sevent.color)
            .attr('x', x)
            .attr('y', top)
    }
    drawSubImage(eventWarp, event: IEventTime, x) {
        let y  = this.barPadding * 5 + this.barWidth * 2.5;
        eventWarp.append('image')
            .attr('class', 'sub_event_image')
            .attr('xlink:href', event.eventType === 'on' ? logPng : (event.eventType === 'netread' ? netReadPng : netWritePng))
            .attr('x', x + 23)
            .attr('y', y)
            .attr('width', 15)
            .attr('height', 15);
    }
    drawNumRect(subRectWarp, left, subLeft, list: IEventTime[], lineColor) {
        let num = list.length;
        let types = _.chain(list).map('type').uniq().join('_').value();
        let classValue = '';
        if (!_.some(list, item => this.subEvents.indexOf(item.type) > -1)) {
            classValue = 'disabled';
        }
        subRectWarp.attr('data-types', types);
        subRectWarp.append('rect')
            .attr('class', `event_num_rect ${classValue}`)
            .attr('width', this.subEventNumWidth)
            .attr('height', this.barWidth)
            .attr('rx', 3)
            .attr('ry', 3)
            .attr('x', subLeft)
            .attr('y', this.barPadding * 3 + this.barWidth);
        subRectWarp.append('text')
            .text(num)
            .attr('class', `event_num_text ${classValue}`)
            .attr('x', subLeft + (num > 9 ? 3 : 7))
            .attr('y', this.barPadding * 3 + this.barWidth * 1.5 + 4);
        subRectWarp.append('line')
            .attr('class', 'event_dash dash_line')
            .attr('stroke', lineColor)
            .attr('x1', parseFloat(left))
            .attr('x2', subLeft + this.subEventNumWidth / 2)
            .attr('y1', this.barPadding + this.barWidth)
            .attr('y2', this.barPadding * 3 + this.barWidth);
    }
    addNumRectMouseEvent() {
        d3.selectAll('.sub_num_warp')
            .on('mouseenter', function(this: any) {
                let id = d3.select(this).attr('id');
                let lineId = id.replace(/subRectNum/, 'line');
                d3.select(`#${lineId}`).attr('style', 'stroke-width: 4px');
                let classValue = d3.select(this).select('.event_num_rect').attr('class');
                if (!classValue.includes('disabled')) {
                    d3.select(this).select('.event_num_rect').attr('style', 'stroke: #1890ff');
                }
                // d3.select(this).select('.event_num_text').attr('style', 'stroke: #1890ff');
            })
            .on('mouseleave', function(this: any) {
                let id = d3.select(this).attr('id');
                let lineId = id.replace(/subRectNum/, 'line');
                d3.select(`#${lineId}`).attr('style', 'stroke-width: 1px');
                d3.select(this).select('.event_num_rect').attr('style', 'stroke: #dcdcdc');
                // d3.select(this).select('.event_num_text').attr('style', 'stroke: #333333');
            });
    }
    changeSubEventType(events: string[]) {
        let _this = this;
        this.subEvents = events;
        d3.selectAll('.sub_num_warp').each(function() {
            let types = d3.select(this).attr('data-types');
            let classValue = '';
            if (_this.subEvents.length > 0) {
                if (!_.some(_this.subEvents, type => types.indexOf(type) > -1)) {
                    classValue = 'disabled';
                }
            }
            d3.select(this).select('.event_num_rect').attr('class', `event_num_rect ${classValue}`);
            d3.select(this).select('.event_num_text').attr('class', `event_num_text ${classValue}`);
        });
    }
    // 每次渲染或者筛选brush时，对时间占比很小绘制虚线的事件进行group聚合并绘制对应的虚线
    drawThreadEventDashLine(idx: number, lineData: IEventTime[], barWarp: any) {
        const _this = this;
        this.chartWarpHeight += this.threadHeight * 2;
        this.svg.attr('height', this.chartWarpHeight > this.svgHeight ? this.chartWarpHeight : this.svgHeight);
        barWarp.attr('height', this.threadHeight * 3);

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
        // 将缩略事件中邻近的on的事件进行归并
        _.forEach(group2LineData, (list: IEventTime[], key) => {
            let subOnList: any = [];
            let subOnListIndex = 0;
            let subOn: any = {};
            _.forEach(list, (item: IEventTime, idx) => {
                if (item.eventType === 'on') {
                    if (subOnList[subOnListIndex]) {
                        subOn.time += item.time;
                        if (item.log) {
                            subOn.log = item.log;
                        }
                        if (idx === list.length - 1) {
                            subOnList[subOnListIndex].push(item.startTime, subOn, true);
                            subOnListIndex++;
                            subOn = {};
                        }
                    } else {
                        subOnList[subOnListIndex] = [];
                        subOnList[subOnListIndex].push(item.startTime);
                        subOn = {...item}
                    }
                } else {
                    if (subOnList[subOnListIndex]) {
                        subOnList[subOnListIndex].push(item.startTime, subOn, false);
                        subOnListIndex++;
                        subOn = {};
                    }
                }
            });
            _.forEach(subOnList, opt => {
                let idx1 = _.findIndex(list, {startTime: opt[0]});
                let idx2 = _.findIndex(list, {startTime: opt[1]});
                if (opt.length > 3 && idx2 - idx1 > 1) {
                    let length = opt[3] ? idx2 - idx1 + 1 : idx2 - idx1;
                    list.splice(idx1, length, opt[2]);
                }
            });
        });
        // console.log('group2LineData', group2LineData);
        _.forEach(group2LineData, (list, left) => {
            const lineWarp = barWarp.append('g').attr('class', 'event_dash_warp');
            const lineColor = this.getEventLineColor(list);
            lineWarp.append('rect')
                .attr('id', `rect_id_${idx}_${left}`)
                .attr('class', 'event_dash_rect')
                .attr('x', parseInt(left) - 1)
                .attr('y', this.barPadding - 2)
                .attr('width', 3)
                .attr('height', this.barWidth + 4);
            lineWarp.append('line')
                .attr('id', `line_id_${idx}_${left}`)
                .attr('class', `event_dash ${list.length > 1 ? 'dash_line' : ''}`)
                .attr('data-idxs', _.join(_.map(list, 'idx'), '_'))
                .attr('data-types', _.join(_.map(list, 'eventType'), '_'))
                .attr('stroke', lineColor)
                .attr('x1', left)
                .attr('x2', left)
                .attr('y1', this.barPadding)
                .attr('y2', this.barPadding + this.barWidth);
            lineWarp
                .on('mouseenter', function(this: any) {
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
        this.groupLineData[idx] = group2LineData;

        let accrueLeft = 0;
        const middleWidth = this.svgWidth / 2 + this.nameWidth / 2;
        _.forEach(group2LineData, (list, left) => {
            const lineColor = this.getEventLineColor(list);
            if (parseFloat(left) < middleWidth) {
                let space = accrueLeft > 0 ? 5 : 0;
                // 全部绘制的时候需要增量记录已经偏移的accrueLeft
                // let subLeft = parseFloat(left) > accrueLeft ? parseFloat(left) : accrueLeft + space;
                let subLeft = parseFloat(left) - accrueLeft > this.subEventWidth * 3 ? parseFloat(left) : accrueLeft + space;
                const subRectWarp = barWarp.append('g').attr('id', `subRectNum_id_${idx}_${left}`).attr('class', 'sub_num_warp');
                this.drawNumRect(subRectWarp, left, subLeft, list, lineColor);
                subLeft += this.subEventNumWidth;
                accrueLeft = subLeft;
            }
        });
        let accrueLeft2 = 0;
        _.forEachRight(group2LineData, (list, left) => {
            const lineColor = this.getEventLineColor(list);
            if (parseFloat(left) > middleWidth) {
                let space = accrueLeft2 > 0 ? 5 : 0;
                // let subLeft = parseFloat(left) < accrueLeft2 ? parseFloat(left) - space : (accrueLeft2 > 0 ? accrueLeft2 - space : parseFloat(left) - space);
                let subLeft = accrueLeft2 - parseFloat(left) > this.subEventWidth * 3 ? parseFloat(left) - space : (accrueLeft2 > 0 ? accrueLeft2 - space : parseFloat(left) - space);
                const subRectWarp = barWarp.append('g').attr('id', `subRectNum_id_${idx}_${left}`).attr('class', 'sub_num_warp');
                subLeft = subLeft - this.subEventNumWidth;
                accrueLeft2 = subLeft;
                this.drawNumRect(subRectWarp, left, subLeft, list, lineColor);
            }
        });
        this.addNumRectMouseEvent();
        d3.selectAll('.sub_num_warp').on('click', function(e) {
            let classValue = d3.select(this).select('.event_num_rect').attr('class');
            if (!classValue.includes('disabled')) {
                let id = d3.select(this).attr('id');
                let idx = id.split('_')[2];
                let left = id.split('_')[3];
                let grouplineData = _this.groupLineData[idx]
                let subEventList = grouplineData[left];
                let position = d3.select(this).select('.event_num_rect').attr('x');
                let subRectWarp = d3.select(`#subEventRectWarp${idx}`);
                subRectWarp.html('');
                _this.drawSubEventWarp(subRectWarp, subEventList, idx, position);
            }
        });

        const subRectWarp = barWarp.append('g').attr('id', `subEventRectWarp${idx}`);
        let position1 = d3.select(`#subRectNum_id_${idx}_${_.keys(group2LineData)[0]}`).select('.event_num_rect').attr('x');
        let subEventList = group2LineData[_.keys(group2LineData)[0]];
        this.drawSubEventWarp(subRectWarp, subEventList, idx, position1);

        // let accrueLeft = 0;
        // _.forEach(group2LineData, (list, left) => {
        //     const lineColor = this.getEventLineColor(list);
        //     let startLeft = 0, leftEnd = 0;
        //     if (parseFloat(left) < middleWidth) {
        //         let space = accrueLeft > 0 ? 10 : 0;
        //         // 全部绘制的时候需要增量记录已经偏移的accrueLeft
        //         let subLeft = parseFloat(left) > accrueLeft ? parseFloat(left) + space : accrueLeft + space;
        //         // let subLeft;
        //         // if (parseFloat(left) < accrueLeft && accrueLeft - parseFloat(left) > subEventWidth * 4) {
        //         //     subLeft = parseFloat(left) + space
        //         // } else {
        //         //     subLeft =  accrueLeft + space;
        //         // }
        //         startLeft = subLeft;
        //         const subRectWarp = barWarp.append('g').attr('id', `subRect_id_${idx}_${left}`).attr('class', 'sub_event_warp');
        //         let drawList = list;
        //         if (list.length > 2) {
        //             drawList = list.slice(0, 2);
        //         }
        //         _.forEach(drawList, (event: IEventTime) => {
        //             let sevent: IEvent = _.find(eventList, {value: event.eventType}) as IEvent;
        //             const eventWarp = subRectWarp.append('g')
        //                 .attr('id', `id_${idx}_${event.idx}`)
        //                 .attr('data-type', event.eventType)
        //                 .attr('class', 'event_warp');
        //             eventWarp.append('rect')
        //                 .attr('class', 'event_rect')
        //                 .attr('width', subEventWidth)
        //                 .attr('height', this.barWidth)
        //                 .attr('fill', sevent.fillColor || '#1890ff')
        //                 .attr('x', subLeft)
        //                 .attr('y', this.barPadding * 3 + this.barWidth);
                    
        //             this.drawSubRectText(eventWarp, sevent, event, subLeft + 2, this.barPadding * 3 + this.barWidth + this.barWidth / 2 - 8);
        //             if(event.type === 'net') {
        //                 this.drawSubImage(eventWarp, event, subLeft);
        //             }
        //             subLeft += subEventWidth;
        //             accrueLeft = subLeft;
        //         });
        //         if(list.length > 2) {
        //             subRectWarp.append('rect')
        //                 .attr('data-left', left)
        //                 .attr('data-position', subLeft)
        //                 .attr('class', 'more_event_rect')
        //                 .attr('width', 10)
        //                 .attr('height', this.barWidth)
        //                 .attr('x', subLeft)
        //                 .attr('y', this.barPadding * 3 + this.barWidth);
        //             subRectWarp.append('text')
        //                 .text('>') 
        //                 .attr('class', 'more_event_rect_text')
        //                 .attr('x', subLeft + 2)
        //                 .attr('y', this.barPadding * 3 + this.barWidth + this.barWidth / 2 + 5);
        //             subLeft += 10;
        //             accrueLeft = subLeft;
        //         }

        //         leftEnd = subLeft;
        //         this.drawSubPath(subRectWarp, left, this.barPadding + this.barWidth, startLeft, leftEnd, this.barPadding * 3 + this.barWidth, lineColor);
        //     }
        // });
        // let accrueLeft2 = 0;
        // _.forEachRight(group2LineData, (list, left) => {
        //     const lineColor = this.getEventLineColor(list);
        //     let startLeft = 0, leftEnd = 0;
        //     if (parseFloat(left) > middleWidth) {
        //         let space = accrueLeft2 > 0 ? 10 : subEventWidth;
        //         let subLeft = parseFloat(left) < accrueLeft2 ? parseFloat(left) - space : (accrueLeft2 > 0 ? accrueLeft2 - space : parseFloat(left) - space);
        //         // let subLeft;
        //         // if (parseFloat(left) > accrueLeft2 && parseFloat(left) - accrueLeft2 > subEventWidth * 4) {
        //         //     subLeft = parseFloat(left) - space
        //         // } else {
        //         //     subLeft =  accrueLeft2 > 0 ? accrueLeft2 - space : parseFloat(left) - space;
        //         // }
        //         leftEnd = subLeft;
        //         const subRectWarp = barWarp.append('g').attr('id', `subRect_id_${idx}_${left}`).attr('class', 'sub_event_warp');
        //         let drawList = list;
        //         if (list.length > 2) {
        //             drawList = list.slice(0, 2);
        //         }
        //         subLeft = subLeft - drawList.length * subEventWidth - (list.length > 2 ? 10 : 0);
        //         accrueLeft2 = subLeft;
        //         startLeft = subLeft;
        //         _.forEach(drawList, (event: IEventTime) => {
        //             let sevent: IEvent = _.find(eventList, {value: event.eventType}) as IEvent;
        //             const eventWarp = subRectWarp.append('g')
        //                 .attr('id', `id_${idx}_${event.idx}`)
        //                 .attr('data-type', event.eventType)
        //                 .attr('class', 'event_warp');
        //             eventWarp.append('rect')
        //                 .attr('class', 'event_rect')
        //                 .attr('width', subEventWidth)
        //                 .attr('height', this.barWidth)
        //                 .attr('fill', sevent.fillColor)
        //                 .attr('x', subLeft)
        //                 .attr('y', this.barPadding * 3 + this.barWidth);

        //             this.drawSubRectText(eventWarp, sevent, event, subLeft + 2, this.barPadding * 3 + this.barWidth + this.barWidth / 2 - 8);
        //             if(event.type === 'net') {
        //                 this.drawSubImage(eventWarp, event, subLeft);
        //             }
        //             subLeft += subEventWidth;
        //         });
        //         if (list.length > 2) {
        //             subRectWarp.append('rect')
        //                 .attr('data-left', left)
        //                 .attr('data-position', subLeft)
        //                 .attr('class', 'more_event_rect')
        //                 .attr('width', 10)
        //                 .attr('height', this.barWidth)
        //                 .attr('x', subLeft)
        //                 .attr('y', this.barPadding * 3 + this.barWidth);
        //             subRectWarp.append('text')
        //                 .text('>') 
        //                 .attr('class', 'more_event_rect_text')
        //                 .attr('x', subLeft + 2)
        //                 .attr('y', this.barPadding * 3 + this.barWidth + this.barWidth / 2 + 5);
        //             subLeft += 10;
        //         }
                
        //         this.drawSubPath(subRectWarp, left, this.barPadding + this.barWidth, startLeft, leftEnd, this.barPadding * 3 + this.barWidth, lineColor);
        //         // this.drawSubPath(subRectWarp, left, this.barPadding + this.barWidth, startLeft+subEventWidth, leftEnd+subEventWidth, this.barPadding * 3 + this.barWidth, lineColor);
        //     }
        // });
        
    }

    // 绘制全局移动的tooltip line
    drawTooltipLine(chartWarp: any, height) {
        let lineWarp = chartWarp.append('g').attr('class', 'tooltip_line_warp').attr('opacity', 0);
        lineWarp.append('line')
            .attr('class', 'tooltip_line')
            .attr('x1', 0)
            .attr('x2', 0)
            .attr('y1', 0)
            .attr('y2', height);
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

    // 绘制时间筛选两端的时间text-tspan
    private drawXAxisWEText(xAxisWarp: any, type: 1 | 0, time) {
        let _this = this;
        const circleSymbol = d3.symbol().size(30).type(d3.symbolCircle);
        const position = this.xScale(time);
        xAxisWarp.append('text')
            .attr('class', 'handle-text')
            .text(type === 1 ? 'IO Start Time' : 'IO End Time')
            .attr('x', type === 1 ? position + 2 : position - 60)
            .attr('y', 10);
        xAxisWarp.append('path')
            .attr('class', 'xaxis_circle')
            .attr('d', circleSymbol)
            .attr('transform', `translate(${position}, 15)`);
    }
    // 绘制底部时间轴
    drawxAxisBrush() {
        const _this = this;
        const bottomXAxisWarp = d3.select('#bottom_xaxis_svg');
        bottomXAxisWarp.html(null);
        // 返回起始和终止两个时间构造线性时间比例尺
        // 定义xAxis坐标轴的比例尺
        let xAxisWarp = bottomXAxisWarp.append('g').attr('id', 'xAxis_warp').attr('transform', `translate(0, 0)`);
        this.xScale2 = d3.scaleLinear().domain(this.xScale.domain()).range([this.nameWidth, this.svgWidth]);
        const xAxis = d3.axisBottom(this.xScale2)
            .tickFormat(d => timeFormat(d as Date));

        // xAxisWarp.append('rect').attr('x', this.nameWidth).attr('y', this.svgWidth).attr('width', this.barWarpWidth).attr('height', 10)
       
        xAxisWarp.append('g')
            .attr('class', 'xaxis-x')
            .attr('transform', `translate(0, 15)`)
            .call(xAxis);

        let requestTime = _.filter(this.lineTimeList, item => item.type === 'request');
        this.drawXAxisWEText(xAxisWarp, 1, requestTime[0].time); // 'IO Start Time'
        this.drawXAxisWEText(xAxisWarp, 0, requestTime[1].time); // 'IO End Time'
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
    // 监听事件的click事件
    private changeEventRectColor(id, active) {
        let eventType = d3.select(`#${id}`).attr('data-type');
        let sevent = _.find(eventList, {value: eventType});
        sevent && d3.select(`#${id}`).select('.event_rect').attr('fill', active ? sevent.activeColor : sevent.fillColor);
    }
    handleEventRectClick(e: any, showActive = false) {
        let id = e.currentTarget.id;
        let temp = id.split('_');
        const evt: any = this.data[temp[1]].eventList[temp[2]];
        evt.threadName = this.data[temp[1]].name;
        evt.tid = this.data[temp[1]].tid;
        this.eventClick(evt);
        if (showActive) {
            if (this.activeEventDomId && !d3.select(`#${this.activeEventDomId}`).empty()) {
                this.changeEventRectColor(this.activeEventDomId, false);
            }
            this.activeEventDomId = id;
            this.changeEventRectColor(this.activeEventDomId, true);
        } else {
            if (this.activeEventDomId && !d3.select(`#${this.activeEventDomId}`).empty()) {
                this.changeEventRectColor(this.activeEventDomId, false);
            }
            this.activeEventDomId = '';
        }
    }
    addEventClickListener() {
        const _this = this;
        d3.selectAll('.event_warp').on('click', (e) => {
            _this.handleEventRectClick(e, true);
        });
        d3.selectAll('.start_icon').on('click', function(e) {
            let tid = d3.select(this).attr('data-tid');
            let timestamp = d3.select(this).attr('data-timestamp');
            _this.handleStarClick(tid, timestamp);
        });
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


    drawSpanTree() {
        let _this = this;
        const colorsList: any[] = ['#f1dbc1', '#f1d6b7', '#efb77b', '#ec9a3e', '#e97a00'];
        const colorScale = d3.scaleQuantize().domain([0, 1]).range(colorsList)

        if (this.spanList.length === 0) {
            return;
        }
        
        const inner: any = document.getElementById('span_svg');
        inner.innerHTML = '';
        
        _.forEach(this.spanList, item => {
            item.timeRate = item.time / this.timeRangeDiff
        });

        const spanHeight = 38; 
        const spanRectHeight = 28;
        const spanRectY = (spanHeight - spanRectHeight) / 2;


        // const spanSize = document.getElementById('camera')?.getBoundingClientRect();
        // const sizeHeight = spanSize?.height as number;
        // const spanChartHeight = 28 * maxLevel > sizeHeight ? 28 * maxLevel : sizeHeight;
        const spanChartHeight = 38 * this.spanList.length;
        this.spanSvg = d3.select('#span_svg').attr('height', spanChartHeight);
        
        const spanTreeWarp = this.spanSvg.append('g').attr('id', 'span_tree_warp').attr('transform', `translate(0, 0)` );
        const circleSymbol = d3.symbol().size(20).type(d3.symbolCircle);
        _.forEach(this.spanList, (item, idx) => {
            const levelWarp = spanTreeWarp.append('g').attr('class', 'span_level').attr('transform', `translate(0, ${spanHeight * idx})`);
            let textLeft = idx > 6 ? 12 * 6 + 5 : 12 * idx + 5;
            levelWarp.append('path')
                .attr('class', 'circle_symbol')
                .attr('d', circleSymbol)
                .attr('transform', `translate(${textLeft}, ${spanHeight / 2})`);
            levelWarp.append('text')
                .text(idx === 0 ? '入口span' : '子span')
                .attr('class', 'span_title')
                .attr('x', textLeft + 8)
                .attr('y', spanHeight / 2 + 3);
            
            let line = d3.line().x(d => d[0]).y(d => d[1]);
            if (idx > 0 && idx < 7) {
                let preLeft = textLeft - 12;
                let lineData: [number, number][] = [[preLeft, -15],[preLeft, spanHeight / 2],[textLeft - 4, spanHeight / 2]];
                levelWarp.append('svg:path')
                    .attr('d', line(lineData))
                    .attr('class', 'span_title_line');
            } else if (idx > 6){
                let lineData: [number, number][] = [[textLeft, -15], [textLeft, spanHeight / 2 - 2]];
                levelWarp.append('svg:path')
                    .attr('d', line(lineData))
                    .attr('class', 'span_title_line');
            }

            levelWarp.append('rect')
                .attr('class', 'span_rect_bg')
                .attr('data-tid', item.tid)
                .attr('data-stime', item.startTime)
                .attr('data-etime', item.endTime)
                .attr('width', this.barWarpWidth)
                .attr('height', spanRectHeight)
                .attr('x', this.nameWidth)
                .attr('y', spanRectY)
                .attr('rx', spanRectHeight / 2)
                .attr('ry', spanRectHeight / 2);
            

            let left = this.xScale(item.startTime) as number;
            let right = this.xScale(item.endTime) as number;
            levelWarp.append('rect')
                .attr('class', 'span_rect')
                .attr('data-tid', item.tid)
                .attr('data-stime', item.startTime)
                .attr('data-etime', item.endTime)
                .attr('width', item.timeRate < 0.01 ? 2 : right - left)
                .attr('height', spanRectHeight)
                .attr('x', left)
                .attr('y', spanRectY)
                .attr('rx', spanRectHeight / 2)
                .attr('ry', spanRectHeight / 2)
                .attr('fill', colorScale(item.timeRate));
                
            levelWarp.append('text')
                .text(item.name)
                .attr('class', 'span_rect_text')
                .attr('x', this.nameWidth + 15)
                .attr('y', spanHeight / 2 + 4)
            levelWarp.append('text')
                .text(parseFloat(item.time).toFixed(2) + 'ms')
                .attr('class', 'span_rect_text')
                .attr('x', this.svgWidth - 70)
                .attr('y', spanHeight / 2 + 4)
        });

        // 绘制tooltip的虚线
        this.drawTooltipLine(spanTreeWarp, spanChartHeight);

        d3.selectAll('.span_rect_bg').on('click', function(e) {
            const tid = d3.select(this).attr('data-tid');
            const stime = d3.select(this).attr('data-stime');
            const etime = d3.select(this).attr('data-etime');
            _this.handleSpanRectClick(tid, stime, etime);
        });
        d3.selectAll('.span_rect').on('click', function(e) {
            const tid = d3.select(this).attr('data-tid');
            const stime = d3.select(this).attr('data-stime');
            const etime = d3.select(this).attr('data-etime');
            _this.handleSpanRectClick(tid, stime, etime);
        });
    }
    // 点击span在下方Event视图中显示框选区域
    handleSpanRectClick(tid, stime, etime) {
        d3.select('#span_chart').attr('style', 'display: none');
        d3.select('#event_chart').attr('style', 'height: 95%; display: block');
        d3.select('.bottom_xaxis_warp').attr('style', 'display: block;height: 35px');
        this.parentRef.props.onEventDetail(true);
        this.parentRef.setState({ showEventChart: true });

        // 在spna对应的tid的线程上绘制对应的时间框选
        let left = this.xScale(parseFloat(stime));
        let right = this.xScale(parseFloat(etime));
        document.getElementById(`thread_warp_${tid}`)?.scrollIntoView();
        d3.select(`#thread_warp_${tid}`)
            .select('.bar_warp')
            .append('rect')
            .attr('class', 'span_event_rect event_span_shining')
            .attr('fill', '#ffffff00')
            .attr('x', left)
            .attr('y', 1)
            .attr('width', right - left < 2 ? 2 : right - left)
            .attr('height', this.threadHeight)
        
        // 在底部时间轴上绘制对应的span线段，标识span执行时间
        if (!d3.select('.xaxis_span_line').empty()) {
            d3.select('.xaxis_span_line').remove();
        }
        let bleft = this.xScale2(parseFloat(stime));
        let bright = this.xScale2(parseFloat(etime));
        d3.select('#bottom_xaxis_svg').select('#xAxis_warp').append('line')
            .attr('class', 'xaxis_span_line')
            .attr('x1', bleft)
            .attr('x2', bright)
            .attr('y1', 15)
            .attr('y2', 15);

        // 监听等点击svg的时候移除当前shining的span框
        const svgDom = d3.select('#camera_svg');
        svgDom.on('click', function() {
            d3.selectAll('.event_span_shining').remove();
            svgDom.on('click', null);
        });
    }

    changeSize = () => {
        this.draw();
    }
}

export default EasyCamera;