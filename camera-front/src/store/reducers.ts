
import { combineReducers } from 'redux';
import { global } from './modules/global';


const rootReducer = combineReducers({
    config: (state = {}) => state,
    global
});
export default rootReducer;
