# **Test Plans**

For information on what each field means, refer to:

1. [Metadata](../proto/metadata.md)
2. [Test Plans](../proto/testplan.md)


## Retrieve all test plans
Method: `GET`

Path: `/api/v1/testplans`

Response:
```json
{
    "count": 2,
    "items": [
        {
            "identity": {
                "id": "4c5c11cd300a665",
                "type": "testplan",
                "version": 1,
                "createdBy": "author",
                "updatedBy": "author",
                "creationTime": 1614505965,
                "updateTime": 1614505965
            },
            "projectId": "4c2f2b65400a665",
            "name": "test plan 1"
        },
        {
            "identity": {
                "id": "4c658d70800b9c5",
                "type": "testplan",
                "version": 1,
                "createdBy": "author",
                "updatedBy": "author",
                "creationTime": 1614605401,
                "updateTime": 1614605401
            },
            "projectId": "4c2f2b65400a665",
            "name": "Test Plan Name",
            "description": "Test Plan description"
        }
    ]
}
```


## Create a new test plan
Method: `POST`

Path: `/api/v1/testplans`

Request:    
```json
{
    "name": "Test Plan Name", // Mandatory
    "description": "Test Plan description"
}
```
Response:
```json
{
    "identity": {
        "id": "4c658d70800b9c5",
        "type": "testplan",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614605401,
        "updateTime": 1614605401
    },
    "projectId": "4c2f2b65400a665",
    "name": "Test Plan Name",
    "description": "Test Plan description"
}
```

## Update a test plan
Method: `PUT`

Path: `/api/v1/testplans/{identity.id}`



Request:    
```json
{
    "name": "Test Plan Name",
    "description": "New test Plan description",
    "projectId": "4c2f2b65400a665"
}
```
Response:
```json
{
    "identity": {
        "id": "4c658d70800b9c5",
        "type": "testplan",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614605401,
        "updateTime": 1614609576
    },
    "projectId": "4c2f2b65400a665",
    "name": "Test Plan Name",
    "description": "New test Plan description"
}
```

## Get a single test plan
Method: `GET`

Path: `/api/v1/testplans/{identity.id}`

Response:
```json
{
    "identity": {
        "id": "4c658d70800b9c5",
        "type": "testplan",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614605401,
        "updateTime": 1614609576
    },
    "projectId": "4c2f2b65400a665",
    "name": "Test Plan Name",
    "description": "New test Plan description"
}
```

## Delete a testplan
Method: `DELETE`

Path: `/api/v1/testplans/{identity.id}`
