import React, { Component } from 'react';
import Frame from '../components/Frame';
import Validator from '../components/Validators';

class Network extends Component {
  render() {
    return (
      <Frame header="Network">
        <div id="network" className="page-container">
          <div className="row network-header">
            <h3>Network Information</h3>
            <button className="btn btn-blue">Create Validator</button>
          </div>
          <div className="row">
            <div className="col network-info">
              <div className="row">
                <div className="col abs-center">
                  <div>
                    <p>
                      Block Height
                      <span>273,762</span>
                    </p>
                    <p>
                      Network Participation
                      <span>75%</span>
                    </p>
                    <p>
                      No. of Validators
                      <span>0/128</span>
                    </p>
                  </div>
                </div>
                <div className="col abs-center">
                  <div>
                    <p>
                      Block Height
                      <span>273,762</span>
                    </p>
                    <p>
                      Network Participation
                      <span>75%</span>
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <div className="abs-center">
              <div className="col network-info network-summary abs-center">
                <div>
                  <div className="row">
                    <div className="col">Validator Summary</div>
                  </div>
                  <div className="row">
                    <div className="col">
                      <p className="vsum-a">ACTIVE</p>
                      <p>0</p>
                    </div>
                    <div className="col">
                      <p className="vsum-s">STARTING</p>
                      <p>0</p>
                    </div>
                    <div className="col">
                      <p className="vsum-e">ERROR</p>
                      <p>0</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div className="validators">
            <div className="row">
              <h3>Validators</h3>
            </div>
            <div className="row">
              <Validator
                balance={100}
                publicKey={1314234234333313143}
                epoch={412}
                status="active"
              />
              <Validator
                balance={100}
                publicKey={1314234234333313143}
                epoch={412}
                status="starting"
              />
              <Validator
                balance={100}
                publicKey={1314234234333313143}
                epoch={412}
                status="error"
              />
            </div>
          </div>
        </div>
      </Frame>
    );
  }
}

export default Network;
