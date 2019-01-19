# Security rules

Building serverless apps often requires the security layer for controlling the access permissions to specific data. In Enbase, to secure your database , you can use the Security Rules, strongly inspired by Google's Firebase Security Rules. By defining the special Javascript statements, you can control access to specific collections and documents in many cases. 

## Quick start - the real-world example
For example, we have the `tasks` collection, and users can:
- read documents, where task is owned by given user
- create documents, when task has properties: `name`, `done`, `userId` and userId must match to logged user id
- update documents, when task has properties: `name`, `done`, `userId` and userId must match to logged user id
- delete documents, when task is owned by given user

The security rules for following assumptions looks like this:
- read - `document.userId == user.id`
- create - `document.hasOwnProperty('userId') && document.userId == user.id && document.hasOwnProperty('name') && document.hasOwnProperty('done')`
- update - `document.hasOwnProperty('userId') && document.userId == user.id && document.hasOwnProperty('name') && document.hasOwnProperty('done')`
- delete - `document.userId == user.id`

## Reference
- `user` - json object or null, represents currently logged in user
- `action` - string, represents called action. Can be `read`, `create`, `update`, `delete` or `stream`
- `document` - json object, represents requested document
- `id` - string, represents requestes document id
- `get` - function: `(collection: string, query: object) => []`, query resource in given collection with given query
