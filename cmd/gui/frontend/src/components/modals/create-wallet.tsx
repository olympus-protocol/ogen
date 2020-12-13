import React, { useState } from 'react';
import NewWalletModal from './new-wallet';

export default function CreateWallet() {
  const [modalOpen, setModalOpen] = useState(false);

  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);

  return (
    <>
      <NewWalletModal isOpen={modalOpen} onClose={closeModal} />
      <div id="add-wallet-modal" className="modal-container">
        <div className="modal-header">
          <span>Add Wallet</span>
          <span className="fas-icon">times</span>
        </div>
        <div className="modal-content abs-center">
          <button className="add-btn btn btn-blue" onClick={openModal}>
            Create Wallet
          </button>
          <span>- or -</span>
          <button className="add-btn btn btn-blue">
            Migrate from Polis Core
          </button>
          <button className="add-btn btn btn-blue">Import Mnemonic</button>
        </div>
      </div>
    </>
  );
}
