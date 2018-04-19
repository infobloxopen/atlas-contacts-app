# atlas-contacts-app

_This generated README.md file loosely follows a [popular template](https://gist.github.com/PurpleBooth/109311bb0361f32d87a2)._

One paragraph of project description goes here.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

#### Local setup

Need `mysql` or `mariadb` installed.

E.g.

Install database engine:

``` sh
    sudo dnf install mariadb mariadb-server
```

Start database server:

``` sh
    systemctl start mariadb
```

Create `atlas-contacts-app` database:

``` sh
    mysql -u root
```

``` sh
    MariaDB [(none)]> CREATE DATABASE atlas-contacts-app;
```

Create necessary table:

``` sh
    mysql -u root atlas-contacts-app < ./migrations/0001_contacts.sql
```

### Installing

#### Build the project

``` sh
make vendor
make
```

If this process finished with errors it's likely that docker doesn't allow to mount host directory in its container. Therefore you are proposed to run `su -c "setenforce 0"` command to fix this issue.

#### Local setup

Run GRPC server:

``` sh
./bin/server -dsn root:@tcp/atlas-contacts-app
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

Open a seperate terminal session where execute `eval $(minikube docker-env)`. This is needed to make these images available for local kubernetes without pushing them to global repo.

Then:

``` sh
make image
make up
```

To shutdown and cleanup:

``` sh
make down
make image-clean
rm -rf bin
```

#### Try local Kubernetes atlas-contacts-app

``` sh
curl http://$(minikube ip):31500/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "email_address": "mike@gmail.com"}'
```

``` sh
curl http://$(minikube ip):31500/atlas-contacts-app/v1/contacts -d '{"first_name": "Bob", "email_address": "john@gmail.com"}'
```

``` sh
curl http://$(minikube ip):31500/atlas-contacts-app/v1/contacts?_filter='first_name=="Mike"'
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
