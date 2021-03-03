
[![Go Report Card](https://goreportcard.com/badge/github.com/curious-kitten/scratch-post)](https://goreportcard.com/report/github.com/curious-kitten/scratch-post) ![Actions](https://github.com/curious-kitten/scratch-post/workflows/Pull%20Requests/badge.svg)

# Scratch Post
A free, open source test management platform. 

You can 

# Testing structure

The test structures are:
 * [Projects](./docs/proto/execution.md)
 * [Scenarios](./docs/proto/scenario.md)
 * [Executions](./docs/proto/execution.md) 
 * [Test Plans](./docs/proto/testplan.md)

# Using Scratch Post

Building the project:

`make build-app`

This will create a binary under `build/_bin` 

The project requires 3 configuration files to be used. All these files can be generate using the binary.
```bash
./scratch-post generate -h
generate is used to generate the configurations needed to run scratch-post

Usage:
  scratch-post generate [command]

Available Commands:
  admin-db-config admin-db-config generates JSON file for configuring the database to store administrative information
  api-config      api-config generates JSON file for configuring the REST API
  test-db-config  test-db-config generates JSON file for configuring the database to store test information

Flags:
  -h, --help   help for generate

Use "scratch-post generate [command] --help" for more information about a command.
```

1. Generating the REST API config
    ```bash
    ./scratch-post generate api-config -h
    api-config generates JSON file for configuring the REST API

    Usage:
    scratch-post generate api-config [flags]

    Flags:
        --adminPrefix string   prefix for all admin endpoints (default "/admin")
        --executions string    executions endpoint (default "/executions")
        --file string          file which will contain the configuration (default "apiconfig.json")
    -h, --help                 help for api-config
        --port string          port for the server (default "9090")
        --probes string        probes endpoints (default "/probes")
        --projects string      projects endpoint (default "/projects")
        --rootPrefix string    prefix for all api endpoints (default "/api/v1")
        --scenarios string     scenarios endpoint (default "/scenarios")
        --testplans string     testplans endpoint (default "/testplans")
        --users string         users endpoint. Is part of the admin endpoints (default "/users")
    ```

1. Generating the Test DB config. *address* and *database* are mandatory.
    ```bash
    ./scratch-post generate test-db-config -h
    test-db-config generates JSON file for configuring the Mongo database.
            This data base is used to store test information. 
            Information provided by this config file is:
            - address:     the URL to connect to the instance 
            - database:    the specific database to be used in the instance
            - collections: a map which you can use to specify what collection each scratch-post item type can use

    Usage:
    scratch-post generate test-db-config [flags]

    Flags:
        --address string      testdb server address
        --database string     mongo database name
        --executions string   collection name to be used for executions (default "executions")
        --file string         file which will contain the configuration (default "testdb.json")
    -h, --help                help for test-db-config
        --projects string     collection name to be used for projects (default "projects")
        --scenarios string    collection name to be used for scenarios (default "scenarios")
        --testplans string    collection name to be used for testplans (default "testplans")
    ```
    :grey_exclamation: The DB type used is MongoDB. If you don't have Mongo instance available, you can create a free instance at https://cloud.mongodb.com/

1. Generating the Admin DB config. *address* is mandatory
    ```bash
    ./scratch-post generate admin-db-config -h
    admin-db-config generates JSON file for configuring the Postgress database.
            This data base is used to store test information. 
            Information provided by this config file is:
            - address: the URL to connect to the instance

    Usage:
    scratch-post generate admin-db-config [flags]

    Flags:
        --address string    testdb server address
        --file string       file which will contain the configuration (default "admindb.json")
    -h, --help              help for admin-db-config
        --maxIdle int       maximum idle connections (default 5)
        --maxLifetime int   maximum lifetime (default 5)
        --maxOpen int       maximum open connections (default 5)
    ```
    :grey_exclamation: The DB type used is Postgress. If you don't have a Postgress instance available, you can create a free instance at https://www.elephantsql.com/

1. To start the app you can use: 
    ```bash
    ./scratch-post start -h
    Starts the server for managing test cases

    Usage:
    scratch-post start [flags]

    Flags:
        --admindb string        Path to admin DB config settings (default "admindb.json")
        --apiconfig string      Path to API config settings (default "apiconfig.json")
    -h, --help                  help for start
        --isJWT                 Sets the authentication type to JWT. Default is session ID
        --scenarios string      collection name to be used for scenarios (default "scenarios")
        --securityFile string   Path to file which contains the JWT security string (default "security.txt")
        --testdb string         Path to DB config settings (default "testdb.json")
    ```

    If you already have the config files, you can also use `make run`