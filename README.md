# housekeeper: A tool for automating the image creation pipeline for OpenStack.

<!--TOC-->
- [Installation](#installation)
  - [Linux](#linux)
- [Usage](#usage)
  - [List](#list)
  - [Cleanup](#cleanup)
  - [Delete](#delete)
  - [Publish](#publish)
<!--/TOC-->
## Installation
### Linux
Download and install binary from [releases](https://github.com/hornwind/openstack-image-housekeeper/releases):
```bash
curl -sSLO "https://github.com/hornwind/openstack-image-housekeeper/releases/download/v0.2.0/openstack-image-housekeeper_0.2.0_linux_amd64.tar.gz"
tar -zxvf openstack-image-housekeeper_0.2.0_linux_amd64.tar.gz housekeeper
sudo mv housekeeper /usr/local/bin/housekeeper
sudo chown root:root /usr/local/bin/housekeeper
sudo chmod 0755 /usr/local/bin/housekeeper
```
## Usage
### List
`housekeeper list` prints Name, ID, CreatedAt, Protected, Hidden and Tags of your private images. Supports setting values through environment variables.
```
NAME:
   housekeeper list - List of available images

USAGE:
   housekeeper list [command options] [arguments...]

OPTIONS:
   --loglevel value  configure log level (default: "info") [$HOUSEKEEPER_LOG_LEVEL]
   --help, -h        show help
```
example output:
```
Name: test_image
ID: e6637019-e80c-49b1-84ff-1bbe97cfcd64
CreatedAt: 2023-07-06 15:05:32 +0000 UTC
Protected false
Hidden false
Tags:
  780e66832a83e72c8bf49684976340e61a30506a
  master
```
### Cleanup
Runs cleanup by name of image.\
`housekeeper cleanup gitlab_dev_16.2.2`

Performs idempotent cleanup of existing images by name. Keeps the latest image based on the git commit sha in the image tags. If unable to retrieve the latest N commits, it retains the last built image. Images with the 'public' attribute remain unaffected. Supports setting values through environment variables.

```
NAME:
   housekeeper cleanup - Cleanup images by name

USAGE:
   housekeeper cleanup [command options] [arguments...]

OPTIONS:
   --scandepth value  configure git scan depth (default: 10) [$HOUSEKEEPER_SCAN_DEPTH]
   --dry-run          run without dangerous activity (default: false) [$HOUSEKEEPER_DRY_RUN]
   --loglevel value   configure log level (default: "info") [$HOUSEKEEPER_LOG_LEVEL]
   --help, -h         show help
```
### Delete
`housekeeper delete <uuid>`
Deletes one or more private images by their UUIDs separated by spaces.
```bash
housekeeper delete f25148bb-fc89-4787-abfa-4889e455c3f8 8b2d978b-da7f-4ddd-839e-27fbbecb4de2
```
### Publish
Publishes an image by its UUID.
All images with the same name are first set to the following state: `visibility: private`, `protected: false`, `hidden: false`.\
Then, the image being published is set to the `visibility: public` state, with the `protected` and `hidden` values determined by the respective `--protected` and `--hidden` flags, defaulting to `false`.\
Supports setting values through environment variables.
```
NAME:
   housekeeper publish - Publication image by id

USAGE:
   housekeeper publish [command options] [arguments...]

OPTIONS:
   --dry-run         run without dangerous activity (default: false) [$HOUSEKEEPER_DRY_RUN]
   --protected       set image protected (default: false) [$HOUSEKEEPER_SET_PROTECTED]
   --hidden          set image hidden (default: false) [$HOUSEKEEPER_SET_HIDDEN]
   --loglevel value  configure log level (default: "info") [$HOUSEKEEPER_LOG_LEVEL]
   --help, -h        show help
```
