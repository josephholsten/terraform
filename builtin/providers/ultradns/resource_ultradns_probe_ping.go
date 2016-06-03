package ultradns

import (
	"fmt"
	"log"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUltradnsProbePing() *schema.Resource {
	return &schema.Resource{
		Create: resourceUltradnsProbePingCreate,
		Read:   resourceUltradnsProbePingRead,
		Update: resourceUltradnsProbePingUpdate,
		Delete: resourceUltradnsProbePingDelete,

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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"packets": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
						"packet_size": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
						"limits": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
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
							},
						},
					},
				},
			},
			// Computed
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUltradnsProbePingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := makePingProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_ping configuration: %v", err)
	}

	log.Printf("[INFO] ultradns_probe_ping create: %#v, detail: %#v", r, r.Details.Detail)
	resp, err := client.Probes.Create(r.Key().RRSetKey(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("create failed: %v", err)
	}

	uri := resp.Header.Get("Location")
	d.Set("uri", uri)
	d.SetId(uri)
	log.Printf("[INFO] ultradns_probe_ping.id: %v", d.Id())

	return resourceUltradnsProbePingRead(d, meta)
}

func resourceUltradnsProbePingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := makePingProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_ping configuration: %v", err)
	}

	log.Printf("[DEBUG] ultradns_probe_ping read: %#v", r)
	probe, _, err := client.Probes.Find(r.Key())
	log.Printf("[DEBUG] ultradns_probe_ping response: %#v", probe)

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

	return populateResourceDataFromPingProbe(probe, d)
}

func resourceUltradnsProbePingUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := makePingProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_ping configuration: %v", err)
	}

	log.Printf("[INFO] ultradns_probe_ping update: %+v", r)
	_, err = client.Probes.Update(r.Key(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("update failed: %s", err)
	}

	return resourceUltradnsProbePingRead(d, meta)
}

func resourceUltradnsProbePingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := makePingProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_ping configuration: %s", err)
	}

	log.Printf("[INFO] ultradns_probe_ping delete: %+v", r)
	_, err = client.Probes.Delete(r.Key())
	if err != nil {
		return fmt.Errorf("delete failed: %s", err)
	}

	return nil
}

// Resource Helpers

type pingProbeResource struct {
	Name string
	Zone string
	ID   string

	Agents     []string
	Interval   string
	PoolRecord string
	Threshold  int
	Type       udnssdk.ProbeType

	Details *udnssdk.ProbeDetailsDTO
}

func makePingProbeResource(d *schema.ResourceData) (pingProbeResource, error) {
	p := pingProbeResource{}
	p.Zone = d.Get("zone").(string)
	p.Name = d.Get("name").(string)
	p.ID = d.Id()
	p.Interval = d.Get("interval").(string)
	p.PoolRecord = d.Get("pool_record").(string)
	p.Threshold = d.Get("threshold").(int)
	p.Type = "PING"

	as, ok := d.GetOk("agents")
	if !ok {
		return p, fmt.Errorf("ultradns_probe_ping.agents not ok: %+v", d.Get("agents"))
	}
	for _, a := range as.([]interface{}) {
		p.Agents = append(p.Agents, a.(string))
	}

	pp, ok := d.GetOk("ping_probe")
	if !ok {
		return p, fmt.Errorf("ultradns_probe_ping.ping_probe not ok: %+v", d.Get("ping_probe"))
	}

	pps := pp.([]interface{})
	if len(pps) >= 1 {
		if len(pps) > 1 {
			return p, fmt.Errorf("ping_probe: only 0 or 1 blocks alowed, got: %#v", len(pps))
		}
		p.Details = makeProbeDetails(pps[0])
	}

	return p, nil
}

func (p pingProbeResource) RRSetKey() udnssdk.RRSetKey {
	return p.Key().RRSetKey()
}

func (p pingProbeResource) ProbeInfoDTO() udnssdk.ProbeInfoDTO {
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

func (p pingProbeResource) Key() udnssdk.ProbeKey {
	return udnssdk.ProbeKey{
		Zone: p.Zone,
		Name: p.Name,
		ID:   p.ID,
	}
}

func makeProbeDetails(configured interface{}) *udnssdk.ProbeDetailsDTO {
	data := configured.(map[string]interface{})
	// Convert limits from flattened set format to mapping.
	ls := make(map[string]udnssdk.ProbeDetailsLimitDTO)
	for _, limit := range data["limits"].([]interface{}) {
		l := limit.(map[string]interface{})
		name := l["name"].(string)
		ls[name] = udnssdk.ProbeDetailsLimitDTO{
			Warning:  l["warning"].(int),
			Critical: l["critical"].(int),
			Fail:     l["fail"].(int),
		}
	}
	res := udnssdk.ProbeDetailsDTO{
		Detail: udnssdk.PingProbeDetailsDTO{
			Limits:     ls,
			PacketSize: data["packet_size"].(int),
			Packets:    data["packets"].(int),
		},
	}
	return &res
}

func populateResourceDataFromPingProbe(p udnssdk.ProbeInfoDTO, d *schema.ResourceData) error {
	d.SetId(p.ID)
	d.Set("pool_record", p.PoolRecord)
	d.Set("interval", p.Interval)
	d.Set("agents", p.Agents)
	d.Set("threshold", p.Threshold)

	var pp []map[string]interface{}
	dp, err := mapFromDetails(p)
	if err != nil {
		return fmt.Errorf("ProbeInfo.details could not be unmarshalled: %v, Details: %#v", err, p.Details)
	}
	pp = append(pp, dp)

	err = d.Set("ping_probe", pp)
	if err != nil {
		return fmt.Errorf("ping_probe set failed: %v, from %#v", err, pp)
	}
	return nil
}

func mapFromDetails(raw udnssdk.ProbeInfoDTO) (map[string]interface{}, error) {
	d, err := raw.Details.PingProbeDetails()
	if err != nil {
		return nil, err
	}
	ls := make([]map[string]interface{}, 0, len(d.Limits))
	for name, lim := range d.Limits {
		l := map[string]interface{}{
			"name":     name,
			"warning":  lim.Warning,
			"critical": lim.Critical,
			"fail":     lim.Fail,
		}
		ls = append(ls, l)
	}
	m := map[string]interface{}{
		"limits":      ls,
		"packets":     d.Packets,
		"packet_size": d.PacketSize,
	}
	return m, nil
}
