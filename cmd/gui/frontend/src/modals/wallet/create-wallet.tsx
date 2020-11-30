import React from 'react';

class CreateWallet extends React.Component<{}>{
    render() {
        return (
            <div id="add-wallet-modal" className="modal-container">
                <div className="modal-header">
                    <span>Add Wallet</span>
                    <span className="fas-icon">times</span>
                </div>
                <div className="modal-content abs-center">
                    <button className="add-btn btn btn-blue">Create Wallet</button>
                    <span>- or -</span>
                    <button className="add-btn btn btn-blue">Migrate from Polis Core</button>
                    <button className="add-btn btn btn-blue">Import Mnemonic</button>
                </div>
            </div>
        );
    }
}

export default CreateWallet;