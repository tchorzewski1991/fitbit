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

If you don't have your local instalation of Go follow the installation guide from: https://go.dev
## Run locally

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

Build the node

```bash
make build
```

Start the node

```bash
./fitbit
```

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

