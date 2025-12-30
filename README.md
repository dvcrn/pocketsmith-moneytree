
# Pocketsmith Moneytree Sync

A tool to sync transactions from Moneytree (Japan) to Pocketsmith.

## Features

- Automatically syncs transactions from Moneytree to Pocketsmith
- Converts Japanese full-width characters to half-width for better readability
- Detects duplicate transactions to avoid double entries
- Updates account balances automatically

## Setup

### Required Environment Variables


MONEYTREE_USERNAME=your_username
MONEYTREE_PASSWORD=your_password  
MONEYTREE_API_KEY=your_api_key
POCKETSMITH_TOKEN=your_token


### Command Line Flags

Alternatively, you can provide credentials via command line flags:


./pocketsmith-moneytree -username=xxx -password=xxx -apikey=xxx -pocketsmith-token=xxx

### Run with docker (recommended)

`docker run -e MONEYTREE_API_KEY=xxx MONEYTREE_USERNAME=xxx -e MONEYTREE_PASSWORD=xxx -e POCKETSMITH_TOKEN=xxx ghcr.io/dvcrn/pocketsmith-moneytree:latest`

## Building


go build


## Running


./pocketsmith-moneytree
