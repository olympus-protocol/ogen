import React from 'react';
import QR from '../../assets/img/qr.svg';
import ArrowRight from '../../assets/img/arrow-right-solid.svg';
interface Props {

}

const TxHistory: React.FC<Props> = ({ }) => {
    return (
        <div id="receive" className="wallet-container abs-center">
            <div>
                <h1>Wallet 1</h1>
                <div className="abs-center">
                    <div className="receive-img">
                        <img src={QR} alt="" />
                    </div>
                </div>
                <div className="abs-center receive-row">
                    <p className="mr-3">Set Amount</p>
                    <div className="input-group mb-3 abs-center">
                        <input type="text" className="form-control receive-amount" aria-label="Receive Amount" aria-describedby="receive-amount" />
                        <div className="input-group-append">
                            <div className="btn btn-outline-secondary" >
                                <span className="fas-icon">arrow-right</span>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="receive-row">
                    <p className="mr-3">PXF6vvX9VyNJGVn6Hyeut7wcaCUFbNnwzB</p>
                    <span className="far-icon">clipboard</span>
                </div>
            </div>
        </div>
    );
}

export default TxHistory;