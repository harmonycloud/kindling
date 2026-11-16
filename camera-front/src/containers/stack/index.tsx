import { useEffect, useState, useRef } from 'react';
import { Upload, Button } from 'antd';
import type { UploadProps } from 'antd';
import _ from 'lodash';
import StackChart from './stackChart';
import './index.less';

import { IOption } from './types';
import list from './mock2.json';

function Stack() {
    const [listData, setListData] = useState<any>(list);
    const [stackData, setStackData] = useState<any[]>([]);
    const [timeRange, setTimeRange] = useState<number[]>([]);

    const dataHandle = (data, times, level, flag) => {
        level++;
        _.forEach(data, (item: any, idx) => {
            flag++;
            item.id = level + '-' + idx + flag;
            item.level = level;
            item.time = item.to - item.from;
            item.timeRate = item.time / (times[1] - times[0]);
            if (item.child && item.child.length > 0) {
                dataHandle(item.child, times, level, flag);
            }
        });
    }
    useEffect(() => {
        let tempTime: any[] = [];
        _.forEach(listData, (item: any) => {
            tempTime = tempTime.concat(_.map(item.flames, 'from'), _.map(item.flames, 'to'));
        });
        const times = [_.min(tempTime), _.max(tempTime)];

        let data = _.cloneDeep(listData);
        _.forEach(data, (item: any) => {
            dataHandle(item.flames, times, 0, 0);
        });
        console.log(data);

        setStackData(data);
        setTimeRange(times)
    }, [listData])

    const props: UploadProps = {
        showUploadList: false,
        beforeUpload(file) {
            // console.log(file);
            return false;
        },
        onChange(info) {
            // console.log(info);
            const fileReader = new FileReader();
            fileReader.onload = () => {
                let result = JSON.parse(fileReader.result as string);
                setListData(result);
                console.log(result);
            }
            fileReader.readAsBinaryString(info.file);
            // if (info.file.status !== 'uploading') {
            //     console.log(info.file, info.fileList);
            // }
            // if (info.file.status === 'done') {
            //     console.log(info.file);
            //     // message.success(`${info.file.name} file uploaded successfully`);
            // } else if (info.file.status === 'error') {
            //     // message.error(`${info.file.name} file upload failed.`);
            // }
        },
    };
    const option: IOption = {
        data: stackData,
        timeRange: timeRange
    }
    return (
        <div className='stack_warp'>
            <header className='stack_header'>
                <div className='stack_header_text'>
                    堆栈分析
                </div>
                <div className='stack_header_operation'>
                    <Upload {...props}>
                        <Button>上传JSON文件</Button>
                    </Upload>
                </div>
            </header>
            <div className='stack_body'>
                <div className='stack_content'>
                    <StackChart option={option}/>
                </div>
            </div>
        </div>
    )
}
export default Stack;