package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/pinkroccade/terraform-provider-mcs/internal/provider"
)

var version = "dev"

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/pinkroccade/mcs",
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
