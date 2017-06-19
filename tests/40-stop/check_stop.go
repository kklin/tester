package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/minion/supervisor/images"

	log "github.com/Sirupsen/logrus"
)

func main() {
	if err := exec.Command("quilt", "stop", "-containers").Run(); err != nil {
		log.WithError(err).Fatal("FAILED, couldn't run stop command")
	}

	log.Info("Sleeping thirty seconds for `quilt stop -containers` to take effect")
	time.Sleep(30 * time.Second)

	c, err := client.New(api.DefaultSocket)
	if err != nil {
		log.WithError(err).Fatal("FAILED, couldn't get quiltctl client")
	}
	defer c.Close()

	machines, err := c.QueryMachines()
	if err != nil {
		log.WithError(err).Fatal("FAILED, couldn't query the machines")
	}

	extraContainers := false
	for _, m := range machines {
		containersRaw, err := exec.Command("quilt", "ssh", m.StitchID,
			"docker", "ps", "--format", "{{.Names}}").Output()
		if err != nil {
			log.WithError(err).Fatal("FAILED, couldn't run docker ps")
		}

		fmt.Printf("Containers on machine %s:\n", m.StitchID)
		fmt.Println(string(containersRaw))
		names := strings.Split(strings.TrimSpace(string(containersRaw)), "\n")
		extraContainers = extraContainers || len(filterQuiltContainers(names)) > 0
	}

	if extraContainers {
		log.Fatal("FAILED, unexpected containers")
	}

	fmt.Println("PASSED")
}

var quiltContainers = map[string]struct{}{
	images.Etcd:          {},
	images.Ovncontroller: {},
	images.Ovnnorthd:     {},
	images.Ovsdb:         {},
	images.Ovsvswitchd:   {},
	images.Registry:      {},
	"minion":             {},
}

func filterQuiltContainers(containers []string) (filtered []string) {
	for _, c := range containers {
		if _, ok := quiltContainers[c]; !ok && c != "" {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
