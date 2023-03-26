# Shard Store

### Composition
* Filegateway - a backend for frontend with an api to upload and download files
* Filestore - a backend that is actually responsible for storing the data

### Configuration
#### Filegateway default settings
```env
FG_APP_NAME=filegateway
FG_APP_ENV=local
FG_HTTP_PORT=8080
FG_MAX_FILE_SIZE=10485760 // 10Mb
FG_NUMBER_OF_CHUNKS=3
FG_STORAGE_SERVERS="localhost:9000;localhost:9001;localhost:9002"
FG_STORAGE_SERVER_TIMEOUT="10s"
```
#### Filestore default settings
```env
FS_APP_NAME=filestore
FS_APP_ENV=local
FS_GRPC_PORT=9000
FS_REFLECTION_API=true
```
Number of servers and should be greater or equal to the number of chunks.
Filegateway obviously needs to know all the addresses of the file servers.

### Usage
Look at Makefile
