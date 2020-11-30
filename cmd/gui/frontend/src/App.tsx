import React from 'react';
import { BrowserRouter as Router, Route } from 'react-router-dom';
import './theme/css/styles.css';
import 'bootstrap/dist/css/bootstrap.css';

import DashboardContainer from './containers/dashboard'
import Wallet from './containers/wallet';
import Settings from './containers/settings';
import Network from './containers/network';
import CreateWallet from './modals/wallet/create-wallet';
import Migrate from './modals/wallet/migrate';
import ImportWallet from './modals/wallet/import-wallet';
import CreateValidator from './modals/validators/CreateValidator';

function App() {
  return (
    <div id="App">
      <Router>
        <Route exact path="/" component={DashboardContainer} />
        <Route exact path="/wallet" component={Wallet} />
        <Route exact path="/settings" component={Settings} />
        <Route exact path="/network" component={Network} />
        <Route exact path="/modals/add-wallet" component={CreateWallet} />
        <Route exact path="/modals/migrate-wallet" component={Migrate} />
        <Route exact path="/modals/import-wallet" component={ImportWallet} />
        <Route exact path="/modals/create-validator" component={CreateValidator} />
      </Router>
    </div>
  );
}

export default App;
