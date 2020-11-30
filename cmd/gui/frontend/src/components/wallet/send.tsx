import React from 'react';

interface Props {

}

const Send: React.FC<Props> = () => {
    return (
        <div id="send" className="wallet-container">
            <div className="send-row">
                <h3>Selected Wallet</h3>
                <p>Wallet 1</p>
            </div>

            <div className="send-row">
                <h3>Amount</h3>
                <div className="abs-center">
                    <input type="number" />
                    <p className="ml-2 mt-3">POLIS</p>
                </div>
            </div>

            <div className="send-row">
                <h3>Recepient Address</h3>
                <div className="abs-center send-address">
                    <input type="text" />
                </div>
            </div>

            <div className="send-row">
                <div className="mt-3 abs-center send-info">
                    <div className="send-info-item">
                        <p className="send-info-title">Fee </p>
                        <p>~0.00000037 POLIS</p>
                    </div>
                    <div className="send-info-item">
                        <p className="send-info-title">After Fee </p>
                        <p>~800.99999963 POLIS</p>
                    </div>
                </div>
            </div>
            
            <div className="send-row">
                <button className="btn btn-white btn-send">
                    SEND
                </button>
            </div>
        </div>
    );
}

export default Send;