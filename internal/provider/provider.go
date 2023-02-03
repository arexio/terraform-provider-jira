package provider

import (
	"context"

	jira "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var version = "dev"

// New returns a *schema.Provider.
func New() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JIRA_DOMAIN", nil),
				Description: "Your Jira domain name. " +
					"It can also be sourced from the `JIRA_DOMAIN` environment variable.",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JIRA_USER_EMAIL", nil),
				Description: "Your Jira user email. " +
					"It can also be sourced from the `JIRA_USER_EMAIL` environment variable.",
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JIRA_TOKEN", nil),
				Description: "Your Jira user token. " +
					"It can also be sourced from the `JIRA_TOKEN` environment variable.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"jira_user":             newUser(),
			"jira_group":            newGroup(),
			"jira_group_membership": newGroupMembership(),
		},
	}

	provider.ConfigureContextFunc = configureProvider(provider.TerraformVersion)

	return provider
}

func configureProvider(tfVersion string) schema.ConfigureContextFunc {
	return func(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
		client, err := jira.New(nil, rd.Get("domain").(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}

		client.Auth.SetBasicAuth(rd.Get("email").(string), rd.Get("token").(string))

		return client, nil
	}
}
