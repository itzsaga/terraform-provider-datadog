package datadog

import (
	"context"
	"fmt"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var integrationAzureMutex = sync.Mutex{}

func resourceDatadogIntegrationAzure() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.",
		CreateContext: resourceDatadogIntegrationAzureCreate,
		ReadContext:   resourceDatadogIntegrationAzureRead,
		UpdateContext: resourceDatadogIntegrationAzureUpdate,
		DeleteContext: resourceDatadogIntegrationAzureDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"tenant_name": {
				Description: "Your Azure Active Directory ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"client_id": {
				Description: "Your Azure web application ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"client_secret": {
				Description: "(Required for Initial Creation) Your Azure web application secret key.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"host_filters": {
				Description: "String of host tag(s) (in the form `key:value,key:value`) defines a filter that Datadog will use when collecting metrics from Azure. Limit the Azure instances that are pulled into Datadog by using tags. Only hosts that match one of the defined tags are imported into Datadog. e.x. `env:production,deploymentgroup:red`",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"automute": {
				Description: "Silence monitors for expected Azure VM shutdowns.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
		},
	}
}

func resourceDatadogIntegrationAzureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	tenantName, clientId, err := utils.TenantAndClientFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	integrations, httpresp, err := apiInstances.GetAzureIntegrationApiV1().ListAzureIntegration(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error listing azure integration")
	}
	if err := utils.CheckForUnparsed(integrations); err != nil {
		return diag.FromErr(err)
	}
	for _, integration := range integrations {
		if integration.GetTenantName() == tenantName && integration.GetClientId() == clientId {
			d.Set("tenant_name", integration.GetTenantName())
			d.Set("client_id", integration.GetClientId())
			d.Set("automute", integration.GetAutomute())
			hostFilters, exists := integration.GetHostFiltersOk()
			if exists {
				d.Set("host_filters", hostFilters)
			}
			return nil
		}
	}
	return diag.Errorf("error getting an Azure integration: tenant_name=%s", tenantName)
}

func resourceDatadogIntegrationAzureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationAzureMutex.Lock()
	defer integrationAzureMutex.Unlock()

	tenantName := d.Get("tenant_name").(string)
	clientID := d.Get("client_id").(string)

	iazure := buildDatadogAzureIntegrationDefinition(d, tenantName, clientID, false)

	if _, httpresp, err := apiInstances.GetAzureIntegrationApiV1().CreateAzureIntegration(auth, *iazure); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating an Azure integration")
	}

	d.SetId(fmt.Sprintf("%s:%s", iazure.GetTenantName(), iazure.GetClientId()))

	return resourceDatadogIntegrationAzureRead(ctx, d, meta)
}

func resourceDatadogIntegrationAzureUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationAzureMutex.Lock()
	defer integrationAzureMutex.Unlock()

	existingTenantName, existingClientID, err := utils.TenantAndClientFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	iazure := buildDatadogAzureIntegrationDefinition(d, existingTenantName, existingClientID, true)

	if _, httpresp, err := apiInstances.GetAzureIntegrationApiV1().UpdateAzureIntegration(auth, *iazure); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating an Azure integration")
	}

	d.SetId(fmt.Sprintf("%s:%s", iazure.GetNewTenantName(), iazure.GetNewClientId()))

	return resourceDatadogIntegrationAzureRead(ctx, d, meta)
}

func resourceDatadogIntegrationAzureDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationAzureMutex.Lock()
	defer integrationAzureMutex.Unlock()

	tenantName, clientID, err := utils.TenantAndClientFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	iazure := buildDatadogAzureIntegrationDefinition(d, tenantName, clientID, false)

	if _, httpresp, err := apiInstances.GetAzureIntegrationApiV1().DeleteAzureIntegration(auth, *iazure); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting an Azure integration")
	}

	return nil
}

func buildDatadogAzureIntegrationDefinition(terraformDefinition *schema.ResourceData, tenantName string, clientID string, update bool) *datadogV1.AzureAccount {
	datadogDefinition := datadogV1.NewAzureAccount()
	// Required params
	datadogDefinition.SetTenantName(tenantName)
	datadogDefinition.SetClientId(clientID)
	// Optional params
	hostFilters := terraformDefinition.Get("host_filters")
	datadogDefinition.SetHostFilters(hostFilters.(string))
	automute := terraformDefinition.Get("automute")
	datadogDefinition.SetAutomute(automute.(bool))

	clientSecret, exists := terraformDefinition.GetOk("client_secret")
	if exists {
		datadogDefinition.SetClientSecret(clientSecret.(string))
	}
	// Only do the following if building for the Update
	if update {
		newTenantName, exists := terraformDefinition.GetOk("tenant_name")
		if exists {
			datadogDefinition.SetNewTenantName(newTenantName.(string))
		}
		newClientID, exists := terraformDefinition.GetOk("client_id")
		if exists {
			datadogDefinition.SetNewClientId(newClientID.(string))
		}
	}
	return datadogDefinition
}
