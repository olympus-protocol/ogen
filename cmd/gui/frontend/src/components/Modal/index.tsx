import React, { ReactNode } from 'react';
import ReachModal from 'react-modal';

type ModalHeaderProps = {
  children?: ReactNode;
};

export function ModalHeader({ children }: ModalHeaderProps) {
  return <div className="modal-header">{children}</div>;
}

type ModalBodyProps = {
  children?: ReactNode;
};

export function ModalBody({ children }: ModalBodyProps) {
  return <div className="modal-body">{children}</div>;
}

type ModalProps = {
  isOpen: boolean;
  onClose?: () => void;
  children?: ReactNode;
};

export default function Modal({ isOpen, onClose, children }: ModalProps) {
  return (
    <ReachModal isOpen={isOpen} onRequestClose={onClose}>
      {children}
    </ReachModal>
  );
}
