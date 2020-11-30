import React, { Component } from 'react';
import Frame from '../components/frame/frame';
import Balance from '../components/dashboard/balance';
import Dao from '../components/dashboard/dao';
import News from '../components/dashboard/news';
import History from '../components/dashboard/history';

class DashboardContainer extends Component {
    renderBody() {
        return (
            <div id="dashboard" className="page-container">
                <div className="dashboard-container">
                    <div className="row">
                        <div className="col-lg-6">
                            <Balance />
                        </div>
                        <div className="col-lg-6">
                            <Dao />
                        </div>
                    </div>

                    <div className="row">
                        <div className="col-lg-6">
                            <History />
                        </div>
                        <div className="col-lg-6">
                            <News />
                        </div>
                    </div>
                </div>
            </div>
        );
    }
    render() {
        return (
            <Frame body={this.renderBody()} header={"dashboard"} />
        );
    }
}

export default DashboardContainer;