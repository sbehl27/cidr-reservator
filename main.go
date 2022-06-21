// main.go
package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	//_ = resourceServerCreate(nil, nil)
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return Provider()
		},
	})

	//cidrs := map[string]string{"lappen": "10.5.0.0/26", "lulut": "10.5.1.0/24", "test": "10.5.0.64/26"}
	//calculator, _ := cidrCalculator.New(&cidrs, 24, "10.5.0.0/16")
	//netmask, _ := calculator.GetNextNetmask()
	////connector := connector.New("test-cidr-reservator", "10.5.0.0/16")
	////netmask, err := cidrCalculator.GetNextNetmask()
	//println(netmask)
}
