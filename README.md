# sched-load

![Build Status](https://travis-ci.org/dhrapson/sched-load.svg)

This repo contains a utility for managing scheduled uploads to IaaS object stores.
The challenge being:
* the source systems vary in operating system & bit architecture
* the source systems will have a native scheduled job (e.g. cron / windows scheduler setup on them) & need an executable to run, that actions the upload
* the uploaded files might come from a variety of locations and timezones
* the source system may want to indicate that no file is expected until further notice (e.g. before source system upgrade)
* the upload may fail for some reason
* as owner of the IaaS object stores, expecting regular upload of files, we must be able to notify when an expected file upload did not arrive

To address these challenges, the source code contained within is written in Go Lang,
and has the ability to upload files to S3 (first release) & to upload a schedule file indicating
the schedule on which the object store owner should expect files.

The project is building in [TravisCI](https://travis-ci.org/dhrapson/sched-load) and is managed via a public [Tracker](https://www.pivotaltracker.com/n/projects/1941641) board

## To build

Fetch & build from github.com
`go get github.com/dhrapson/sched-load`

OR if changing locally

```
go install
sched-load status
```

## To run

* Set AWS credentials in one of the standard ways: .aws/credentials or env vars
* Run `sched-load help` for running instructions

## To test

```
TEST_AWS_ACCESS_KEY_ID='your_aws_key_id' \
TEST_AWS_SECRET_ACCESS_KEY='your_aws_secret_key' \
go test -v .
```
