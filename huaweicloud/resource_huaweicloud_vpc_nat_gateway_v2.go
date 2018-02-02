package huaweicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/huawei-clouds/golangsdk"
	"github.com/huawei-clouds/golangsdk/openstack/vpc/v2/natgateways"
)

func resourceVpcNatGatewayV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceVpcNatGatewayV2Create,
		Read:   resourceVpcNatGatewayV2Read,
		Update: resourceVpcNatGatewayV2Update,
		Delete: resourceVpcNatGatewayV2Delete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"spec": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				ValidateFunc: resourceNatGatewayV2ValidateSpec,
			},
			"tenant_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"router_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"internal_network_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVpcNatGatewayV2Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	vpcV2Client, err := config.vpcV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating HuaweiCloud vpc client: %s", err)
	}

	createOpts := &natgateways.CreateOpts{
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		Spec:              d.Get("spec").(string),
		TenantID:          d.Get("tenant_id").(string),
		RouterID:          d.Get("router_id").(string),
		InternalNetworkID: d.Get("internal_network_id").(string),
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	natGateway, err := natgateways.Create(vpcV2Client, createOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error creatting Nat Gateway: %s", err)
	}

	log.Printf("[DEBUG] Waiting for HuaweiCloud Nat Gateway (%s) to become available.", natGateway.ID)

	stateConf := &resource.StateChangeConf{
		Target:     []string{"ACTIVE"},
		Refresh:    waitForNatGatewayActive(vpcV2Client, natGateway.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error creating HuaweiCloud Nat Gateway: %s", err)
	}

	d.SetId(natGateway.ID)

	return resourceVpcNatGatewayV2Read(d, meta)
}

func resourceVpcNatGatewayV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	vpcV2Client, err := config.vpcV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating HuaweiCloud vpc client: %s", err)
	}

	natGateway, err := natgateways.Get(vpcV2Client, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "Nat Gateway")
	}

	d.Set("name", natGateway.Name)
	d.Set("description", natGateway.Description)
	d.Set("spec", natGateway.Spec)
	d.Set("router_id", natGateway.RouterID)
	d.Set("internal_network_id", natGateway.InternalNetworkID)
	d.Set("tenant_id", natGateway.TenantID)

	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceVpcNatGatewayV2Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	vpcV2Client, err := config.vpcV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating HuaweiCloud vpc client: %s", err)
	}

	var updateOpts natgateways.UpdateOpts

	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}
	if d.HasChange("description") {
		updateOpts.Description = d.Get("description").(string)
	}
	if d.HasChange("spec") {
		updateOpts.Spec = d.Get("spec").(string)
	}

	log.Printf("[DEBUG] Update Options: %#v", updateOpts)

	_, err = natgateways.Update(vpcV2Client, d.Id(), updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("Error updating Nat Gateway: %s", err)
	}

	return resourceVpcNatGatewayV2Read(d, meta)
}

func resourceVpcNatGatewayV2Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	vpcV2Client, err := config.vpcV2Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating HuaweiCloud vpc client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    waitForNatGatewayDelete(vpcV2Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error deleting HuaweiCloud Nat Gateway: %s", err)
	}

	d.SetId("")
	return nil
}

func waitForNatGatewayActive(vpcV2Client *golangsdk.ServiceClient, nId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := natgateways.Get(vpcV2Client, nId).Extract()
		if err != nil {
			return nil, "", err
		}

		log.Printf("[DEBUG] HuaweiCloud Nat Gateway: %+v", n)
		if n.Status == "ACTIVE" {
			return n, "ACTIVE", nil
		}

		return n, "", nil
	}
}

func waitForNatGatewayDelete(vpcV2Client *golangsdk.ServiceClient, nId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Attempting to delete HuaweiCloud Nat Gateway %s.\n", nId)

		n, err := natgateways.Get(vpcV2Client, nId).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				log.Printf("[DEBUG] Successfully deleted HuaweiCloud Nat gateway %s", nId)
				return n, "DELETED", nil
			}
			return n, "ACTIVE", err
		}

		err = natgateways.Delete(vpcV2Client, nId).ExtractErr()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				log.Printf("[DEBUG] Successfully deleted HuaweiCloud Nat Gateway %s", nId)
				return n, "DELETED", nil
			}
			return n, "ACTIVE", err
		}

		log.Printf("[DEBUG] HuaweiCloud Nat Gateway %s still active.\n", nId)
		return n, "ACTIVE", nil
	}
}

var Specs = [4]string{"1", "2", "3", "4"}

func resourceNatGatewayV2ValidateSpec(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	for i := range Specs {
		if value == Specs[i] {
			return
		}
	}
	errors = append(errors, fmt.Errorf("%q must be one of %v", k, Specs))
	return
}
