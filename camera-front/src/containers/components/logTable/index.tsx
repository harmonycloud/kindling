import React from 'react';
import isArray from 'lodash/isArray';
import _ from 'lodash';
import style from './index.module.less';

interface IProps {
    list: any[]
}
class LogTable extends React.Component<IProps, any> {
    constructor(props) {
        super(props);
        this.state = {
            list: [],
            ifListObj: false,
            listObj: {}
        };
    }
    UNSAFE_componentWillMount() {
        this.initData(this.props.list);
    }
    UNSAFE_componentWillReceiveProps(nextpProps) {
        if(!_.isEqual(nextpProps.list,this.props.list))
            this.initData(nextpProps.list);
    }
    initData = (list) => {
        if (isArray(list)) {
            this.setState({
                list: list,
                ifListObj: false
            });
        } else {
            this.setState({
                list: list.result ? list.result : [],
                listObj: list,
                ifListObj: true
            });
        }
    }

    expandLog = () => {
        let { listObj } = this.state;
        listObj.expand = !listObj.expand;
        this.setState({
            list: listObj.expand ? listObj.msgList : listObj.sliceMsgList,
            listObj: listObj
        });
    }
    render() {
        return (
            <table className={style.log_table}>
                <tbody>
                    {
                        this.state.list.length <= 1 && _.isEmpty(this.state.list[0])===true ?
                        <tr key={1}>
                            <td>
                                <span className={style.index}>1</span>
                                <span className={style.err}>暂无数据</span>
                            </td>
                        </tr>
                        :
                        this.state.list.map((item, index) => <tr key={index}>
                            <td>
                                <span className={style.index}>{index + 1}</span>
                                <span className={style.err}>{item}</span>
                            </td>
                        </tr>)
                    }
                    {
                        (this.state.ifListObj && this.state.listObj.ifExpand) ? <tr>
                            <td className={style.log_expand}>
                                <span onClick={this.expandLog}>
                                    {
                                        this.state.listObj.expand ? '收起日志列表' : '展开所有日志列表'
                                    }
                                </span>
                            </td>
                        </tr> : null
                    }
                </tbody>
            </table>
        );
    }
}
export default LogTable;