# google-dynamic-dns
A Go script to update Dynamic DNS information in Google Domains

## Requirements
* A key generated from https://ipinfo.io (Limit 1000 requests/day)
* Credentials from Google Domains Dynamic DNS
* Go Language

## Setup
* Copy the content of `sample-configs.json` to `configs.json` and update the content
* Run `go build` and schedule it with `cron`
