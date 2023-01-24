package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSensitiveDataScannerRuleSimple(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSensitiveDataScannerRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSensitiveDataScannerRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerRuleExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_sensitive_data_scanner_rule.foo", "description", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_sensitive_data_scanner_rule.foo", "is_enabled", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_sensitive_data_scanner_rule.foo", "name", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_sensitive_data_scanner_rule.foo", "pattern", "UPDATE ME"),
				),
			},
		},
	})
}

func testAccCheckDatadogSensitiveDataScannerRule(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_rule" "foo" {
    description = "UPDATE ME"
    excluded_attributes = ["username"]
    is_enabled = "UPDATE ME"
    name = "UPDATE ME"
    pattern = "UPDATE ME"
    tags = "UPDATE ME"
    text_replacement {
    number_of_chars = "UPDATE ME"
    replacement_string = "UPDATE ME"
    type = "UPDATE ME"
    }
}`, uniq)
}

func testAccCheckDatadogSensitiveDataScannerRuleDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := SensitiveDataScannerRuleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func SensitiveDataScannerRuleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_sensitive_data_scanner_rule" {
				continue
			}

			_, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogSensitiveDataScannerRuleExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := sensitiveDataScannerRuleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func sensitiveDataScannerRuleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_sensitive_data_scanner_rule" {
			continue
		}

		_, httpResp, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving monitor")
		}
	}
	return nil
}
