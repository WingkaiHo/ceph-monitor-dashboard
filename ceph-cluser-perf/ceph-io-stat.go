// ceph-io-stat
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"ceph-cluser-perf/node"
)

type ceph_node_item struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Exists       int    `json:exists`
	Type         string `json:"type"`
	Type_id      int    `json:"type_id"`
	Children     []int  `json:"children"`
	Status       string `json:"status"`
	Reweight     string `json:"reweight"`
	Crush_weight string `json:"crush_weight"`
	Depth        int    `json:"depth"`
}

type ceph_cluste_tree_struct struct {
	Nodes []ceph_node_item `json:"nodes"`
	Stray []string         `json:"stray"`
}

type perf_stats_struct struct {
	Commit_latency_ms int64 `json:"commit_latency_ms"`
	Apply_latency_ms  int64 `json:"apply_latency_ms"`
}

type osd_perf_info struct {
	Id         int               `json:"id"`
	Perf_stats perf_stats_struct `json:"perf_stats"`
}

type cluster_osds_perf_infos struct {
	Osd_perf_infos []osd_perf_info `json:"osd_perf_infos"`
}

type osd_metadata_struct struct {
	Arch               string `json:"arch"`
	Back_addr          string `json:"back_addr"`
	Ceph_version       string `json:"ceph_version"`
	Cpu                string `json:"cpu"`
	Distro             string `json:"distro"`
	Distro_codename    string `json:"distro_codename"`
	Distro_description string `json:"distro_description"`
	Distro_version     string `json:"distro_version"`
	Front_addr         string `json:"front_addr"`
	Hb_back_addr       string `json:"hb_back_addr"`
	Hb_front_addr      string `json:"hb_front_addr"`
	Hostname           string `json:"hostname"`
	Kernel_version     string `json:"kernel_version"`
	Mem_swap_kb        int64  `json:"mem_swap_kb"`
	Mem_total_kb       int64  `json:"mem_total_kb"`
	OS                 string `json:"os"`
	Osd_data           string `json:"osd_data"`
	Osd_journal        string `json:"osd_journal"`
}

type osd_info_struct struct {
	Id         int                 `json:"id"`
	Up_down    string              `json:"up_down"`
	In_out     string              `json:"in_out"`
	Perf_stats perf_stats_struct   `json:"perf_stats"`
	Metadata   osd_metadata_struct `json:"metadata"`
}

/**
 * @brief iostat struct
 */
type io_stats_sum_struct struct {
	Num_bytes                      int64 `json:"num_bytes"`
	Num_objects                    int64 `json:"num_objects"`
	Num_object_clones              int64 `json:"num_object_clones"`
	Num_object_copies              int64 `json:"num_object_copies"`
	Num_objects_missing_on_primary int64 `json:"num_objects_missing_on_primary"`
	Num_objects_degraded           int64 `json:"num_objects_degraded"`
	Num_objects_unfound            int64 `json:"num_objects_unfound"`
	Num_objects_dirty              int64 `json:"num_objects_dirty"`
	Num_whiteouts                  int64 `json:"num_whiteouts"`
	Num_read                       int64 `json:"num_read"`
	Num_read_kb                    int64 `json:"num_read_kb"`
	Num_write                      int64 `json:"num_write"`
	Num_write_kb                   int64 `json:"num_write_kb"`
	Num_scrub_errors               int64 `json:"num_scrub_errors"`
	Num_shallow_scrub_errors       int64 `json:"num_shallow_scrub_errors"`
	Num_deep_scrub_errors          int64 `json:"num_deep_scrub_errors"`
	Num_objects_recovered          int64 `json:"num_objects_recovered"`
	Num_bytes_recovered            int64 `json:"num_bytes_recovered"`
	Num_keys_recovered             int64 `json:"num_keys_recovered"`
	Num_objects_omap               int64 `json:"num_objects_omap"`
	Num_objects_hit_set_archive    int64 `json:"num_objects_hit_set_archive"`
}

type io_perf_statistic_struct struct {
	Stamp          string `json:"stamp"`
	Iops           int64  `json:"iops"`
	Rd_iops        int64  `json:"rd_iops"`
	Wd_iops        int64  `json:wd_iops`
	Read_speed_kb  int64  `json:"read_speed_kb"`
	Write_speed_kb int64  `json"write_speed_kb"`
}

type pg_stats_sum_struct struct {
	Stat_sum        io_stats_sum_struct `json:"Stat_sum"`
	Log_size        int64               `json:"log_size"`
	Ondisk_log_size int64               `json:"ondisk_log_size"`
}

type osd_stats_sum_struct struct {
	Kb_used             int64 `json:"kb_used"`
	Kb_avail            int64 `json:"kb_avail"`
	Snap_trim_queue_len int   `json:"snap_trim_queue_len"`
	Num_snap_trimming   int   `json:"num_snap_trimming"`
}

type pg_dump_sum_struct struct {
	Version           int    `json:"version"`
	Stamp             string `json:"stamp"`
	Last_osdmap_epoch int    `json:"last_osdmap_epoch"`
	Last_pg_scan      int    `json:"last_pg_scan"`
	Full_ratio        float64 `json:"full_ratio"`
	Near_full_ratio   float64 `json:"near_full_ratio"`

	Pg_stats_sum   pg_stats_sum_struct  `json:"pg_stats_sum"`
	Osd_stats_sum  osd_stats_sum_struct `json:"osd_stats_sum"`
	Pg_stats_delta pg_stats_sum_struct  `json:"pg_stats_delta"`
}

type pool_io_stats_struct struct {
	Poolid   int                 `json:"poolid"`
	Stat_sum io_stats_sum_struct `json:"stat_sum"`
}

type log_store_stats struct {
	Bytes_total  int64  `json:"bytes_total"`
	Bytes_sst    int64  `json:"bytes_sst"`
	Bytes_log    int64  `json:"bytes_log"`
	Bytes_misc   int64  `json:"bytes_misc"`
	Last_updated string `json:"last_updated"`
}

type mon_service_stat struct {
	Name 				string			`json:"name"`
	// disk info
	Kb_total            int64           `json:"kb_total"`
	Kb_last_updatedused int64           `json:"kb_used"`
	Kb_avail            int64           `json:"kb_avail"`
	Avail_percent       int             `json:"avail_percent"`
	Last_updated        string          `json:"last_updated"`
	Store_stats         log_store_stats `json:"store_stats"`
	Health              string          `json:"health"`
}

// ceph status
type cluster_health_services struct {
	Mons []mon_service_stat `json:"mons"`
}

type health_struct struct {
	Health_services []cluster_health_services `json:"health_services"`
}

type health_status struct {
	Health health_struct `json:"health"`
	//Summary        []string          `json:"summary"`
	Timechecks     timechecks_struct `json:"timechecks"`
	Overall_status string            `json:"overall_status"`
	//Detail         []string          `json:"detail"`
}

type timechecks_struct struct {
	Epoch        int    `json:"epoch"`
	Round        int    `json:"round"`
	Round_status string `json:"round_status"`
}

type ceph_health_status struct {
	Health         health_status     `json:"health"`
	Fsid           string            `json:"fsid"`
	Election_epoch int               `json:"election_epoch"`
	Quorum         []int             `json:"quorum"`
	Quorum_names   []string          `json:"quorum_names"`
	Monmap         monmap_struct     `json:"monmap"`
	Osdmap         osdmap_struct     `json:"osdmap"`
	Pgmap          pgmap_struct      `json:"pgmap"`
	Mdsmap         mdsmap_map        `json:"mdsmap"`
}


type monmap_item struct {
	Rank int    `json:"rank"`
	Name string `json:"name"`
	Addr string `json:"addr"`
}

type monmap_struct struct {
	Epoch    int           `json:"epoch"`
	Fsid     string        `json:"fsid"`
	Modified string        `json:"modified"`
	Created  string        `json:"created"`
	Mons     []monmap_item `json:"mons"`
}

// cluster mds stat
type mdsmap_map struct {
	Epoch   int   `json:"epoch"`
	Up      int   `json:"up"`
	In      int   `json:"in"`
	Max     int   `json:"max"`
	By_rank []int `json:"by_rank"`
}

type osdmap_item struct {
	Num_osds    int64  `json:"num_osds"`
	Num_up_osds int64  `json:"num_up_osds"`
	Num_in_osds int64  `json:"num_in_osds"`
	Full        bool `json:"full"`
	Nearfull    bool `json:"nearfull"`
}

// cluster osd stat
type osdmap_struct struct {
	Osdmap osdmap_item `json:"osdmap"`
}

type pgs_state struct {
	State_name string `json:"state_name"`
	Count      int    `json:"count"`
}

type pgmap_struct struct {
	Pgs_by_state []pgs_state `json:"pgs_by_state"`
	Version      int         `json:"version"`
	Num_pgs      int         `json:"num_pgs"`
	Data_bytes   int64       `json:"data_bytes"`
	Bytes_used   int64       `json:"bytes_used"`
	Bytes_avail  int64       `json:"bytes_avail"`
	Bytes_total  int64       `json:"bytes_total"`
}

type mon_info_struct struct {
	Stats mon_service_stat `json:"stat"`
	Addr  string           `json:"addr"`
}

type ceph_summary_df_stat struct {
	Total_bytes int64 `json:"total_bytes"`
	Total_used_bytes int64 `json:"total_used_bytes"`
	Total_avail_bytes int64 `json:"total_avail_bytes"`
}

type ceph_pool_df_stat struct {
	Kb_used int64 `json:"kb_used"`
	Bytes_used int64 `json:"bytes_used"`
	Max_avail int64 `json:"max_avail"`
	Objects int64 `json:"objects"`
}

type ceph_pool_info struct {
	Name string `json"name"`
	Id int `json:"id"`
	Stats  ceph_pool_df_stat `json:"stats"`
}

type ceph_df_result struct {
	Stats ceph_summary_df_stat `json:"stats"`
	Pools []ceph_pool_info `json:"pools"`
}


type ceph_osd_info struct {
	Id 		int `json:"id"`
	Name		string `json:"name"`
	Type    string `json:"type"`
	Type_id  int   `json:"type_id"`
	Crush_weight float64 `json:"crush_weight"`
	Depth 	 int `json:"depth"`
	Reweight float64 `json:"reweight"`
	Kb		 int64 `json:"kb"`
	Kb_used  int64 `json:"kb_used"`
	Kb_avail int64 `json:"kb_avail"`
	Utilization float64 `json:"utilization"`
	Var float64 `json:"var"`
}

type ceph_osd_df_result struct {
	Nodes  []ceph_osd_info `json:"nodes"`
}

type ceph_bucket_info struct {
	Id 		int `json:"id"`
	Name		string `json:"name"`
	Type    string `json:"type"`
	Type_id  int   `json:"type_id"`
	Crush_weight float64 `json:"crush_weight"`
	Depth	int `json:"depth"`
	Exists	int `json:"exists"`
	Status	string `json:"status"`
	Reweight float64 `json:"reweight"`
	Primary_affinity float64 `json:"primary_affinity"`
}

type ceph_osd_tree_result struct {
	Nodes []ceph_bucket_info `json:"nodes"`
}

type ceph_pool_detail struct {
	Name string `json:"pool_name"`
	PG_num int `json"pg_num"`
	Size int `json:"size"`
}


var old_sum_status pg_dump_sum_struct
var new_sum_status pg_dump_sum_struct
var old_pool_status_array []pool_io_stats_struct
var new_pool_status_array []pool_io_stats_struct
var pool_info_array []ceph_pool_info 
var ceph_last_time int64
var hostname string

/**
 * @brief Output integet value to collectd.
 */
func get_submit_stat_str(hostname, plugin, plugin_ins, str_type, str_type_ins string, value, time_value int64) string {
	stat := fmt.Sprintf("PUTVAL %s/%s-%s/%s-%s %d:%d\n", hostname, plugin, plugin_ins, str_type, str_type_ins, time_value,
		value)
	return stat
}

func trim(args []string) []string {
	s := make([]string, 0)
	for _, x := range args {
		if x != "" {
			s = append(s, x)
		}
	}
	return s
}

func exec_command(cmd_name string, arg string) (string, error) {
	var arg_array []string
	var res bytes.Buffer

	if arg != "" {
		arg_array = strings.Fields(arg)
	} else {
		arg_array = make([]string, 0)
	}

	cmd := exec.Command(cmd_name, arg_array...)
	cmd.Stdout = &res
	err := cmd.Run()

	return res.String(), err

}

func is_master_monitor() bool {
	var stat ceph_health_status

	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	str_json, err := exec_command("ceph", "status -f json")
	if err != nil {
		//		fmt.Println(err)
		return false
	}

	err = json.Unmarshal([]byte(str_json), &stat)
	if err != nil {
		return false
	}

	if len(stat.Quorum_names) > 0 {
		if strings.Contains(hostname, stat.Quorum_names[0]) {
			return true
		}
	}

	return false
}

func Init_summary_stat() error {
	str_stat, err := exec_command("ceph", "pg dump summary -f json")

	if err != nil {
		//fmt.Println("1")
		return err
	}

	err = json.Unmarshal([]byte(str_stat), &old_sum_status)
	if err != nil {
		//fmt.Println(err.Error())
		return err
	}

	str_stat, err = exec_command("ceph", "pg dump pools -f json")
	if err != nil {
		//fmt.Println("3")
		return err
	}

	err = json.Unmarshal([]byte(str_stat), &old_pool_status_array)
	if err != nil {
//		fmt.Println("4")
		return err
	}

	tm := time.Now()
	ceph_last_time = tm.Unix()
	return nil
}

func Get_io_status() {
	
	if is_master_monitor() == false {
		return
	}
	// init status at first time
	if ceph_last_time == 0 {
		Init_summary_stat()
		return
	}
	tm := time.Now()
	curr_time := tm.Unix()
	get_pool_df(curr_time)
	get_pools_io_stats(curr_time)
	get_every_pool_io_stats(curr_time)
	get_osds_perf(curr_time)
	get_cluster_health(curr_time)
	get_osd_df(curr_time)
	get_osd_stat(curr_time)
	get_pool_detail(curr_time)
	ceph_last_time = curr_time

}

func cal_io_stat(curr_time int64, id int, old_sum_status, new_sum_status io_stats_sum_struct, name string) {
	var io_perf io_perf_statistic_struct
	var read int64
	var read_kb int64
	var write int64
	var write_kb int64

	f := bufio.NewWriter(os.Stdout)
	delta_time := curr_time - ceph_last_time
	
	if new_sum_status.Num_read > old_sum_status.Num_read {
		read = new_sum_status.Num_read - old_sum_status.Num_read
	} else {
		read = 0
	}

	if new_sum_status.Num_write > old_sum_status.Num_write {
		write = new_sum_status.Num_write - old_sum_status.Num_write
	} else {
		write = 0
	}

	if new_sum_status.Num_read_kb > old_sum_status.Num_read_kb {
		read_kb = new_sum_status.Num_read_kb - old_sum_status.Num_read_kb
	} else {
		read_kb = 0
	}

	if new_sum_status.Num_write_kb > old_sum_status.Num_write_kb {
		write_kb = new_sum_status.Num_write_kb - old_sum_status.Num_write_kb
	} else {
		write_kb = 0
	}

	io_perf.Iops = (read + write) / delta_time
	io_perf.Rd_iops = read / delta_time
	io_perf.Wd_iops = write / delta_time
	io_perf.Read_speed_kb = read_kb / delta_time
	io_perf.Write_speed_kb = write_kb / delta_time

	stat := get_submit_stat_str("ceph_cluster", "pool", name, "gauge", "iops", io_perf.Iops, curr_time)
	stat += get_submit_stat_str("ceph_cluster", "pool", name, "gauge", "read_iops", io_perf.Rd_iops, curr_time)
	stat += get_submit_stat_str("ceph_cluster", "pool", name, "gauge", "write_iops", io_perf.Wd_iops, curr_time)
	stat += get_submit_stat_str("ceph_cluster", "pool", name, "gauge", "write_speed", (io_perf.Write_speed_kb * 1024), curr_time)
	stat += get_submit_stat_str("ceph_cluster", "pool", name, "gauge", "read_speed", (io_perf.Read_speed_kb * 1024), curr_time)

	f.Write([]byte(stat))
	f.Flush()
}

func get_pool_name(id, index int) (string){
	
	if index < len(pool_info_array) {
		if (id == pool_info_array[index].Id) {
			return pool_info_array[index].Name
		}
	}
	
	for i := range(pool_info_array) {
		if (id == pool_info_array[i].Id) {
			return pool_info_array[i].Name
		}
	}
	
	return ""
}

func get_every_pool_io_stats(curr_time int64) {
	var name string
	
	if ceph_last_time == 0 {
		return
	}

	str_stat, err := exec_command("ceph", "pg dump pools -f json")

	err = json.Unmarshal([]byte(str_stat), &new_pool_status_array)
	if err != nil {
		return
	}

	for i := range old_pool_status_array {
		name = get_pool_name(old_pool_status_array[i].Poolid, i)
		if name == "" {
			continue
		}
		
		if i < len(new_pool_status_array) && old_pool_status_array[i].Poolid == new_pool_status_array[i].Poolid {
			cal_io_stat(curr_time, old_pool_status_array[i].Poolid, 
			old_pool_status_array[i].Stat_sum, new_pool_status_array[i].Stat_sum, name)
			continue
		}

		for j := range new_pool_status_array {
			if old_pool_status_array[i].Poolid == new_pool_status_array[j].Poolid {
				cal_io_stat(curr_time, old_pool_status_array[i].Poolid, 
				old_pool_status_array[i].Stat_sum, new_pool_status_array[j].Stat_sum, name)
				break
			}
		}
	}
	
	old_array_len := len(old_pool_status_array)
	new_array_len := len(new_pool_status_array)
	if (old_array_len != new_array_len) {
		old_pool_status_array = make([]pool_io_stats_struct, new_array_len)
	}
	
	for i := range new_pool_status_array {
		old_pool_status_array[i] = new_pool_status_array[i]
	}
}

func get_pools_io_stats(curr_time int64) {
	var io_perf io_perf_statistic_struct
	f := bufio.NewWriter(os.Stdout)

	if ceph_last_time == 0 {
		return
	}

	str_stat, err := exec_command("ceph", "pg dump summary -f json-pretty")

	err = json.Unmarshal([]byte(str_stat), &new_sum_status)
	if err != nil {
		return
	}

	if curr_time <= ceph_last_time {
		return
	}

	delta_time := curr_time - ceph_last_time
	read := new_sum_status.Pg_stats_sum.Stat_sum.Num_read - old_sum_status.Pg_stats_sum.Stat_sum.Num_read
	write := new_sum_status.Pg_stats_sum.Stat_sum.Num_write - old_sum_status.Pg_stats_sum.Stat_sum.Num_write

	read_kb := new_sum_status.Pg_stats_sum.Stat_sum.Num_read_kb - old_sum_status.Pg_stats_sum.Stat_sum.Num_read_kb
	write_kb := new_sum_status.Pg_stats_sum.Stat_sum.Num_write_kb - old_sum_status.Pg_stats_sum.Stat_sum.Num_write_kb

	io_perf.Iops = (read + write) / delta_time
	io_perf.Rd_iops = read / delta_time
	io_perf.Wd_iops = write / delta_time
	io_perf.Read_speed_kb = read_kb / delta_time
	io_perf.Write_speed_kb = write_kb / delta_time

	stat := get_submit_stat_str("ceph_cluster", "summary", "pool", "gauge", "iops", io_perf.Iops, curr_time)
	stat += get_submit_stat_str("ceph_cluster", "summary", "pool", "gauge", "read_iops", io_perf.Rd_iops, curr_time)
	stat += get_submit_stat_str("ceph_cluster", "summary", "pool", "gauge", "write_iops", io_perf.Wd_iops, curr_time)
	stat += get_submit_stat_str("ceph_cluster", "summary", "pool", "gauge", "write_speed", (io_perf.Write_speed_kb * 1024), curr_time)
	stat += get_submit_stat_str("ceph_cluster", "summary", "pool", "gauge", "read_speed", (io_perf.Read_speed_kb * 1024), curr_time)

	f.Write([]byte(stat))
	f.Flush()

	old_sum_status = new_sum_status
}

func get_cluster_health(curr_time int64) {
	var ceph_stat ceph_health_status
	var stat string
	var str_health string
	
	f := bufio.NewWriter(os.Stdout)
	
	if ceph_last_time == 0 {
		return
	}

	str_stat, err := exec_command("ceph", "status -f json")

	err = json.Unmarshal([]byte(str_stat), &ceph_stat)
	if err != nil {
		return
	}
	
	if len(ceph_stat.Health.Health.Health_services) > 0 {		
		if len (ceph_stat.Health.Health.Health_services[0].Mons) > 0 {
			str_health = ceph_stat.Health.Health.Health_services[0].Mons[0].Health;
			if  str_health == "HEALTH_OK" {
				stat = get_submit_stat_str("ceph_cluster", "summary", "monitor", "gauge", "health", 1, curr_time)
			} else if str_health == "HEALTH_WARN" {
				stat = get_submit_stat_str("ceph_cluster", "summary", "monitor", "gauge", "health", 0, curr_time)
			} else {
				stat = get_submit_stat_str("ceph_cluster", "summary", "monitor", "gauge", "health", -1, curr_time)
			}	
			mon_num := len(ceph_stat.Health.Health.Health_services[0].Mons)
			stat += get_submit_stat_str("ceph_cluster", "summary", "monitor", "gauge", "num", int64(mon_num), curr_time)
		}
		
		str_health = ceph_stat.Health.Overall_status
		if  str_health == "HEALTH_OK" {
			stat += get_submit_stat_str("ceph_cluster", "summary", "cluster", "gauge", "health", 1, curr_time)
		} else if str_health == "HEALTH_WARN" {
			stat += get_submit_stat_str("ceph_cluster", "summary", "cluster", "gauge", "health", 0, curr_time)
		} else {
			stat += get_submit_stat_str("ceph_cluster", "summary", "cluster", "gauge", "health", -1, curr_time)
		}
		
	}
	
	// print the osd status 
	stat += get_submit_stat_str("ceph_cluster", "summary", "osd", "gauge", "num_osds", 
	     ceph_stat.Osdmap.Osdmap.Num_osds, curr_time)
		
	stat += get_submit_stat_str("ceph_cluster", "summary", "osd", "gauge", "num_up_osds", 
	     ceph_stat.Osdmap.Osdmap.Num_up_osds, curr_time)
	
	stat += get_submit_stat_str("ceph_cluster", "summary", "osd", "gauge", "num_in_osds", 
	     ceph_stat.Osdmap.Osdmap.Num_in_osds, curr_time)
	
	f.Write([]byte(stat))
	f.Flush()

}

func get_osds_perf(curr_time int64) {
	var all_osd_perf cluster_osds_perf_infos
	f := bufio.NewWriter(os.Stdout)
	var stat string

	if ceph_last_time == 0 {
		return
	}

	str_stat, err := exec_command("ceph", "osd perf -f json")

	err = json.Unmarshal([]byte(str_stat), &all_osd_perf)
	if err != nil {
		return
	}

	for i := range all_osd_perf.Osd_perf_infos {
		id := all_osd_perf.Osd_perf_infos[i].Id
		stat += get_submit_stat_str("ceph_cluster", "osd", strconv.Itoa(id), "gauge", "commit_latency",
			all_osd_perf.Osd_perf_infos[i].Perf_stats.Commit_latency_ms, curr_time)
		stat += get_submit_stat_str("ceph_cluster", "osd", strconv.Itoa(id), "gauge", "apply_latency",
			all_osd_perf.Osd_perf_infos[i].Perf_stats.Apply_latency_ms, curr_time)
	}

	f.Write([]byte(stat))
	f.Flush()
}

func get_pool_df(curr_time int64) {
	var ceph_df ceph_df_result
	var stat string
	
	f := bufio.NewWriter(os.Stdout)

	if ceph_last_time == 0 {
		return
	}

	str_stat, err := exec_command("ceph", "df -f json")

	err = json.Unmarshal([]byte(str_stat), &ceph_df)
	if err != nil {
		return
	}
	
	stat += get_submit_stat_str("ceph_cluster", "summary", "cluster", "gauge", "total_bytes",
		ceph_df.Stats.Total_bytes, curr_time)
	stat += get_submit_stat_str("ceph_cluster", "summary", "cluster", "gauge", "total_used_bytes",
		ceph_df.Stats.Total_used_bytes, curr_time)
	stat += get_submit_stat_str("ceph_cluster", "summary", "cluster", "gauge", "total_avail_bytes",
		ceph_df.Stats.Total_avail_bytes, curr_time)
	
	for i := range ceph_df.Pools {
		stat += get_submit_stat_str("ceph_cluster", "pool", ceph_df.Pools[i].Name, 
		"gauge", "used_bytes", ceph_df.Pools[i].Stats.Bytes_used, curr_time)
		
		stat += get_submit_stat_str("ceph_cluster", "pool", ceph_df.Pools[i].Name,
		"gauge", "avail_bytes", ceph_df.Pools[i].Stats.Max_avail, curr_time)
		
		stat += get_submit_stat_str("ceph_cluster", "pool", ceph_df.Pools[i].Name,
		"gauge", "objects", ceph_df.Pools[i].Stats.Objects, curr_time)
	}
	
	f.Write([]byte(stat))
	f.Flush()
	pool_info_array = ceph_df.Pools
}

func get_osd_df(curr_time int64) { 
	var ceph_osd_df ceph_osd_df_result
	var stat string
	
	f := bufio.NewWriter(os.Stdout)

	if ceph_last_time == 0 {
		return
	}

	str_stat, err := exec_command("ceph", "osd df -f json")

	err = json.Unmarshal([]byte(str_stat), &ceph_osd_df)
	if err != nil {
		return
	}
	
	for i := range ceph_osd_df.Nodes {
		id := ceph_osd_df.Nodes[i].Id
		stat += get_submit_stat_str("ceph_cluster", "osd", strconv.Itoa(id), "gauge", "capacity_kb",
			ceph_osd_df.Nodes[i].Kb, curr_time)
		stat += get_submit_stat_str("ceph_cluster", "osd", strconv.Itoa(id), "gauge", "used_kb",
			ceph_osd_df.Nodes[i].Kb_used, curr_time)	
		stat += get_submit_stat_str("ceph_cluster", "osd", strconv.Itoa(id), "gauge", "avail_kb",
			ceph_osd_df.Nodes[i].Kb_avail, curr_time)
		stat += node.Get_submit_float_stat_str("ceph_cluster", "osd", strconv.Itoa(id), "gauge", "utilization",
			ceph_osd_df.Nodes[i].Utilization, curr_time)
	}
	
	f.Write([]byte(stat))
	f.Flush()
}

func get_osd_stat(curr_time int64) {
	var osd_tree ceph_osd_tree_result
	var stat string
	var value int64
	
	f := bufio.NewWriter(os.Stdout)

	if ceph_last_time == 0 {
		return
	}

	str_stat, err := exec_command("ceph", "osd tree -f json")

	err = json.Unmarshal([]byte(str_stat), &osd_tree)
	if err != nil {
		return
	}
	
	for i := range osd_tree.Nodes {
		id := osd_tree.Nodes[i].Id
		
		// check if is osd bucket 
		if id  >= 0  {
			if osd_tree.Nodes[i].Status == "up" {
				value = 1
			} else if osd_tree.Nodes[i].Status == "in" {
				value = 0
			} else {
				value = -1
			}
			stat += get_submit_stat_str("ceph_cluster", "osd", strconv.Itoa(id), "gauge", "status",
			value, curr_time)
		}
	}
	
	f.Write([]byte(stat))
	f.Flush()
}


func get_pool_detail(curr_time int64) {
	var pool_detail_array []ceph_pool_detail
	var stat string
	
	f := bufio.NewWriter(os.Stdout)

	if ceph_last_time == 0 {
		return
	}

	str_stat, err := exec_command("ceph", "osd pool ls detail -f json")

	err = json.Unmarshal([]byte(str_stat), &pool_detail_array)
	if err != nil {
		return
	}
	
	for i := range pool_detail_array {
		stat += get_submit_stat_str("ceph_cluster", "pool", pool_detail_array[i].Name, "gauge", "replication",
			int64(pool_detail_array[i].Size), curr_time)
		
		stat += get_submit_stat_str("ceph_cluster", "pool", pool_detail_array[i].Name, "gauge", "pg_num",
			int64(pool_detail_array[i].PG_num), curr_time)
	
	}
	
	f.Write([]byte(stat))
	f.Flush()
}
