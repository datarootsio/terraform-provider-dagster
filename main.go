package main

import (
	"fmt"
	"os"

	"github.com/datarootsio/terraform-provider-dagster/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
)

const providerAddress = "registry.terraform.io/datarootsio/dagster"

// Run "go generate" to generate the docs
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --rendered-provider-name Dagster --provider-name dagster

func main() {
	providerServer := providerserver.NewProtocol6(&provider.DagsterProvider{})

	err := tf6server.Serve(providerAddress, providerServer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start starting plugin server: %s", err)
		os.Exit(1)
	}
}
