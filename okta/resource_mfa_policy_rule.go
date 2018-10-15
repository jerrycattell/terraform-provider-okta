package okta

import (
	articulateOkta "github.com/articulate/oktasdk-go/okta"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceMfaPolicyRule() *schema.Resource {
	return &schema.Resource{
		Exists:   resourcePolicyRuleExists,
		Create:   resourceMfaPolicyRuleCreate,
		Read:     resourceMfaPolicyRuleRead,
		Update:   resourceMfaPolicyRuleUpdate,
		Delete:   resourceMfaPolicyRuleDelete,
		Importer: createPolicyRuleImporter(),

		Schema: buildRuleSchema(map[string]*schema.Schema{
			"enroll": &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"CHALLENGE", "LOGIN", "NEVER"}, false),
				Description:  "Should the user be enrolled the first time they LOGIN, the next time they are CHALLENGEd, or NEVER?",
			},
		}),
	}
}

func resourceMfaPolicyRuleCreate(d *schema.ResourceData, m interface{}) error {
	if err := ensureNotDefaultRule(d); err != nil {
		return err
	}

	client := getClientFromMetadata(m)
	template := buildMfaPolicyRule(d, client)
	rule, err := createRule(d, m, template, passwordPolicyRule)
	if err != nil {
		return err
	}

	d.SetId(rule.ID)
	err = validatePriority(template.Priority, rule.Priority)
	if err != nil {
		return err
	}

	return resourceMfaPolicyRuleRead(d, m)
}

func resourceMfaPolicyRuleRead(d *schema.ResourceData, m interface{}) error {
	rule, err := getPolicyRule(d, m)
	if err != nil {
		return err
	}

	return syncRuleFromUpstream(d, rule)
}

func resourceMfaPolicyRuleUpdate(d *schema.ResourceData, m interface{}) error {
	if err := ensureNotDefaultRule(d); err != nil {
		return err
	}

	client := getClientFromMetadata(m)
	template := buildMfaPolicyRule(d, client)

	rule, err := updateRule(d, m, template)
	if err != nil {
		return err
	}

	err = validatePriority(template.Priority, rule.Priority)
	if err != nil {
		return err
	}

	return resourceMfaPolicyRuleRead(d, m)
}

func resourceMfaPolicyRuleDelete(d *schema.ResourceData, m interface{}) error {
	if err := ensureNotDefaultRule(d); err != nil {
		return err
	}

	client := getClientFromMetadata(m)
	_, err := client.Policies.DeletePolicyRule(d.Get("policyid").(string), d.Id())

	return err
}

// build password policy rule from schema data
func buildMfaPolicyRule(d *schema.ResourceData, client *articulateOkta.Client) *articulateOkta.MfaRule {
	rule := client.Policies.MfaRule()
	rule.Name = d.Get("name").(string)
	rule.Status = d.Get("status").(string)
	if priority, ok := d.GetOk("priority"); ok {
		rule.Priority = priority.(int)
	}

	rule.Conditions = &articulateOkta.PolicyConditions{
		Network: getNetwork(d),
		People:  getUsers(d),
	}

	if enroll, ok := d.GetOk("enroll"); ok {
		rule.Actions = &articulateOkta.MfaRuleActions{
			Enroll: &articulateOkta.Enroll{
				Self: enroll.(string),
			},
		}
	}

	return &rule
}
