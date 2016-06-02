package ultradns

import (
	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ULTRADNS_USERNAME", nil),
				Description: "UltraDNS Username.",
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ULTRADNS_PASSWORD", nil),
				Description: "UltraDNS User Password",
			},
			"baseurl": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ULTRADNS_BASEURL", nil),
				Default:     udnssdk.DefaultLiveBaseURL,
				Description: "UltraDNS Base URL",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"ultradns_dirpool":      resourceUltradnsDirpool(),
			"ultradns_notification": resourceUltradnsNotification(),
			"ultradns_probe":        resourceUltradnsProbe(),
			"ultradns_record":       resourceUltraDNSRecord(),
			"ultradns_tcpool":       resourceUltradnsTcpool(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
		BaseURL:  d.Get("baseurl").(string),
	}

	return config.Client()
}
