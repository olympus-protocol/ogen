interface Wallet {
  NewWallet: (name: string, mnemonic: string, password: string) => void;
  GetAvailableWallets: () => any;
  OpenWallet: (name: string, password: string) => void;
  GetBalance: () => any;
}

function getWallet(): Wallet {
  const win = window as any;
  return win.backend && win.backend.wallet;
}

export async function fetchWallets(): Promise<
  { [key: string]: string } | undefined
> {
  const wallet = getWallet();
  if (!wallet) return;

  return await wallet.GetAvailableWallets();
}

export async function openWallet(
  name: string,
  password: string
): Promise<string | undefined> {
  const wallet = getWallet();
  if (!wallet) return;

  await wallet.OpenWallet(name, password);

  return name;
}

export async function newWallet(name: string, password: string): Promise<void> {
  const wallet = getWallet();
  if (!wallet) return;

  await wallet.NewWallet(name, '', password);
}

export async function walletBalance(): Promise<any> {
  const wallet = getWallet();
  if (!wallet) return;

  return await wallet.GetBalance();
}
