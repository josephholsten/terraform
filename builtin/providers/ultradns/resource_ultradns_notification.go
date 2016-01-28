package ultradns

import (
	"fmt"
	"log"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/schema"
)

func schemaNotification() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
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
				"email": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"poolRecords": &schema.Schema{
					Type:     schema.TypeList,
					Optional: false,
					Elem:     schemaPoolRecords(),
				},
			},
		},
	}
}
func schemaPoolRecords() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"poolRecord": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"notification": schemaNotificationInfo(),
			},
		},
	}
}
func schemaNotificationInfo() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"probe": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"record": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"scheduled": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
		},
	}
}
func resourceUltraDNSNotificationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	n := newNotificationResource(d)

	log.Printf("[INFO] ultradns_notification create: %#v", n)
	r, err := client.Notifications.Create(n.Key(), n.notificationDTO())
	if err != nil {
		return fmt.Errorf("ultradns_notification created failed: %s", err)
	}

	uri := r.Header.Get("Location")
	d.Set("uri", uri)
	d.SetId(uri)
	log.Printf("[INFO] Notification ID: %s", d.Id())

	return resourceUltraDNSNotificationRead(d, meta)
}

func resourceUltraDNSNotificationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)
	n := newNotificationResource(d)
	notification, _, err := client.Notifications.Find(n.Key())

	if err != nil {
		uderr, ok := err.(*udnssdk.ErrorResponseList)
		if ok {
			for _, r := range uderr.Responses {
				// 70002 means Notifications Not Found
				if r.ErrorCode == 70002 {
					d.SetId("")
					return nil
				}
				return fmt.Errorf("ultradns_notification not found: %s", err)
			}
		}
		return fmt.Errorf("ultradns_notification not found: %s", err)
	}
	var prs []map[string]interface{}
	for _, e := range notification.PoolRecords {
		prs = append(prs, map[string]interface{}{
			"poolRecord":   e.PoolRecord,
			"notification": e.Notification,
		})
	}
	// FIXME: shouldn't this call populate?
	return nil
}

func resourceUltraDNSNotificationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	n := newNotificationResource(d)

	log.Printf("[INFO] UltraDNS Notification update configuration: %#v", n)
	_, err := client.Notifications.Update(n.Key(), n.notificationDTO())
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to update UltraDNS Notification: %s", err)
	}

	return resourceUltraDNSNotificationRead(d, meta)
}

func resourceUltraDNSNotificationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*udnssdk.Client)

	n := newNotificationResource(d)

	log.Printf("[INFO] ultradns_notification delete: %#v", n)
	_, err := client.Notifications.Delete(n.Key())
	if err != nil {
		return fmt.Errorf("ultradns_notification delete failed: %s", err)
	}

	return nil
}

type notificationResource struct {
	Name        string
	Zone        string
	Email       string
	PoolRecords []udnssdk.NotificationPoolRecord
}

func newNotificationResource(d *schema.ResourceData) notificationResource {
	n := notificationResource{}
	n.Zone = d.Get("zone").(string)
	n.Name = d.Get("name").(string)
	n.Email = d.Get("email").(string)

	prs := d.Get("poolRecords").(*schema.Set).List()
	for _, e := range prs {
		rd := e.(*schema.ResourceData)
		pr := newNotificationPoolRecord(rd)
		n.PoolRecords = append(n.PoolRecords, pr)
	}
	return n
}

func (n notificationResource) Key() udnssdk.NotificationKey {
	return udnssdk.NotificationKey{
		Name:  n.Name,
		Zone:  n.Zone,
		Email: n.Email,
	}
}

func (n notificationResource) RRSetKey() udnssdk.RRSetKey {
	return n.Key().RRSetKey()
}

func (n notificationResource) notificationDTO() udnssdk.NotificationDTO {
	return udnssdk.NotificationDTO{
		Email:       n.Email,
		PoolRecords: n.PoolRecords,
	}
}

func newNotificationInfo(d *schema.ResourceData) udnssdk.NotificationInfoDTO {
	info := udnssdk.NotificationInfoDTO{}
	info.Probe = d.Get("probe").(bool)
	info.Record = d.Get("record").(bool)
	info.Scheduled = d.Get("scheduled").(bool)
	return info
}

func newNotificationPoolRecord(d *schema.ResourceData) udnssdk.NotificationPoolRecord {
	pr := udnssdk.NotificationPoolRecord{}
	pr.PoolRecord = d.Get("poolrecord").(string)
	n := d.Get("notification").(*schema.ResourceData)
	pr.Notification = newNotificationInfo(n)
	return pr
}
