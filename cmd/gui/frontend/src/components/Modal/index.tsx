import React, { ReactNode } from "react";
import ReachModal from "react-modal";

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
