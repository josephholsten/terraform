package ultradns

import (
	"fmt"
	"testing"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccUltradnsProbeHTTPBasic(t *testing.T) {
	var record udnssdk.RRSet
	domain := "ultradns.phinze.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUltradnsRecordAndHTTPProbeDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltradnsRecordAndHTTPProbeBasic, domain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_tcpool.probe-http-test", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "name", "probe-http-test"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "pool_record", "192.168.0.11"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "agents.0", "DALLAS"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "agents.1", "AMSTERDAM"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "interval", "ONE_MINUTE"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "threshold", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.method", "GET"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.url", "http://localhost/index"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltradnsRecordAndHTTPProbeMaximal, domain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_tcpool.probe-http-test", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "name", "probe-http-test"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "pool_record", "192.168.0.11"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "agents.0", "DALLAS"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "agents.1", "AMSTERDAM"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "interval", "ONE_MINUTE"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "threshold", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.method", "POST"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.url", "http://localhost/index"),

					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.0.name", "connect"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.0.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.0.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.0.fail", "3"),

					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.1.name", "avgConnect"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.1.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.1.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.1.fail", "3"),

					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.2.name", "run"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.2.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.2.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.2.fail", "3"),

					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.3.name", "avgRun"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.3.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.3.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.transaction.0.limit.3.fail", "3"),

					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.total_limits.0.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.total_limits.0.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe_http.http", "http_probe.0.total_limits.0.fail", "3"),

				),
			},
		},
	})
}

func testAccCheckUltradnsRecordAndHTTPProbeDestroy(s *terraform.State) error {
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

const testAccCheckUltradnsRecordAndHTTPProbeBasic = `
resource "ultradns_tcpool" "probe-http-test" {
  zone  = "%s"
  name  = "probe-http-test"

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

resource "ultradns_probe_http" "http" {
  zone = "%s"
  name = "probe-http-test"

  pool_record = "192.168.0.11"

  agents = ["DALLAS", "AMSTERDAM"]

  interval  = "ONE_MINUTE"
  threshold = 1

  http_probe {
    transaction {
      method           = "GET"
      url              = "http://localhost/index"
    }
  }

  depends_on = ["ultradns_tcpool.probe-http-test"]
}
`

const testAccCheckUltradnsRecordAndHTTPProbeMaximal = `
resource "ultradns_tcpool" "probe-http-test" {
  zone  = "%s"
  name  = "probe-http-test"

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

resource "ultradns_probe_http" "http" {
  zone = "%s"
  name = "probe-http-test"

  pool_record = "192.168.0.11"

  agents = ["DALLAS", "AMSTERDAM"]

  interval  = "ONE_MINUTE"
  threshold = 1

  http_probe {
    transaction {
      method           = "POST"
      url              = "http://localhost/index"
      transmitted_data = "{}"
      follow_redirects = true

      limit {
        name = "connect"

        warning  = 1
        critical = 2
        fail     = 3
      }
      limit {
        name = "avgConnect"

        warning  = 1
        critical = 2
        fail     = 3
      }
      limit {
        name = "run"

        warning  = 1
        critical = 2
        fail     = 3
      }
      limit {
        name = "avgRun"

        warning  = 1
        critical = 2
        fail     = 3
      }
    }

    total_limits {
      warning  = 1
      critical = 2
      fail     = 3
    }
  }

  depends_on = ["ultradns_tcpool.probe-http-test"]
}
`
