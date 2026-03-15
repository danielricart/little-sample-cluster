# little-sample-cluster
Litle Application with a fullly-fledged deployment and cluster.

NOTE: 
Originally, the sample payload for `PUT` had a typo and it was written as `dateOfBrith`(notice the `r` <> `i`). This has been fixed in the implementation. 

## Application
This is a simple application that exposes the following HTTP based APIs:

Description: Save/updates a given user name and date of birth in a database.

Request: 
```
PUT /hello/<username> { “dateOfBirth”: “YYYY-MM-DD” }

Response: 204 No Content
```

Note:
- Username should only be letters.
- YYYY-MM-DD must be a date before today's date.

Description: Returns a birthday message.
```
Request: Get /hello/<username>

Response: 200 Ok
```

Response examples:

A. If username’s birthday is in N days:

```
{ “message”: “Hello, <username>! Your birthday is in N day(s)”}
```

B. If username’s birthday is today:

```
{ “message”: “Hello, <username>! Happy birthday!” }
```

Note: Use the storage or DB of your choice.
 
