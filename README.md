# atlas-contacts-app

_This generated README.md file loosely follows a [popular template](https://gist.github.com/PurpleBooth/109311bb0361f32d87a2)._

One paragraph of project description goes here.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

Install go dep

``` sh
go get -u github.com/golang/dep/cmd/dep
```

Need `postgresql` installed locally or running inside docker container within `contacts` database created.
You can change the name of database in `dsn` parameter in `server` app.
Table creation in this example is omitted as it will be done autmotically by `gorm`.

### Installing

#### Build the project

``` sh
make vendor
make
```

If this process finished with errors it's likely that docker doesn't allow to mount host directory in its container.
Therefore you are proposed to run `su -c "setenforce 0"` command to fix this issue.

#### Local setup

Please note that you should have the following ports opened on you local workstation: `:8080 :9090 :8088 :8089`.
If they are busy - please change them via corresponding parameters of `gateway` and `server` binaries.

Run GRPC server:

``` sh
go run ./cmd/server/main.go -db "host=localhost port=5432 user=postgres password=postgres sslmode=disable dbname=contacts"
```

Run GRPC gateway:

``` sh
go run ./cmd/gateway/*.go
```

#### Try atlas-contacts-app

For Multi-Account environment, Authorization token is required. You can generate it using https://jwt.io/ with following Payload:
```
{
  "AccountID": YourAccountID
}
```

``` sh
curl -H "Grpc-Metadata-Authorization: Token $JWT" \
http://localhost:8080/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "email_address": "mike@gmail.com"}'
```

``` sh
curl -H "Grpc-Metadata-Authorization: Token $JWT" \
http://localhost:8080/atlas-contacts-app/v1/contacts -d '{"first_name": "Bob", "email_address": "john@gmail.com"}'
```

``` sh
curl -H "Grpc-Metadata-Authorization: Token $JWT" \
http://localhost:8080/atlas-contacts-app/v1/contacts?_filter='first_name=="Mike"'
```

Note, that JWT should contain AccountID field.

#### Local Kubernetes setup

##### Prerequisites

Make sure nginx is deployed in you K8s. Otherwise you can deploy it using

``` sh
make nginx-up
```

##### Deployment
Upload atlas-contacts-app images into minikube
```
make push-minikube
```

To deploy atlas-contacts-app use

``` sh
make up
```

To deploy authN stub, clone atlas-stubs repo and then execute deployment script inside authn-stub package:

``` sh
	cd $GOPATH/src/github.com/infobloxopen && git clone https://github.com/infobloxopen/atlas-stubs.git
	cd atlas-stubs/authn-stub && make && make up
```

This will start AuthN stub that maps `User-And-Pass` header on JWT tokens, with following meaning:

```
admin1:admin -> AccountID=1
admin2:admin -> AccountID=2
```

##### Usage

Try it out by executing following curl commangs:

``` sh
curl -k -H "User-And-Pass: admin1:admin" \
https://minikube/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "primary_email": "mike@gmail.com"}'
```

``` sh
curl -k -H "User-And-Pass: admin1:admin" \
https://minikube/atlas-contacts-app/v1/contacts -d '{"first_name": "Bob", "primary_email": "john@gmail.com"}'
```

``` sh
curl -k -H "User-And-Pass: admin1:admin" \
https://minikube/atlas-contacts-app/v1/contacts?_filter='first_name=="Mike"'
```

##### Pagination (page token)

**DISCLAIMER**: it is intended only for demonstration purposes and should not be emulated.

Contacts App implements pagination in by adding application **specific** page token implementation.

Actually the service supports "composite" pagination in a specific way:

- limit and offset are still supported but without page token

- if an user requests page token and provides limit then limit value will be used as a step for all further requests
		`page_token = null & limit = 2 -> page_token=base64(offset=2:limit=2)`
		
- if an user requests page token and provides offset then only first time the provided offset is applied
		`page_token = null & offset = 2 & limit = 2 -> page_token=base64(offset=4:limit=2)`

Get all contacts: `GET http://localhost:8080/atlas-contacts-app/v1/contacts`
```json
{
  "results": [
    {
      "emails": [
        {
          "address": "one@mail.com",
          "id": "1"
        }
      ],
      "first_name": "Mike",
      "id": "1",
      "primary_email": "one@mail.com"
    },
    {
      "emails": [
        {
          "address": "two@mail.com",
          "id": "2"
        }
      ],
      "first_name": "Mike",
      "id": "2",
      "primary_email": "two@mail.com"
    },
    {
      "emails": [
        {
          "address": "three@mail.com",
          "id": "3"
        }
      ],
      "first_name": "Mike",
      "id": "3",
      "primary_email": "three@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK"
  }
}
```

Default pagination (supported by atlas-app-toolkit): `GET http://localhost:8080/atlas-contacts-app/v1/contacts?_limit=1&_offset=1`
```json
{
  "results": [
    {
      "emails": [
        {
          "address": "two@mail.com",
          "id": "2"
        }
      ],
      "first_name": "Mike",
      "id": "2",
      "primary_email": "two@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK"
  }
}
```

Request **specific** page token: `GET http://localhost:8080/atlas-contacts-app/v1/contacts?_page_token=null&_limit=2`
```json
{
  "results": [
    {
      "emails": [
        {
          "address": "one@mail.com",
          "id": "1"
        }
      ],
      "first_name": "Mike",
      "id": "1",
      "primary_email": "one@mail.com"
    },
    {
      "emails": [
        {
          "address": "two@mail.com",
          "id": "2"
        }
      ],
      "first_name": "Mike",
      "id": "2",
      "primary_email": "two@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK",
    "_page_token": "NDo0"
  }
}
```

Get next page via page token: `GET http://localhost:8080/atlas-contacts-app/v1/contacts?_page_token=NDo0`
```json
{
  "results": [
    {
      "emails": [
        {
          "address": "three@mail.com",
          "id": "3"
        }
      ],
      "first_name": "Mike",
      "id": "3",
      "primary_email": "three@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK",
    "_page_token": "NTo0"
  }
}
```

Get next page: `GET http://localhost:8080/atlas-contacts-app/v1/contacts?_page_token=NTo0`
The `"_page_token": "null"` means there are no more pages
```json
{
  "success": {
    "status": 200,
    "code": "OK",
    "_page_token": "null"
  }
}
```

## Deployment

Add additional notes about how to deploy this application. Maybe list some common pitfalls or debugging strategies.

## Running the tests

Explain how to run the automated tests for this system.

```
Give an example
```

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags).
