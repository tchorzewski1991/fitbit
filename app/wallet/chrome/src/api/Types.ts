export type Genesis = {
    chain_id: number;
    tx_per_block: number;
    difficulty: number;
    mining_reward: number;
    gas_price: number;
    balances: {
        [account: string]: number
    };
}

export type Account = {
    id: string;
    name: string;
    nonce: number;
    balance: number;
}

export type Accounts = Array<Account>
