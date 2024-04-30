package testutils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func FetchValueFromState(resourceName string, attributeName string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource (%s) not found in state", resourceName)
		}

		val, ok := rs.Primary.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("error fetching attribute (%s) from resource (%s)", attributeName, resourceName)
		}

		*target = val

		return nil
	}
}
