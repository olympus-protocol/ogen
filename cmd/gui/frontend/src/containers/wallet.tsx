import React, { Component } from 'react';
import Frame from '../components/frame/frame';
import History from '../components/wallet/history';
import Receive from '../components/wallet/receive';
import Send from '../components/wallet/send';

interface IState {
    activeTab: number,
    tab1: string,
    tab2: string,
    tab3: string,
}

class Wallet extends Component<{}, IState> {

    constructor(props: any) {
        super(props);
        this.state = {
            activeTab: 0,
            tab1: "wallet-tab-1",
            tab2: "wallet-tab-2",
            tab3: "wallet-tab-3"
        }
    }

    componentDidMount() {
        this.setDefaultTab();
    }
    
    setDefaultTab() {
        let selectedTab = this.state.tab1 + " wallet-tab-active";
        this.setState({
            tab1: selectedTab
        })
    }

    toggleSelect(num: number) {
        let selectedTab = "";
        switch (num) {
            case 0:
                selectedTab = this.state.tab1 + " wallet-tab-active";
                this.setState({
                    activeTab: num,
                    tab1: selectedTab,
                    tab2: "wallet-tab-2",
                    tab3: "wallet-tab-3"
                });
                break;
            case 1:
                selectedTab = this.state.tab2 + " wallet-tab-active";
                this.setState({
                    activeTab: num,
                    tab1: "wallet-tab-1",
                    tab2: selectedTab,
                    tab3: "wallet-tab-3",
                });
                break;
            case 2:
                selectedTab = this.state.tab3 + " wallet-tab-active";
                this.setState({
                    activeTab: num,
                    tab1: "wallet-tab-1",
                    tab2: "wallet-tab-2",
                    tab3: selectedTab
                });
                break;
        }
    }

    renderBody() {
        return (
            <div id="wallet" className="page-container">
                <div className="wallet-tabs">
                    <div className={this.state.tab1} onClick={() => this.toggleSelect(0)}>
                        Transaction History
                    </div>
                    <div className={this.state.tab2} onClick={() => this.toggleSelect(1)}>
                        Send
                    </div>
                    <div className={this.state.tab3} onClick={() => this.toggleSelect(2)}>
                        Receive
                    </div>
                </div>
                {
                    this.state.activeTab === 0 ?
                        <History />
                        : this.state.activeTab === 1 ?
                            <Send />
                            : this.state.activeTab === 2 ?
                                <Receive /> : ''
                }
            </div>
        );
    }
    render() {
        return <Frame body={this.renderBody()} header={"wallet"} />
    }
}

export default Wallet;