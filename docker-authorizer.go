package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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

	addr := "127.0.0.1:8000"
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}
	fmt.Println("Listening at ", addr )
	log.Fatal(srv.ListenAndServe())
}


func (service *DockerAuthorizer) ClientInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containers, err := service.dockerEngine.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			panic(err)
		}

		for _,  _ = range containers {
			//for _, container := range containers {
			//id := container.ID
			//image := container.Image
			//ip := container.NetworkSettings.Networks["bridge"].IPAddress
		}

		config := vault.DefaultConfig()
		client, err := vault.NewClient(config)
		secret, err := client.Logical().Read("secret/data/" + service.kvPrefix + "/example" )
		if err != nil {
			fmt.Printf("Errror communicating with Vault %s", err)
			fmt.Fprint(w, err)
			return
		}
		fmt.Println(secret)
		//data := secret.Data["data"].(map[string]interface{})
		//fmt.Println(data["test"])
		//value := data["test"].(string)

		wrappingClient, err := client.Clone()
		wrappingClient.SetWrappingLookupFunc(
			func(method string, lookupPath string) string {
				return "5m"
			},
		)
		renewable := false
		tokenCreate := vault.TokenCreateRequest{
			Policies: []string{"default"},
			NumUses: 1,
			TTL: "5m",
			ExplicitMaxTTL: "5m",
			Renewable: &renewable,
		}
		vaultAuth := wrappingClient.Auth()
		vaultToken := vaultAuth.Token()
		token, _ := vaultToken.Create(&tokenCreate)
		fmt.Fprintln(w,token.WrapInfo.Token, "\n",token.WrapInfo.WrappedAccessor)
	}
}
