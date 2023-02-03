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

func newGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: createGroup,
		ReadContext:   readGroup,
		UpdateContext: schema.NoopContext,
		DeleteContext: deleteGroup,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage group identities and creating and deleting groups.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the group.",
			},
		},
	}
}

func createGroup(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	group, err := expandGroup(rd)
	if err != nil {
		return diag.FromErr(err)
	}
	groupResp, _, err := api.Group.Create(ctx, group.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	rd.SetId(groupResp.Name)

	return readGroup(ctx, rd, m)
}

func expandGroup(rd *schema.ResourceData) (*models.GroupScheme, error) {
	group := &models.GroupScheme{}

	if rd.HasChange("name") {
		group.Name = rd.Get("name").(string)
	}

	return group, nil
}

func readGroup(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	groups, resp, err := api.Group.Bulk(ctx, &models.GroupBulkOptionsScheme{
		GroupNames: []string{rd.Id()},
	}, 0, 1)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if groups.Total == 0 {
		rd.SetId("")
		return nil
	}

	group := groups.Values[0]

	result := multierror.Append(
		rd.Set("group_id", group.GroupID),
		rd.Set("name", group.Name),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteGroup(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	resp, err := api.Group.Delete(ctx, rd.Id())
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
