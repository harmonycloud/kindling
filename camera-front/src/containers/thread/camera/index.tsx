import React from 'react';
import { Button, Switch, Empty, Tooltip, message } from 'antd';
import { GatewayOutlined, RollbackOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import tooltipPng from '@/assets/images/tooltip.png';
import tooltipPng2 from '@/assets/images/tooltip2.png';
import _ from 'lodash';
import './index.less';
import { IOption } from './types';
import Camera from './draw2';
import EasyCamera from './easyDraw';
import { toggleProfile } from '@/request';

interface IProps {
    option: IOption;
    onViewChange: (type: 'complex' | 'sample') => void;
    onEventDetail: (show: boolean) => void;
    style?: React.CSSProperties;
}
interface IState {
    supportAddLine: boolean;
    supportTrace: boolean;
    supportBrush: boolean;
    showJavaLock: boolean;
    showRunq: boolean;
    showLog: boolean;
    showEventChart: boolean;
    installLoading: boolean;
    events: string[];
}
class CameraWarp extends React.Component<IProps, IState> {
    constructor(props: IProps) {
        super(props);
        this.state = {
            supportAddLine: false,
            supportBrush: false,
            supportTrace: true,
            showRunq: true,
            showJavaLock: true,
            showLog: true,
            showEventChart: false,
            installLoading: false,
            events: ['on', 'net', 'file']
        }
    }
    camera = new Camera({ ...this.props.option, parentRef: this });
    easyCamera = new EasyCamera({ ...this.props.option, parentRef: this });
    observer: any = null;

    componentDidMount() {
        this.print();
        this.addMutationObserver();
    }
    componentDidUpdate(prevProps: Readonly<IProps>): void {
        if (!_.isEqual(prevProps.option.data, this.props.option.data)) {
            this.camera = new Camera({ ...this.props.option, parentRef: this });
            this.easyCamera = new EasyCamera({ ...this.props.option, parentRef: this });
            this.setState({
                supportAddLine: false,
                supportBrush: false,
                supportTrace: true,
                showJavaLock: true,
                showLog: true,
                showEventChart: false
            }, () => {
                this.print();
            })
        }
        if (prevProps.option.showComplex !== this.props.option.showComplex) {
            this.print();
        }
    }
    componentWillUnmount() {
        this.setState = () => {
            return;
        };
        if (this.observer) {
            this.observer = null;
        }
    }

    addMutationObserver = () => {
        let MutationObserver = window.MutationObserver;
        // || window.WebKitMutationObserver || window.MozMutationObserver;
        let element = document.getElementById('right_thread_warp')
        this.observer = new MutationObserver((mutationList) => {
            const { option } = this.props;
            if (option.showComplex) {
                this.camera.changeSize();
            } else {
                this.easyCamera.changeSize();
            }
        });
        this.observer.observe(element, { attributes: true, attributeFilter: ['style', 'class'], attributeOldValue: true })
    }

    print = () => {
        const { option } = this.props;
        // console.log('print', option.showComplex, option.data);
        if (option.data && option.data.length > 0) {
            if (option.showComplex) {
                document.getElementById('event_chart')!.style.display = 'block';
                this.camera.draw();
            } else {
                this.easyCamera.draw();
            }
        }
    }

    toggleSupportAddLine = () => {
        this.setState((prevState) => ({
            supportAddLine: !prevState.supportAddLine
        }));
        this.state.supportAddLine ? this.camera.removeAddLine() : this.camera.supportAddLine();
    }
    traceAnaliysis = () => {
        if (this.props.option.data.length > 0) {
            this.setState((prevState) => ({
                supportTrace: !prevState.supportTrace
            }), () => {
                this.camera.startTrace(this.state.supportTrace);
            })
        }
    }
    closeTraceAnaliysis = () => {
        this.setState({
            supportTrace: false
        });
    }
    toggleChartBrush = () => {
        this.camera.addChartBrush();
        // this.setState((prevState) => ({
        //     supportBrush: !prevState.supportBrush
        // }));
        // this.state.supportBrush ? this.camera.removeChartBrush() : this.camera.addChartBrush();
    }
    resetChartBrush = () => {
        this.camera.resetChartBrush()
    }
    toggleRunqBtn = () => {
        this.setState((prevState) => ({
            showRunq: !prevState.showRunq
        }));
        if (this.props.option.showComplex) {
            this.state.showRunq ? this.camera.hideRunQ() : this.camera.showRunQ();
        } else {
            this.state.showRunq ? this.easyCamera.hideRunQ() : this.easyCamera.showRunQ();
        }
    }
    toggleJavaLockBtn = () => {
        this.setState((prevState) => ({
            showJavaLock: !prevState.showJavaLock
        }));
        if (this.props.option.showComplex) {
            this.state.showJavaLock ? this.camera.hideJavaLock() : this.camera.showJavaLock();
        } else {
            this.state.showJavaLock ? this.easyCamera.hideJavaLock() : this.easyCamera.showJavaLock();
        }
    }
    toggleLogBtn = () => {
        this.setState((prevState) => ({
            showLog: !prevState.showLog
        }));
        if (this.props.option.showComplex) {
            this.state.showLog ? this.camera.hideLog() : this.camera.showLog();
        } else {
            this.state.showLog ? this.easyCamera.hideLog() : this.easyCamera.showLog();
        }
    }

    showSpanChart = () => {
        this.props.onEventDetail(false);
        this.setState({ showEventChart: false })
        this.easyCamera.hideEventChart();
    }

    changeEvents = (event) => {
        let { events } = this.state;
        if (events.indexOf(event) > -1) {
            events.splice(events.indexOf(event), 1);
        } else {
            events.push(event)
        }
        this.setState({ events });
        this.easyCamera.changeSubEventType(events);
    }

    installProfile = () => {
        this.setState({ installLoading: true });
        const params = {
            operation: 'start_attach_agent',
            pid: this.props.option.trace.labels.pid
        };
        toggleProfile(params).then(res => {
            if (res.data.Code === 1) {
                message.success('启动成功');
            } else {
                message.warning(res.data.Msg);
            }
            this.setState({ installLoading: false });
        });
    }

    render() {
        const { option, onViewChange } = this.props;
        const { supportAddLine, supportBrush, showRunq, showJavaLock, showLog, supportTrace, showEventChart, installLoading, events } = this.state;
        return (
            <div id="camera_chart_warp">
                <div className='header'>
                    {
                        option.showComplex ? <div>
                            <Button size="small" onClick={this.toggleSupportAddLine} style={{ marginRight: 10 }}>{supportAddLine ? '取消关键时刻' : '添加关键时刻'}</Button>
                            <Button size="small" onClick={this.traceAnaliysis} type={supportTrace ? 'primary' : 'default'} ghost={supportTrace} style={{ marginRight: 10 }}>
                                Trace分析
                                <Tooltip title="开启Trace分析后，线程轴仅展示与本次请求业务关联的线程，收起其他系统线程"><QuestionCircleOutlined /></Tooltip>
                            </Button>
                            <Button size="small" onClick={() => onViewChange('sample')}>简易视图</Button>
                        </div> : <div>
                            <Button size="small" onClick={() => onViewChange('complex')} style={{ marginRight: 10 }}>复杂视图</Button>
                            {
                                showEventChart && <Button size="small" onClick={this.showSpanChart}>返回</Button>
                            }
                        </div>
                    }
                    <div className='toggle_print'>
                        {
                            option.showComplex && <React.Fragment>
                                <Tooltip title="区域缩放">
                                    <GatewayOutlined className={`operate_icon ${supportBrush && 'active'}`} onClick={this.toggleChartBrush} />
                                </Tooltip>
                                <Tooltip title="区域缩放还原">
                                    <RollbackOutlined className='operate_icon' onClick={this.resetChartBrush} />
                                </Tooltip>
                            </React.Fragment>
                        }
                        {
                            showEventChart && !option.showComplex && <React.Fragment>
                                <span>CPU事件</span>
                                <Switch checked={events.indexOf('on') > -1} onChange={() => this.changeEvents('on')} size='small' />
                                <span>Net事件</span>
                                <Switch checked={events.indexOf('net') > -1} onChange={() => this.changeEvents('net')} size='small' />
                                <span>File事件</span>
                                <Switch checked={events.indexOf('file') > -1} onChange={() => this.changeEvents('file')} size='small' />
                            </React.Fragment>
                        }
                        {
                            option.showComplex || showEventChart ? <React.Fragment>
                                <span>runq</span>
                                <Switch checked={showRunq} onChange={this.toggleRunqBtn} size='small' />
                                <span>Java Lock</span>
                                <Switch checked={showJavaLock} onChange={this.toggleJavaLockBtn} size='small' />
                                <span>Log事件</span>
                                <Switch checked={showLog} onChange={this.toggleLogBtn} size='small' />
                            </React.Fragment> : null
                        }
                        <div className='operate_tooltip'>
                            <QuestionCircleOutlined className='operate_icon' />
                            <img alt='图例' src={option.showComplex ? tooltipPng : tooltipPng2} />
                        </div>
                    </div>
                </div>
                <div id='camera' className='camera_charts'>
                    <span className='chart_title'>{(option.showComplex || showEventChart) ? 'Trace执行线程事件' : 'Trace中span执行消耗'}</span>
                    {
                        (option.data && option.data.length > 0) ? <>
                            <svg id='top_time_svg'></svg>
                            {
                                option.showComplex ? null : (option.spanList.length > 0 ? <div id='span_chart'>
                                    <svg id='span_svg'></svg>
                                </div> : <div className='empty_span_warp'>
                                    <div className='empty_span_text'>未检测到支持Span分析的Tracing探针，为不影响使用，请您及时安装。(点击查看<a href='http://kindling.harmonycloud.cn/docs/reference/kindling-java-agent/kindling-java-compatibility/' target="_blank">支持的探针列表</a>)</div>
                                    <div className='empty_span_text'>或点击下方按钮，让Trace Profiling自动为您的应用(目前只支持Java应用)安装探针。</div>
                                    <Button type="primary" onClick={this.installProfile} loading={installLoading}>立即自动安装</Button>
                                </div>)
                            }
                            <div className='main_chart' id="event_chart" style={option.showComplex ? {} : { height: '60%' }}>
                                <svg id='camera_svg'></svg>
                            </div>
                            <div className='bottom_xaxis_warp' style={option.showComplex ? {} : { height: '35px' }}>
                                <svg id='bottom_xaxis_svg'></svg>
                            </div>
                            <div id='tooltip_warp' className='camera_tooltip'></div>
                        </> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
                    }
                </div>
            </div>
        )
    }

}

// function dataEqual(prevProps, nextProps) {
//     return _.isEqual(prevProps.option.data, nextProps.option.data);
// }
// export default React.memo(forwardRef(CameraWarp), dataEqual);
export default CameraWarp;