package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sbehl27-org/terraform-provider-cidr-reservator/internal/provider/cidrCalculator"
	"github.com/sbehl27-org/terraform-provider-cidr-reservator/internal/provider/connector"
	"strconv"
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
	baseCidr := strings.Replace(strings.ReplaceAll(idContent[1], "_", "/"), "/", ".", 3)
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

//TODO: Throw error if netmask_id already exists!!!
func resourceServerCreate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceServerCreateWithRetry(ctx, data, m, 0)
}

func resourceServerCreateWithRetry(ctx context.Context, data *schema.ResourceData, m interface{}, retry int8) diag.Diagnostics {
	var diags diag.Diagnostics
	networkConfig, gcpConnector, err := readRemote(ctx, data, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if data.Get("id") != nil {
		return diags
	}
	netmaskId := data.Get("netmask_id").(string)
	if _, contains := networkConfig.Subnets[netmaskId]; contains {
		return diag.Errorf("The netmaskId %s already exists, but does not belong to your Terraform state!!!", netmaskId)
	}
	prefixLength := int8(data.Get("prefix_length").(int))
	if err != nil {
		return diag.FromErr(err)
	}
	newCidrCalculator, err := cidrCalculator.New(&networkConfig.Subnets, prefixLength, gcpConnector.BaseCidrRange)
	if err != nil {
		return diag.FromErr(err)
	}
	nextNetmask, err := newCidrCalculator.GetNextNetmask()
	if err != nil {
		return diag.FromErr(err)
	}
	networkConfig.Subnets[netmaskId] = nextNetmask
	err = gcpConnector.WriteRemote(networkConfig, ctx)
	if err != nil {
		if retry < 4 {
			return resourceServerCreateWithRetry(ctx, data, m, retry+1)
		} else {
			return diag.FromErr(err)
		}
	}
	data.SetId(fmt.Sprintf("%s:%s:%s", gcpConnector.BucketName, gcpConnector.BaseCidrRange, netmaskId))
	err = data.Set("netmask", nextNetmask)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceServerRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	//var diags diag.Diagnostics
	//_, _, err := readRemote(ctx, data, m)
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//return diags
	return nil
}

//TODO: implement with retry
func resourceServerUpdate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	networkConfig, gcpConnector, err := readRemote(ctx, data, m)
	if err != nil {
		return diag.FromErr(err)
	}
	netmaskId := data.Get("netmask_id").(string)
	prefixLength, err := strconv.ParseInt(data.Get("prefix_length").(string), 10, 8)
	if err != nil {
		return diag.FromErr(err)
	}
	delete(networkConfig.Subnets, "netmask_id")
	newCidrCalculator, err := cidrCalculator.New(&networkConfig.Subnets, int8(prefixLength), gcpConnector.BaseCidrRange)
	if err != nil {
		return diag.FromErr(err)
	}
	nextNetmask, err := newCidrCalculator.GetNextNetmask()
	if err != nil {
		return diag.FromErr(err)
	}
	networkConfig.Subnets[netmaskId] = nextNetmask
	err = gcpConnector.WriteRemote(networkConfig, ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(fmt.Sprintf("%s:%s:%s", gcpConnector.BucketName, gcpConnector.BaseCidrRange, netmaskId))
	err = data.Set("netmask", nextNetmask)
	if err != nil {
		return nil
	}
	return diags
}

//TODO: implement with retry!
func resourceServerDelete(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	networkConfig, gcpConnector, err := readRemote(ctx, data, m)
	if err != nil {
		return diag.FromErr(err)
	}
	netmaskId := data.Get("id").(string)
	if netmaskId != "" {
		return diags
	}
	delete(networkConfig.Subnets, netmaskId)
	err = gcpConnector.WriteRemote(networkConfig, ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
