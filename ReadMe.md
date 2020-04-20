
# clickhouse-backup

[![Build Status](https://travis-ci.org/AlexAkulov/clickhouse-backup.svg?branch=master)](https://travis-ci.org/AlexAkulov/clickhouse-backup)
[![GoDoc](https://godoc.org/github.com/AlexAkulov/clickhouse-backup?status.svg)](http://godoc.org/github.com/AlexAkulov/clickhouse-backup)
[![Telegram](https://img.shields.io/badge/telegram-join%20chat-3796cd.svg)](https://t.me/clickhousebackup)
[![Docker Image](https://img.shields.io/docker/pulls/alexakulov/clickhouse-backup.svg)](https://hub.docker.com/r/alexakulov/clickhouse-backup)

Tool for easy ClickHouse backup and restore with S3 and GCS support

## Features

- Easy creating and restoring backups of all or specific tables
- Efficient storing of multiple backups on the file system
- Most efficient AWS S3/GCS uploading and downloading with streaming compression
- Support of incremental backups on remote storages

## Limitations

- ClickHouse above 1.1.54390 is supported
- Only MergeTree family tables engines
- Backup of 'Tiered storage' or `storage_policy` IS NOT SUPPORTED!
- Maximum backup size on remote storages is 5TB
- Maximum number of parts on AWS S3 is 10,000 (increase part_size if your database is more than 1TB)

## Download

- Download the latest binary from the [releases](https://github.com/AlexAkulov/clickhouse-backup/releases) page and decompress with:

```shell
tar -zxvf clickhouse-backup.tar.gz
```

- Use the official tiny Docker image and run it like:

```shell
docker run --rm -it --network host -v "/var/lib/clickhouse:/var/lib/clickhouse" \
   -e CLICKHOUSE_PASSWORD="password" \
   -e S3_BUCKET="clickhouse-backup" \
   -e S3_ACCESS_KEY="access_key" \
   -e S3_SECRET_KEY="secret" \
   alexakulov/clickhouse-backup --help
```

- Bulid from the sources:

```shell
GO111MODULE=on go get github.com/AlexAkulov/clickhouse-backup
```

## Usage

```
NAME:
   clickhouse-backup - Tool for easy backup of ClickHouse with cloud support

USAGE:
   clickhouse-backup <command> [-t, --tables=<db>.<table>] <backup_name>

VERSION:
   unknown

DESCRIPTION:
   Run as 'root' or 'clickhouse' user

COMMANDS:
     tables          Print list of tables
     create          Create new backup
     upload          Upload backup to remote storage
     list            Print list of backups
     download        Download backup from remote storage
     restore         Create schema and restore data from backup
     delete          Delete specific backup
     default-config  Print default config
     freeze          Freeze tables
     clean           Remove data in 'shadow' folder
     help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config FILE, -c FILE  Config FILE name. (default: "/etc/clickhouse-backup/config.yml")
   --help, -h              show help
   --version, -v           print the version
```

### Default Config

Config file location can be defined by ```$CLICKHOUSE_BACKUP_CONFIG```

All options can be overwritten via environment variables

```yaml
general:
  remote_storage: s3           # REMOTE_STORAGE
  disable_progress_bar: false  # DISABLE_PROGRESS_BAR
  backups_to_keep_local: 0     # BACKUPS_TO_KEEP_LOCAL
  backups_to_keep_remote: 0    # BACKUPS_TO_KEEP_REMOTE
clickhouse:
  username: default            # CLICKHOUSE_USERNAME
  password: ""                 # CLICKHOUSE_PASSWORD
  host: localhost              # CLICKHOUSE_HOST
  port: 9000                   # CLICKHOUSE_PORT
  timeout: 5m                  # CLICKHOUSE_TIMEOUT
  data_path: ""                # CLICKHOUSE_DATA_PATH
  skip_tables:                 # CLICKHOUSE_SKIP_TABLES
    - system.*
  timeout: 5m                  # CLICKHOUSE_TIMEOUT
  freeze_by_part: false        # CLICKHOUSE_FREEZE_BY_PART
s3:
  access_key: ""                   # S3_ACCESS_KEY
  secret_key: ""                   # S3_SECRET_KEY
  bucket: ""                       # S3_BUCKET
  endpoint: ""                     # S3_ENDPOINT
  region: us-east-1                # S3_REGION
  acl: private                     # S3_ACL
  force_path_style: false          # S3_FORCE_PATH_STYLE
  path: ""                         # S3_PATH
  disable_ssl: false               # S3_DISABLE_SSL
  part_size: 104857600             # S3_PART_SIZE
  compression_level: 1             # S3_COMPRESSION_LEVEL
  # supports 'tar', 'lz4', 'bzip2', 'gzip', 'sz', 'xz'
  compression_format: gzip         # S3_COMPRESSION_FORMAT
  # empty (default), AES256, or aws:kms
  sse: AES256                      # S3_SSE
  disable_cert_verification: false # S3_DISABLE_CERT_VERIFICATION
gcs:
  credentials_file: ""         # GCS_CREDENTIALS_FILE
  credentials_json: ""         # GCS_CREDENTIALS_JSON
  bucket: ""                   # GCS_BUCKET
  path: ""                     # GCS_PATH
  compression_level: 1         # GCS_COMPRESSION_LEVEL
  compression_format: gzip     # GCS_COMPRESSION_FORMAT
cos:
  url: ""                      # COS_URL
  timeout: 2m                  # COS_TIMEOUT
  secret_id: ""                # COS_SECRET_ID
  secret_key: ""               # COS_SECRET_KEY
  path: ""                     # COS_PATH
  compression_format: gzip     # COS_COMPRESSION_FORMAT
  compression_level: 1         # COS_COMPRESSION_LEVEL
  debug: false                 # COS_DEBUG
```

## ATTENTION!

Never change files permissions in `/var/lib/clickhouse/backup`.
This path contains hard links. Permissions on all hard links to the same data on disk are always identical.
That means that if you change the permissions/owner/attributes on a hard link in backup path, permissions on files with which ClickHouse works will be changed too.
That might lead to data corruption.

## Examples

### Simple cron script for daily backup and uploading
```bash
#!/bin/bash
BACKUP_NAME=my_backup_$(date -u +%Y-%m-%dT%H-%M-%S)
clickhouse-backup create $BACKUP_NAME
clickhouse-backup upload $BACKUP_NAME
```

### More use cases of clickhouse-backup
- [How to convert MergeTree to ReplicatedMegreTree](Examples.md#how-to-convert-mergetree-to-replicatedmegretree)
- [How to store backups on NFS or another server](Examples.md#how-to-store-backups-on-nfs-or-another-server)
- [How to move data to another clickhouse server](Examples.md#how-to-move-data-to-another-clickhouse-server)
- [How to reduce number of partitions](Examples.md#How-to-reduce-number-of-partitions)
- [How to monitor that backups created and uploaded correctly](Examples.md#how-to-monitor-that-backups-created-and-uploaded-correctly)
- [How to backup sharded cluster with Ansible](Examples.md#how-to-backup-sharded-cluster-with-ansible)
- [How to backup database with several terabytes of data](Examples.md#how-to-backup-database-with-several-terabytes-of-data)
- [How to use clickhouse-backup in Kubernetes](Examples.md#how-to-use-clickhouse-backup-in-kubernetes)
