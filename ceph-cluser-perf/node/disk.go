// disk
package node

import (
	"bufio"
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"os"
	"strings"
	"time"
)

var old_diskstat_array []linuxproc.DiskStat
var new_diskstat_array []linuxproc.DiskStat
var disk_last_time int64

const UINT_MAX = 4294967295

func Get_submit_float_stat_str(hostname, plugin, plugin_ins, str_type, str_type_ins string, value float64, time_value int64) string {
	stat := fmt.Sprintf("PUTVAL %s/%s-%s/%s-%s %d:%f\n", hostname, plugin, plugin_ins, str_type, str_type_ins, time_value,
		value)
	return stat
}

func cal_disk_io_stat(old_stats, new_stats linuxproc.DiskStat, curr_time int64) {
	var stats linuxproc.DiskStat
	var read_wait float64
	var write_wait float64
	var str_stat string

	time_delta := curr_time - disk_last_time

	if new_stats.ReadIOs < old_stats.ReadIOs {
		stats.ReadIOs = 1 + new_stats.ReadIOs + (UINT_MAX - old_stats.ReadIOs)
	} else {
		stats.ReadIOs = new_stats.ReadIOs - old_stats.ReadIOs
	}

	if new_stats.WriteIOs < old_stats.WriteIOs {
		stats.WriteIOs = 1 + new_stats.WriteIOs + (UINT_MAX - old_stats.WriteIOs)
	} else {
		stats.WriteIOs = new_stats.WriteIOs - old_stats.WriteIOs
	}

	if new_stats.ReadMerges < old_stats.ReadMerges {
		stats.ReadMerges = 1 + new_stats.ReadMerges + (UINT_MAX - old_stats.ReadMerges)
	} else {
		stats.ReadMerges = new_stats.ReadMerges - old_stats.ReadMerges
	}

	if new_stats.WriteMerges < old_stats.WriteMerges {
		stats.WriteMerges = 1 + new_stats.WriteMerges + (UINT_MAX - old_stats.WriteMerges)
	} else {
		stats.WriteMerges = new_stats.WriteMerges - old_stats.WriteMerges
	}

	if new_stats.ReadSectors < old_stats.ReadSectors {
		stats.ReadSectors = 1 + new_stats.ReadSectors + (UINT_MAX - old_stats.ReadSectors)
	} else {
		stats.ReadSectors = new_stats.ReadSectors - old_stats.ReadSectors
	}

	if new_stats.WriteSectors < old_stats.WriteSectors {
		stats.WriteSectors = 1 + new_stats.WriteSectors + (UINT_MAX - old_stats.WriteSectors)
	} else {
		stats.WriteSectors = new_stats.WriteSectors - old_stats.WriteSectors
	}

	if new_stats.WriteTicks < old_stats.WriteTicks {
		stats.WriteTicks = 1 + new_stats.WriteTicks + (UINT_MAX - old_stats.WriteTicks)
	} else {
		stats.WriteTicks = new_stats.WriteTicks - old_stats.WriteTicks
	}

	if new_stats.ReadTicks < old_stats.ReadTicks {
		stats.ReadTicks = 1 + new_stats.ReadTicks + (UINT_MAX - old_stats.ReadTicks)
	} else {
		stats.ReadTicks = new_stats.ReadTicks - old_stats.ReadTicks
	}

	rd_iops := float64(stats.ReadIOs) / float64(time_delta)
	wd_iops := float64(stats.WriteIOs) / float64(time_delta)
	read_speed := float64(stats.ReadSectors*512) / float64(time_delta)
	write_speed := float64(stats.WriteSectors*512) / float64(time_delta)
	rd_merge_iops := float64(stats.ReadMerges) / float64(time_delta)
	wd_merge_iops := float64(stats.WriteMerges) / float64(time_delta)
	if stats.ReadIOs != 0 {
		read_wait = float64(stats.ReadTicks) / float64(stats.ReadIOs)
	}

	if stats.WriteIOs != 0 {
		write_wait = float64(stats.WriteTicks) / float64(stats.WriteIOs)
	}

	str_stat = Get_submit_float_stat_str(hostname, "disk", new_stats.Name, "gauge", "rd_iops", rd_iops, curr_time)
	str_stat += Get_submit_float_stat_str(hostname, "disk", new_stats.Name, "gauge", "wd_iops", wd_iops, curr_time)
	str_stat += Get_submit_float_stat_str(hostname, "disk", new_stats.Name, "gauge", "rd_iops_merge", rd_merge_iops, curr_time)
	str_stat += Get_submit_float_stat_str(hostname, "disk", new_stats.Name, "gauge", "wd_iops_merge", wd_merge_iops, curr_time)
	str_stat += Get_submit_float_stat_str(hostname, "disk", new_stats.Name, "gauge", "rd_spd", read_speed, curr_time)
	str_stat += Get_submit_float_stat_str(hostname, "disk", new_stats.Name, "gauge", "wd_spd", write_speed, curr_time)
	str_stat += Get_submit_float_stat_str(hostname, "disk", new_stats.Name, "gauge", "read_wait", read_wait, curr_time)
	str_stat += Get_submit_float_stat_str(hostname, "disk", new_stats.Name, "gauge", "write_wait", write_wait, curr_time)

	f := bufio.NewWriter(os.Stdout)
	if f != nil {
		f.Write([]byte(str_stat))
		f.Flush()
	}
}

func get_disk_stats_array() ([]linuxproc.DiskStat, error) {
	var used int

	stat, err := linuxproc.ReadDiskStats("/proc/diskstats")
	if err != nil {
		return nil, err
	}

	result := make([]linuxproc.DiskStat, len(stat))

	used = 0
	for i := range stat {
		if strings.Contains(stat[i].Name, "ram") == false && strings.Contains(stat[i].Name, "loop") == false {
			result[used] = stat[i]
			used++
		}
	}
	return result[0:used], nil
}

func init_old_disk_stats() {
	var err error
	old_diskstat_array, err = get_disk_stats_array()
	if err != nil {
		return
	}

	disk_last_time = time.Now().Unix()
}

func Get_disk_stats() {
	var err error
	if disk_last_time == 0 {
		init_old_disk_stats()
		return
	}

	curr_time := time.Now().Unix()
	new_diskstat_array, err = get_disk_stats_array()
	if err != nil {
		return
	}
	for i := range old_diskstat_array {
		if i < len(new_diskstat_array) && old_diskstat_array[i].Name == new_diskstat_array[i].Name {
			cal_disk_io_stat(old_diskstat_array[i], new_diskstat_array[i], curr_time)
			continue
		}

		for j := range new_diskstat_array {
			if old_diskstat_array[i].Name == new_diskstat_array[j].Name {
				cal_disk_io_stat(old_diskstat_array[i], new_diskstat_array[j], curr_time)
				break
			}
		}
	}
	disk_last_time = curr_time
	old_diskstat_array = new_diskstat_array
}
