import React from 'react';
import { useWalletState } from '../../state/wallets/hooks';

export default function Balance() {
  const { balance } = useWalletState();

  return (
    <div className="dashboard-item dashboard-balance abs-center">
      <div>
        <span style={{ fontSize: `${1.5}em` }}>Balance</span>
        <br />
        <div className="dashboard-balance-info">
          <span style={{ fontSize: `${2.5}em` }}>
            {balance?.Confirmed} POLIS
          </span>
          <br />
          <span style={{ fontSize: `${1.5}em` }}>(703.665342 USD)</span>
        </div>
        <span>Unconfirmed: 1240.000 POLIS</span>
        <br />
        <span>Latest Block: 684161</span>
      </div>
    </div>
  );
}
