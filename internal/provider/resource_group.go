package provider

import (
	"context"
	"net/http"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: createGroup,
		ReadContext:   readGroup,
		DeleteContext: deleteGroup,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage group identities and creating and deleting groups.",
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
	api := m.(*Client)

	group, err := expandGroup(rd)
	if err != nil {
		return diag.FromErr(err)
	}
	groupResp, _, err := api.Jira.Group.Create(ctx, group.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	rd.SetId(groupResp.Name)

	return readGroup(ctx, rd, m)
}

func expandGroup(d *schema.ResourceData) (*models.GroupScheme, error) {
	group := &models.GroupScheme{}

	if d.IsNewResource() {
		group.Name = d.Get("name").(string)
	}

	return group, nil
}

func readGroup(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*Client)

	groups, resp, err := api.Jira.Group.Bulk(ctx, &models.GroupBulkOptionsScheme{
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
	api := m.(*Client)

	resp, err := api.Jira.Group.Delete(ctx, rd.Id())
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
