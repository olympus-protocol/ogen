import React from 'react';
import { NavLink } from "react-router-dom";

interface Props {
    selected: string
}

const Sidebar: React.FC<Props> = ({ selected }) => {
    return (
        <div className="sidebar">
            <NavLink exact to="/">
                <div className="sidebar-logo abs-center">
                    <img className="brand" src={"/assets/img/olympus-logo.png"} alt="" />
                </div>
            </NavLink>

            <div className="sidebar-logo-alt abs-center">
                <img className="brand" src={"/assets/img/olympus.png"} alt="" />
            </div>
            <ul>
                <NavLink exact to="/" className="sidebar-li" activeClassName="sidebar-li-active">
                    <li><span className="fas-icon">tachometer-alt</span><span className="sidebar-li-text">Dashboard</span></li>
                </NavLink>
                <NavLink exact to="/wallet" className="sidebar-li" activeClassName="sidebar-li-active">
                    <li><span className="fas-icon">wallet</span><span className="sidebar-li-text">Wallet</span></li>
                </NavLink>
                <div className="sidebar-li">
                    <li><span className="fas-icon">users</span><span className="sidebar-li-text">Governance</span></li>
                </div>
                <NavLink exact to="/settings" className="sidebar-li" activeClassName="sidebar-li-active">
                    <li><span className="fas-icon">cog</span><span className="sidebar-li-text">Settings</span></li>
                </NavLink>
                <NavLink exact to="/network" className="sidebar-li" activeClassName="sidebar-li-active">
                    <li><span className="fas-icon">network-wired</span><span className="sidebar-li-text">Network</span></li>
                </NavLink>
            </ul >
        </div >
    );
}

export default Sidebar;