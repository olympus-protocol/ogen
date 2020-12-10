import { useCallback } from "react";
import { useDispatch, useSelector } from "react-redux";
import { walletBalance } from "../../backend/wallet";
import { AppState } from "../index";
import {
  fetchUserBalance,
  fetchUserWallets,
  selectWallet as selectUserWallert,
} from "./actions";

export function useWalletState() {
  return useSelector((state: AppState) => state.wallet);
}

export function useWalletActionCreators() {
  const dispatch = useDispatch<any>();

  const fetchWallets = useCallback(() => dispatch(fetchUserWallets()), [
    dispatch,
  ]);

  const selectWallet = useCallback(
    (wallet) => dispatch(selectUserWallert(wallet)),
    [dispatch]
  );

  const updateWalletInfo = async () => {
    const balance = await walletBalance()

    dispatch(fetchUserBalance(balance))
  }

  return {
    fetchWallets,
    selectWallet,
    updateWalletInfo
  };
}
