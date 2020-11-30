import React from 'react';

export class NewsComponent extends React.Component<any, any> {
    render() {
        return (
            <div className="abs-center">
                <div className="dashboard-news dashboard-item-alt">
                    <h3>Latest News</h3>
                    <div className="row">
                        <div className="dashboard-news-item-main">
                            <div>
                                <img src="https://polispay.org/wp-content/uploads/2020/09/1qZoZFv8RJLoQPrjGXuwctg.png" alt="main" />
                            </div>
                            <div className="abs-center">
                                <div className="dashboard-news-item-main-text">
                                    <p className="dashboard-news-item-main-title">Polis Core 1.6.4 Mandatory Update</p>
                                    <p>by Nadya Castilleja | Aug 18, 2020 </p>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className="row">
                        <div className="dashboard-news-item">
                            <span className="dashboard-news-item-title">Polis Core 1.6.4 Mandatory Update</span>
                            <span>by Nadya Castilleja | Aug 18, 2020 </span>
                        </div>
                    </div>
                    <div className="row">
                        <div className="dashboard-news-item">
                            <span className="dashboard-news-item-title">Polis Core 1.6.4 Mandatory Update</span>
                            <span>by Nadya Castilleja | Aug 18, 2020 </span>
                        </div>
                    </div>
                    <div className="row">
                        <div className="dashboard-news-item">
                            <span className="dashboard-news-item-title">Polis Core 1.6.4 Mandatory Update</span>
                            <span>by Nadya Castilleja | Aug 18, 2020 </span>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}
