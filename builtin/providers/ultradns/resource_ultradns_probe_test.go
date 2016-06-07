package ultradns

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUltradnsProbe(t *testing.T) {
	// domain := os.Getenv("ULTRADNS_DOMAIN")
	domain := "ultradns.phinze.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccTcpoolCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testCfgProbeMinimal, domain, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ultradns_probe.it", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_probe.it", "name", "test-probe"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "pool_record", "10.4.0.1"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "type", "PING"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "agents.0", "DALLAS"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "agents.1", "AMSTERDAM"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "interval", "ONE_MINUTE"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "threshold", "1"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.packets", "15"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.packet_size", "56"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.#", "2"),

					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.354186460.name", "lossPercent"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.354186460.warning", "1"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.354186460.critical", "2"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.354186460.fail", "3"),

					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.466411754.name", "total"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.466411754.warning", "2"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.466411754.critical", "3"),
					resource.TestCheckResourceAttr("ultradns_probe.it", "ping_probe.0.limit.466411754.fail", "4"),
				),
			},
		},
	})
}

const testCfgProbeMinimal = `
resource "ultradns_tcpool" "test-probe" {
  zone  = "%s"
  name  = "test-probe"

  ttl   = 30
  description = "traffic controller pool with probes"

  run_probes    = true
  act_on_probes = true
  max_to_lb     = 2

  rdata {
    host = "10.4.0.1"

    state          = "NORMAL"
    run_probes     = true
    priority       = 1
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  rdata {
    host = "10.4.0.2"


    state          = "NORMAL"
    run_probes     = true
    priority       = 2
    failover_delay = 0
    threshold      = 1
    weight         = 2
  }

  backup_record_rdata = "10.4.0.3"
}

resource "ultradns_probe" "it" {
  zone  = "%s"
  name  = "test-probe"

  pool_record = "10.4.0.1"

  type   = "PING"
  agents = ["DALLAS", "AMSTERDAM"]

  interval  = "ONE_MINUTE"
  threshold = 1

  ping_probe {
    packets    = 15
    packet_size = 56

    limit {
      name     = "lossPercent"
      warning  = 1
      critical = 2
      fail     = 3
    }

    limit {
      name     = "total"
      warning  = 2
      critical = 3
      fail     = 4
    }
  }

  depends_on = ["ultradns_tcpool.test-probe"]
}
`
