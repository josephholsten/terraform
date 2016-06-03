package ultradns

import (
	"fmt"
	"log"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUltradnsProbeHTTP() *schema.Resource {
	return &schema.Resource{
		Create: resourceUltradnsProbeHTTPCreate,
		Read:   resourceUltradnsProbeHTTPRead,
		Update: resourceUltradnsProbeHTTPUpdate,
		Delete: resourceUltradnsProbeHTTPDelete,

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
			"http_probe": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transaction": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"method": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
									},
									"url": &schema.Schema{
										Type:     schema.TypeString,
										Required: true,
									},
									"transmitted_data": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
									},
									"follow_redirects": &schema.Schema{
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"limit": &schema.Schema{
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
						"total_limits": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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

func resourceUltradnsProbeHTTPCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := makeHTTPProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_http configuration: %v", err)
	}

	log.Printf("[INFO] ultradns_probe_http create: %#v, detail: %#v", r, r.Details.Detail)
	resp, err := client.Probes.Create(r.Key().RRSetKey(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("create failed: %v", err)
	}

	uri := resp.Header.Get("Location")
	d.Set("uri", uri)
	d.SetId(uri)
	log.Printf("[INFO] ultradns_probe_http.id: %v", d.Id())

	return resourceUltradnsProbeHTTPRead(d, meta)
}

func resourceUltradnsProbeHTTPRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := makeHTTPProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_http configuration: %v", err)
	}

	log.Printf("[DEBUG] ultradns_probe_http read: %#v", r)
	probe, _, err := client.Probes.Find(r.Key())
	log.Printf("[DEBUG] ultradns_probe_http response: %#v", probe)

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

	return populateResourceDataFromHTTPProbe(probe, d)
}

func resourceUltradnsProbeHTTPUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := makeHTTPProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_http configuration: %v", err)
	}

	log.Printf("[INFO] ultradns_probe_http update: %+v", r)
	_, err = client.Probes.Update(r.Key(), r.ProbeInfoDTO())
	if err != nil {
		return fmt.Errorf("update failed: %s", err)
	}

	return resourceUltradnsProbeHTTPRead(d, meta)
}

func resourceUltradnsProbeHTTPDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	r, err := makeHTTPProbeResource(d)
	if err != nil {
		return fmt.Errorf("Could not load ultradns_probe_http configuration: %s", err)
	}

	log.Printf("[INFO] ultradns_probe_http delete: %+v", r)
	_, err = client.Probes.Delete(r.Key())
	if err != nil {
		return fmt.Errorf("delete failed: %s", err)
	}

	return nil
}

// Resource Helpers

type httpProbeResource struct {
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

func makeHTTPProbeResource(d *schema.ResourceData) (httpProbeResource, error) {
	p := httpProbeResource{}
	p.Zone = d.Get("zone").(string)
	p.Name = d.Get("name").(string)
	p.ID = d.Id()
	p.Interval = d.Get("interval").(string)
	p.PoolRecord = d.Get("pool_record").(string)
	p.Threshold = d.Get("threshold").(int)
	p.Type = udnssdk.HTTPProbeType

	as, ok := d.GetOk("agents")
	if !ok {
		return p, fmt.Errorf("ultradns_probe_http.agents not ok: %+v", d.Get("agents"))
	}
	for _, a := range as.([]interface{}) {
		p.Agents = append(p.Agents, a.(string))
	}

	pp, ok := d.GetOk("http_probe")
	if !ok {
		return p, fmt.Errorf("ultradns_probe_http.http_probe not ok: %+v", d.Get("http_probe"))
	}

	pps := pp.([]interface{})
	if len(pps) >= 1 {
		if len(pps) > 1 {
			return p, fmt.Errorf("http_probe: only 0 or 1 blocks alowed, got: %#v", len(pps))
		}
		p.Details = makeHTTPProbeDetails(pps[0])
	}

	return p, nil
}

func (p httpProbeResource) RRSetKey() udnssdk.RRSetKey {
	return p.Key().RRSetKey()
}

func (p httpProbeResource) ProbeInfoDTO() udnssdk.ProbeInfoDTO {
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

func (p httpProbeResource) Key() udnssdk.ProbeKey {
	return udnssdk.ProbeKey{
		Zone: p.Zone,
		Name: p.Name,
		ID:   p.ID,
	}
}

func makeHTTPProbeDetails(configured interface{}) *udnssdk.ProbeDetailsDTO {
	data := configured.(map[string]interface{})
	// Convert limits from flattened set format to mapping.

	rawTs := data["transaction"].([]interface{})
	ts := []udnssdk.Transaction{}
	for _, rt := range rawTs {
		t := rt.(map[string]interface{})
		ls := make(map[string]udnssdk.ProbeDetailsLimitDTO)
		for _, limit := range t["limit"].([]interface{}) {
			l := limit.(map[string]interface{})
			name := l["name"].(string)
			ls[name] = makeProbeDetailsLimit(l)
		}
		trans := udnssdk.Transaction{
			Method:          t["method"].(string),
			URL:             t["url"].(string),
			TransmittedData: t["transmitted_data"].(string),
			FollowRedirects: t["follow_redirects"].(bool),
			Limits: ls,
		}
		ts = append(ts, trans)
	}
	d := udnssdk.HTTPProbeDetailsDTO{
		Transactions: ts,
	}
	rawLims := data["total_limits"].([]interface{})
	// TODO: validate 0 or 1 total_limits
	if len(rawLims) > 0 {
		lim := rawLims[0]
		l := makeProbeDetailsLimit(lim)
		d.TotalLimits = &l
	}
	res := udnssdk.ProbeDetailsDTO{
		Detail: d,
	}
	return &res
}

func makeProbeDetailsLimit(configured interface{}) udnssdk.ProbeDetailsLimitDTO {
	l := configured.(map[string]interface{})
	return udnssdk.ProbeDetailsLimitDTO{
		Warning:  l["warning"].(int),
		Critical: l["critical"].(int),
		Fail:     l["fail"].(int),
	}
}

func populateResourceDataFromHTTPProbe(p udnssdk.ProbeInfoDTO, d *schema.ResourceData) error {
	d.SetId(p.ID)
	d.Set("pool_record", p.PoolRecord)
	d.Set("interval", p.Interval)
	d.Set("agents", p.Agents)
	d.Set("threshold", p.Threshold)

	var pp []map[string]interface{}
	dp, err := mapFromHTTPDetails(p.Details)
	if err != nil {
		return fmt.Errorf("ProbeInfo.details could not be unmarshalled: %v, Details: %#v", err, p.Details)
	}
	pp = append(pp, dp)

	err = d.Set("http_probe", pp)
	if err != nil {
		return fmt.Errorf("http_probe set failed: %v, from %#v", err, pp)
	}
	return nil
}

func mapFromHTTPDetails(raw *udnssdk.ProbeDetailsDTO) (map[string]interface{}, error) {
	d, err := raw.HTTPProbeDetails()
	if err != nil {
		return nil, err
	}
	ts := make([]map[string]interface{}, 0, len(d.Transactions))
	for _, rt := range d.Transactions {
		t := map[string]interface{}{
			"method":           rt.Method,
			"url":              rt.URL,
			"transmitted_data": rt.TransmittedData,
			"follow_redirects": rt.FollowRedirects,
		}
		ts = append(ts, t)
	}
	m := map[string]interface{}{
		"transaction": ts,
	}
	return m, nil
}
