import React, {useEffect, useState} from "react";
import {Genesis} from "../types/Genesis";
import axios from "axios";

export default function ShowGenesis(){
    const [genesis, setGenesis] = useState<Genesis | null>(null)

    useEffect(() => {
        fetchGenesis()
    }, [])

    const fetchGenesis = async () => {
        try {
            const res = await axios.get<Genesis>("http://localhost:3000/v1/genesis")
            setGenesis(res.data)
        } catch (e) {
            console.error(e)
        }
    }

    const balances = genesis ? Object.keys(genesis.balances).map((account,idx) => {
        return <p key={idx}>Account: {account} | Balance: {genesis.balances[account]}</p>
    }) : null;

    return (
        <div>
            {genesis ? <div>
                <p>Chain ID:      {genesis.chain_id}</p>
                <p>Tx per block:  {genesis.tx_per_block}</p>
                <p>Difficulty:    {genesis.difficulty}</p>
                <p>Mining reward: {genesis.mining_reward}</p>
                <p>Gas price:     {genesis.gas_price}</p>
                <p>Balances: </p> {balances}
            </div> : <p>Loading...</p> }
        </div>
    )
}
