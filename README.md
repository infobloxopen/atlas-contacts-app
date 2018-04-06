# atlas-contacts-app

_This generated README.md file loosely follows a [popular template](https://gist.github.com/PurpleBooth/109311bb0361f32d87a2)._

One paragraph of project description goes here.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

#### Local setup:
Need `mysql` or `mariadb` installed.

E.g.

Install database engine:
```
    $ sudo dnf install mariadb mariadb-server
```

Start database server:
```
    $ systemctl start mariadb
```

Create `atlas-contacts-app` database:

```
    $ mysql -u root
```

```
    MariaDB [(none)]> CREATE DATABASE atlas-contacts-app;
```

Create necessary table:
```
    $ mysql -u root atlas-contacts-app < ./migrations/0001_contacts.sql
```


### Installing

#### Local setup:

Build the project.
```
make
```

Run GRPC gateway:
```
./bin/gateway
```

Run GRPC server:
```
./bin/server -dsn root:@tcp/atlas-contacts-app
```

Try atlas-contacts-app:
```
curl http://localhost:8080/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "email_address": "mike@gmail.com"}'
```

```
curl http://localhost:8080/atlas-contacts-app/v1/contacts -d '{"first_name": "Bob", "email_address": "john@gmail.com"}'
```

```
curl http://localhost:8080/atlas-contacts-app/v1/contacts?_filter='first_name=="Mike"'
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
