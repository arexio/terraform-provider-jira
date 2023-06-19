package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/arexio/terraform-provider-jira/internal/provider"
)

// Ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Generate documentation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	// Set descriptions to support Markdown syntax,
	// this will be used in document generation.
	schema.DescriptionKind = schema.StringMarkdown

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.New,
	})
}
