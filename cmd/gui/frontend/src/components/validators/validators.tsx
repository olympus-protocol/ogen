import React from 'react';

interface ValidatorProps {
  balance: number;
  publicKey: number;
  epoch: number;
  status: string;
}

export default function Validator({
  status,
  balance,
  publicKey,
  epoch,
}: ValidatorProps) {
  return (
    <div className="col-lg-6">
      <div className="validator">
        <div className="row">
          <span className="validator-title">Balance</span>
          <span>{balance} POLIS</span>
        </div>
        <div className="row">
          <span className="validator-title">Public Key</span>
          <span>{publicKey}</span>
        </div>
        <div className="row">
          <span className="validator-title">Active Since (epoch)</span>
          <span>{epoch}</span>
        </div>
        <div className="row validator-status">
          <div className={`validator-${status}`} />
        </div>
      </div>
    </div>
  );
}
