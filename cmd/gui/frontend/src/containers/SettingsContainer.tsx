import React, { Component } from 'react';
import AppFrame from '../components/common/AppFrame';

class SettingsContainer extends Component {
    renderBody() {
        return (
            <div id="settings" className="page-container">
                <div className="settings-container">
                    <h3>Settings</h3>
                    <div className="settings-item">
                        <div className="abs-center">
                            <span className="fas-icon" style={{ fontSize: 1.5 + 'em' }}>language</span>
                            <span className="ml-3 mb-1">Language</span>
                        </div>
                        <div className="mr-3 abs-center">
                            <select>
                                <option value="en">English</option>
                                <option value="es">Spanish</option>
                                <option value="de">German</option>
                            </select>
                        </div>
                    </div>
                    <div className="settings-item">
                        <div className="abs-center">
                            <span className="fas-icon" style={{ fontSize: 1.5 + 'em' }}>percent</span>
                            <span className="ml-3 mb-1">Rates Source</span>
                        </div>
                        <div className="mr-3 abs-center">
                            <select>
                                <option value="en">Obol</option>
                            </select>
                        </div>
                    </div>
                    <div className="settings-item">
                        <div className="abs-center">
                            <span className="fas-icon" style={{ fontSize: 1.5 + 'em' }}>project-diagram</span>
                            <span className="ml-3 mb-1">Modify Node URL</span>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
    render() {
        return (
            <AppFrame body={this.renderBody()} header={"Settings"} />
        );
    }
}

export default SettingsContainer;