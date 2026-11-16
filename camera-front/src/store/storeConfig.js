import { createStore, applyMiddleware } from 'redux';
import { persistStore, persistReducer } from 'redux-persist';
// import thunkMiddleware from 'redux-thunk';
import storage from 'redux-persist/lib/storage/session';
import rootReducer from './reducers';

const persistConfig = {
    key: 'root',
    storage: storage,
    whitelist: ['global', 'business'],
};

const persistedReducer = persistReducer(persistConfig, rootReducer);
let createStoreWithMiddleware = applyMiddleware()(createStore);
/**
 * 创建store
 * @param  {[type]} initialState [description]
 * @return {[type]}              [description]
 */
function configureStore(initialState) {
    // console.log('initialState', initialState);
    // const store = createStore(rootReducer, initialState)
    const store = createStoreWithMiddleware(
        persistedReducer,
        initialState,
        window.__REDUX_DEVTOOLS_EXTENSION__ && window.__REDUX_DEVTOOLS_EXTENSION__()
    );
    //redux调试代码
    // window.devToolsExtension && window.devToolsExtension({ actionCreators })

    //热加载,及时跟新reducer
    // if (module.hot) {
    //     // Enable Webpack hot module replacement for reducers
    //     module.hot.accept('./reducers', () => {
    //         const nextReducer = require('./reducers');
    //         store.replaceReducer(nextReducer);
    //     });
    // }
    let persistor = persistStore(store);

    return { store, persistor };
}

export const { persistor, store } = configureStore({});
