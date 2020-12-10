import { createAction, createAsyncThunk } from "@reduxjs/toolkit";
import { fetchWallets } from "../../backend/wallet";

export const fetchUserWallets = createAsyncThunk(
    'wallets/fetchUserWallets',
    async () => {
        return await fetchWallets()
    }
)

export const selectWallet = createAction<string>("wallets/selectWallet");

export const fetchUserBalance = createAction<any>("wallets/fetchUserBalance") 
