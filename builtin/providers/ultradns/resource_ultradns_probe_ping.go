package ultradns

import (
	"fmt"
	"io/ioutil"
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

	r, err := newPingProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_ping configuration: %v", err)
	}

	log.Printf("[INFO] ultradns_probe_ping create: %#v, detail: %#v", r, r.Details.Detail)
	resp, err := client.Probes.Create(r.Key().RRSetKey(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("create failed: %v", err)
	}
	httpBody, err := ioutil.ReadAll(resp.Body)
	log.Printf("[INFO] ultradns_probe_ping create resp: %+v body: %v", resp, string(httpBody))

	uri := resp.Header.Get("Location")
	d.Set("uri", uri)
	d.SetId(uri)
	log.Printf("[INFO] ultradns_probe_ping.id: %v", d.Id())

	return resourceUltradnsProbePingRead(d, meta)
}

func resourceUltradnsProbePingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := newPingProbeResource(d)
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

	r, err := newPingProbeResource(d)
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

	r, err := newPingProbeResource(d)
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

type pingProbeResource struct {
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

func newPingProbeResource(d *schema.ResourceData) (pingProbeResource, error) {
	p := pingProbeResource{}
	// zoneName
	p.Zone = d.Get("zone").(string)
	// ownerName
	p.Name = d.Get("name").(string)
	// id
	p.ID = d.Id()

	p.Interval = d.Get("interval").(string)
	p.PoolRecord = d.Get("pool_record").(string)
	p.Threshold = d.Get("threshold").(int)
	p.Type = "PING"

	// agents
	as, ok := d.GetOk("agents")
	if !ok {
		return p, fmt.Errorf("ultradns_probe_ping.agents not ok: %+v", d.Get("agents"))
	}
	for _, e := range as.([]interface{}) {
		p.Agents = append(p.Agents, e.(string))
	}

	// details
	// TODO: validate p.Type is in typeToAttrKeyMap.Keys
	drd, ok := d.GetOk("ping_probe")
	if !ok {
		return p, fmt.Errorf("ultradns_probe_ping.ping_probe not ok: %+v", d.Get("ping_probe"))
	}
	p.Details = makeProbeDetails(drd)

	return p, nil
}

func makeProbeDetails(drd interface{}) *udnssdk.ProbeDetailsDTO {
	probelist := drd.([]interface{})
	probedetails := probelist[0].(map[string]interface{})
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
	// d := udnssdk.PingProbeDetails{
	// 	Limits:     ls
	// 	PacketSize: probedetails["packet_size"]
	// 	Packets:    probedetails["packets"]
	// }
	d := map[string]interface{}{
		"limits":     ls,
		"packetSize": probedetails["packet_size"],
		"packets":    probedetails["packets"],
	}
	return &udnssdk.ProbeDetailsDTO{
		Detail: d,
	}
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

func populateResourceDataFromPingProbe(p udnssdk.ProbeInfoDTO, d *schema.ResourceData) error {
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
		// var dp map[string]interface{}
		// err = json.Unmarshal(p.Details.GetData(), &dp)
		dp, err := mapFromDetails(p.Details)
		if err != nil {
			return err
		}
		var dps []map[string]interface{}
		dps = append(dps, dp)

		err = d.Set("ping_probe", dps)
		if err != nil {
			return fmt.Errorf("Error setting details: %v", err)
		}
	}
	return nil
}

func mapFromDetails(raw *udnssdk.ProbeDetailsDTO) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	// err := json.Unmarshal(d.GetData(), &m)
	d, err := raw.PingProbeDetails()
	m["packets"] = d.Packets
	m["packet_size"] = d.PacketSize
	var ls []map[string]interface{}
	for name, lim := range d.Limits {
		l := make(map[string]interface{})
		l["name"] = name
		l["warning"] = lim.Warning
		l["critical"] = lim.Critical
		l["fail"] = lim.Fail
		ls = append(ls, l)
	}
	m["limits"] = ls
	return m, err
}
