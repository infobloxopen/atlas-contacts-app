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
./bin/server -dsn "host=localhost port=5432 user=postgres password=postgres sslmode=disable dbname=contacts"
```

Run GRPC gateway:

``` sh
./bin/gateway
```

#### Try atlas-contacts-app

``` sh
curl http://localhost:8080/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "email_address": "mike@gmail.com"}'
```

``` sh
curl http://localhost:8080/atlas-contacts-app/v1/contacts -d '{"first_name": "Bob", "email_address": "john@gmail.com"}'
```

``` sh
curl http://localhost:8080/atlas-contacts-app/v1/contacts?_filter='first_name=="Mike"'
```

#### Local Kubernetes setup

##### Prerequisites

Make sure nginx is deployed in you K8s. Otherwise you can deploy it using

``` sh
make nginx-up
```

or by running

``` sh
kubectl apply -f kube/nginx.yaml
```

If you launching atlas-contacts-app for the first time you need to create `contacts` namespace for it in Kubernetes. This can be done by running

``` sh
kubectl apply -f kube/ns.yaml
```

or

``` sh
kubectl create ns contacts
```

##### Deployment

To deploy atlas-contacts-app use

``` sh
make up
```

or as alternative you can run

``` sh
kubectl apply -f kube/kube.yaml
```

##### Usage

Try it out by executing following curl commangs:

``` sh
curl https://minikube/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "email_address": "mike@gmail.com"}'
```

``` sh
curl https://minikube/atlas-contacts-app/v1/contacts -d '{"first_name": "Bob", "email_address": "john@gmail.com"}'
```

``` sh
curl https://minikube/atlas-contacts-app/v1/contacts?_filter='first_name=="Mike"'
```

##### Pagination

Contacts App implements pagination in by adding application specific page token implementation.

Actually the service supports "composite" pagination in a specific way:

- limit and offset are still supported but without page token

- if an user requests page token and provides limit then limit value will be used as a step for all further requests
		`page_toke = null & limit = 2 -> page_token=base64(offset=2:limit=2)`
		
- if an user requests page token and provides offset then only first time the provided offset is applied
		`page_token = null & offset = 2 & limit = 2 -> page_token=base64(offset=4:limit=2)`

```shell
GET http://localhost:8080/atlas-contacts-app/v1/contacts

HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Date: Fri, 25 May 2018 16:40:12 GMT
Content-Length: 603

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
    },
    {
      "emails": [
        {
          "address": "four@mail.com",
          "id": "4"
        }
      ],
      "first_name": "Mike",
      "id": "4",
      "primary_email": "four@mail.com"
    },
    {
      "emails": [
        {
          "address": "five@mail.com",
          "id": "5"
        }
      ],
      "first_name": "Mike",
      "id": "5",
      "primary_email": "five@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK"
  }
}
```

```
GET http://localhost:8080/atlas-contacts-app/v1/contacts?_limit=2&_offset=2

HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Date: Fri, 25 May 2018 16:43:39 GMT
Content-Length: 274

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
    },
    {
      "emails": [
        {
          "address": "four@mail.com",
          "id": "4"
        }
      ],
      "first_name": "Mike",
      "id": "4",
      "primary_email": "four@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK"
  }
}
```

```
GET http://localhost:8080/atlas-contacts-app/v1/contacts?_page_token=null&_limit=4

HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Grpc-Metadata-Status-Page-Info-Page_token: NDo0
Date: Fri, 25 May 2018 16:54:40 GMT
Content-Length: 513

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
    },
    {
      "emails": [
        {
          "address": "four@mail.com",
          "id": "4"
        }
      ],
      "first_name": "Mike",
      "id": "4",
      "primary_email": "four@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK",
    "_page_token": "NDo0"
  }
}
```

```
GET http://localhost:8080/atlas-contacts-app/v1/contacts?_page_token=NDo0

HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Grpc-Metadata-Status-Page-Info-Page_token: NTo0
Date: Fri, 25 May 2018 16:55:09 GMT
Content-Length: 182

{
  "results": [
    {
      "emails": [
        {
          "address": "five@mail.com",
          "id": "5"
        }
      ],
      "first_name": "Mike",
      "id": "5",
      "primary_email": "five@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK",
    "_page_token": "NTo0"
  }
}
```

```
GET http://localhost:8080/atlas-contacts-app/v1/contacts?_page_token=NTo0

HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Content-Type: application/grpc
Grpc-Metadata-Status-Page-Info-Page_token: null
Date: Fri, 25 May 2018 16:55:28 GMT
Content-Length: 59

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
