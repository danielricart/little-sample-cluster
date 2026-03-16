little-sample-cluster
-

Little Application with a fullly-fledged deployment, k8s cluster and architectural description

> [!NOTE]  
> Originally, the sample payload for `PUT` had a typo, and it was written as `dateOfBrith`(notice the `r` <> `i`). This has been fixed in the implementation. 

# Application

The details about the application implementation and design decisions are documented in [`DECISIONS.md`](./DECISIONS.md). 

The details on how to run and maintain the application are found in [`RUN.md`](./RUN.md)

The code can be found in [`main.go`](./main.go) as the entry point and the packages are in the folder [`pkg`](./pkg). 

## Design decisions:
- GET response body will be:
```json
{ "dateOfBirth": "YYYY-MM-DD" }
```

- If GET username is not found, response will be 404
- if PUT payload does not match, response will be 400
- capitalization of usernames is respected: `newuser` is different than `Newuser`, `NewUser` or `NEWUSER`

## Description 

This is a simple application that exposes the following HTTP based APIs:

Description: Save/updates a given user name and date of birth in a database.

Request: 
```
PUT /hello/<username> { "dateOfBirth": "YYYY-MM-DD" }

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
{ "message": "Hello, <username>! Your birthday is in N day(s)"}
```

B. If username’s birthday is today:

```
{ "message": "Hello, <username>! Happy birthday!" }
```

Note: Use the storage or DB of your choice.
 
# Infrastructure 

## Production Architecture

Check [`PRODUCTION.md`](./PRODUCTION.md) and [`architecture_layer1.png`](./architecture_layer1.png) [`architecture_layer2.png`](./architecture_layer2.png) [`architecture_layer3.png`](./architecture_layer3.png) for architecture description and diagrams. Run `python3 architecture.py` to recreate the diagrams. 

# dependencies
## architecture diagrams

```shell
brew install graphviz # sudo apt install graphviz
pip3 install diagrams
```


# AI Usage
- to speed up boilerplate for github action creation, e2e tests and multi-case tests.
- to speed up http method filtering. some mass refactoring on tests and metaresearch about some issues I've found. 
- Architecture.py has been suggested by ClaudeCode based on the content of PRODUCTION.md. Then I edited manually to make it correct. 
- demo/deploy.sh has been worked in collab with ClaudeCode.
