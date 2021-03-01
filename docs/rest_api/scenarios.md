
# **Scenarios**

For information on what each field means, refer to:

1. [Metadata](../proto/metadata.md)
2. [Scenarios](../proto/scenario.md)


## Retrieve all scenarios
Method: `GET`

Path: `/api/v1/scenarios`

Response:
```json
{
    "count": 2,
    "items": [
        {
            "identity": {
                "id": "4c2f2b65400a665",
                "type": "scenario",
                "version": 1,
                "createdBy": "author",
                "updatedBy": "author",
                "creationTime": 1614035154,
                "updateTime": 1614035154
            },
            "name": "Test Name"
        },
        {
            "identity": {
                "id": "4c65280ca00b9c5",
                "type": "scenario",
                "version": 1,
                "createdBy": "author",
                "updatedBy": "author",
                "creationTime": 1614601248,
                "updateTime": 1614601548
            },
            "name": "Scenario Name",
            "description": "New description"
        }
    ]
}
```


## Create a new scenario
Method: `POST`

Path: `/api/v1/scenarios`

Request:    
```json
{
    "name":"Example Scenario",
    "description": "Description of the scenario",
    "prerequisites": "Maybe start the app?",
    "projectId":"4c2f2b65400a665",
    "labels": ["test label"],
    "steps":[
        {
            "position":1,
            "name": "login",
            "action": "user logs in with correct credentials",
            "expectedOutcome": "login action is performed successfully"
        }
    ]
}
```
Response:
```json
{
    "identity": {
        "id": "4c658344000b9c5",
        "type": "scenario",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614604984,
        "updateTime": 1614604984
    },
    "projectId": "4c2f2b65400a665",
    "name": "Example Scenario",
    "description": "Description of the scenario",
    "prerequisites": "Maybe start the app?",
    "steps": [
        {
            "position": 1,
            "name": "login",
            "action": "user logs in with correct credentials",
            "expectedOutcome": "login action is performed successfully"
        }
    ],
    "labels": [
        "test label"
    ]
}
```

## Update a scenario
Method: `PUT`

Path: `/api/v1/scenarios/{identity.id}`



Request:    
```json
{
    "name":"Example Scenario",
    "description": "New description of the scenario",
    "prerequisites": "Maybe start the app?",
    "projectId":"4c2f2b65400a665",
    "labels": ["test label"],
    "steps":[
        {
            "position":1,
            "name": "login",
            "action": "user logs in with correct credentials",
            "expectedOutcome": "login action is performed successfully"
        }
    ]
}
```
Response:
```json
{
    "identity": {
        "id": "4c658344000b9c5",
        "type": "scenario",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614604984,
        "updateTime": 1614605122
    },
    "projectId": "4c2f2b65400a665",
    "name": "Example Scenario",
    "description": "New description of the scenario",
    "prerequisites": "Maybe start the app?",
    "steps": [
        {
            "position": 1,
            "name": "login",
            "action": "user logs in with correct credentials",
            "expectedOutcome": "login action is performed successfully"
        }
    ],
    "labels": [
        "test label"
    ]
}
```

## Get a single scenario
Method: `GET`

Path: `/api/v1/scenarios/{identity.id}`

Response:
```json
{
    "name":"Example Scenario",
    "description": "New description of the scenario",
    "prerequisites": "Maybe start the app?",
    "projectId":"4c2f2b65400a665",
    "labels": ["test label"],
    "steps":[
        {
            "position":1,
            "name": "login",
            "action": "user logs in with correct credentials",
            "expectedOutcome": "login action is performed successfully"
        }
    ]
}
```

## Delete a scenario
Method: `DELETE`

Path: `/api/v1/scenarios/{identity.id}`