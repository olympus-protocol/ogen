import { configureStore, getDefaultMiddleware } from '@reduxjs/toolkit';

import wallet from './wallets/reducer';

const store = configureStore({
  reducer: {
    wallet,
  },
  middleware: [...getDefaultMiddleware({ thunk: true })],
});

export default store;

export type AppState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
