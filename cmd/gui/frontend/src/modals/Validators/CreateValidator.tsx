import React from 'react';

class CreateValidator extends React.Component {
    render() {
        return (
            <div id="modal-create-validator" className="modal-container">
                <div className="modal-header">
                    <span>Create Validator</span>
                    <span className="fas-icon">times</span>
                </div>
                <div className="modal-content abs-center">
                    <p>Please paste the output of the following command in the space below.</p>
                    <div className="code-box">
                        <span>genvalidatorkey $VALIDATOR_AMOUNT</span>
                    </div>
                    <textarea id="validator-textarea" name="validator-textarea" rows={10} cols={100}>
                    </textarea>
                </div>
                <button className="modal-bottom-button">
                    Start Generated Validators
                </button>
            </div>
        );
    }
}

export default CreateValidator;