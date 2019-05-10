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

### Stored Role Information

The _authorizer_ expects authorizing system data to stored in a _KV version 2_ store under `secret/docker/{role}`.  A
role will contain the following keys:

* `network` (`string`): The network the requesting network _must_ be attached to in order to be authorized.
* `image` (`string`): The container which must be making the request.
* `policies` (`string`): The policy name which the wrapped response will contain a token for.

### Configuring the authorizer

Both _Docker_ and _Vault_ libraries are initialized via normal means.  This means all the normal client settings maybe
effected via the environment variables.

## Constraints

* *Direct Container <-> Authorizer Connection* The authorizer utilizes the requesting address to verify the container is
making the request.  As a result, if the service proxied or not able to directly connect, the requesting container will
be rejected.

* *Docker Daemon is trusted* This assumes your Docker Daemon is an authorative source and able to verify the containers.


## Building and Contributing

Contributions are welcome.   Buildings should be as easy as `go build`.  On OSX you will need to build and deploy the
container to get around the setup of [Docker for Mac](https://docs.docker.com/v17.12/docker-for-mac/install/)'s NATing.

> docker build . --tag docker-authorizer:test
