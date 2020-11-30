import React from 'react';

export default class HistoryComponent extends React.Component<any, any> {
    render() {
        return (
            <div className="dashboard-txhistory dashboard-item-alt">
                <h3>Transaction History</h3>
                <div className="table-responsive">
                    <table className="table">
                        <thead className="thead-blue">
                            <tr>
                                <th scope="col"></th>
                                <th scope="col">Date</th>
                                <th scope="col">Destination Address</th>
                                <th scope="col">Amount (POLIS)</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <th scope="row">
                                    <span className="fas-icon">arrow-circle-down</span>
                                </th>
                                <td>14/09/2020 05:04</td>
                                <td>(PVqWvxqvNNyy1d9ns4CH6CBhjZEpZLDLN3)</td>
                                <td>2.304000</td>
                            </tr>
                            <tr>
                                <th scope="row">
                                    <span className="fas-icon">arrow-circle-down</span>
                                </th>
                                <td>14/09/2020 05:04</td>
                                <td>(PVqWvxqvNNyy1d9ns4CH6CBhjZEpZLDLN3)</td>
                                <td>2.304000</td>
                            </tr>
                            <tr>
                                <th scope="row">
                                    <span className="fas-icon">arrow-circle-down</span>
                                </th>
                                <td>14/09/2020 05:04</td>
                                <td>(PVqWvxqvNNyy1d9ns4CH6CBhjZEpZLDLN3)</td>
                                <td>2.304000</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        );
    }
}
