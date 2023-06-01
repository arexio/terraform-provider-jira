package provider

import (
	"context"

	"github.com/ctreminiom/go-atlassian/admin"
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
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("JIRA_TOKEN", nil),
				Description: "Your Jira user token. " +
					"It can also be sourced from the `JIRA_TOKEN` environment variable.",
			},
			"admin_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ADMIN_TOKEN", nil),
				Description: "Your Jira admin token. " +
					"It can also be sourced from the `ADMIN_TOKEN` environment variable.",
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

type Client struct {
	Jira  *jira.Client
	Admin *admin.Client
}

func configureProvider(tfVersion string) schema.ConfigureContextFunc {
	return func(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var (
			cli Client
			err error
		)

		cli.Jira, err = jira.New(nil, rd.Get("domain").(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}
		cli.Jira.Auth.SetBasicAuth(rd.Get("email").(string), rd.Get("token").(string))

		cli.Admin, err = admin.New(nil)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		cli.Admin.Auth.SetBearerToken(rd.Get("admin_token").(string))

		return &cli, nil
	}
}
