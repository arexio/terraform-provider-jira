package provider

import (
	"context"
	"net/http"

	jira "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: createUser,
		ReadContext:   readUser,
		UpdateContext: schema.NoopContext,
		DeleteContext: deleteUser,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage user identities and creating and deleting users.",
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Email address of the user.",
			},
		},
	}
}

func createUser(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	user, err := expandUser(rd)
	if err != nil {
		return diag.FromErr(err)
	}
	userResp, _, err := api.User.Create(ctx, &models.UserPayloadScheme{
		EmailAddress: user.EmailAddress,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	rd.SetId(userResp.AccountID)

	return readUser(ctx, rd, m)
}

func expandUser(rd *schema.ResourceData) (*models.UserScheme, error) {
	user := &models.UserScheme{}

	if rd.HasChange("email") {
		user.EmailAddress = rd.Get("email").(string)
	}

	return user, nil
}

func readUser(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	user, resp, err := api.User.Get(ctx, rd.Id(), nil)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		rd.Set("account_id", user.AccountID),
		rd.Set("account_type", user.AccountType),
		rd.Set("email", user.EmailAddress),
		rd.Set("display_name", user.DisplayName),
		rd.Set("active", user.Active),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteUser(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	resp, err := api.User.Delete(ctx, rd.Id())
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
