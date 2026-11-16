import moment from 'moment';



//存储localStorage
export const setStore = (name: string, content: string) => {
    if (!name) return;
    if (typeof content !== 'string') {
        content = JSON.stringify(content);
    }
    window.localStorage.setItem(name, content);
};
//获取localStorage
export const getStore = (name: string) => {
    if (!name) return;
    return window.localStorage.getItem(name);
};


export const dateTime = 'YYYY-MM-DD HH:mm:ss';
export const defaultTimeInfo = {
    rangeValue: '1day',
    from: moment(new Date().getTime() - 24 * 3600000).format(dateTime),
    to: moment(new Date()).format(dateTime),
    diff: 24 * 60 * 60000,
    isRecent: true
}
// 如果是最近时间，则需要更新from to
export const translateRecentTime = (timeInfo: any) => {
    if (timeInfo.isRecent) {
        return {
            rangeValue: timeInfo.rangeValue,
            from: moment(new Date().getTime() - timeInfo.diff).format(dateTime),
            to: moment(new Date()).format(dateTime),
            diff: timeInfo.diff,
            isRecent: timeInfo.isRecent
        }
    } else {
        return timeInfo;
    }
}