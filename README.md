# Hasura Storage

Hasura storage is a service that adds a storage service on top of hasura and any s3-compatible storage service. The goal is to be able to leverage the cloud storage service while also leveraging hasura features like its graphql API, permissions, actions, presets, etc...

## Workflows

To understand what hasura-storage does we can look at the two main workflows to upload and retrieve files.

### Uploading files

When a user wants to upload a file hasura-storage will first check with hasura if the user is allowed to do so, if it the file will be uploaded to s3 and, on completion, file metadata will be stored in hasura.

``` mermaid
sequenceDiagram
    actor User
    autonumber
    User->>+hasura-storage: upload file
    hasura-storage->>+hasura: check permissions
    hasura->>-hasura-storage: return if user can upload file
    hasura-storage->>+s3: upload file
    s3->>-hasura-storage: file information
    hasura-storage->>+hasura: file metadata
    hasura->>-hasura-storage: success
    hasura-storage->>-User: file metadata
```

### Retrieving files

Similarly, when retrieving files, hasura-storage will first check with hasura if the user has permissions to retrieve the file and if the user is allowed, it will forward the file to the user:

``` mermaid
sequenceDiagram
    actor User
    autonumber
    User->>+hasura-storage: request file
    hasura-storage->>+hasura: check permissions
    hasura->>-hasura-storage: return if user can access file
    hasura-storage->>+s3: retrieve file
    s3->>-hasura-storage: file
    hasura-storage->>-User: "file"
```

## Features

The main features of the service are:

- leverage hasura's permissions to allow users to upload/retrieve files
- upload files to any s3-compatible service
- dowload files from any s3-compatible service
- create presigned URLs to grant temporary access
- caching information to integrate with caches and CDNs (cache headers, etag, conditional headers, etc)

## OpenAPI

The service comes with an [OpenAPI definition](/blob/main/controller/openapi.yaml) which you can also see [online](https://editor.swagger.io/?url=https://raw.githubusercontent.com/nhost/hasura-storage/main/controller/openapi.yaml).

## Using the service

Easiest way to get started is by using [nhost](https://nhost.io)'s free tier but if you want to self-host you can easily do it yourself as well.

### Self-hosting the service

Requirements:

1. [hasura](https://hasura.io) running, which in turns needs [postgres or any other supported database](https://hasura.io/docs/latest/graphql/core/databases/index/#supported-databases).
2. An s3-compatible service. For instance, [AWS S3](https://aws.amazon.com/s3/), [minio](https://min.io), etc...

A fully working example using docker-compose can be found [here](/blob/main/build/dev/docker/).


## Contributing

If you need help or want to contribute it is recommended to read the [contributing](/blob/main/CONTRIBUTING.md) information first. In addition, if you plan to contribute with code it is also encouraged to read the [development](/blob/main/DEVELOPMENT.md) guide.
