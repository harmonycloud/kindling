import React from 'react';
import ReactEcharts from 'echarts-for-react';
import * as echarts from 'echarts';
import { Empty } from 'antd';
import { isEqual } from 'lodash';
import 'echarts/lib/component/legend';


const emptyStyle = {
    margin: 0,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center'
};

function Chart(props) {
    const {
        option,
        height = '100%',
        width = '100%',
        events = null,
        noDataImg = null,
        marginTop = '',
    } = props;

    return (
        option.series.length > 0 && option.series[0].data.length > 0 ? (
            <ReactEcharts
                option={option}
                style={{ height, width, marginTop }}
                onEvents={events}
                notMerge></ReactEcharts>
        ) : (
            <Empty
                image={noDataImg || Empty.PRESENTED_IMAGE_SIMPLE}
                style={{ height, ...emptyStyle }}
                description=''></Empty>
        )
    );
}

function dataEqual(prevProps, nextProps) {
    return isEqual(prevProps.option, nextProps.option);
}
export default React.memo(Chart, dataEqual);
// export default Chart;