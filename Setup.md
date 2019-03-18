# Cain

Cain is a project which is part of the official cassandra helm chart. We have extended the project to have incremental backups, OS level encryption and native swift object storage support. Requirements for the new feature set is reflected upon the helm chart that we have included in the repository. 

## Initial Step

To begin using cain should be to build the docker image of cain using the makefile and dockerfiles included in the project. While testing I’ve used Travis to build the image and upload it to a public docker hub repository to host it. Make note of the image registry repository that you have pushed to as it is required in the helm chart to state where the image is hosted. 

## Second Step

Edit the helm chart according to your own requirements as it very flexible.
There are two main configurations in the helm chart called;  backup and incremental;
```
backup:
  enabled: true

  schedule:
  - keyspace: k1
    cron: "*/15 * * * *"
  - keyspace: k2
    cron: "*/15 * * * *"

  image:
    repos: host/cain
    tag: latest

  extraArgs: []

  env:
  - name: SWIFT_API_USER
    value: user
  - name: SWIFT_API_KEY
    value: key
  - name: SWIFT_AUTH_URL
    value: auth_url
  - name: SWIFT_TENANT
    value: tenant
  - name: SWIFT_API_DOMAIN
    value: domain

  resources:
    requests:
      memory: 1Gi
      cpu: 1
    limits:
      memory: 1Gi
      cpu: 1

  destination: swift://container/path/to/folder

```
* enabled: enables the snapshot backups of the system
* schedule: for each keyspace in the system, there should be a cronjob to take snapshots with given time intervals
* image: repository and tag of the cain image that has been pushed to a registry. (ex: barium/cain)
* extraArgs: optional extra arguments to push to kubernetes
* env: environment variables to be used with the cronjobs. Here we state the SWIFT API credentials for the backups to connect to swift. All five variables are required and not optional
* resources: as each cronjob gets triggered, this setting states the resources for the temporary container
* destination: backup destination, it should start with swift://, followed by the designated container and then the path that you wish the backups are stored in.
 
 ```
incremental:
  enabled: true

  schedule:
  - keyspace: k1
    cron: "*/15 * * * *"
  - keyspace: k2
    cron: "*/15 * * * *"

  image:
    repos: host/cain
    tag: latest

  extraArgs: []

  env:
  - name: SWIFT_API_USER
    value: user
  - name: SWIFT_API_KEY
    value: key
  - name: SWIFT_AUTH_URL
    value: auth_url
  - name: SWIFT_TENANT
    value: tenant
  - name: SWIFT_API_DOMAIN
    value: domain

  resources:
    requests:
      memory: 1Gi
      cpu: 1
    limits:
      memory: 1Gi
      cpu: 1

  destination: swift://container/path/to/folder

```
* enabled: enables the incremental backups of the system
* schedule: for each keyspace in the system, there should be a cronjob to copy the incremental backup files with given time intervals
* image: repository and tag of the cain image that has been pushed to a registry. (ex: barium/cain)
* extraArgs: optional extra arguments to push to kubernetes
* env: environment variables to be used with the cronjobs. Here we state the SWIFT API credentials for the backups to connect to swift. All five variables are required and not optional
* resources: as each cronjob gets triggered, this setting states the resources for the temporary container
* destination: backup destination, it should start with swift://, followed by the designated container and then the path that you wish the inceremental backups are stored in.


## Encryption
Only external requirement for the code to work is to have a kubernetes secret named `backup-secret` with a single Data called `passphrase`.

Simple example (taken from kubernetes docs): 
```
apiVersion: v1
kind: Secret
metadata:
  name: backup-secret
  namespace: cassandra
type: Opaque
data:
  passphrase: MWYyZDFlMmU2N2Rm
```

The example passphrase written above is encoded using base64 and any key that would be set should be encoded as such.

This key is used by the cain to encrypt all files symmetrically during the backup operation with `AES256` and it will also be used during the restore operation to decrypt the files to their plaintext versions.

## Restore

While encrypting cain uses this cmd function to encrypt files
```
gpg --homedir /tmp --batch --cipher-algo AES256 --passphrase <passphrase> -o <output> --symmetric <fileToEncrypt>
```

So before the restore operation you should decrypt each file using the following command;
```
gpg --homedir /tmp --batch --cipher-algo AES256 --passphrase <passphrase> -o <output> -d <fileToDecrypt>
```

After all files are decrypted to their plaintext a restore operation is exactly the same as following the offical cassandra restore guidelines.