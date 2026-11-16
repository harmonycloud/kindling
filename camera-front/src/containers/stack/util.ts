import _ from 'lodash';
import * as d3 from 'd3';
import moment from 'moment';

export const textHandle = (text: string, num: number) => {
    if (text && text.length > num) {
        return text.substring(0, num) + '...';
    } else {
        return text;
    }
}

export const timeFormat = function(time: Date) {
    return d3.timeFormat("%M:%S.%L")(time);
};

export const timeNSFormat = (time: number) => {
    let msTime = Math.floor(time / 1000000);
    let usTime = time % 1000000;
    return d3.timeFormat("%M:%S.%L")(new Date(msTime)) + `.${Math.floor(usTime / 1000)}`;
};