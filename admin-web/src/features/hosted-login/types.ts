export type PasswordValues = {
  email: string;
  password: string;
};

export type RegisterValues = PasswordValues & {
  username?: string;
};

export type PhoneValues = {
  phone: string;
  code: string;
};

export type UnifiedLoginConfig = {
  brandName: string;
  brandMark: string;
  subtitle: string;
  enableRegister: boolean;
  enablePhone: boolean;
  enableGitHub: boolean;
  enableGoogle: boolean;
  enableWallet: boolean;
};

export type WalletProvider = {
  request: (input: { method: string; params?: unknown[] }) => Promise<unknown>;
};

export type WalletOption = {
  id: string;
  name: string;
  icon?: string;
  provider: WalletProvider;
};

export type EIP6963ProviderDetail = {
  info: {
    uuid: string;
    name?: string;
    icon?: string;
    rdns?: string;
  };
  provider: WalletProvider;
};
