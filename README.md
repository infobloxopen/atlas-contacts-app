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

To enable multiaccounting example, run following command:

``` sh
make auths-up
```

which will start AuthS[Tub] that maps `Username` header on JWT tokens, with following meaning:

admin1: AccountID=1
admin2: AccountID=2

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
curl -H "Username: admin1" \
https://minikube/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "email_address": "mike@gmail.com"}'
```

``` sh
curl -H "Username: admin1" \
https://minikube/atlas-contacts-app/v1/contacts -d '{"first_name": "Bob", "email_address": "john@gmail.com"}'
```

``` sh
curl -H "Username: admin1" \
https://minikube/atlas-contacts-app/v1/contacts?_filter='first_name=="Mike"'
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
