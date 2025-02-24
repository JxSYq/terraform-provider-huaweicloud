package huaweicloud

import (
	"testing"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHuaweiCloudRdsFlavorV3DataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHuaweiCloudRdsFlavorV3DataSource_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsFlavorV3DataSourceID("data.huaweicloud_rds_flavors_v3.flavor"),
				),
			},
		},
	})
}

func testAccCheckRdsFlavorV3DataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmtp.Errorf("Can't find rds data source: %s ", n)
		}

		if rs.Primary.ID == "" {
			return fmtp.Errorf("Rds data source ID not set ")
		}

		return nil
	}
}

var testAccHuaweiCloudRdsFlavorV3DataSource_basic = `

data "huaweicloud_rds_flavors_v3" "flavor" {
  db_type = "PostgreSQL"
  db_version = "10"
  instance_mode = "ha"
}
`
