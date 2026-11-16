import * as d3 from 'd3';
import { D3BrushEvent } from 'd3';
import _ from 'lodash';
import { ILegend } from './types';
import { dataHandle, textHandle, timeFormat, DateTimeFormat, colors } from './util';


// 绘制Legend
const drawLegend = (legendWarp: any, data: ILegend[]) => {
    let legendWidthSum = 10;
    _.forEach(data, (legend: ILegend) => {
        let legendG = legendWarp.append('g').attr('class', 'legend_name').attr('transform', `translate(${legendWidthSum}, 10)`);
        legendG.append('rect')
            .attr('width', 10)
            .attr('height', 10)
            .attr('rx', 2)
            .attr('ry', 2)
            .attr('x', 0)
            .attr('y', 0)
            .attr('fill', legend.color)
        legendG.append('text')
            .text(textHandle(legend.name, 15))
            .attr('x', function () {
                let textlen: number = this.getComputedTextLength();
                legendWidthSum += 10 + 5 + textlen + 15;
                return 10 + 5;
            })
            .attr('y', 9)
            .style('font-size', '14px');
    });
}



// TODO 在内部移动时tooltip也要跟着动
const showTooltip = (x: number, y: number) => {
    let tooltipDom = document.getElementById('tooltip_warp');  
    if (tooltipDom) {
        // if (!ifShowToolip) {
        //     tooltipDom.setAttribute('style', `display: block;left: ${x}px;top: ${y}px`);
        // } else {    
        //     tooltipDom.setAttribute('style', `left: ${x}px;top: ${y}px`);
        // }
    }
}
const moveTooltip = (e: any) => {
    let { clientX, clientY } = e;
    showTooltip(clientX, clientY);
}
const hideTooltip = () => {
    let tooltipDom = document.getElementById('tooltip_warp');  
    if (tooltipDom) {
        tooltipDom.setAttribute('style', `display: none;`);
    }   
}



interface CDraw {
    svg: any;
    nameWidth: number;
    svgWidth: number;

    drawxAxis: () => void;
}

function cameraObj(this: CDraw) {}
cameraObj.prototype.draw = function(data: any, option: any) {
    let _this = this;
    _this.nameWidth = option?.nameWidth || 120; // 侧边轴上的名称宽度包括跟柱图的间隔20
    const barWidth = option?.barWidth || 26; // 柱图的高度
    const barPadding = option?.barPadding || 12; // 柱图的上下padding
    const padding = option?.padding || 20;

    const { legendList, dashList } = dataHandle(data);

    const svg = d3.select('#camera_svg');
    _this.svg = svg;
    // legend相关基础参数
    const legendWarp = svg.append('g').attr('id', 'legend_warp').attr('transform', `translate(0, 0)` );

    const chartWarp = svg.append('g').attr('id', 'chart_warp').attr('transform', `translate(0, 32)` );
    
    const cameraSize = document.getElementById('camera')?.getBoundingClientRect();
    _this.svgWidth = cameraSize?.width as number;

    const threadHeight = barWidth + barPadding * 2; // 柱图高度加上下间隔组成对应rect的高度，撑开thread_warp

    const brush = d3.brushX()
        .extent([[0, 0], [this.svgWidth, threadHeight]])
        .on('brush', () => {});
    // 绘制主体柱图区域
    _.forEach(data, (item, idx: number) => {
        let threadWarp = chartWarp.append('g').attr('id', `thread_warp_${idx}`).attr('transform', `translate(0, ${threadHeight * idx})` );
        threadWarp.append('rect').attr('class', 'thread_warp').attr('height', threadHeight);
        threadWarp
            .append('text')
            .text(`${item.name}`)
            .attr('class', 'left_name')
            .attr('x', function () {
                let textlen: number = this.getComputedTextLength();
                console.log(textlen);
                return _this.nameWidth - 20 - textlen;
            })
            .attr('y', barPadding + barWidth / 2 + 5);

        let barWarp = threadWarp.append('g').attr('class', 'bar_warp');
        let allEventTimeValue = _.sum(_.map(item.events, 'time'));
        _.forEach(item.events, (event, idx2:number) => {
            let timeWidth = event.time / allEventTimeValue * (this.svgWidth - _this.nameWidth - padding);
            let left = _this.nameWidth + (event.start / allEventTimeValue * (this.svgWidth - _this.nameWidth - padding));
            if (event.time / allEventTimeValue > 0.01) {
                barWarp.append('rect')
                    .attr('id', `id_${idx}_${idx2}`)
                    .attr('class', 'event_rect')
                    .attr('data-name', item.name)
                    .attr('data-event', event.name)
                    .attr('width', timeWidth)
                    .attr('height', barWidth)
                    .attr('fill', colors[idx2])
                    .attr('x', left)
                    .attr('y', barPadding)
            } else {
                barWarp.append('line')
                    .attr('class', 'event_dash')
                    .attr('x1', left)
                    .attr('x2', left)
                    .attr('y1', barPadding)
                    .attr('y2', barPadding + barWidth)
            }
        });
        // 日志Icon绘制，默认初始不显示
        var symbol = d3.symbol()
            .size(100)
            .type(d3.symbolTriangle);
        _.forEach(item.logEvents, (logs, idx) => {
            let position = _this.nameWidth + (logs.time / allEventTimeValue * (this.svgWidth - _this.nameWidth - padding));
            threadWarp.append('path')
                .attr('class', 'log_icon')
                .attr('d', symbol)
                .attr('transform', `translate(${position}, ${barPadding + barWidth})`)
                .attr('opacity', 0)
                .attr('fill', '#333333');
        });

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
        
        // 设置d3 brush
        barWarp.call(brush).call(brush.move); // x.range()
    });

    // 绘制顶部legend
    drawLegend(legendWarp, legendList);
    // 绘制底部时间条形选择框
    this.drawxAxis();
    // 绘制鼠标移上去时的全局tooltip虚线
    const chartWarpSize = document.getElementById('chart_warp')?.getBoundingClientRect();
    chartWarp.append('line')
        .attr('class', 'tooltip_line')
        .attr('opacity', 0)
        .attr('x1', 10)
        .attr('x2', 10)
        .attr('y1', 0)
        .attr('y2', chartWarpSize?.height as number)
    
    chartWarp
        .on('mouseenter', function(e) {
            let [x] = d3.pointer(e, this);
            d3.select('.tooltip_line')
                .attr('opacity', 1)
                .attr('x1', x)
                .attr('x2', x);
        })
        .on('mousemove', function(e) {
            let [x] = d3.pointer(e, this);
            d3.select('.tooltip_line')
                .attr('opacity', 1)
                .attr('x1', x)
                .attr('x2', x);
        })
        .on('mouseleave', function(e) {
            d3.select('.tooltip_line').attr('opacity', 0);
        })
        // Tooltip事件监听
        // d3.selectAll('.event_rect').on('mouseenter', (e, d) => {
        //     let { clientX, clientY } = e;
        //     showTooltip(clientX, clientY);
        // });
        // d3.selectAll('.event_rect').on('mousemove', _.debounce(moveTooltip, 10));
        // d3.selectAll('.event_rect').on('mouseleave', () => {
        //     console.log('hide')
        //     hideTooltip();
        // });
}

// 绘制时间筛选两端的时间text-tspan
function drawXAxisWEText(xAxisWarp: any, type: 'start' | 'end', times: string[], x: number) {
    let textWarp: any;
    if (type === 'start') {
        textWarp = xAxisWarp.append('text').attr('class', 'handle-text handle-text-w');
    } else {
        textWarp = xAxisWarp.append('text').attr('class', 'handle-text handle-text-e');
    }
    textWarp.selectAll('tspan')
        .data(times)
        .enter()
        .append('tspan')
        .text((time: string) => time)
        .attr('x', function() {
            if (type === 'start') {
                let textlen: number = this.getComputedTextLength();
                return x - 3 - textlen;
            } else {
                return x + 3;
            }
        })
        .attr('y', (d: string, i: number) => {
            return 12 + 12 * i;
        });
}
// 时间筛选框筛选时更新text-tspan中的时间值
function updateXAxisWEText(textWarp: any, type: 'start' | 'end', times: string[], x: number) {
    textWarp.selectAll('tspan')
        .data(times)
        .text((time: string) => time)
        .attr('x', function() {
            if (type === 'start') {
                let textlen: number = this.getComputedTextLength();
                return x - 3 - textlen;
            } else {
                return x + 3;
            }
        });
}
/**
 * 绘制时间筛选的区域
 */
cameraObj.prototype.drawxAxis = function() {
    const _this = this;
    // 返回起始和终止两个时间构造线性时间比例尺
    const xAxisHeight = 28;
    const initTime = [new Date('2022-07-06 08:00:00'), new Date('2022-07-07 08:00:00')];
    // 定义xAxis坐标轴的比例尺
    let xAxisWarp = _this.svg.append('g').attr('id', 'xAxis_warp').attr('transform', `translate(0, ${300})`);
    let x = d3.scaleLinear()
        .domain(initTime)
        .range([_this.nameWidth, _this.svgWidth - 20]);
    const x2 = d3.scaleTime().domain(x.domain()).range([_this.nameWidth, _this.svgWidth - 20]);
    const xAxis = d3.axisBottom(x)
        // .ticks(d3.timeMonth.every(2))
        .tickFormat(d => timeFormat(d));

    // brush时间监听
    const xAxisBrushed = function(e: any) {
        const selection = e.selection;
        // x.domain(selection.map(x2.invert, x2));
        // 获取brush筛选的时间区间
        let tiemRange = selection.map(x2.invert, x2);
        // d3.select('.handle-text-w').text(timeFormat(tiemRange[0])).attr('x', selection[0]).attr('y', 10);
        // d3.select('.handle-text-e').text(timeFormat(tiemRange[1])).attr('x', selection[1]).attr('y', 10);
        updateXAxisWEText(d3.select('.handle-text-w'), 'start', DateTimeFormat(tiemRange[0], true) as string[], selection[0]);
        updateXAxisWEText(d3.select('.handle-text-e'), 'end',DateTimeFormat(tiemRange[1], true) as string[], selection[1]);
    }
    // 构造xAxis上的brush对象
    const xAxisBrush = d3.brushX()
        .extent([[this.nameWidth, 0], [this.svgWidth - 20, xAxisHeight]])
        .on('brush', xAxisBrushed);

    xAxisWarp.append('rect')
        .attr('class', 'xAxis_rect')
        .attr('width', this.svgWidth - this.nameWidth - 20)
        .attr('height', xAxisHeight)
        .attr('x', this.nameWidth)
        .attr('y', 0)
    xAxisWarp.call(xAxisBrush).call(xAxisBrush.move, x.range());;

    xAxisWarp.append('g')
        .attr('class', 'xaxis_line')
        .attr('transform', `translate(0, ${xAxisHeight})`)
        .call(xAxis);
    let startTimes: string[] = DateTimeFormat(initTime[0], true) as string[];
    drawXAxisWEText(xAxisWarp, 'start', startTimes, this.nameWidth);
    let endTimes: string[] = DateTimeFormat(initTime[1], true) as string[];
    drawXAxisWEText(xAxisWarp, 'end', endTimes, this.svgWidth - 20);
    // .attr('opacity', 0)
    // .attr('opacity', 0)
}

cameraObj.prototype.showLog = function() {
    console.log('show Log');
    d3.selectAll('.log_icon').attr('opacity', 1);
}
cameraObj.prototype.hideLog = function() {
    console.log('hide Log');
    d3.selectAll('.log_icon').attr('opacity', 0);
}

export default cameraObj;