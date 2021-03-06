package okta

import (
	"errors"
	"fmt"

	"github.com/okta/okta-sdk-golang/okta/query"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGroupRead(d *schema.ResourceData, m interface{}) error {
	return findGroup(d.Get("name").(string), d, m)
}

func findGroup(name string, d *schema.ResourceData, m interface{}) error {
	client := getOktaClientFromMetadata(m)
	groups, _, err := client.Group.ListGroups(&query.Params{Q: name})
	if err != nil {
		return fmt.Errorf("failed to query for groups: %v", err)
	}
	if len(groups) > 0 {
		d.SetId(groups[0].Id)
		d.Set("description", groups[0].Profile.Description)
		return nil
	}

	return errors.New("Group not found")

}
