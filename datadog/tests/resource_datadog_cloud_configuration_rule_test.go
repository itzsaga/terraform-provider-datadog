package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

const tfCloudConfRuleName = "datadog_cloud_configuration_rule.acceptance_test"

func TestAccDatadogCloudConfigurationRule_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCloudConfigurationRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudConfigurationCreatedConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationCreatedCheck(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogCloudConfigurationUpdatedConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationUpdatedCheck(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogCloudConfigurationUpdatedMandatoryFieldsConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationUpdatedMandatoryFieldsCheck(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogCloudConfigurationRule_MandatoryFieldsOnly(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCloudConfigurationRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudConfigurationCreatedMandatoryFieldsConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationCreatedMandatoryFieldsCheck(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogCloudConfigurationRule_Import(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCloudConfigurationRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudConfigurationCreatedConfig(ruleName),
			},
			{
				ResourceName:      tfCloudConfRuleName,
				ImportState:       true,
				ImportStateVerify: true,
				Check:             testAccCheckDatadogCloudConfigurationCreatedCheck(accProvider, ruleName),
			},
		},
	})
}

func testAccCheckDatadogCloudConfigurationCreatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = false
  message = "Acceptance test TF rule"
  name    = "%s"
  notifications = [ "@channel" ]
  group_by = [ "@resource" ]
  policy = "package datadog"
  resource_type = "gcp_compute_instance"
  related_resource_types = [ "gcp_compute_disk" ]
  severity = "low"
  tags = [
    "test:acceptance",
    "terraform:true",
  ]
}
`, name)
}

func testAccCheckDatadogCloudConfigurationCreatedCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudConfRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "notifications.0", "@channel"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.0", "@resource"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0", "gcp_compute_disk"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.0", "test:acceptance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.1", "terraform:true"),
	)
}

func testAccCheckDatadogCloudConfigurationUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = true
  message = "Acceptance test TF rule - updated"
  name    = "%s - updated"
  notifications = [ "@channel-upd" ]
  group_by = [ "@resource", "@resource_type" ]
  policy = "package datadog # updated"
  resource_type = "gcp_compute_disk"
  related_resource_types = [ "gcp_compute_instance", "gcp_compute_firewall" ]
  severity = "high"
  tags = [ "test:acceptance-updated" ]
}
`, name)
}

func testAccCheckDatadogCloudConfigurationUpdatedCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudConfRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule - updated"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName+" - updated"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "notifications.0", "@channel-upd"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.0", "@resource"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.1", "@resource_type"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog # updated"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_disk"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0", "gcp_compute_instance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.1", "gcp_compute_firewall"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "high"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.0", "test:acceptance-updated"),
	)
}

func testAccCheckDatadogCloudConfigurationUpdatedMandatoryFieldsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = false
  message = "Acceptance test TF rule - updated again"
  name    = "%s - updated again"
  policy = "package datadog # updated again"
  resource_type = "gcp_compute_instance"
  severity = "medium"
  tags = [ "test:acceptance-updated-again" ]
}
`, name)
}

func testAccCheckDatadogCloudConfigurationUpdatedMandatoryFieldsCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudConfRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule - updated again"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName+" - updated again"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "notifications.0"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "group_by.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog # updated again"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "medium"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.0", "test:acceptance-updated-again"),
	)
}

func testAccCheckDatadogCloudConfigurationCreatedMandatoryFieldsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = false
  message = "Acceptance test TF rule"
  name    = "%s"
  policy = "package datadog"
  resource_type = "gcp_compute_instance"
  severity = "low"
}
`, name)
}

func testAccCheckDatadogCloudConfigurationCreatedMandatoryFieldsCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudConfRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "notifications"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "group_by"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "related_resource_types"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "tags"),
	)
}

func testAccCheckDatadogCloudConfRuleExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, rule := range s.RootModule().Resources {
			_, _, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, rule.Primary.ID)
			if err != nil {
				return fmt.Errorf("received an error retrieving security monitoring rule: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogCloudConfigurationRuleDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_cloud_configuration_rule" {
				_, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, resource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == 404 {
						continue
					}
					return fmt.Errorf("received an error deleting cloud configuration rule: %s", err)
				}
				return fmt.Errorf("cloud configuration rule still exists")
			}
		}
		return nil
	}
}
