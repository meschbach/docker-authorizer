# Docker Authorizer

Got a container?  How ya gonna get your secrets?  Give 'em an app role?  So last year man.

Ya trust your Docker daemon? Ya got big problems if you don't.  Like real big.  How about using the Docker daemon to
authenticate the container?  Lemme say that again:

Authenticate Docker containers to your Vault instance.

## How does it work?

A container reaches out to the _authorizer_ on port 8000, issuing an _HTTP GET_ request for `/`.  The _authorizer_ will
use the client address to find the container from the local Docker instance.  A container containing the property
`org.meschbach/docker-authorizer/role` on an expected _network_ will be given a [wrapped response](https://www.vaultproject.io/docs/concepts/response-wrapping.html)
to retrieve a token.

## Constraints

* *Direct Container <-> Authorizer Connection* The authorizer utilizes the requesting address to verify the container is
making the request.  As a result, if the service proxied or not able to directly connect, the requesting container will
be rejected.

* *Docker Daemon is trusted* This assumes your Docker Daemon is an authorative source and able to verify the containers.


## Building and Contributing

Contributions are welcome.   Buildings should be as easy as `go build`.  On OSX you will need to build and deploy the
container to get around the setup of [Docker for Mac](https://docs.docker.com/v17.12/docker-for-mac/install/)'s NATing.

> docker build . --tag docker-authorizer:test
