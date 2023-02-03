package provider

import (
	"context"
	"net/http"
	"strings"

	jira "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: createGroupMembership,
		ReadContext:   readGroupMembership,
		UpdateContext: schema.NoopContext,
		DeleteContext: deleteGroupMembership,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage group identities and creating and deleting groups.",
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the group.",
			},
			"account_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Account id of the user.",
			},
		},
	}
}

type groupMembershipScheme struct {
	GroupName string
	AccountID string
}

func createGroupMembership(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	membership, err := expandGroupMembership(rd)
	if err != nil {
		return diag.FromErr(err)
	}
	groupResp, _, err := api.Group.Add(ctx, membership.GroupName, membership.AccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	rd.SetId(groupResp.Name)

	return readGroupMembership(ctx, rd, m)
}

func expandGroupMembership(rd *schema.ResourceData) (*groupMembershipScheme, error) {
	membership := &groupMembershipScheme{}

	if rd.HasChange("group_name") {
		membership.GroupName = rd.Get("group_name").(string)
	}

	if rd.HasChange("account_id") {
		membership.AccountID = rd.Get("account_id").(string)
	}

	return membership, nil
}

func readGroupMembership(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	groupName, accountID := groupNameAccountIDFromID(rd.Id())

	user, resp, err := api.User.Get(ctx, accountID, []string{"groups"})
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	found := false
	for _, group := range user.Groups.Items {
		if group.Name == groupName {
			found = true
			break
		}
	}
	if !found {
		rd.SetId("")
		return nil
	}

	result := multierror.Append(
		rd.Set("group_name", groupName),
		rd.Set("account_id", accountID),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteGroupMembership(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	groupName, accountID := groupNameAccountIDFromID(rd.Id())
	resp, err := api.Group.Remove(ctx, groupName, accountID)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

func groupNameAccountIDFromID(id string) (string, string) {
	components := strings.Split(id, ":")
	return components[0], components[1]
}
