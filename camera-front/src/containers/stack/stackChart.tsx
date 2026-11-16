import React, { Component } from 'react';
import _ from 'lodash';
import Stack from './stack';
import { Button } from 'antd';

interface IProps {
    option: any
}
interface IState {
    expandAll: boolean
}
class StackChart extends Component<IProps, IState> {
    constructor(props) {
        super(props);
        this.state = {
            expandAll: false
        };
    }
    stack = new Stack(this.props.option);

    componentDidMount() {
        this.print();
    }
    componentDidUpdate(prevProps: Readonly<IProps>): void {
        if (!_.isEqual(prevProps.option.data, this.props.option.data)) {
            this.stack = new Stack(this.props.option);
            this.print();
        }
    }

    print() {
        const {option} = this.props;
        if (option.data && option.data.length > 0) {
            this.stack.draw();
        }
    }

    handleExpand = () => {
        this.setState((prevState) => ({
            expandAll: !prevState.expandAll
        }));
        this.stack.expandAllStack(this.state.expandAll);
    }

    render() {
        const { expandAll } = this.state;
        return (
            <div className='stack_chart_warp'>
                <div className='stack_chart_header'>
                    <Button size='small' onClick={this.handleExpand}>{expandAll ? '收起' : '展开'}</Button>
                </div>
                
                <div id='stack_chart'>
                    <svg id='top_time_svg'></svg>
                    <div className='main_chart'>
                        <svg id='stack_svg'></svg>
                        <div id='tooltip_warp' className='stack_tooltip'></div>
                    </div>
                </div>
            </div>
            
        )
    }
}

export default StackChart;