import { Component } from 'react';
import { Outlet } from "react-router-dom";
import Header from './components/header';

class HomeWarp extends Component {
    render() {
        return (
            <div className="home">
                <Header />
                <div className="home_content">
                    <Outlet />
                </div>
            </div>
        );
    }
}

export default HomeWarp;