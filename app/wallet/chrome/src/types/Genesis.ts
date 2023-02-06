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
