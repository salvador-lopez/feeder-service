# Feeder service

## Requirements

In order to execute this application in your machine you'll need to have the following software installed:

- Go 1.16 or higher https://golang.org/doc/install
- Docker: https://docs.docker.com/get-docker/
- Docker-compose: https://docs.docker.com/compose/install/

##Init the application

To setup the containers that holds the mongodb infrastructure we need to execute in the project root folder:
```
make init
```
## Run the application

In order to execute the feeder-service app you'll find the required end vars in the Makefile, so you can modify them there if you want. The makefile command:
```
make server-run
```

## Execute tests:
```
make unit-tests
make integration-tests
make acceptance-tests
```

If you use the intellij IDEA (intellij ultimate or only goland) you can execute both tests and the socket-server application through the run configurations stored in the .run folder

## Architecture overview:
- This application was developed using the hexagonal architecture tactical approach of the Domain Driven Design.
- It's also using the CQRS pattern in the application layer (the domain model is shared between Commands and Queries). The reason to have this is that with this approach is very easy to know what actions (Commands) will modify the state of your application

## Folder structure:
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

