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
that will transform data from my running activites to digital tokens which ... I will be able to store later
on my own, private blockchain. Despite the fact this project hasn't got any business value I can see the time
spent on it as a great opportunity to explore new technologies while keeping the right balance of complexity
that blockchain technology brings to the table.

## Features

- TODO

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
./fitbit
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
| --state-accounts-path   | Path to the location where all account <br/>private keys will be stored.      | data/accounts | false    |
| --state-data-path       | Path to the location where all mined <br/>blocks will be stored.              | data/miner    | false    |
| --state-beneficiary     | Beneficiary is the owner of the node. <br/>Account which gains mining reward. | miner         | false    |
| --state-origin-peers    | The origin node we need to <br/>connect to make initial sync.                 | 0.0.0.0:4000  | false    |

To initialize third node we need to run the following command:

```bash
go run ./app/services/node/main.go \
    --node-public-host 0.0.0.0:3002 \
    --node-private-host 0.0.0.0:4002 \
    --state-beneficiary <replace-with-beneficiary-name>.ecdsa \
    --state-data-path data/<replace-with-beneficiary-data-name>
```

This project is shipped with the minimalistic wallet CLI. To generate a new account (public - private
key pair) run the following command:

```bash
go run ./app/wallet/cli/main.go generate --account-name Babajaga --account-path data/accounts
```

Running this command will generate a new ECDSA private key under `data/accounts` path.

TODO: continue from here

Shutdown the node

```bash
ctrl+c
```

Clean up project directory

```bash
make clean
```

## How it works?

TODO: Describe

By default running node should expose following http ports...

## Run tests

To run tests, run the following command

```bash
make tests
```

## License

[MIT](https://choosealicense.com/licenses/mit/)

