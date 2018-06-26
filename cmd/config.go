package cmd

const (
	// ServerAddress is the default address for the gRPC server, if no override is specified in the flags
	ServerAddress = "0.0.0.0:9090"
	// GatewayAddress is the default address for the gateway server, if no override is specified in the flags
	GatewayAddress = "0.0.0.0:8080"
	// InternalAddress is the default address for the internal http server, if no override is specified in the flags
	InternalAddress = "0.0.0.0:8081"
	// DatabaseAddress is the default address for the database, if no override is specified in the flags
	DBConnectionString = "host=localhost port=5432 user=postgres password=postgres sslmode=disable dbname=atlas_contacts_app"
	// SwaggerFile is the file location of the swagger file to serve
	SwaggerFile = "./pkg/pb/contacts.swagger.json"
	// ApplicationID associates a microservice with an application. The atlas
	// contacts application consists of only one service, so we identify both the
	// service and the application as "atlas-contacts-app"
	ApplicationID = "atlas-contacts-app"
)
