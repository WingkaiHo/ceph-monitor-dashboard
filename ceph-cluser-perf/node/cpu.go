// cpu.go
package node

import (
	"bufio"
	"fmt"
	"os"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
)

var hostname string

func submit_cpu(pluging_instance int, pluging_name string, unixTs int64, value uint64) string {
	s := fmt.Sprintf("PUTVAL %s/cpu-%d/absolute-%s %d:%d\n", hostname, pluging_instance,
		pluging_name, unixTs, value)
	return s
}

func submit_cpu_percent(pluging_name string, unixTs int64, value float64) string {
	//	s := fmt.Sprintf("PUTVAL %s/cpu-avg/gauge-%s %d:%2.2f\n", hostname, pluging_name, unixTs, value*100.0)
	s := fmt.Sprintf("PUTVAL %s/cpu-avg/gauge-%s %d:%f\n", hostname, pluging_name, unixTs, value)
	//	s := fmt.Sprintf("PUTVAL %s/cpu-avg/gauge-%s %d:%f\n", hostname, pluging_name, unixTs, value*100.0)

	return s
}

func GetCPUPercent() {
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		fmt.Errorf("stat read fail")
		return
	}

	unixTs := time.Now().Unix()
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()

	s := stat.CPUStatAll
	total := float64(s.User + s.Nice + s.System + s.Idle + s.IOWait + s.IRQ + s.SoftIRQ + s.Steal)

	b := submit_cpu_percent("user", unixTs, float64(s.User)/total)
	b += submit_cpu_percent("nice", unixTs, float64(s.Nice)/total)
	b += submit_cpu_percent("system", unixTs, float64(s.System)/total)
	b += submit_cpu_percent("idle", unixTs, float64(s.Idle)/total)
	b += submit_cpu_percent("iowait", unixTs, float64(s.IOWait)/total)
	b += submit_cpu_percent("irq", unixTs, float64(s.IRQ)/total)
	b += submit_cpu_percent("softirq", unixTs, float64(s.SoftIRQ)/total)
	b += submit_cpu_percent("steal", unixTs, float64(s.Steal)/total)

	f.Write([]byte(b))
}

func GetCPU() {
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		fmt.Errorf("stat read fail")
		return
	}

	unixTs := time.Now().Unix()
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()
	for i := 0; i < len(stat.CPUStats); i++ {
		s := stat.CPUStats[i]
		b := submit_cpu(i, "user", unixTs, s.User)
		b += submit_cpu(i, "nice", unixTs, s.Nice)
		b += submit_cpu(i, "system", unixTs, s.System)
		b += submit_cpu(i, "idle", unixTs, s.Idle)
		b += submit_cpu(i, "iowait", unixTs, s.IOWait)
		b += submit_cpu(i, "irq", unixTs, s.IRQ)
		b += submit_cpu(i, "softirq", unixTs, s.SoftIRQ)
		b += submit_cpu(i, "steal", unixTs, s.Steal)
		f.Write([]byte(b))
	}
}
