# Live streaming server

## Get started
* Install postgress sql. Create user, database with values from file `./zconfig/zconfig.go`
* Change file `./zconfig/zconfig.go`
* Build and run static files server from folder `./static`.
* Build and run main server.

## Documentation
Read files in folder `./document`

## Folder structure conventions
* Package main can import all packages
* A package doesn't import same or lower level packages.  