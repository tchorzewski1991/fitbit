# Fitbit

[![Go Report Card](https://goreportcard.com/badge/github.com/tchorzewski1991/fitbit)](https://goreportcard.com/report/github.com/tchorzewski1991/fitbit)

This project is a reference blockchain implementation that was built for educational purpose.
Although feature rich this project should not be considered production ready as there is still
much to work on.

Blockchain has gained some momentum recently. As I have been fascinated by this technology since
my first exposure to it I want to actively learn and explore its applications in various industries.


## A bit of a context

Why Fitbit? Regardless of being a passionate about software development I am also a huge advocate of any
kind of sport activities - especially running. I thought one day - how cool would it be to write a project
that will transform data from my running activities to digital tokens which ... I will be able to store and
use later on my own, private blockchain. Despite the fact this project hasn't got any business value I can
see the time spent on it as a great opportunity to explore new technologies while keeping the right balance
of complexity that blockchain technology brings to the table.

## Functionalities 

- Node
  - Public API
    - Provides info about genesis file
    - Provides list of account balances
    - Provides balance of specific account
    - Provides list of uncommited transactions
    - Provides uncommited transactions of specific account
    - Handles submission of wallet transactions
  - Private API
    - Provides list of known peers
    - Provides list of blocks by height
    - Provides list of uncommited transactions
    - Handles submission and sync of new peer
    - Handles submission and sync of transactions
    - Handles submission and sync of new block proposal
- CLI Wallet
  - Provides ability to generate new account
  - Provides ability to generate public address of account
  - Handles submission of wallet transactions
- Chrome Wallet
  - In progress

## Prerequisites

Running this project locally assumes you have installed version of Go `1.19` or higher.

In order to make sure your environment is setup properly run the following command:

```bash
go version
```

If you don't have your local installation of Go follow the installation guide from: https://go.dev

## Running the Project

Clone the project

```bash
git clone https://github.com/tchorzewski1991/fitbit.git
```

Go to the project directory

```bash
cd fitbit
```

Install dependencies

```bash
make tidy
```

Start the primary node with:

```bash
make run-primary-node
```

Start the second node with:

```bash
make run-miner-node
```

From technical point of view node is a device that is connected to the blockchain network
and participates in the consensus process. In the production ready blockchains there are
different types on nodes with a very specific roles, like mining, validation, etc.

Fitbit blockchain does not define different types of nodes, so every node we run is just
a full node with complete copy of the blockchain ledger. This node validates transactions
before they are added to the blockchain.

We are not restricted to run 2 nodes. We can run as many nodes as we want.
The previous command is just a shortcut for the `go run` which starts a new
node:

```bash
go run ./app/services/node/main.go
```

Starting the node requires however a bit of configuration.

The following table describes the full set of flags available while setting up the node:

| Flag                    | Description                                                                   | Default       | Required |
|-------------------------|-------------------------------------------------------------------------------|---------------|----------|
| --node-public-host      | Host address of the public API                                                | 0.0.0.0:3000  | true     |
| --node-private-host     | Host address of the private API                                               | 0.0.0.0:4000  | true     |
| --node-read-timeout     | -                                                                             | 5s            | false    |
| --node-write-timeout    | -                                                                             | 5s            | false    |
| --node-idle-timeout     | -                                                                             | 5s            | false    |
| --node-shutdown-timeout | -                                                                             | 5s            | false    |
| --state-accounts-path   | Path to the location where all account <br/>private keys will be stored. (*)  | data/accounts | false    |
| --state-data-path       | Path to the location where all mined <br/>blocks will be stored.         (*)  | data/miner    | false    |
| --state-beneficiary     | Beneficiary is the owner of the node. <br/>Account which gains mining reward. | miner         | false    |
| --state-origin-peers    | The origin node we need to <br/>connect to make initial sync.                 | 0.0.0.0:4000  | false    |

Just a quick note on '*'-ed descriptions above. I am perfectly aware of the flaws of this architectural decisions
and this is not the shape of the project we want to ship to production. Nevertheless, for reference implementation
it is just enough. Thanks to the simple implementation of JSON file based storage we got a noticeable advantage
of quick feedback loop.

Although program won't stop you, consider the fact you will need a new set of
public - private key pair (a.k.a. Account) for the third beneficiary.

This project is shipped with the minimalistic wallet CLI which gets you covered.
To generate a new Account (public - private key pair) run the following command:

```bash
go run ./app/wallet/cli/main.go generate --account-name babajaga --account-path data/accounts
```

Running this command will generate a new ECDSA private key under `data/accounts/babajaga.ecdsa` path.

To initialize third node we need to run the following command:

```bash
go run ./app/services/node/main.go \
    --node-public-host 0.0.0.0:3002 \
    --node-private-host 0.0.0.0:4002 \
    --state-beneficiary babajaga.ecdsa \
    --state-data-path data/babajaga
```

We are running currently a small p2p network of 3 independent nodes:

| Name          | Ports                                                  |
|---------------|--------------------------------------------------------|
| Primary node  | Public API: 0.0.0.0:3000<br/>Private API: 0.0.0.0:4000 |
| Miner node    | Public API: 0.0.0.0:3001<br/>Private API: 0.0.0.0:4001 |
| Babajaga node | Public API: 0.0.0.0:3002<br/>Private API: 0.0.0.0:4002 |

Fitbit blockchain does not support fully decentralized environment. Every new node needs
to sync-up with the chain on startup, which basically means we need to point to at least 
one origin-node. Current setup leverages only one origin-node which is a primary node.
This setting is shipped by default, but can be  overwritten by using `--state-origin-peers` 
configuration flag.

To start the mining competition a new transaction has to be sent to at least 1 out of 3 running nodes.

Fitbit wallet CLI provides a simple interface for sending transactions:

```bash
go run ./app/wallet/cli/main.go send --help
```

```
Sends a new transaction

Usage:
  app send [flags]

Flags:
  -d, --data bytesHex   Transaction data to send.
  -n, --nonce uint      Transaction number.
  -c, --tip uint        Transaction tip to add.
  -t, --to string       The receiver of the transaction.
  -u, --url string      The url of the public node. (default "http://localhost:3000")
  -v, --value uint      Transaction value to send.

Global Flags:
  -a, --account-name string   The name of the account. (default "private.ecdsa")
  -p, --account-path string   The path to the account private key. (default "data/accounts")
```

Example command for sending transaction:

```bash
go run ./app/wallet/cli/main.go send \
  --account-name miner \
  --to 0xDBE46a0b3BF1543c9FD7e4BFbD1a054406b62D7d \
  --nonce 1 \
  --value 100 \
  --tip 10
```

The command above will send one new transaction between **miner** and account identified by public 
address **0xDBE46a0b3BF1543c9FD7e4BFbD1a054406b62D7d**. Before transaction will be sent to node's public API 
it will be cryptographically signed by the private key owned by **miner**. Keep in mind that public address 
of the transaction sender does not need to be provided explicitly as it can be derived from the sender's private key.
This is why wallet cli uses the name of the account as one of its parameters to `send` command  instead of 
explicit `--from` flag.

Shutdown the node

```bash
ctrl+c
```

Clean up project directory

```bash
make clean
```

## How it works?

TODO

## Run tests

To run tests, run the following command

```bash
make tests
```

## License

[MIT](https://choosealicense.com/licenses/mit/)

