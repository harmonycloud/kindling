import { createAction, handleActions } from 'redux-actions';
import { getStore, defaultTimeInfo } from '@/services/util';

export const types = {
    SET_UserInfo: 'SET_UserInfo',
    SET_Theme: 'SET_Theme',
    SET_TimeInfo: 'SET_TimeInfo'
};

const reducers: any = {};

interface IState {
    theme: string;
    timeInfo: any;
}

const theme = getStore('theme') || 'light';
const initialState: IState = {
    theme: theme,
    timeInfo: defaultTimeInfo
};


reducers[types.SET_TimeInfo] = (state: any, action: any) => {
    return { ...state, ...action.payload };
};

export const global = handleActions(reducers, initialState);

export const setTheme = createAction(types.SET_Theme);
export const setTimeInfo = createAction(types.SET_TimeInfo);