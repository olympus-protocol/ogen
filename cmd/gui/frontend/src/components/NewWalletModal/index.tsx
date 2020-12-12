import React from 'react';
import { newWallet } from '../../backend/wallet';
import useForm from '../../hooks/useForm';
import useGoToPath from '../../hooks/useGoToPath';
import Modal, { ModalHeader, ModalBody } from '../Modal';

type NewWalletModalProps = {
  isOpen: boolean;
  onClose: () => void;
};

export default function NewWalletModal({
  isOpen,
  onClose,
}: NewWalletModalProps) {
  const { form, handleTextInput }: any = useForm();
  const goToHome = useGoToPath('/');

  const onSubmit = async (e: any) => {
    e.preventDefault();

    if (!form.name || !form.password) {
      return;
    }

    try {
      await newWallet(form.name, form.password);
      goToHome();
    } catch (e) {
      console.log(e);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalHeader>
        <h2>Create New Wallet</h2>
      </ModalHeader>
      <ModalBody>
        <form onSubmit={onSubmit}>
          <label htmlFor="name">Wallet Name</label>
          <input
            className="form-control"
            type="text"
            name="name"
            onChange={handleTextInput}
          />

          <label htmlFor="name">Wallet Password</label>
          <input
            className="form-control"
            type="password"
            name="password"
            onChange={handleTextInput}
          />

          <input
            className="btn btn-primary"
            type="submit"
            value="Create Wallet"
          />
        </form>
      </ModalBody>
    </Modal>
  );
}
