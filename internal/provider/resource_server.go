package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sbehl27-org/terraform-provider-cidr-reservator/internal/provider/cidrCalculator"
	"github.com/sbehl27-org/terraform-provider-cidr-reservator/internal/provider/connector"
	"strings"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"prefix_length": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"base_cidr": {
				Type:     schema.TypeString,
				Required: true,
			},
			"netmask_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"netmask": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: importState,
		},
	}
}

func importState(ctx context.Context, data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	idContent := strings.Split(data.Id(), ":")
	reservatorBucket := idContent[0]
	baseCidr := idContent[1]
	netmaskId := idContent[2]
	gcpConnector := connector.New(reservatorBucket, baseCidr)
	networkConfig, err := gcpConnector.ReadRemote(ctx)
	if err != nil {
		return nil, err
	}
	subnet, contains := networkConfig.Subnets[netmaskId]
	if !contains {
		return nil, errors.New(fmt.Sprintf("Netmask with id %s does not exist!", netmaskId))
	}
	prefixLength := strings.Split(subnet, "/")[1]
	data.Set("base_cidr", baseCidr)
	data.Set("netmask_id", netmaskId)
	data.Set("prefix_length", prefixLength)
	return []*schema.ResourceData{data}, nil
}

func readRemote(ctx context.Context, data *schema.ResourceData, m interface{}) (*connector.NetworkConfig, *connector.GcpConnector, error) {
	cidrProviderBucket := m.(string)
	gcpConnector := connector.New(cidrProviderBucket, data.Get("base_cidr").(string))
	networkConfig, err := gcpConnector.ReadRemote(ctx)
	if err != nil {
		if err.Error() == "storage: object doesn't exist" {
			err = nil
			networkConfig = &connector.NetworkConfig{Subnets: make(map[string]string)}
		} else {
			return nil, nil, err
		}
	}
	//netmaskId := data.Get("netmask_id").(string)
	//data.SetId("")
	//if netmask, contains := networkConfig.Subnets[netmaskId]; contains {
	//	data.SetId(netmaskId)
	//	err = data.Set("netmask", netmask)
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//}
	return networkConfig, &gcpConnector, nil
}

func retry(toRetry func() error) error {
	n := 0
	var err error
	for n < 4 {
		err = toRetry()
		if err == nil {
			break
		}
		n++
	}
	return err
}

func resourceServerCreate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	err := retry(innerResourceServerCreate(ctx, data, m))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func innerResourceServerCreate(ctx context.Context, data *schema.ResourceData, m interface{}) func() error {
	return func() error {
		networkConfig, gcpConnector, err := readRemote(ctx, data, m)
		if err != nil {
			return err
		}
		if data.Get("id") != nil {
			return nil
		}
		netmaskId := data.Get("netmask_id").(string)
		if _, contains := networkConfig.Subnets[netmaskId]; contains {
			return fmt.Errorf("The netmaskId %s already exists, but does not belong to your Terraform state!!!", netmaskId)
		}
		prefixLength := int8(data.Get("prefix_length").(int))
		if err != nil {
			return err
		}
		newCidrCalculator, err := cidrCalculator.New(&networkConfig.Subnets, prefixLength, gcpConnector.BaseCidrRange)
		if err != nil {
			return err
		}
		nextNetmask, err := newCidrCalculator.GetNextNetmask()
		if err != nil {
			return err
		}
		networkConfig.Subnets[netmaskId] = nextNetmask
		err = gcpConnector.WriteRemote(networkConfig, ctx)
		if err != nil {
			return err
		}
		data.SetId(fmt.Sprintf("%s:%s:%s", gcpConnector.BucketName, gcpConnector.BaseCidrRange, netmaskId))
		err = data.Set("netmask", nextNetmask)
		if err != nil {
			return err
		}
		return nil
	}
}

func resourceServerRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceServerUpdate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	err := retry(innerResourceServerUpdate(ctx, data, m))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func innerResourceServerUpdate(ctx context.Context, data *schema.ResourceData, m interface{}) func() error {
	return func() error {
		networkConfig, gcpConnector, err := readRemote(ctx, data, m)
		if err != nil {
			return err
		}
		netmaskId := data.Get("netmask_id").(string)
		prefixLength := int8(data.Get("prefix_length").(int))
		if err != nil {
			return err
		}
		delete(networkConfig.Subnets, "netmask_id")
		newCidrCalculator, err := cidrCalculator.New(&networkConfig.Subnets, int8(prefixLength), gcpConnector.BaseCidrRange)
		if err != nil {
			return err
		}
		nextNetmask, err := newCidrCalculator.GetNextNetmask()
		if err != nil {
			return err
		}
		networkConfig.Subnets[netmaskId] = nextNetmask
		err = gcpConnector.WriteRemote(networkConfig, ctx)
		if err != nil {
			return err
		}
		data.SetId(fmt.Sprintf("%s:%s:%s", gcpConnector.BucketName, gcpConnector.BaseCidrRange, netmaskId))
		err = data.Set("netmask", nextNetmask)
		if err != nil {
			return err
		}
		return nil
	}
}

func resourceServerDelete(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	err := retry(innerResourceServerDelete(ctx, data, m))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func innerResourceServerDelete(ctx context.Context, data *schema.ResourceData, m interface{}) func() error {
	return func() error {
		networkConfig, gcpConnector, err := readRemote(ctx, data, m)
		if err != nil {
			return err
		}
		netmaskId := data.Get("netmask_id").(string)
		delete(networkConfig.Subnets, netmaskId)
		err = gcpConnector.WriteRemote(networkConfig, ctx)
		if err != nil {
			return err
		}
		return nil
	}
}
