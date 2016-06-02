package ultradns

import (
	"fmt"
	"testing"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccUltradnsProbePingBasic(t *testing.T) {
	var record udnssdk.RRSet
	// domain := os.Getenv("ULTRADNS_DOMAIN")
	domain := "ultradns.phinze.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUltradnsRecordAndPingProbeDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltradnsRecordAndPingProbeBasic, domain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_tcpool.probe-test", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "name", "probe-test"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "pool_record", "192.168.0.11"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "agents.0", "DALLAS"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "agents.1", "AMSTERDAM"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "interval", "ONE_MINUTE"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "threshold", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.packets", "15"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.packet_size", "56"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.limits.0.name", "lossPercent"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.limits.0.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.limits.0.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.limits.0.fail", "3"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.limits.1.name", "total"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.limits.1.warning", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.limits.1.critical", "3"),
					resource.TestCheckResourceAttr("ultradns_probe_ping.ping", "ping_probe.0.limits.1.fail", "4"),
				),
			},
		},
	})
}

/*
func TestAccUltraDNSRecord_Updated(t *testing.T) {
	var record udnssdk.RRSet
	domain := os.Getenv("ULTRADNS_DOMAIN")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUltraDNSRecordDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltraDNSRecordConfig_basic, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_record.foobar", &record),
					testAccCheckUltraDNSRecordAttributes(&record),
					resource.TestCheckResourceAttr(
						"ultradns_record.foobar", "name", "terraform"),
					resource.TestCheckResourceAttr(
						"ultradns_record.foobar", "zone", domain),
					resource.TestCheckResourceAttr(
						"ultradns_record.foobar", "rdata.0", "192.168.0.10"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltraDNSRecordConfig_new_value, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_record.foobar", &record),
					testAccCheckUltraDNSRecordAttributesUpdated(&record),
					resource.TestCheckResourceAttr(
						"ultradns_record.foobar", "name", "terraform"),
					resource.TestCheckResourceAttr(
						"ultradns_record.foobar", "zone", domain),
					resource.TestCheckResourceAttr(
						"ultradns_record.foobar", "rdata.0", "192.168.0.11"),
					resource.TestCheckResourceAttr(
						"ultradns_record.foobar", "string_profile", testProfile),
				),
			},
		},
	})
}
*/
func testAccCheckUltradnsRecordAndPingProbeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*udnssdk.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ultradns_record" {
			continue
		}

		k := udnssdk.RRSetKey{
			Zone: rs.Primary.Attributes["zone"],
			Name: rs.Primary.Attributes["name"],
			Type: rs.Primary.Attributes["type"],
		}

		_, err := client.RRSets.Select(k)
		if err == nil {
			return fmt.Errorf("Record still exists")
		}
	}

	return nil
}

/*
func testAccCheckUltraDNSRecordAttributes(record *udnssdk.RRSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if record.RData[0] != "192.168.0.11" {
			return fmt.Errorf("Bad content: %v", record.RData)
		}

		return nil
	}
}


func testAccCheckUltraDNSRecordAttributesUpdated(record *udnssdk.RRSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if record.RData[0] != "192.168.0.11" {
			return fmt.Errorf("Bad content: %v", record.RData)
		}

		return nil
	}
}

func testAccCheckUltraDNSRecordExists(n string, record *udnssdk.RRSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*udnssdk.Client)
		foundRecord, err := client.RRSets.ListAllRRSets(rs.Primary.Attributes["zone"], rs.Primary.Attributes["name"], rs.Primary.Attributes["type"])

		if err != nil {
			return err
		}

		if foundRecord[0].OwnerName != rs.Primary.Attributes["hostname"] {
			return fmt.Errorf("Record not found: %+v,\n %+v\n", foundRecord, rs.Primary.Attributes)
		}

		*record = foundRecord[0]

		return nil
	}
}
*/
const testAccCheckUltradnsRecordAndPingProbeBasic = `
resource "ultradns_tcpool" "probe-test" {
  zone  = "%s"
  name  = "probe-test"

  ttl   = 30
  description = "traffic controller pool with probes"

  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2

  rdata {
    host           = "192.168.0.11"

    state          = "NORMAL"
    run_probes     = true
    priority       = 1
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  rdata {
    host           = "192.168.0.12"

    state          = "NORMAL"
    run_probes     = true
    priority       = 2
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  backup_record_rdata = "192.168.0.1"
}

resource "ultradns_probe_ping" "ping" {
  zone  = "%s"
  name  = "probe-test"

  pool_record = "192.168.0.11"

  agents = ["DALLAS", "AMSTERDAM"]

  interval  = "ONE_MINUTE"
  threshold = 1

  ping_probe {
    packets    = 15
    packet_size = 56

    limits {
      name     = "lossPercent"
      warning  = 1
      critical = 2
      fail     = 3
    }

    limits {
      name     = "total"
      warning  = 2
      critical = 3
      fail     = 4
    }
  }

  depends_on = ["ultradns_tcpool.probe-test"]
}
`
