# atlas-contacts-app

_This generated README.md file loosely follows a [popular template](https://gist.github.com/PurpleBooth/109311bb0361f32d87a2)._

One paragraph of project description goes here.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

**Install go dep**

``` sh
go get -u github.com/golang/dep/cmd/dep
```

**Install atlas-db-controller for managing database resources**

  * To create atlas-db custom resource and controller follow [link](https://github.com/infobloxopen/atlas-db/blob/master/README.md). **Make sure to create one if not already exists**.
  * To create database server instance, database and schema/tables resources locally to be used by contacts app, follow
  ```
    export GIT_USER=<YOUR_GIT_USER_NAME>
    export GIT_TOKEN=<YOUR_GIT_USER_TOKEN>
    cat deploy/contacts-localdb.yaml | envsubst '$GIT_USER$GIT_TOKEN' | kubectl apply -f -
  ```
  To create GIT_TOKEN if not exists follow [link](https://github.com/settings/tokens)

  * To use RDS database instance, database and schema/tables resources used by contacts app, modify `contacts-rds.yaml` and run"
  ```
    kubectl create -f ./deploy/contacts-rds.yaml
  ```

### Local development setup

Please note that you should have the following ports opened on you local workstation: `:8080 :8081 :9090 :5432`.
If they are busy - please change them via corresponding parameters of `gateway` and `server` binaries or postgres container run.

Run PostgresDB:

```sh
docker run --name contacts-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=contacts -p 5432:5432 -d postgres:9.4
```

Table creation should be done manually by running the migrations scripts. Scripts can be found at `./db/migrations/`

Create vendor directory with required golang packages
``` sh
make vendor
```

Run App server:

``` sh
go run ./cmd/server/*.go -db "host=localhost port=5432 user=postgres password=postgres sslmode=disable dbname=contacts"
```

#### Try atlas-contacts-app

For Multi-Account environment, Authorization token (Bearer) is required. You can generate it using https://jwt.io/ with following Payload:
```
{
  "AccountID": YourAccountID
}
```

Example:
```
{
  "AccountID": 1
}
```
Bearer
``` sh
export JWT="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2NvdW50SUQiOjF9.GsXyFDDARjXe1t9DPo2LIBKHEal3O7t3vLI3edA7dGU"
```

Request examples:
``` sh
curl -H "Authorization: Bearer $JWT" \
http://localhost:8080/v1/contacts -d '{"first_name": "Mike", "primary_email": "mike@example.com"}'
```

``` sh
curl -H "Authorization: Bearer $JWT" \
http://localhost:8080/v1/contacts -d \
'{"first_name": "Robert", "primary_email": "robert@example.com", "nicknames": ["bob", "robbie"]}'
```

``` sh
curl -H "Authorization: Bearer $JWT" \
http://localhost:8080/v1/contacts?_filter='first_name=="Mike"'
```
Note, that `JWT` should contain AccountID field.

#### Build docker images

``` sh
make
```
Will be created docker images 'infoblox/contacts-gateway' and 'infoblox/contacts-server'.

If this process finished with errors it's likely that docker doesn't allow to mount host directory in its container.
Therefore you are proposed to run `su -c "setenforce 0"` command to fix this issue.

### Local Kubernetes setup

##### Prerequisites

Make sure nginx is deployed in your K8s. Otherwise you can deploy it using

``` sh
make nginx-up
```

##### Deployment
To deploy atlas-contacts-app use

``` sh
make up
```
Will be used latest Docker Hub images: 'infoblox/contacts-gateway:latest', 'infoblox/contacts-server:latest'.

To deploy authN stub, clone atlas-stubs repo (https://github.com/infobloxopen/atlas-stubs.git) and then execute deployment script inside authn-stub package or:

``` sh
curl https://raw.githubusercontent.com/infobloxopen/atlas-stubs/master/authn-stub/deploy/authn-stub.yaml | kubectl apply -f -
```

This will start AuthN stub that maps `User-And-Pass` header on JWT tokens, with following meaning:

```
admin1:admin -> AccountID=1
admin2:admin -> AccountID=2
```

##### Usage

Try it out by executing following curl commands:

``` sh
# Create some profiles
curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/profiles -d '{"name": "personal", "notes": "Used for personal aims"}' | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/profiles -d '{"name": "work", "notes": "Used for work aims"}' | jq

# Create some groups assigned to profiles
curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/groups -d '{"name": "schoolmates", "profile_id": 1}' | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/groups -d '{"name": "family", "profile_id": 1}' | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/groups -d '{"name": "accountants", "profile_id": 2}' | jq

# Add some contacts assigned to profiles and groups
curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "primary_email": "mike@gmail.com", "profile_id": 1, "groups": [{"id": 1, "name": "schoolmates", "profile_id": 1}, {"id": 2, "name": "family", "profile_id": 1}], "home_address": {"city": "Minneapolis", "state": "Minnesota", "country": "US"}}' | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/contacts -d '{"first_name": "John", "primary_email": "john@gmail.com", "profile_id": 2, "work_address": {"city": "St.Paul", "state": "Minnesota", "country": "US"}}' | jq

# Read created resources
curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/profiles  | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/groups | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/contacts | jq
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

Get all contacts: `GET http://localhost:8080/v1/contacts`
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

Default pagination (supported by atlas-app-toolkit): `GET http://localhost:8080/v1/contacts?_limit=1&_offset=1`
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

Request **specific** page token: `GET http://localhost:8080/v1/contacts?_page_token=null&_limit=2`
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

Get next page via page token: `GET http://localhost:8080/v1/contacts?_page_token=NDo0`
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

Get next page: `GET http://localhost:8080/v1/contacts?_page_token=NTo0`
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
