package provider

import (
	"context"
	"fmt"
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

func newMembershipFromID(id string) *membershipScheme {
	components := strings.Split(id, ":")
	if len(components) < 2 {
		return &membershipScheme{}
	}

	return &membershipScheme{
		GroupName: components[0],
		AccountID: components[1],
	}
}

func expandGroupMembership(d *schema.ResourceData) (*membershipScheme, error) {
	membership := &membershipScheme{}

	if d.IsNewResource() {
		membership.GroupName = d.Get("group_name").(string)
		membership.AccountID = d.Get("account_id").(string)
	}

	return membership, nil
}

type membershipScheme struct {
	GroupName string
	AccountID string
}

func createGroupMembership(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	membership, err := expandGroupMembership(rd)
	if err != nil {
		return diag.FromErr(err)
	}
	groupResp, resp, err := api.Group.Add(ctx, membership.GroupName, membership.AccountID)
	if err != nil {
		return diag.Errorf("creating status code: %s [groupname: %s accountID: %s]", resp.Code, membership.GroupName, membership.AccountID)
	}

	rd.SetId(fmt.Sprintf("%s:%s", groupResp.Name, membership.AccountID))

	return readGroupMembership(ctx, rd, m)
}

func readGroupMembership(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	membership := newMembershipFromID(rd.Id())

	user, resp, err := api.User.Get(ctx, membership.AccountID, []string{"groups"})
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	found := false
	for _, group := range user.Groups.Items {
		if group.Name == membership.GroupName {
			found = true
			break
		}
	}
	if !found {
		rd.SetId("")
		return nil
	}

	result := multierror.Append(
		rd.Set("group_name", membership.GroupName),
		rd.Set("account_id", membership.AccountID),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteGroupMembership(ctx context.Context, rd *schema.ResourceData, m any) diag.Diagnostics {
	api := m.(*jira.Client)

	membership := newMembershipFromID(rd.Id())
	resp, err := api.Group.Remove(ctx, membership.GroupName, membership.AccountID)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			rd.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
