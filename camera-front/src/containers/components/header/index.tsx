import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Menu, message, Switch } from 'antd';
import { RadarChartOutlined, LoadingOutlined } from '@ant-design/icons';
import style from './header.module.less';

import logo from '@/assets/images/logo.png';
import { toggleProfile } from '@/request';
import { getStore, setStore } from '@/services/util';

interface IProps {
    history: any;
    location: any;
}
interface IState {
    activeMenuKey: string;
}
function Header(props: any) {
    const navigate = useNavigate();
    const [theme, setTheme] = useState<string>(getStore('theme') as string | 'light');
    const [activeMenuKey, setActiveMenuKey] = useState<string>('');
    const [profileStatus, setProfileStatus] = useState<boolean>(false);
    const [loading, setLoading] = useState<boolean>(false);

    const clickMenu = ({ key }: {key: string}) => {
        let path = '';
        switch (key) {
            case 'thread':
                path = '/thread';
                break;
            case 'stack':
                path = '/stack';
                break;
            default:
                path = '/';
        }
        navigate(path);
    }

    const requestProfile = (operation) => {
        const params = {
            operation: operation
        };
        setLoading(true);
        toggleProfile(params).then(res => {
            setLoading(false);
            if (res.data.Code === 1) {
                if (operation === 'status') {
                    setProfileStatus(res.data.Msg !== 'stopped');
                } else {
                    setProfileStatus(operation !== 'stop');
                }
            } else {
                message.warning(res.data.Msg);
            }
        });

    }
    
    const changeProfileStatus = (value) => {
        let operation = value ? 'start' : 'stop';
        requestProfile(operation);
    }
    const changeTheme = (value) => {
        let t = value ? 'dark' : 'light';
        setTheme(t);
        setStore('theme', t);
        let body = document.getElementsByTagName('body')[0];
        body.className = `${t}-theme`;
    }

    useEffect(() => {
        requestProfile('status');
    }, []);

    return (
        <div className={style.home_header}>
            <div className={style.home_header_left}>
                <img src={logo} height='30px' alt='' />
                <Menu mode='horizontal' selectedKeys={[activeMenuKey]} style={{ border: 'none' }} onClick={clickMenu} overflowedIndicator={null}>
                    <Menu.Item key="thread" icon={<RadarChartOutlined />}>
                        线程分析
                    </Menu.Item>
                    {/* <Menu.Item key="stack" icon={<RadarChartOutlined />}>
                        堆栈分析
                    </Menu.Item> */}
                </Menu>
            </div>
            <div className={style.home_header_right}>
                <span>启动Trace检测：</span>
                <Switch checked={profileStatus} onChange={changeProfileStatus}></Switch>
                {
                    loading && <LoadingOutlined style={{ marginLeft: 10 }}/>
                }
                &nbsp;&nbsp;&nbsp;&nbsp;
                {/* <Switch checkedChildren='暗黑' unCheckedChildren='明亮' checked={theme ==='dark'} onChange={changeTheme}></Switch> */}
            </div>
        </div>
    );
}

export default Header;