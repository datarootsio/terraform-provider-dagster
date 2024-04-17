//go:build tools
// +build tools

package main

import (
	_ "github.com/Khan/genqlient"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
