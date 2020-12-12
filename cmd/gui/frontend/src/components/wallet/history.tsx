import React from 'react';

export default function History() {
  return (
    <div id="tx-history" className="wallet-container">
      <div className="input-group mb-3">
        <label className="mr-3 mt-2" htmlFor="txSearch">
          Search
        </label>
        <input
          type="text"
          className="form-control"
          placeholder="Look by date, transaction type, address, label, etc..."
          aria-label="txSearch"
          aria-describedby="txSearch"
        />
        <div className="input-group-append">
          <div id="txSearch" className="btn btn-outline-secondary">
            <span className="fas-icon">search</span>
          </div>
        </div>
      </div>
      <div className="table-responsive">
        <table className="table">
          <thead className="thead-blue">
            <tr>
              <th scope="col" />
              <th scope="col">Date</th>
              <th scope="col">Type</th>
              <th scope="col">Destination Address</th>
              <th scope="col">Amount (POLIS)</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <th scope="row">
                <span className="fas-icon">arrow-circle-down</span>
              </th>
              <td>14/09/2020 05:04</td>
              <td>Payment to yourself</td>
              <td>(PVqWvxqvNNyy1d9ns4CH6CBhjZEpZLDLN3)</td>
              <td>2.304000</td>
            </tr>
            <tr>
              <th scope="row">
                <span className="fas-icon">arrow-circle-down</span>
              </th>
              <td>14/09/2020 05:04</td>
              <td>Payment to yourself</td>
              <td>(PVqWvxqvNNyy1d9ns4CH6CBhjZEpZLDLN3)</td>
              <td>2.304000</td>
            </tr>
            <tr>
              <th scope="row">
                <span className="fas-icon">arrow-circle-down</span>
              </th>
              <td>14/09/2020 05:04</td>
              <td>Payment to yourself</td>
              <td>(PVqWvxqvNNyy1d9ns4CH6CBhjZEpZLDLN3)</td>
              <td>2.304000</td>
            </tr>
            <tr>
              <th scope="row">
                <span className="fas-icon">arrow-circle-down</span>
              </th>
              <td>14/09/2020 05:04</td>
              <td>Payment to yourself</td>
              <td>(PVqWvxqvNNyy1d9ns4CH6CBhjZEpZLDLN3)</td>
              <td>2.304000</td>
            </tr>
            <tr>
              <th scope="row">
                <span className="fas-icon">arrow-circle-down</span>
              </th>
              <td>14/09/2020 05:04</td>
              <td>Payment to yourself</td>
              <td>(PVqWvxqvNNyy1d9ns4CH6CBhjZEpZLDLN3)</td>
              <td>2.304000</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div className="row export-btn">
        <button className="btn btn-blue">Export .csv data</button>
      </div>
    </div>
  );
}
