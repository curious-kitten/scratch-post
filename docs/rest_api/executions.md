
# **Executions**

For information on what each field means, refer to:

1. [Metadata](../proto/metadata.md)
2. [Executions](../proto/execution.md)

## Retrieve all executions
Method: `GET`

Path: `/api/v1/executions`

Response:
```json
{
    "count": 2,
    "items": [
        {
            "identity": {
                "id": "4c5c13ad900a665",
                "type": "execution",
                "version": 1,
                "createdBy": "author",
                "updatedBy": "author",
                "creationTime": 1614506042,
                "updateTime": 1614506042
            },
            "projectId": "4c2f2b65400a665",
            "scenarioId": "4c2f3bd7f00a665",
            "testPlanId": "4c5c11cd300a665",
            "name": "test plan 1",
            "steps": [
                {
                    "definition": {
                        "position": 1,
                        "name": "login",
                        "action": "user logs in with correct credentials",
                        "expectedOutcome": "login action is performed successfully"
                    }
                }
            ]
        },
        {
            "identity": {
                "id": "4c65ffcc900b9c5",
                "type": "execution",
                "version": 1,
                "createdBy": "author",
                "updatedBy": "author",
                "creationTime": 1614610085,
                "updateTime": 1614610294
            },
            "projectId": "4c2f2b65400a665",
            "scenarioId": "4c658344000b9c5",
            "testPlanId": "4c658d70800b9c5",
            "status": 1,
            "steps": [
                {
                    "definition": {
                        "position": 1,
                        "name": "login",
                        "action": "user logs in with correct credentials",
                        "expectedOutcome": "login action is performed successfully"
                    },
                    "status": 1,
                    "ActualResult": "Login did not succeed"
                }
            ]
        }
    ]
}
```


## Create a new execution
Method: `POST`

Path: `/api/v1/executions`

Request:    
```json
{
    "projectId": "4c2f2b65400a665",
    "scenarioId": "4c658344000b9c5",
    "testPlanId": "4c658d70800b9c5"
}
```
Response:
```json
{
    "identity": {
        "id": "4c65ffcc900b9c5",
        "type": "execution",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614610085,
        "updateTime": 1614610085
    },
    "projectId": "4c2f2b65400a665",
    "scenarioId": "4c658344000b9c5",
    "testPlanId": "4c658d70800b9c5",
    "steps": [
        {
            "definition": {
                "position": 1,
                "name": "login",
                "action": "user logs in with correct credentials",
                "expectedOutcome": "login action is performed successfully"
            }
        }
    ]
}
```

## Update a execution
Method: `PUT`

Path: `/api/v1/executions/{identity.id}`



Request:    
```json
{
    "identity": {
        "id": "4c65ffcc900b9c5",
        "type": "execution",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614610085,
        "updateTime": 1614610085
    },
    "projectId": "4c2f2b65400a665",
    "scenarioId": "4c658344000b9c5",
    "testPlanId": "4c658d70800b9c5",
    "steps": [
        {
            "definition": {
                "position": 1,
                "name": "login",
                "action": "user logs in with correct credentials",
                "expectedOutcome": "login action is performed successfully"
            },
            "status": 1,
            "actualResult": "Login did not succeed"
        }
    ]
}
```
Response:
```json
{
    "identity": {
        "id": "4c65ffcc900b9c5",
        "type": "execution",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614610085,
        "updateTime": 1614610294
    },
    "projectId": "4c2f2b65400a665",
    "scenarioId": "4c658344000b9c5",
    "testPlanId": "4c658d70800b9c5",
    "status": 1,
    "steps": [
        {
            "definition": {
                "position": 1,
                "name": "login",
                "action": "user logs in with correct credentials",
                "expectedOutcome": "login action is performed successfully"
            },
            "status": 1,
            "ActualResult": "Login did not succeed"
        }
    ]
}
```

## Get a single execution
Method: `GET`

Path: `/api/v1/executions/{identity.id}`

Response:
```json
{
    "identity": {
        "id": "4c65ffcc900b9c5",
        "type": "execution",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614610085,
        "updateTime": 1614610294
    },
    "projectId": "4c2f2b65400a665",
    "scenarioId": "4c658344000b9c5",
    "testPlanId": "4c658d70800b9c5",
    "status": 1,
    "steps": [
        {
            "definition": {
                "position": 1,
                "name": "login",
                "action": "user logs in with correct credentials",
                "expectedOutcome": "login action is performed successfully"
            },
            "status": 1,
            "ActualResult": "Login did not succeed"
        }
    ]
}
```
