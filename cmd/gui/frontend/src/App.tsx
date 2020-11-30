import React from 'react';
import { BrowserRouter as Router, Route } from 'react-router-dom';
import './assets/css/styles.css';
import 'bootstrap/dist/css/bootstrap.css';

import DashboardContainer from './containers/DashboardContainer'
import WalletContainer from './containers/WalletContainer';
import SettingsContainer from './containers/SettingsContainer';
import NetworkContainer from './containers/NetworkContainer';
import Add from './modals/Wallet/Add';
import Migrate from './modals/Wallet/Migrate';
import ImportMnemonic from './modals/Wallet/ImportMnemonic';
import CreateValidator from './modals/Validators/CreateValidator';

function App() {
  return (
    <div id="App">
      <Router>
        <Route exact path="/" component={DashboardContainer} />
        <Route exact path="/wallet" component={WalletContainer} />
        <Route exact path="/settings" component={SettingsContainer} />
        <Route exact path="/network" component={NetworkContainer} />
        <Route exact path="/modals/add-wallet" component={Add} />
        <Route exact path="/modals/migrate-wallet" component={Migrate} />
        <Route exact path="/modals/import-wallet" component={ImportMnemonic} />
        <Route exact path="/modals/create-validator" component={CreateValidator} />
      </Router>
    </div>
  );
}

export default App;
