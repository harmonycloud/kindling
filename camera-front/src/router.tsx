
import { Routes, Route, Navigate } from "react-router-dom";
import HomeWarp from './containers';
import Thread from './containers/thread';
import Stack from './containers/stack';

const routes = (
    <Routes>
        <Route path="/" element={<HomeWarp />}>
            <Route path="/" element={<Navigate to="thread"/>}></Route>
            <Route path="thread" element={<Thread />}></Route>
            <Route path="stack" element={<Stack />}></Route>
        </Route>
    </Routes>
);

export default routes;
