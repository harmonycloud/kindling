import { CSSProperties, useEffect, useState } from 'react';
import { Button, Select, Tooltip } from 'antd';
import { ArrowLeftOutlined, ArrowRightOutlined } from '@ant-design/icons';
import RangeCalendar from 'rc-calendar/lib/RangeCalendar';
import TimePickerPanel from 'rc-time-picker/lib/Panel';
import { dateTime } from '@/services/util';
import zhCN from 'rc-calendar/lib/locale/zh_CN';
import 'moment/locale/zh-cn';
import _ from 'lodash';
import moment from 'moment';

import './index.less';
import 'rc-calendar/assets/index.css';
import 'rc-time-picker/assets/index.css';

const { Option } = Select;
type ITime = {
    text: string;
    value: string;
    inputValue: string;
    diff: number;
}
// 时间长度区域select
const timeRange: ITime[] = [
    {
        text: '30分钟',
        value: '30min',
        inputValue: '最近30分钟',
        diff: 30 * 60000
    }, {
        text: '1小时',
        value: '1hour',
        inputValue: '最近1小时',
        diff: 60 * 60000
    }, {
        text: '12小时',
        value: '12hour',
        inputValue: '最近12小时',
        diff: 12 * 60 * 60000
    }, {
        text: '1天',
        value: '1day',
        inputValue: '最近1天',
        diff: 24 * 60 * 60000
    }, {
        text: '3天',
        value: '3day',
        inputValue: '最近3天',
        diff: 3 * 24 * 60 * 60000
    }, {
        text: '7天',
        value: '7day',
        inputValue: '最近7天',
        diff: 7 * 24 * 60 * 60000
    }, {
        text: '自定义',
        value: 'custom',
        inputValue: '2022-04-11 00:00:00 - 2022-04-11 12:12:12',
        diff: 24 * 60 * 60000
    }
]
interface Props {
    timeInfo?: any; // 时间对象
    onChange?: (timeInfo) => void;
    placement?: 'left' | 'right';
}
function CustomDatePicker(props: Props) {
    const { timeInfo } = props;
    const [rangeValue, setRangeValue] = useState<string>('1day');   // 时间长度区域select的value
    const [displayBlock, setDisplayBlock] = useState<'block' | 'none'>('none');   // 自定义时间控件的display
    const [diff, setDiff] = useState<number>(24 * 60 * 60000);   // 当前时间间隔
    const [inputValue, setInputValue] = useState<string>('最近1天');   // input显示的文案
    const [selectTimeRange, setSelectTimeRange] = useState<Array<any>>([moment().subtract(1, 'days'), moment()]);   // 所选时间范围
    const [isRecent, setIsRecent] = useState<boolean>(true);    // 是否是最近时间
    const [isDisableOk, setIsDisableOk] = useState<boolean>(false);  // 时间选择确定按钮是否禁用
    const [preStates, setPreStates] = useState<any>({});     // select中选择自定义时存储之前的状态

    const { placement = 'left' } = props;
    const datePickerStyle: CSSProperties = {
        [placement]: '0px',
        position: 'absolute',
        zIndex: 999,
    };
    /**
     * 改变时间长度区域select
     * 若value选择的是自定义，则显示自定义时间选择组件，并且保存之前的state到preStates，在取消自定义时间选择后要恢复到之前的state
     */
    const changeRange = (value) => {
        const timeRangeItem: ITime = _.find(timeRange, { value }) as ITime;
        let to = Date.now();
        let from = to - timeRangeItem.diff;
        // 若为自定义，保存之前state并更新from to，不使用默认1天，取selectTimeRange中的时间参数
        if (value === 'custom') {
            from = selectTimeRange[0].valueOf();
            to = selectTimeRange[1].valueOf();
            let preStates = {
                rangeValue,
                diff,
                inputValue,
                selectTimeRange,
                isRecent
            }
            setPreStates(preStates);
        }
        setRangeValue(value);
        setDisplayBlock(value === 'custom' ? 'block' : 'none');
        setDiff(to - from);
        setInputValue(value === 'custom' ? `${moment(from).format(dateTime)} - ${moment(to).format(dateTime)}` : timeRangeItem.inputValue);
        setSelectTimeRange([moment(from), moment(to)]);
        setIsRecent(value === 'custom' ? false : true);
        const timeObj = {
            rangeValue: value,
            from: moment(from).format(dateTime),
            to: moment(to).format(dateTime),
            diff: to - from,
            isRecent: value === 'custom' ? false : true
        }
        value !== 'custom' && props.onChange && props.onChange(timeObj);
    }
    /**
     * 回退时间
     * isRecent为最近时间的则要在最新时间减去diff作为回退时间区间
     * isRecent不是最近时间的则在原有时间区间基础上减diff
     */
    const subTime = () => {
        let to, from;
        if (isRecent) {
            to = Date.now() - diff;
            from = to - diff;
            const inputValue = `${moment(from).format(dateTime)} - ${moment(to).format(dateTime)}`;
            setInputValue(inputValue);
            setSelectTimeRange([moment(from), moment(to)]);
            setIsRecent(false);
        } else {
            to = selectTimeRange[1].valueOf() - diff;
            from = to - diff;
            const inputValue = `${moment(from).format(dateTime)} - ${moment(to).format(dateTime)}`;
            setInputValue(inputValue);
            setSelectTimeRange([moment(from), moment(to)]);
        }
        const timeObj = {
            rangeValue: rangeValue,
            from: moment(from).format(dateTime),
            to: moment(to).format(dateTime),
            diff: diff,
            isRecent: false
        }
        props.onChange && props.onChange(timeObj);
    }
    /**
     * 前进时间
     * 根据当前时间区间加diff作为前进时间区间
     * 若to超过当前时间，则rangeValue显示为最近xx，isRecent置为true
     * 若为自定义且to超过当前时间，则to始终为当前时间，from为to - diff
     */
    const addTime = () => {
        const nowTime = Date.now();
        const to = selectTimeRange[1].valueOf() + diff;
        let timeObj: any = {
            rangeValue: rangeValue,
            diff: diff,
            isRecent: false
        }
        if (to >= nowTime) {
            const from = nowTime - diff;
            timeObj.from = moment(from).format(dateTime);
            timeObj.to = moment(nowTime).format(dateTime);
            let inputValue = '';
            if (rangeValue === 'custom') {
                inputValue = `${moment(from).format(dateTime)} - ${moment(nowTime).format(dateTime)}`;
            } else {
                const timeRangeItem: ITime = _.find(timeRange, { value: rangeValue }) as ITime;
                inputValue = timeRangeItem.inputValue;
                setIsRecent(true);
                timeObj.isRecent = true;
            }
            setInputValue(inputValue);
            setSelectTimeRange([moment(from), moment(nowTime)]);
        } else {
            const from = to - diff;
            timeObj.from = moment(from).format(dateTime);
            timeObj.to = moment(to).format(dateTime);
            const inputValue = `${moment(from).format(dateTime)} - ${moment(to).format(dateTime)}`;
            setInputValue(inputValue);
            setSelectTimeRange([moment(from), moment(to)]);
        }
        props.onChange && props.onChange(timeObj);
    }
    /**
     * 选择自定义时间选择的回调
     * 时间不在5分钟～7天之内的，确定按钮置灰
     */
    const handleChange = (data) => {
        if (data.length > 1) {
            const chooseDiff = data[1].valueOf() - data[0].valueOf();
            const minDuration = 5 * 60000;
            const maxDuration = 7 * 24 * 60 * 60000;
            setIsDisableOk(chooseDiff < minDuration || chooseDiff > maxDuration);
        } else {
            setIsDisableOk(false);
        }
        setSelectTimeRange(data);
    }
    // 确定自定义时间选择的回调
    const onStandaloneOk = (data) => {
        if (isDisableOk) return;
        const to = data[1];
        const from = data[0];
        const diff = to.valueOf() - from.valueOf();
        const inputValue = `${from.format(dateTime)} - ${to.format(dateTime)}`
        setDisplayBlock('none');
        setDiff(diff);
        setInputValue(inputValue);
        const timeObj = {
            rangeValue: rangeValue,
            from: from.format(dateTime),
            to: to.format(dateTime),
            diff: diff,
            isRecent: false
        }
        props.onChange && props.onChange(timeObj);
    }
    // 点击遮罩隐藏自定义时间选择组件并回退之前的时间状态
    const cancelRangeCalendar = () => {
        setRangeValue(preStates.rangeValue);
        setDiff(preStates.diff);
        setInputValue(preStates.inputValue);
        setSelectTimeRange(preStates.selectTimeRange);
        setIsRecent(preStates.isRecent);
        setDisplayBlock('none');
    }
    /**
     * 构造回填到选择框内容
     * @param text 当前时间长度的文本
     */
    const selectRender = (text) => {
        return (
            <div className='select_option_warp'>
                <span className='text_content'>{text}</span>
                <span>{inputValue}</span>
            </div>
        )
    }
    useEffect(() => {
        if (!_.isEmpty(timeInfo)) {
            const timeRangeItem = _.find(timeRange, { value: timeInfo.rangeValue }) as ITime;
            let to = moment(timeInfo.to).valueOf();
            let from = moment(timeInfo.from).valueOf();
            let inputValue = timeRangeItem.inputValue;
            if (timeInfo.isRecent) { // 是最近的 更新from to
                to = Date.now();
                from = to - timeInfo.diff;
            } else {
                inputValue = `${timeInfo.from} - ${timeInfo.to}`;
            }
            setRangeValue(timeInfo.rangeValue);
            setDiff(timeInfo.diff);
            setInputValue(inputValue);
            setIsRecent(timeInfo.isRecent);
            setSelectTimeRange([moment(from), moment(to)]);
        }
    }, [timeInfo])
    return (
        <div className='custom_date_warp'>
            <div className='operate_warp'>
                <Select value={rangeValue}
                    style={{ width: 400 }}
                    optionLabelProp='label'
                    showArrow={false}
                    onSelect={changeRange}>
                    {
                        _.map(timeRange, (item, index) => (
                            <Option key={index} value={item.value} label={selectRender(item.text)}>
                                <div className='select_option_warp'>
                                    <span className='text_content'>{item.text}</span>
                                    <span>{item.value === 'custom' ? '从自定义时间选择器中选择' : item.inputValue}</span>
                                </div>
                            </Option>
                        ))
                    }
                </Select>
                <Tooltip placement='bottom' title={`查看后移${(_.find(timeRange, { value: rangeValue }) as ITime).text}时间`}>
                    <Button icon={<ArrowLeftOutlined />} className='f-ml8' onClick={subTime} />
                </Tooltip>
                <Tooltip placement='bottom' title={`查看前移${(_.find(timeRange, { value: rangeValue }) as ITime).text}时间`}>
                    <Button icon={<ArrowRightOutlined />} disabled={isRecent} onClick={addTime} className='f-ml8' />
                </Tooltip>
            </div>
            <div className={isDisableOk ? 'calender_disable_ok' : ''} style={{ display: displayBlock, ...datePickerStyle }}>
                <div style={{ position: 'relative' }}>
                    <RangeCalendar
                        showToday={false}
                        showWeekNumber={false}
                        disabledDate={(current) => current && current >= moment()}
                        selectedValue={selectTimeRange}
                        onOk={onStandaloneOk}
                        locale={zhCN}
                        showOk
                        // showClear
                        onChange={handleChange}
                        format={dateTime}
                        timePicker={<TimePickerPanel />}
                    />
                    <div className='out_range_tip' style={{ display: isDisableOk ? 'block' : 'none' }}>自定义时间选择范围在5分钟~7天之内</div>
                </div>
            </div>
            <div className='date-mask' style={{ display: displayBlock }} onClick={cancelRangeCalendar}></div>
        </div>
    )
}

export default CustomDatePicker;