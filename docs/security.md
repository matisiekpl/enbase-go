# Security rules

Building serverless apps often requires the security layer for controlling the access permissions to specific data. In Enbase, to secure your database , you can use the Security Rules, strongly inspired by Google's Firebase Security Rules. By defining the special Javascript statements, you can control access to specific collections and documents in many cases. 

## Quick start - the real-world example
For example, we have the `tasks` collection, and users can:
- read documents, where task is owned by given user
- create documents, when task has properties: `name`, `done`, `userId` and userId must match to logged user id
- update documents, when task has properties: `name`, `done`, `userId` and userId must match to logged user id
- delete documents, when task is owned by given user
