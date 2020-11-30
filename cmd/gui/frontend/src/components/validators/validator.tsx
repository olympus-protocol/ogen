import React from 'react';

interface ValidatorProps {
    balance: number,
    publicKey: number,
    epoch: number,
    status: string
}

export default class Validator extends React.Component<ValidatorProps, any> {
    render() {
        const { status } = this.props
        return (
            <div className="col-lg-6">
                <div className="validator">
                    <div className="row">
                        <span className="validator-title">Balance</span><span>{this.props.balance} POLIS</span>
                    </div>
                    <div className="row">
                        <span className="validator-title">Public Key</span><span>{this.props.publicKey}</span>
                    </div>
                    <div className="row">
                        <span className="validator-title">Active Since (epoch)</span><span>{this.props.epoch}</span>
                    </div>
                    <div className="row validator-status">
                        <div className={"validator-" + status} />
                    </div>
                </div>
            </div>
        );
    }
}
