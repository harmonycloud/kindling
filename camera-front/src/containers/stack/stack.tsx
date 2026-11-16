
import * as d3 from 'd3';
import _ from 'lodash';
import { Modal, Tree } from 'antd';
import { textHandle, timeNSFormat } from './util';
import { IOption } from './types';

const TreeNode = Tree.TreeNode;
class Stack {
    data: any[] = [];
    timeRange: number[] = [];
    svgId: string;
    svg: any;
    svgWidth: number = 500;     // 绘制svg区域的宽度
    nameWidth: number;          // 线程名称区域的宽度
    barWidth: number;
    barPadding: number;         // 柱状上下padding距离 
    padding: number;            // svg绘制区域的padding
    topTimeheight: number = 30;
    chartWarpHeight: number = 200;
    stackHeight: any = {};
    colors: string[] = ['#F79C01', '#D43530', '#F68100', '#D34F1E', '#F65F04', '#F3A900', '#F55018', '#DD812B', '#F5250E', '#FBE200', '#F8B402'];

    ifShowTooltipLine: boolean = false;
    expandThreadIds: number[] = [];

    xScale: any;                
    xAxis: any;                 
    xScale2: any;          
    xAxisWarp: any;
    chartBrush: any;

    constructor (option: IOption) {
        this.data = option?.data;
        this.timeRange = option.timeRange;
        this.svgId = option?.svgId || 'stack_svg';
        this.nameWidth = option?.nameWidth || 150; // 侧边轴上的名称宽度包括跟柱图的间隔20
        this.barPadding = option?.barPadding || 8; // 柱图的上下padding
        this.padding = option?.padding || 20;
        this.barWidth = option?.barWidth || 18; // 柱图的高度

        
    }

    draw() {
        let _this = this;
        const inner: any = document.getElementById(this.svgId);
        inner.innerHTML = '';

        // 柱图高度加上下间隔组成对应rect的高度，撑开thread_warp
        const threadHeight = this.barWidth + this.barPadding * 2; 

        const cameraSize = document.getElementById('stack_svg')?.getBoundingClientRect();
        this.svgWidth = cameraSize?.width as number;
        
        this.svg = d3.select(`#${this.svgId}`);
        const chartWarp = this.svg.append('g').attr('id', 'chart_warp').attr('transform', `translate(0, 0)` );

        this.xScale = d3.scaleLinear()
            .domain(this.timeRange)
            .range([this.nameWidth, this.svgWidth]);
        this.xAxis = d3.axisBottom(this.xScale)

        console.log(this.data);
        this.chartWarpHeight = threadHeight * this.data.length;
        _.forEach(this.data, (item, idx: number) => {
            let threadWarp = chartWarp.append('g')
                .attr('id', `thread_warp_${item.threadId}`)
                .attr('class', 'thread_warp')
                .attr('height', threadHeight)
                .attr('transform', `translate(0, ${threadHeight * idx})` );
            threadWarp.append('rect')
                .attr('class', 'thread_rect_warp')
                .attr('fill', () => idx % 2 === 0 ? '#F9F9F9' : '#FFFFFF')
                .attr('width', this.svgWidth)
                .attr('height', threadHeight);
            const threadText = threadWarp
                .append('text')
                .text(`${textHandle(item.threadName, 15)}`)
                .attr('data-threaId', item.threadId)
                .attr('class', `stack_left_name`)
                .attr('x', function (this: any) {
                    let textlen: number = this.getComputedTextLength();
                    return _this.nameWidth - 20 - textlen;
                })
                .attr('y', this.barPadding + this.barWidth / 2 + 5);
            if (item.threadName.length > 15) {
                threadText.append('title').text(item.threadName);
            }
            const barWarp = threadWarp.append('g').attr('class', 'bar_warp');
            const level = 1;

            const stackWarp = barWarp.selectAll(`.stack_rect_warp`)
                .data(item.flames)
                .enter()
                .append('g')
                .attr('data-threaId', item.threadId)
                .attr('class', d => `stack_rect_warp_${d.id} stack_rect_warp`);
            this.drawStackRect(stackWarp, level);
            this.stackHeight[item.threadId] = threadHeight;

            d3.selectAll('.stack_rect_warp').on('click', function(e, stack: any) {
                console.log(e, stack);
                const threadId = d3.select(this).attr('data-threaId');
                _this.drawStack(threadId, stack);
                _this.resetThreadHeight();
            });
            d3.selectAll('.stack_left_name').on('click', function(e, stack: any) {
                // console.log(e, stack);
                const threadId = parseInt(d3.select(this).attr('data-threaId'));
                const list = _.filter(_this.data, item => item.threadId === threadId);
                if (_this.expandThreadIds.indexOf(threadId) === -1) {
                    _this.expandThreadIds.push(threadId);
                    _this.expandStack(list);
                    _this.resetThreadHeight();
                } else {
                    _this.expandThreadIds.splice(_this.expandThreadIds.indexOf(threadId), 1);
                    _this.collapseStack(list);
                }
            });
            this.addMouseListener();
        });

        this.drawTopTimeWarp();
        this.drawTooltipLine(chartWarp);
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

    drawStackRect = (stackWarp: any, level) => {
        const _this = this;
        stackWarp.append('rect')
            .attr('class', 'stack_rect')
            .attr('width', d => d.timeRate < 0.01 ? 2 : (this.svgWidth - this.nameWidth) * d.timeRate)
            .attr('height', this.barWidth)
            .attr('x', d => this.xScale(d.from))
            .attr('y', this.barPadding + this.barWidth * (level - 1))
            .attr('fill', function(d) {
                let tempId = parseInt(d.id.split('-')[1]);
                return _this.colors[tempId % 10];
            });

        stackWarp.append('text')
            .text(d => d.name)
            .attr('class', 'stack_rect_text')
            .attr('x', d => this.xScale(d.from) + 5)
            .attr('y', this.barPadding + 12 + this.barWidth * (level - 1))
            .style('display', function(this: any, d: any) {
                let textlen: number = this.getComputedTextLength();
                let w = (_this.svgWidth - _this.nameWidth) * d.timeRate;
                return w - textlen > 10 ? 'block' : 'none';
            });
    }
    // 绘制对应堆栈对应的火焰图区块
    drawStack(threadId, stack, remove = true) {
        const _this = this;
        const { level: parentLevel, id: parentId, child: stackList } = stack;
        const barWarp = d3.select(`#thread_warp_${threadId}`).select('.bar_warp');
        if (barWarp.select(`.sub_stack_rect_warp_${parentId}`).empty()) {
            if (!stackList || stackList.length === 0) return;
            const level = stackList[0].level;
            const subBarWarp = barWarp.select(`.stack_rect_warp_${parentId}`)
            .append('g')
            .attr('class', `sub_stack_rect_warp_${parentId}`)
            
            const stackWarp = subBarWarp.selectAll(`.stack_rect_warp`)
                .data(stackList)
                .enter()
                .append('g')
                .attr('data-threaId', threadId)
                .attr('class', d => `stack_rect_warp_${d.id} stack_rect_warp`);
            this.drawStackRect(stackWarp, level);

            let maxHeight = this.barPadding * 2  + this.barWidth * level;
            if (maxHeight > this.stackHeight[threadId]) {
                this.stackHeight[threadId] = maxHeight;
            } else {
                maxHeight = this.stackHeight[threadId];
            }
            d3.select(`#thread_warp_${threadId}`).attr('height', maxHeight);
            d3.select(`#thread_warp_${threadId}`)
                .select('.thread_rect_warp')
                .transition()
                .attr('height', maxHeight);

            d3.selectAll('.stack_rect_warp').on('click', function(e, stack: any) {
                e.stopPropagation();
                const threadId = d3.select(this).attr('data-threaId');
                _this.drawStack(threadId, stack);
                _this.resetThreadHeight();
            });
            this.addMouseListener();
        } else {
            if (remove) {
                barWarp.select(`.sub_stack_rect_warp_${parentId}`).remove();
                d3.select(`#thread_warp_${threadId}`)
                    .select('.thread_rect_warp')
                    .transition()
                    .attr('height', this.barPadding * 2  + this.barWidth * parentLevel);
                d3.select(`#thread_warp_${threadId}`).attr('height', this.barPadding * 2  + this.barWidth * parentLevel);
                this.resetThreadHeight();
            }
        }
    }
    // 堆栈的tooltip绘制
    addMouseListener() {
        d3.selectAll('.stack_rect').on('mouseenter', (e, d: any) => {
            let { pageX, pageY } = e;
            let dom = `<div class="stack_info">
                        <div class="title">${d.name}</div>
                        <div class="info">
                            <span>开始时间：${d.from}</span>
                            <span>持续时间：${d.time}</span>
                        </div>
                    </div>`;
            d3.select('#tooltip_warp').html(dom);

            const left = pageX < 300 ? pageX + 10 : pageX - 10;
            const tranX = pageX < 300 ? '0%' : '-100%';
            d3.select('#tooltip_warp')
                .attr('class', 'stack_tooltip show_toolip')
                .attr('style', `left: ${left}px;top: ${pageY - 60}px; transform: translate(${tranX}, -50%)`);
        })
        .on('mouseleave', () => {
            d3.select('#tooltip_warp').attr('class', 'stack_tooltip hide_toolip')
        });
    }
    // 递归遍历展开堆栈时调用drawStack
    dataHandle = (threadId, data) => {
        this.drawStack(threadId, data, false);
        if (data.child && data.child.length > 0) {
            _.forEach(data.child, item => {
                this.dataHandle(threadId, item);
            });
        }
    }
    // 展开堆栈
    expandStack(list: any[]) {
        _.forEach(list, item => {
            _.forEach(item.flames, opt => {
                this.dataHandle(item.threadId, opt);
            });
        });
    }
    // 收起堆栈
    collapseStack(list: any[]) {
        _.forEach(list, item => {
            const barWarp = d3.select(`#thread_warp_${item.threadId}`).select('.bar_warp');
            _.forEach(item.flames, opt => {
                barWarp.select(`.sub_stack_rect_warp_${opt.id}`).remove();
                // barWarp.select(`.sub_stack_rect_warp_${opt.id}`).style('display', 'none');
            });
            const minHeight = this.barPadding * 2  + this.barWidth * 1;
            this.stackHeight[item.threadId] = minHeight;
            d3.select(`#thread_warp_${item.threadId}`)
                .select('.thread_rect_warp')
                .transition()
                .attr('height', minHeight);
            d3.select(`#thread_warp_${item.threadId}`).attr('height', minHeight);
        });
        this.resetThreadHeight();
    }
    // 展开或收起所有的堆栈
    expandAllStack(expand: boolean) {
        if (expand) {
            this.expandThreadIds = [];
            this.collapseStack(this.data);
        } else {
            this.expandThreadIds = _.map(this.data, 'threadId')
            this.expandStack(this.data);
            this.resetThreadHeight();
        }
    }

    getTimeStack = (time, data) => {
        _.forEach(data, item => {
            if (item.from < time && time < item.to) {
                item.armed = true;
            }
            if (item.child && item.child.length > 0) {
                this.getTimeStack(time, item.child);
            }
        });
    }

    renderTreeNodes = (threadId, data) => {
        return data && data.map((item) => {
          if (item.child && item.child.length > 0) {
            if (item.armed) {
                return (
                    <TreeNode title={`${item.name}_${threadId}_${item.id}`} key={`${threadId}_${item.id}`}>
                          {this.renderTreeNodes(threadId, item.child)}
                    </TreeNode>
                );
            }
          }
          return item.armed && <TreeNode title={`${item.name}_${threadId}_${item.id}`} key={`${threadId}_${item.id}`} />;
        });
    }
    showArmedStack = (time, data) => {
        const tempStyle: React.CSSProperties = {
            overflowX: 'auto',
            overflowY: 'auto',
            maxHeight: '620px'
        };
        Modal.info({
            title: `${timeNSFormat(time)}`,
            width: 850,
            content: (
                <div style={tempStyle}>
                    <Tree defaultExpandAll>
                        {
                            data.map(item => _.some(item.flames, opt => opt.armed) ? <TreeNode title={item.threadName} key={item.threadId}>
                                {
                                    this.renderTreeNodes(item.threadId, item.flames)
                                }
                            </TreeNode> : null)
                        }
                    </Tree>
                </div>
            ),
            okText: '关闭',
            onOk() {},
        });
    }

    // 绘制time tooltip line
    drawTopTooltip(chartWarp: any) {
        const _this = this;
        let lineWarp = chartWarp.append('g').attr('class', 'tooltip_line_warp').attr('opacity', 0);
        const FirTime = timeNSFormat(this.xScale.invert(this.nameWidth));
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

        lineWarp.on('click', function(e: any) {
            const { offsetX } = e; 
            let time = _this.xScale.invert(offsetX);
            let testData = _.cloneDeep(_this.data);
            _.forEach(testData, item => {
                _this.getTimeStack(time, item.flames);
            });
            _this.showArmedStack(time, testData);
        });
    }

    // tooltip - 鼠标mouseenter和mousemove的公共事件
    private tooltipLineAction(x: number) {
        d3.selectAll('.tooltip_line_warp')
            .attr('opacity', 1)
            .attr('transform', `translate(${x})` );
        let time = timeNSFormat(this.xScale.invert(x));
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

    // 展开堆栈之后线程高度改变之后重新计算每个线程位置的坐标
    resetThreadHeight() {
        let allHeight = 0;
        _.forEach(this.data, (item) => {
            let threadHeight = this.stackHeight[item.threadId];
            d3.select(`#thread_warp_${item.threadId}`).transition().attr('transform', `translate(0, ${allHeight})`);
            allHeight += parseInt(threadHeight); 
        });
        this.chartWarpHeight = allHeight;
        this.svg.select('.tooltip_line_warp').select('.tooltip_line').attr('y2', this.chartWarpHeight);
        this.svg.attr('height', _.sum(_.values(this.stackHeight)));
    }
}
export default Stack;