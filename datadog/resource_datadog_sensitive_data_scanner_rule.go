package datadog

import (
	"context"
	"strconv"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogSensitiveDataScannerRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog SensitiveDataScannerRule resource. This can be used to create and manage Datadog sensitive_data_scanner_rule.",
		ReadContext:   resourceDatadogSensitiveDataScannerRuleRead,
		CreateContext: resourceDatadogSensitiveDataScannerRuleCreate,
		UpdateContext: resourceDatadogSensitiveDataScannerRuleUpdate,
		DeleteContext: resourceDatadogSensitiveDataScannerRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the rule.",
			},
			"excluded_attributes": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Attributes excluded from the scan.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether or not the rule is enabled.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the rule.",
			},
			"pattern": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Not included if there is a relationship to a standard pattern.",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of tags.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"text_replacement": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Object describing how the scanned event will be replaced.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number_of_chars": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Required if type == 'partial_replacement_from_beginning' or 'partial_replacement_from_end'. It must be > 0.",
						},
						"replacement_string": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Required if type == 'replacement_string'.",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Type of the replacement text. None means no replacement. hash means the data will be stubbed. replacement_string means that one can chose a text to replace the data. partial_replacement_from_beginning allows a user to partially replace the data from the beginning, and partial_replacement_from_end on the other hand, allows to replace data from the end.",
						},
					},
				},
			},
		},
	}
}

func resourceDatadogSensitiveDataScannerRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResp, "error calling ListScanningGroups")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	return updateSensitiveDataScannerRuleState(d, &resp)
}

func resourceDatadogSensitiveDataScannerRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	body := buildSensitiveDataScannerRuleRequestBody(d)

	resp, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().CreateScanningRule(auth, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating SensitiveDataScannerRule")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateSensitiveDataScannerRuleState(d, &resp)
}

func buildSensitiveDataScannerRuleRequestBody(d *schema.ResourceData) *datadogV2.SensitiveDataScannerRuleCreateRequest {
	attributes := datadogV2.NewSensitiveDataScannerRuleAttributesWithDefaults()

	if description, ok := d.GetOk("description"); ok {
		attributes.SetDescription(description.(string))
	}
	excludedAttributes := []string{}
	for _, s := range d.Get("excluded_attributes").([]interface{}) {
		excludedAttributes = append(excludedAttributes, s.(string))
	}
	attributes.SetExcludedAttributes(excludedAttributes)

	if isEnabled, ok := d.GetOk("is_enabled"); ok {
		attributes.SetIsEnabled(isEnabled.(bool))
	}

	if name, ok := d.GetOk("name"); ok {
		attributes.SetName(name.(string))
	}

	if pattern, ok := d.GetOk("pattern"); ok {
		attributes.SetPattern(pattern.(string))
	}
	tags := []string{}
	for _, s := range d.Get("tags").([]interface{}) {
		tags = append(tags, s.(string))
	}
	attributes.SetTags(tags)

	textReplacement := datadogV2.NewSensitiveDataScannerTextReplacementWithDefaults()

	if numberOfChars, ok := d.GetOk("number_of_chars"); ok {
		numberOfCharsInt, _ := strconv.ParseInt(numberOfChars.(string), 10, 64)
		textReplacement.SetNumberOfChars(numberOfCharsInt)
	}

	if replacementString, ok := d.GetOk("replacement_string"); ok {
		textReplacement.SetReplacementString(replacementString.(string))
	}

	if typeVar, ok := d.GetOk("type"); ok {
		typeVarItem, _ := datadogV2.NewSensitiveDataScannerTextReplacementTypeFromValue(typeVar.(string))
		textReplacement.SetType(*typeVarItem)
	}
	attributes.SetTextReplacement(*textReplacement)

	req := datadogV2.NewSensitiveDataScannerRuleCreateRequestWithDefaults()
	req.Data = *datadogV2.NewSensitiveDataScannerRuleCreateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogSensitiveDataScannerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()

	body := buildSensitiveDataScannerRuleUpdateRequestBody(d)

	resp, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().UpdateScanningRule(auth, id, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating SensitiveDataScannerRule")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateSensitiveDataScannerRuleState(d, &resp)
}

func buildSensitiveDataScannerRuleUpdateRequestBody(d *schema.ResourceData) *datadogV2.SensitiveDataScannerRuleUpdateRequest {
	attributes := datadogV2.NewSensitiveDataScannerRuleAttributesWithDefaults()

	if description, ok := d.GetOk("description"); ok {
		attributes.SetDescription(description.(string))
	}
	excludedAttributes := []string{}
	for _, s := range d.Get("excluded_attributes").([]interface{}) {
		excludedAttributes = append(excludedAttributes, s.(string))
	}
	attributes.SetExcludedAttributes(excludedAttributes)

	if isEnabled, ok := d.GetOk("is_enabled"); ok {
		attributes.SetIsEnabled(isEnabled.(bool))
	}

	if name, ok := d.GetOk("name"); ok {
		attributes.SetName(name.(string))
	}

	if pattern, ok := d.GetOk("pattern"); ok {
		attributes.SetPattern(pattern.(string))
	}
	tags := []string{}
	for _, s := range d.Get("tags").([]interface{}) {
		tags = append(tags, s.(string))
	}
	attributes.SetTags(tags)

	textReplacement := datadogV2.NewSensitiveDataScannerTextReplacementWithDefaults()

	if numberOfChars, ok := d.GetOk("number_of_chars"); ok {
		numberOfCharsInt, _ := strconv.ParseInt(numberOfChars.(string), 10, 64)
		textReplacement.SetNumberOfChars(numberOfCharsInt)
	}

	if replacementString, ok := d.GetOk("replacement_string"); ok {
		textReplacement.SetReplacementString(replacementString.(string))
	}

	if typeVar, ok := d.GetOk("type"); ok {
		typeVarItem, _ := datadogV2.NewSensitiveDataScannerTextReplacementTypeFromValue(typeVar.(string))
		textReplacement.SetType(*typeVarItem)
	}
	attributes.SetTextReplacement(*textReplacement)

	req := datadogV2.NewSensitiveDataScannerRuleUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewSensitiveDataScannerRuleUpdateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogSensitiveDataScannerRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()
	body := datadogV2.NewSensitiveDataScannerRuleDeleteRequestWithDefaults()
	metaVar := datadogV2.NewSensitiveDataScannerMetaVersionOnlyWithDefaults()

	if version, ok := d.GetOk("version"); ok {
		versionInt, _ := strconv.ParseInt(version.(string), 10, 64)
		metaVar.SetVersion(versionInt)
	}
	body.SetMeta(*metaVar)

	_, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().DeleteScanningRule(auth, id, body)
	if err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResp, "error deleting SensitiveDataScannerRule")
	}

	return nil
}

func updateSensitiveDataScannerRuleState(d *schema.ResourceData, resp *datadogV2.SensitiveDataScannerGetConfigResponse) diag.Diagnostics {

	return nil
}
