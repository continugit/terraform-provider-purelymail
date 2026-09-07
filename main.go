package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"gitlab.com/janschuermann/terraform/providers/purelymail/purelymail"
)

var version = "dev"

func main() {
	err := providerserver.Serve(context.Background(), purelymail.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/continugit/purelymail",
	})
	if err != nil {
		log.Fatal(err)
	}
}
