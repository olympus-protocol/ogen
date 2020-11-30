import React, { Component } from 'react';
import AppFrame from '../components/common/AppFrame';
import Balance from '../components/Dashboard/Balance';
import DAO from '../components/Dashboard/DAO';
import News from '../components/Dashboard/News';
import TxHistory from '../components/Dashboard/TxHistory';

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
                            <DAO />
                        </div>
                    </div>

                    <div className="row">
                        <div className="col-lg-6">
                            <TxHistory />
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
            <AppFrame body={this.renderBody()} header={"Dashboard"} />
        );
    }
}

export default DashboardContainer;