package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	vault "github.com/hashicorp/vault/api"
)

type DockerAuthorizer struct {
	dockerEngine *docker.Client
	kvPrefix string
}

func main()  {
	cli, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	service := DockerAuthorizer{cli, "docker"}

	r := mux.NewRouter()
	r.HandleFunc("/", service.ClientInfo())

	addr := "0.0.0.0:8000"
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}
	fmt.Printf("Listening at %s\n", addr )
	log.Fatal(srv.ListenAndServe())
}


func (service *DockerAuthorizer) ClientInfo() http.HandlerFunc {
	config := vault.DefaultConfig()
	client, err := vault.NewClient(config)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		//Figure out the IP address of the request
		requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w,"400 Bad Remote Address", 400)
			return
		}

		//Find containers matching the IP address
		containers, err := service.dockerEngine.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			fmt.Fprint(os.Stderr, "Error retrieving Docker containers", err)
			http.Error(w, "403 Unauthorized", 403)
			return
		}

		match := make([]string,0)
		for _, container := range containers {
			role, ok := container.Labels["org.meschbach/docker-authorizer/role"]
			if !ok {
				continue
			}

			//Extract role configuration
			config, err := client.Logical().Read("secret/data/" + service.kvPrefix + "/" + role )
			if err != nil {
				fmt.Printf("Errror communicating with Vault %s", err)
				continue
			}
			if config == nil {
				fmt.Println("No configuration for role")
				continue
			}

			data := config.Data["data"].(map[string]interface{})
			network := data["network"].(string)

			//Check the requesting address is the same as the client
			containerIP := container.NetworkSettings.Networks[network].IPAddress
			if requestIP != containerIP {
				fmt.Println("IP %s does not match container %s", requestIP, containerIP)
				continue
			}

			//Verify the container matches the role
			if data["image"] != container.Image {
				fmt.Println("Image %s does not match recorded image", container.Image, data["image"])
				continue
			}

			//id := container.ID

			match = append(match, data["policies"].(string))
		}
		if len(match) == 0 {
			fmt.Println("Request does not match any containers")
			http.Error(w, "401 Not Match", 401)
			return
		}
		matchingRole := match[0]

		wrappingClient, err := client.Clone()
		wrappingClient.SetWrappingLookupFunc(
			func(method string, lookupPath string) string {
				return "5m"
			},
		)
		renewable := false
		tokenCreate := vault.TokenCreateRequest{
			Policies: []string{matchingRole},
			NumUses: 1,
			TTL: "5m",
			ExplicitMaxTTL: "5m",
			Renewable: &renewable,
		}
		vaultAuth := wrappingClient.Auth()
		vaultToken := vaultAuth.Token()
		token, err := vaultToken.Create(&tokenCreate)
		if err != nil {
			http.Error(w, "503 Internal Error", 503)
			return
		}
		fmt.Fprintln(w,token.WrapInfo.Token)
	}
}
