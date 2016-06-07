package ultradns

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUltradnsProbe() *schema.Resource {
	return &schema.Resource{
		Create: resourceUltradnsProbeCreate,
		Read:   resourceUltradnsProbeRead,
		Update: resourceUltradnsProbeUpdate,
		Delete: resourceUltradnsProbeDelete,

		Schema: map[string]*schema.Schema{
			// Required
			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pool_record": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"agents": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"threshold": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			// Optional
			"interval": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ping_probe": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     schemaPingProbe(),
			},
			// Computed
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func schemaPingProbe() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"packets": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"packet_size": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"limit": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Set:      hashLimits,
				Elem:     resourceProbeLimits(),
			},
		},
	}
}

func resourceProbeLimits() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"warning": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"critical": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"fail": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceUltradnsProbeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe configuration: %v", err)
	}

	log.Printf("[INFO] ultradns_probe create: %+v", r)
	resp, err := client.Probes.Create(r.Key().RRSetKey(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("create failed: %v", err)
	}

	uri := resp.Header.Get("Location")
	d.Set("uri", uri)
	id := resp.Header.Get("ID")
	d.SetId(id)
	log.Printf("[INFO] ultradns_probe.id: %v", d.Id())

	return resourceUltradnsProbeRead(d, meta)
}

func resourceUltradnsProbeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe configuration: %v", err)
	}

	log.Printf("[DEBUG] ultradns_probe read: %+v", r)
	probe, _, err := client.Probes.Find(r.Key())

	if err != nil {
		uderr, ok := err.(*udnssdk.ErrorResponseList)
		if ok {
			for _, r := range uderr.Responses {
				// 70002 means Probes Not Found
				if r.ErrorCode == 70002 {
					d.SetId("")
					return nil
				}
				return fmt.Errorf("not found: %s", err)
			}
		}
		return fmt.Errorf("not found: %s", err)
	}

	return populateResourceDataFromProbe(probe, d)
}

func resourceUltradnsProbeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe configuration: %v", err)
	}

	log.Printf("[INFO] ultradns_probe update: %+v", r)
	_, err = client.Probes.Update(r.Key(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("update failed: %s", err)
	}

	return resourceUltradnsProbeRead(d, meta)
}

func resourceUltradnsProbeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe configuration: %s", err)
	}

	log.Printf("[INFO] ultradns_probe delete: %+v", r)
	_, err = client.Probes.Delete(r.Key())
	if err != nil {
		return fmt.Errorf("delete failed: %s", err)
	}

	return resourceUltradnsProbeRead(d, meta)
}

// Resource Helpers

func newProbeResource(d *schema.ResourceData) (probeResource, error) {
	p := probeResource{}
	p.Zone = d.Get("zone").(string)
	p.Name = d.Get("name").(string)
	p.ID = d.Id()
	p.Interval = d.Get("interval").(string)
	p.PoolRecord = d.Get("pool_record").(string)
	p.Threshold = d.Get("threshold").(int)
	for _, a := range d.Get("agents").([]interface{}) {
		p.Agents = append(p.Agents, a.(string))
	}

	p.Type = udnssdk.ProbeType(d.Get("type").(string))
	if !isValidType(p.Type) {
		return p, fmt.Errorf("type invalid: %v", p.Type)
	}

	// details
	// TODO: validate p.Type is in typeToAttrKeyMap.Keys
	detailsAttr := d.Get(p.detailsAttribute()).([]interface{})
	if len(detailsAttr) >= 1 {
		// TODO: validate 1 >= len >= 0
		// TODO: case(p.Type) -> makeFooDetail
		probedetails := detailsAttr[0].(map[string]interface{})
		ls := map[string]interface{}{}
		for _, limit := range probedetails["limit"].(*schema.Set).List() {
			l := limit.(map[string]interface{})
			name := l["name"].(string)
			ls[name] = makeProbeDetailsLimit(l)
		}
		p.Details = &udnssdk.ProbeDetailsDTO{
			Detail: map[string]interface{}{
				"limits": ls,
			},
		}
	}

	return p, nil
}

func (p probeResource) detailsAttribute() string {
	return typeToAttrKeyMap[p.Type]
}

var typeToAttrKeyMap = map[udnssdk.ProbeType]string{
	udnssdk.DNSProbeType:      "dns_probe",
	udnssdk.FTPProbeType:      "ftp_probe",
	udnssdk.HTTPProbeType:     "http_probe",
	udnssdk.PingProbeType:     "ping_probe",
	udnssdk.SMTPProbeType:     "smtp_probe",
	udnssdk.SMTPSENDProbeType: "smtpsend_probe",
}

func isValidType(t udnssdk.ProbeType) bool {
	return t == udnssdk.DNSProbeType ||
		t == udnssdk.FTPProbeType ||
		t == udnssdk.HTTPProbeType ||
		t == udnssdk.PingProbeType ||
		t == udnssdk.SMTPProbeType ||
		t == udnssdk.SMTPSENDProbeType
}

func populateResourceDataFromProbe(p udnssdk.ProbeInfoDTO, d *schema.ResourceData) error {
	d.SetId(p.ID)
	d.Set("pool_record", p.PoolRecord)
	d.Set("interval", p.Interval)
	d.Set("agents", p.Agents)
	d.Set("threshold", p.Threshold)

	d.Set("type", p.ProbeType)

	// err := p.Details.Populate(p.ProbeType)
	// if err != nil {
	// 	return fmt.Errorf("Could not populate probe details: %v, ProbeInfo: %#v", err, p)
	// }
	if p.Details != nil {
		var dp map[string]interface{}
		err := json.Unmarshal(p.Details.GetData(), &dp)
		if err != nil {
			return err
		}

		err = d.Set(typeToAttrKeyMap[p.ProbeType], dp)
		if err != nil {
			return fmt.Errorf("Error setting details: %v", err)
		}
	}
	return nil
}
