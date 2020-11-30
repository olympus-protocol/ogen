import React from 'react';
import Dropzone from 'react-dropzone';

class Migrate extends React.Component<{}> {
    render() {

        return (
            <div className="modal-container">
                <div className="modal-header">
                    <span>Migrate From Polis Core</span>
                    <span className="fas-icon">times</span>
                </div>
                <div className="modal-content">
                    <p>Retrieve the xxxxxx.yyy file from $destination and select it on the dialog.</p>
                    <Dropzone onDrop={acceptedFiles => console.log(acceptedFiles)}>
                        {({ getRootProps, getInputProps }) => (
                            <section className="dropzone">
                                <div {...getRootProps()}>
                                    <input {...getInputProps()} />
                                    <p>Drag 'n' drop some files here, or click to select files</p>
                                </div>
                            </section>
                        )}
                    </Dropzone>
                </div>
            </div>
        );
    }
}

export default Migrate;