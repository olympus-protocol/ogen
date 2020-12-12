import React, { ChangeEvent, useEffect, useState } from 'react';
import useGoToPath from '../../hooks/useGoToPath';
import {
  useWalletActionCreators,
  useWalletState,
} from '../../state/wallets/hooks';
import SelectWalletModal from '../SelectWalletModal';

interface HeaderProps {
  header: string;
}

export default function Header({ header }: HeaderProps) {
  const [selected, setSelected] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const { wallets, selectedWallet } = useWalletState();
  const { fetchWallets } = useWalletActionCreators();
  const goToAddWallet = useGoToPath('/modals/add-wallet');

  const handleSelect = (e: ChangeEvent<HTMLSelectElement>) => {
    const { value } = e.target;

    if (value === 'add') {
      goToAddWallet();
    } else if (value) {
      setSelected(value);
      setModalOpen(true);
    }
  };

  const closeModal = () => setModalOpen(false);

  useEffect(() => {
    fetchWallets();
  }, [fetchWallets]);

  return (
    <>
      <SelectWalletModal
        selectedWallet={selected}
        isOpen={modalOpen}
        onClose={closeModal}
      />
      <div className="header abs-center">
        <span>{header}</span>
        <div className="wallet-select">
          <select onChange={handleSelect} value={selectedWallet}>
            <option value="">Select Wallet</option>
            {wallets &&
              Object.keys(wallets).map((wallet) => (
                <option key={wallet} value={wallet}>
                  {wallet}
                </option>
              ))}
            <optgroup label="------------">
              <option value="add">Add Wallet</option>
            </optgroup>
          </select>
        </div>
      </div>
    </>
  );
}
