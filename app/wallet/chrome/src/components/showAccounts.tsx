import React, {useEffect, useState} from "react";
import {Accounts} from "../api/Types";
import axios from "axios";

export default function ShowAccounts() {
    const [accounts, setAccounts] = useState<Accounts | []>([])

    useEffect(() => {
        fetchAccounts()
    }, [])

    const fetchAccounts = async () => {
        try {
            const res = await axios.get<Accounts>("http://localhost:3000/v1/accounts")
            setAccounts(res.data)
        } catch (e) {
            console.error(e)
        }
    }

    return (
        <div>
            {accounts.map((account, idx) => {
                return <div key={idx}>
                    <p>ID: {account.id}</p>
                    <p>Name: {account.name}</p>
                    <p>Nonce: {account.nonce}</p>
                    <p>Balance: {account.balance}</p>
                </div>
            })}
        </div>
    )
}
