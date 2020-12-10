import { createReducer } from "@reduxjs/toolkit";
import { fetchUserBalance, fetchUserWallets, selectWallet } from "./actions";

type Balance = {
  Confirmed: number
  Pending: number
}

type Wallet = {
  [key: string]: string
}

interface WalletState {
  readonly selectedWallet: string | undefined;
  readonly wallets: Wallet | undefined;
  readonly balance: Balance | undefined
}


const initialState: WalletState = {
  selectedWallet: "",
  wallets: {},
  balance: {
    Confirmed: 0,
    Pending: 0
  }
};

export default createReducer<WalletState>(initialState, (builder) =>
  builder
    .addCase(selectWallet, (state, { payload: wallet }) => {
      return {
        ...state,
        selectedWallet: wallet,
      };
    })
    .addCase(fetchUserWallets.fulfilled, (state, { payload: wallets }) => {
      return {
        ...state,
        wallets,
      };
    })
    .addCase(fetchUserBalance, (state, { payload: balance }) => {
      return {
        ...state,
        balance
      }
    })
);
