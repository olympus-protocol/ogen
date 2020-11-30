import React from 'react';

class ImportMnemonic extends React.Component<{}> {

    render() {
        const words = [];
        for (let i = 1; i < 25; i++) {
            words.push({ name: "word_" + i, label: i + "." })
        }
        return (
            <div className="modal-container">
                <div className="modal-header">
                    <span>Import Mnemonic</span>
                    <span className="fas-icon">times</span>
                </div>
                <div className="modal-content abs-center">
                    <p>Write wallet name</p>
                    <input className="wallet-name mb-3" type="text" name="wallet_name" />
                    <p>Please write each word from your mnemonic phrase in the following input fields. Usually 12 words, but it can be up to 24 words in length.</p>
                    <div className="modal-import-grid">
                        <div className="row">
                            {
                                words.map((word, i) => {
                                    return (
                                        <div className="col-md-3 abs-center" key={i}>
                                            <p className="mr-3">{word.label}</p>
                                            <input type="text" name={word.name}/>
                                        </div>
                                    );
                                })
                            }
                        </div>
                    </div>
                </div>
                <button className="modal-bottom-button">
                    Submit
                </button>
            </div>
        );
    }
}

export default ImportMnemonic;