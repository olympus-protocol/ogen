import React, { useState } from "react";
import { openWallet } from "../../backend/wallet";
import { useWalletActionCreators } from "../../state/wallets/hooks";
import Modal from "../Modal";

type SelectWalletModalProps = {
  selectedWallet: string;
  isOpen: boolean;
  onClose: () => void;
};

export default function SelectWalletModal({
  selectedWallet,
  isOpen,
  onClose,
}: SelectWalletModalProps) {
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const { selectWallet, updateWalletInfo } = useWalletActionCreators();

  const onSubmit = async (e: any) => {
    e.preventDefault();

    if (!password) return;

    try {
      await openWallet(selectedWallet, password);
      selectWallet(selectedWallet);
      updateWalletInfo();
      onClose();
    } catch (e: any) {
      setError(e);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <h2>Open {selectedWallet} Wallet</h2>
      <form onSubmit={onSubmit}>
        <label>Enter Wallet Password</label>
        <br />
        <input
          className="form-control"
          type="password"
          placeholder="Enter wallet password"
          onChange={(e) => setPassword(e.target.value)}
        />
        <br />
        {error && <span className="form-control-error">{error}</span>}
        <input className="btn btn-primary" type="submit" value="Submit" />
      </form>
    </Modal>
  );
}
