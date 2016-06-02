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
				Type:     schema.TypeSet,
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

func resourceLimits() *schema.Resource {
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

func schemaPingProbe() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"packets": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"packetSize": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"limits": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     resourceLimits(),
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
	panic(fmt.Sprintf("Probe: %#v ProbeInfo: %#v", r, probe))

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

type probeResource struct {
	Name string
	Zone string
	ID   string

	Agents     []string
	Interval   string
	PoolRecord string
	Threshold  int
	Type       string

	Details *udnssdk.ProbeDetailsDTO
}

func newProbeResource(d *schema.ResourceData) (probeResource, error) {
	p := probeResource{}
	// zoneName
	p.Zone = d.Get("zone").(string)
	// ownerName
	p.Name = d.Get("name").(string)
	// id
	p.ID = d.Id()

	p.Interval = d.Get("interval").(string)
	p.PoolRecord = d.Get("pool_record").(string)
	p.Threshold = d.Get("threshold").(int)
	p.Type = d.Get("type").(string)

	// agents
	as, ok := d.GetOk("agents")
	if !ok {
		return p, fmt.Errorf("ultradns_probe.agents not ok: %+v", d.Get("agents"))
	}
	for _, e := range as.([]interface{}) {
		p.Agents = append(p.Agents, e.(string))
	}

	// details
	// TODO: validate p.Type is in typeToAttrKeyMap.Keys
	drd, ok := d.GetOk(p.detailsAttribute())
	if !ok {
		return p, fmt.Errorf("ultradns_probe.%s not ok: %+v", p.detailsAttribute(), d.Get(p.detailsAttribute()))
	}
	probeset := drd.(*schema.Set)
	var probedetails map[string]interface{}
	probedetails = probeset.List()[0].(map[string]interface{})
	// Convert limits from flattened set format to mapping.
	ls := map[string]interface{}{}
	for _, limit := range probedetails["limits"].([]interface{}) {
		l := limit.(map[string]interface{})
		name := l["name"].(string)
		ls[name] = map[string]interface{}{
			"warning":  l["warning"],
			"critical": l["critical"],
			"fail":     l["fail"],
		}
	}
	probedetails["limits"] = ls
	p.Details = &udnssdk.ProbeDetailsDTO{
		Detail: probedetails,
	}

	return p, nil
}

func (p probeResource) detailsAttribute() string {
	return typeToAttrKeyMap[p.Type]
}

func (p probeResource) RRSetKey() udnssdk.RRSetKey {
	return p.Key().RRSetKey()
}

func (p probeResource) ProbeInfoDTO() udnssdk.ProbeInfoDTO {
	return udnssdk.ProbeInfoDTO{
		ID:         p.ID,
		PoolRecord: p.PoolRecord,
		ProbeType:  p.Type,
		Interval:   p.Interval,
		Agents:     p.Agents,
		Threshold:  p.Threshold,
		Details:    p.Details,
	}
}

var typeToAttrKeyMap = map[string]string{
	"HTTP":      "http_probe",
	"PING":      "ping_probe",
	"FTP":       "ftp_probe",
	"SMTP":      "smtp_probe",
	"SMTP_SEND": "smtpsend_probe",
	"DNS":       "dns_probe",
}

func (p probeResource) Key() udnssdk.ProbeKey {
	return udnssdk.ProbeKey{
		Zone: p.Zone,
		Name: p.Name,
		ID:   p.ID,
	}
}

func populateResourceDataFromProbe(p udnssdk.ProbeInfoDTO, d *schema.ResourceData) error {
	// id
	d.SetId(p.ID)
	// poolRecord
	err := d.Set("pool_record", p.PoolRecord)
	if err != nil {
		return fmt.Errorf("Error setting pool_record: %v", err)
	}
	// interval
	err = d.Set("interval", p.Interval)
	if err != nil {
		return fmt.Errorf("Error setting interval: %v", err)
	}
	// type
	err = d.Set("type", p.ProbeType)
	if err != nil {
		return fmt.Errorf("Error setting type: %v", err)
	}
	// agents
	err = d.Set("agents", p.Agents)
	if err != nil {
		return fmt.Errorf("Error setting agents: %v", err)
	}
	// threshold
	err = d.Set("threshold", p.Threshold)
	if err != nil {
		return fmt.Errorf("Error setting threshold: %v", err)
	}
	// details
	err = p.Details.Populate(p.ProbeType)
	if err != nil {
		return fmt.Errorf("Could not populate probe details: %v, ProbeInfo: %#v", err, p)
	}
	if p.Details != nil {
		var dp map[string]interface{}
		err = json.Unmarshal(p.Details.GetData(), &dp)
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
