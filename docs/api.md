# Enbase API reference
### Models
##### User
```
{
  "email": string,
  "password": string,
  "role": string
}
```

##### Project
```
{
  "name": string,
  "description": string
}
```
##### Database
```
{
  "name": string,
  "description": string,
  "rules": {
    "<collection>:<action>": string
  }
}
```
##### Actions
`create`, `read`, `update`, `delete`

Example
```
{
  "name": "Test",
  "description": "Test database",
  "rules": {
    "pets:create": "true"
  }
}
```

##### Endpoints
`POST /auth/user` - register new user

`POST /auth/session` - sign in

`<method> /system/projects` - operate on projects

`<method> /system/projects/<projectId>` - operate on project

`<method> /system/projects/<projectId>/databases` - operate on databases

`<method> /system/projects/<projectId>/databases/<databaseId>` - operate on database

##### Methods
`POST`, `GET`, `PUT`, `DELETE`
