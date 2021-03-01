# **Projects**


For information on what each field means, refer to:

1. [Metadata](../proto/metadata.md)
2. [Projects](../proto/project.md)


## Retrieve all projects
Method: `GET`

Path: `/api/v1/projects`

Response:
```json
{
    "count": 2,
    "items": [
        {
            "identity": {
                "id": "4c2f2b65400a665",
                "type": "project",
                "version": 1,
                "createdBy": "author",
                "updatedBy": "author",
                "creationTime": 1614035154,
                "updateTime": 1614035154
            },
            "name": "test name 3"
        },
        {
            "identity": {
                "id": "4c65280ca00b9c5",
                "type": "project",
                "version": 1,
                "createdBy": "author",
                "updatedBy": "author",
                "creationTime": 1614601248,
                "updateTime": 1614601548
            },
            "name": "Project Name",
            "description": "New description"
        }
    ]
}
```

## Create a new project
Method: `POST`

Path: `/api/v1/projects`

Request:    
```json
{
    "name": "Project Name", // Mandatory
    "description": "Project description"
}
```
Response:
```json
{
    "identity": {
        "id": "4c65280ca00b9c5",
        "type": "project",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614601248,
        "updateTime": 1614601248
    },
    "name": "Project Name",
    "description": "Project description"
}
```

## Update a project
Method: `PUT`

Path: `/api/v1/projects/{indetity.id}`



Request:    
```json
{
    "name": "Project Name", 
    "description": "New description"
}
```
Response:
```json
{
    "identity": {
        "id": "4c65280ca00b9c5",
        "type": "project",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614601248,
        "updateTime": 1614602820
    },
    "name": "Project Name",
    "description": "New description"
}
```

## Get a single project
Method: `GET`

Path: `/api/v1/projects/{indetity.id}`

Response:
```json
{
    "identity": {
        "id": "4c65280ca00b9c5",
        "type": "project",
        "version": 1,
        "createdBy": "author",
        "updatedBy": "author",
        "creationTime": 1614601248,
        "updateTime": 1614602820
    },
    "name": "Project Name",
    "description": "New description"
}
```

## Delete a project
Method: `DELETE`

Path: `/api/v1/projects/{identity.id}`