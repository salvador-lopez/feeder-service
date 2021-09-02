#Feeder service

##Requirements

In order to execute this application in your machine you'll need to have the following software installed:

- Go 1.16 or higher https://golang.org/doc/install
- Docker: https://docs.docker.com/get-docker/
- Docker-compose: https://docs.docker.com/compose/install/

##Init the application

To setup the containers that holds the mongodb infrastructure we need to execute in the project root folder:
```
docker-compose up -d
```
##Run the application

In order to execute the feeder-service app you'll find the required end vars in the cmd/socket-server/.env.{local|test} file. Copy them and execute the run command as follows from the repository root path:
```
SOCKET_ADDR=localhost:4000;MONGO_URI=mongodb://localhost:27017;MONGO_DATABASE=sku_test;TIMEOUT_IN_SECS=60;LOG_FILE_NAME=server_report_file.txt;MAX_CONCURRENT_CONNECTIONS=5 go run cmd/socket-server/main.go
```

If you want to run the app with the default config values you can execute only as follows:
```
go run cmd/socket-server/main.go
```
The previous command will run the app with the config values defined in the cmd/socket-server/main.go:28 (config.newConfigDefault() factory method)

##Execute tests:
```
SOCKET_ADDR=localhost:5000;MONGO_URI=mongodb://localhost:27017;MONGO_DATABASE=sku_test;TIMEOUT_IN_SECS=2;LOG_FILE_NAME=server_report_file_test.txt;MAX_CONCURRENT_CONNECTIONS=5 go test ./... -tags=unit,integration,acceptance
```

If you want to only execute one kind of tests (unit, integration or acceptance) you can do this modifying the tags flag in the previous command

If you use the intellij IDEA (intellij ultimate or only goland) you can execute both tests and the socket-server application through the run configurations stored in the .run folder

##Architecture overview:
- This application was developed using the hexagonal architecture tactical approach of the Domain Driven Design.
- It's also using the CQRS pattern in the application layer (the domain model is shared between Commands and Queries). The reason to have this is that with this approach is very easy to know what actions (Commands) will modify the state of your application

##Folder structure:
- The entry point of the application lives in the cmd/socket-server folder


- All the code of the application lives in the internal folder, It's separated by modules (sku folder) and inside each module we can find this structure:
  - domain: Here we find all the entities, value objects and the entity repositories (sku, skuId and sku repository) and we will place the domain services if needed.
  Here we have all the domain logic related to guard the consistency of the sku
  

  - application: Here we find the commands and queries (by now only the create sku command is needed)


  - infrastructure/persistence: Here we'll find the repository implementations, in this case the persistence layer is implemented using mongodb


  - infrastructure/io: Here we place all the specific ways to expose our application layer (commands and queries). Now as we're exposing the "create sku command handler" using a socket tcp server we can find the following services:
    - infrastructure/io/socket/tcp/server/server: Here we can find the server that is controlling the signaling and the concurrency of the application. This server executes the readers in a concurrent way limiting the number of allowed concurrent connections and ensuring that
    the graceful shutdown is done when the context is done or when the application is stopped for any reason (ex: signal os.Interrupt is received)
    - infrastructure/io/socket/tcp/sku_reader/sku_reader: This service is creating the connection and start listening to the tcp network and the specified address. As this is a blocking operation this listen has a timeout (the maximum duration of the listen goroutine is the timeout defined when we execute the application)

