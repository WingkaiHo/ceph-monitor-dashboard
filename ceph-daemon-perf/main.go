// ceph-perf-dump project main.go
package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"strconv"
	"bytes"
	"os/exec"
	"encoding/json"
	"time"
	"bufio"
	"flag"
)

var sleepTime int
var debug bool
	
type WBThrottle_struct struct{
	Byte_dirtied int64 `json:"bytes_dirtied"`
	Byte_wb int64 `json:"bytes_wb"`
	Ios_dirtied int64 `json:"ios_dirtied"`
	Ios_wb int64 `json:"ios_wb"`
	Inodes_dirtied int64 `json:"inodes_dirtied"`
	Inodes_wb int64 `json:"inodes_wb"`
}


type statistic_struct struct {
	Avgcount int64 `json:"avgcount"`
	Sum float64 `json:"sum"`
}

type Filestore_struct struct {
	Journal_queue_max_ops int64 `json:"journal_queue_max_ops"`
	Journal_queue_ops int64 `json:"journal_queue_ops"`
	Journal_ops int64 `json:"journal_ops"`
	Journal_queue_max_bytes int64 `json:"journal_queue_max_bytes"`
	Journal_queue_bytes int64 `json:"journal_queue_bytes"`
	Journal_bytes int64 `json:"journal_bytes"`
	Journal_latency statistic_struct `json:"journal_latency"`
	Journal_wr int64 `json:"journal_wr"`
	Journal_wr_bytes statistic_struct `json:"journal_wr_bytes"`
	Journal_full int `json:"journal_full"`
	
	// journal commit 
	Committing int64 `json:"committing"`
	Commitcycle int64 `josn:"commitcycle"`
	Commitcycle_interval statistic_struct `json:"commitcycle_interval"`
	Commitcycle_latency  statistic_struct `json:"commitcycle_latency"`
	
	// op_queue
	Op_queue_max_ops int64 `json:"op_queue_max_ops"`
	Op_queue_ops int64 `json:"op_queue_ops"`
	Ops int64 `json:"ops"`
	Op_queue_max_bytes int64 `json:"op_queue_max_bytes"`
	Op_queue_bytes int64 `json:"op_queue_bytes"`
	Bytes int64 `json:"bytes"`
	Apply_latency statistic_struct `json:"apply_latency"`
	Queue_transaction_latency_avg statistic_struct `json:"queue_transaction_latency_avg"`
}

type Filestore_Leveldb_struct struct {
	Leveldb_get int64 `json:"leveldb_get"`
	Leveldb_transaction int64 `json:"leveldb_transaction"`
	Leveldb_compact int64 `json:"leveldb_compact"`
	Leveldb_compact_range  int64 `json:"leveldb_compact_range"`
	Leveldb_compact_queue_merge int64 `json:"leveldb_compact_queue_merge"`
	Leveldb_compact_queue_len int64 `json:"leveldb_compact_queue_len"`	
}

type OSD_struct struct {
	// osd_pg 
	Op_wip int64 `json:"op_wip"`
	Op int64 `json:"op"`
	Op_in_bytes int64 `json:"op_in_bytes"`
	Op_out_bytes int64 `json:"op_out_bytes"`
	Op_latency statistic_struct `json:"op_latency"`
	Op_process_latency statistic_struct `json:"op_process_latency"`
	
	// osd_pg_client_r
	Op_r int64 `json:"op_r"`
	Op_r_out_bytes int64 `json:"Op_r_out_bytes"`
	Op_r_latency statistic_struct `json:"op_r_latency"`
	Op_r_process_latency statistic_struct `json:"op_r_process_latency"`
	
	// osd_pg_client_w
	Op_w int64 `json:"op_w"`
	Op_w_in_bytes int64 `json:"op_w_in_bytes"`
	Op_w_rlat statistic_struct `json:"op_w_rlat"`
	Op_w_latency statistic_struct `json:"op_w_latency"`
	Op_w_process_latency statistic_struct `json:"op_w_process_latency"`
	
	// osd_pg_client_rw
	Op_rw int64 `json:"op_rw"`
	Op_rw_in_bytes int64 `json:"op_rw_in_bytes"`
	Op_rw_out_bytes int64 `json:"op_rw_out_bytes"`
	Op_rw_rlat statistic_struct `json:"op_rw_rlat"`
	Op_rw_latency statistic_struct `json:"op_rw_latency"`
	Op_rw_process_latency statistic_struct `json:"op_rw_process_latency"`
	
	// osd_pg_cluster_w
	Subop_w int64 `json:"subop_w"`
	Subop_w_in_bytes int64 `json:"subop_w_in_bytes"`
	Subop_w_latency statistic_struct `json:"subop_w_latency"`
	
	// ******
	
}

type throttle_struct struct {
	Val int64	`json:"val"`
	Max int64	`json:"max"`
	Get int64	`json:"get"`
	Get_sum int64 `json:"get_sum"`
	Put int64 `json:"put"`
	Put_sum int64 `json:"put_sum"`
	Wait statistic_struct `json:"wait"`
}

type osd_throttle_perf struct {
	WBThrottle WBThrottle_struct `json:"WBThrottle"`
	Filestore Filestore_struct `json:"filestore"`
	Leveldb Filestore_Leveldb_struct `json:"leveldb"`
	Osd OSD_struct	`json:"osd"`
	Throttle_msgr_dispatch_throttler_client throttle_struct `json:"throttle-msgr_dispatch_throttler-client"`
	Throttle_msgr_dispatch_throttler_cluster throttle_struct `json:"throttle-msgr_dispatch_throttler-cluster"`
	Throttle_osd_client_messages throttle_struct `json:"throttle-osd_client_messages"`
}

type osd_throttle_perf_extend struct {
	osd_id int 
	time int64
	osd_perf_dump osd_throttle_perf
}

type throttle_attr_status struct {
	op_ps int64
	lat_per_op float64
	proce_lat_per_op float64
	in_bytes_ps int64
	out_bytes_ps int64
}

var osd_perf_map map[int]osd_throttle_perf_extend

func init() {
	flag.BoolVar(&debug, "debug", false, "turn on debugging")
	flag.IntVar(&sleepTime, "sleep-time", 10, "Number of seconds between runs")
	flag.Parse()
}

/**
 * @brief Output integet value to collectd.
 */
func get_submit_int_stat_str(procname, plugin, plugin_ins, str_type, str_type_ins string, value, time_value int64) string {
	stat := fmt.Sprintf("PUTVAL proc_%s/%s-%s/%s-%s %d:%d\n", procname, plugin, plugin_ins, str_type, str_type_ins, time_value,
		value)
	return stat
}


func get_submit_float_stat_str(procname, plugin, plugin_ins, str_type, str_type_ins string, value float64, time_value int64) string {
	stat := fmt.Sprintf("PUTVAL proc_%s/%s-%s/%s-%s %d:%f\n", procname, plugin, plugin_ins, str_type, str_type_ins, time_value,
		value)
	return stat
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

func cal_avg_latency(old, curr statistic_struct) (float64){
	var avg float64
	
	avg = 0
	if (curr.Avgcount > old.Avgcount) {
		avg = (curr.Sum - old.Sum) / float64(curr.Avgcount - old.Avgcount)
	} else {
		if (old.Avgcount > 0) {
			avg = curr.Sum / float64(old.Avgcount)
		}
	}
	
	return avg
}

func cal_osd_throttole_or_throttle_msgr_status(old, curr *throttle_struct, 
     time_delta int64) (throttle_attr_status) {
	
	var res throttle_attr_status
	
	res.op_ps = 0
	if curr.Get > old.Get {
		res.op_ps = curr.Get - old.Get 
	}
	
	if curr.Put > old.Put {
		res.op_ps += (curr.Put - old.Put) 
	}
	
	if  curr.Get != curr.Get_sum && curr.Get_sum > old.Get_sum {
		res.out_bytes_ps = (curr.Get_sum - old.Get_sum) / time_delta
	}
	
	if curr.Put != curr.Put_sum && curr.Put_sum > old.Put_sum {
		res.in_bytes_ps = (curr.Put_sum - old.Put_sum) / time_delta
	}
	
	res.lat_per_op = cal_avg_latency(old.Wait, curr.Wait)
	
	return res
}

/** 
 * @brief calculate the osd pg multi level throttle status
 * 
 * @return the level throttle status as osd_pg, osd_pg_client_w, osd_pg_client_r, osd_pg_client_rw(replication sync), osd_pg_cluster_w
 */
func cal_osd_pg_status(old, curr *OSD_struct, time_delta int64)(throttle_attr_status, 
         throttle_attr_status, throttle_attr_status, throttle_attr_status, throttle_attr_status) {
	
	var osd_pg, osd_pg_client_w, osd_pg_client_r, osd_pg_client_rw, osd_pg_cluster_w throttle_attr_status
	
	// osd_pg
	if curr.Op > old.Op {
		osd_pg.op_ps = (curr.Op - old.Op)/time_delta;
	} 
	
	if curr.Op_in_bytes > old.Op_in_bytes {
		osd_pg.in_bytes_ps = (curr.Op_in_bytes - old.Op_in_bytes)/time_delta
	}	
	
	if curr.Op_out_bytes > old.Op_out_bytes {
		osd_pg.out_bytes_ps = (curr.Op_out_bytes - old.Op_out_bytes)/time_delta
	}
	
	osd_pg.lat_per_op = cal_avg_latency(old.Op_latency, curr.Op_latency)
	osd_pg.proce_lat_per_op = cal_avg_latency(old.Op_process_latency,curr.Op_process_latency)
	
	
	// osd_pg_client_w
	if curr.Op_w > old.Op_w {
		osd_pg_client_w.op_ps = (curr.Op_w - old.Op_w)/time_delta
	}
	
	if curr.Op_w_in_bytes > old.Op_w_in_bytes {
		osd_pg_client_w.in_bytes_ps = (curr.Op_w_in_bytes - old.Op_w_in_bytes)/time_delta
	}
	
	osd_pg_client_w.lat_per_op = cal_avg_latency(old.Op_w_latency, curr.Op_w_latency)
	osd_pg_client_w.proce_lat_per_op = cal_avg_latency(old.Op_w_process_latency, curr.Op_w_process_latency)
	
	// osd_pg_client_r
	if curr.Op_r > old.Op_r {
		osd_pg_client_r.op_ps = (curr.Op_r - old.Op_r) / time_delta
	}
	
	
	if curr.Op_r_out_bytes > old.Op_r_out_bytes {
		osd_pg_client_r.in_bytes_ps =  (curr.Op_r_out_bytes - old.Op_r_out_bytes) / time_delta
	}
	
	osd_pg_client_r.lat_per_op = cal_avg_latency(old.Op_r_latency, curr.Op_r_latency)
	osd_pg_client_r.proce_lat_per_op = cal_avg_latency(old.Op_rw_process_latency, curr.Op_r_process_latency)
	
	// osd_pg_client_rw (replication sync)
	if curr.Op_rw > old.Op_rw {
		osd_pg_client_rw.op_ps = (curr.Op_rw - old.Op_rw)/time_delta
	}
	
	if curr.Op_rw_in_bytes > old.Op_rw_in_bytes {
		osd_pg_client_rw.in_bytes_ps =  (curr.Op_rw_in_bytes - old.Op_rw_in_bytes)/time_delta
	}
	
	if curr.Op_rw_out_bytes > old.Op_rw_out_bytes {
		osd_pg_client_rw.in_bytes_ps =  (curr.Op_rw_out_bytes - old.Op_rw_out_bytes)/time_delta
	}
	
	osd_pg_client_rw.lat_per_op = cal_avg_latency(old.Op_rw_latency, curr.Op_rw_latency)
	osd_pg_client_rw.proce_lat_per_op = cal_avg_latency(old.Op_rw_process_latency, curr.Op_rw_process_latency)
	
	// osd_pg_cluster_w
	if curr.Subop_w > old.Subop_w {
		osd_pg_cluster_w.op_ps = (curr.Subop_w - old.Subop_w)/time_delta
	}
	
	if curr.Subop_w_in_bytes > old.Subop_w_in_bytes {
		osd_pg_cluster_w.in_bytes_ps = (curr.Subop_w_in_bytes - old.Subop_w_in_bytes)/time_delta
	}
	
	osd_pg_cluster_w.lat_per_op = cal_avg_latency(old.Subop_w_latency, curr.Subop_w_latency)
	
	
	return osd_pg, osd_pg_client_w, osd_pg_client_r, osd_pg_client_rw, osd_pg_cluster_w
}

func cal_filestore_ops_status(old, curr Filestore_struct, time_delta int64) (throttle_attr_status, throttle_attr_status, throttle_attr_status, throttle_attr_status, throttle_attr_status) {
	
	var filestore_op_queue, filestore_journal_queue, filestore_journal, filestore_commit, filestore_apply throttle_attr_status
	
	// filestore_journal_queue
	if (curr.Journal_queue_ops > old.Journal_queue_ops) {
		filestore_journal_queue.op_ps = (curr.Journal_queue_ops - old.Journal_queue_ops)/time_delta
	} 
	
	if (curr.Journal_queue_bytes > old.Journal_queue_bytes) {
		filestore_journal_queue.in_bytes_ps =(curr.Journal_queue_bytes - old.Journal_queue_bytes)/time_delta
	}
	
	// filestore_journal
	if (curr.Journal_ops > old.Journal_ops) {
		filestore_journal.op_ps = (curr.Journal_ops - old.Journal_ops)/time_delta
	}
	
	if (curr.Journal_bytes > old.Journal_bytes) {
		filestore_journal.in_bytes_ps=(curr.Journal_bytes - old.Journal_bytes)/time_delta
	}
	
	if (curr.Journal_wr_bytes.Sum > old.Journal_wr_bytes.Sum) {
		filestore_journal.out_bytes_ps = int64(curr.Journal_wr_bytes.Sum - old.Journal_wr_bytes.Sum) / time_delta
	}
	
	filestore_journal.lat_per_op = cal_avg_latency(old.Journal_latency, curr.Journal_latency)
	
	// filestore_commit
	if (curr.Commitcycle_latency.Avgcount > old.Commitcycle_latency.Avgcount) {
		filestore_commit.op_ps = (curr.Commitcycle_latency.Avgcount - old.Commitcycle_latency.Avgcount)/time_delta
	}
	filestore_commit.lat_per_op = cal_avg_latency(old.Commitcycle_latency, curr.Commitcycle_latency)
	
	// op_queue_ops
	if (curr.Op_queue_ops > old.Op_queue_ops) {
		filestore_op_queue.op_ps = (curr.Op_queue_ops - old.Op_queue_ops) / time_delta
	} 
	
	if (curr.Op_queue_bytes > old.Op_queue_bytes) {
		filestore_op_queue.in_bytes_ps = (curr.Op_queue_bytes - old.Op_queue_bytes) / time_delta
	}
	
	// filestore_apply
	filestore_apply.lat_per_op = cal_avg_latency(old.Apply_latency, curr.Apply_latency)
	//filestore_apply.op_ps = (curr.Apply_latency.Avgcount - old.Apply_latency.Avgcount) / time_delta
	//filestore_apply.in_bytes_ps = (curr.Bytes - old.Bytes)/time_delta

	
	return filestore_op_queue, filestore_journal_queue, filestore_journal, filestore_commit, filestore_apply
}

func cal_filestore_leveldb_status(old, curr Filestore_Leveldb_struct, time_delta int64)(throttle_attr_status){
	var filestore_leveldb throttle_attr_status
	
	if curr.Leveldb_get > old.Leveldb_get {
		filestore_leveldb.op_ps = curr.Leveldb_get - old.Leveldb_get
	}
	
	if curr.Leveldb_transaction > old.Leveldb_transaction {
		filestore_leveldb.op_ps += (curr.Leveldb_transaction - old.Leveldb_transaction)
	}
	
	filestore_leveldb.op_ps /= time_delta
	return filestore_leveldb
}

func cal_filestore_WBThrottle(old, curr WBThrottle_struct, time_delta int64)(throttle_attr_status) {
	var filestore_wb throttle_attr_status
	
	filestore_wb.in_bytes_ps = (curr.Byte_wb - old.Byte_wb) / time_delta
	filestore_wb.op_ps = (curr.Ios_wb - old.Ios_wb) / time_delta
	
	return filestore_wb
}

func dump_osd_throttle_perf(old_perf, curr_perf *osd_throttle_perf_extend, curr_time int64) {
	var stat string
	f := bufio.NewWriter(os.Stdout)
	
	procname := "osd-" + strconv.Itoa(old_perf.osd_id)
	time_delta := curr_perf.time - old_perf.time
	
	// dump osd_client_messenger status
	osd_client_mesgr := cal_osd_throttole_or_throttle_msgr_status(
	    &old_perf.osd_perf_dump.Throttle_osd_client_messages, 
		&curr_perf.osd_perf_dump.Throttle_osd_client_messages,
		time_delta)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_client_messages", "gauge", "op_ps", osd_client_mesgr.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_client_messages", "gauge", "in_spd", osd_client_mesgr.out_bytes_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_client_messages", "gauge", "out_spd", osd_client_mesgr.in_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_client_messages", "gauge", "latency", osd_client_mesgr.lat_per_op, curr_time)
	
	// dump osd_dispacth_client status
	osd_dispatch_client := cal_osd_throttole_or_throttle_msgr_status(
		&old_perf.osd_perf_dump.Throttle_msgr_dispatch_throttler_client,
		&curr_perf.osd_perf_dump.Throttle_msgr_dispatch_throttler_client,
		time_delta)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_dispatch_client", "gauge", "op_ps", osd_dispatch_client.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_dispatch_client", "gauge", "in_spd", osd_dispatch_client.in_bytes_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_dispatch_client", "gauge", "out_spd", osd_dispatch_client.out_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_dispatch_client", "gauge", "latency", osd_dispatch_client.lat_per_op, curr_time)
	
	// dump osd_dispatch_cluster status
	osd_dispatch_cluster := cal_osd_throttole_or_throttle_msgr_status(
		&old_perf.osd_perf_dump.Throttle_msgr_dispatch_throttler_cluster,
		&old_perf.osd_perf_dump.Throttle_msgr_dispatch_throttler_cluster,
		time_delta)	
	stat += get_submit_int_stat_str(procname, "throttle", "osd_dispatch_cluster", "gauge", "op_ps", osd_dispatch_cluster.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_dispatch_cluster", "gauge", "in_spd", osd_dispatch_cluster.in_bytes_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_dispatch_cluster", "gauge", "out_spd", osd_dispatch_cluster.out_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_dispatch_cluster", "gauge", "latency", osd_dispatch_cluster.lat_per_op, curr_time)
	
	osd_pg, osd_pg_client_w, osd_pg_client_r, osd_pg_client_rep_sync, osd_pg_cluster_w := cal_osd_pg_status(&old_perf.osd_perf_dump.Osd, &curr_perf.osd_perf_dump.Osd, time_delta)
	
	// dump osd_pg status
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg", "gauge", "op_ps", osd_pg.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg", "gauge", "in_spd", osd_pg.in_bytes_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg", "gauge", "out_spd", osd_pg.out_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_pg", "gauge", "latency", osd_pg.lat_per_op, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_pg", "gauge", "process_latency", osd_pg.proce_lat_per_op, curr_time)
	
	// dump osd_pg_client read status
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_client_read", "gauge", "op_ps", osd_pg_client_r.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_client_read", "gauge", "out_spd", osd_pg_client_r.out_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_pg_client_read", "gauge", "latency", osd_pg_client_r.lat_per_op, curr_time)
	
	// dump osd_pg_client write status
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_client_write", "gauge", "op_ps", osd_pg_client_w.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_client_write", "gauge", "in_spd", osd_pg_client_w.out_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_pg_client_write", "gauge", "latency", osd_pg_client_w.lat_per_op, curr_time)
	
	// dump osd_pg_rep_sync (osd_pg_client_rw)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_rep_sync", "gauge", "op_ps", osd_pg_client_rep_sync.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_rep_sync", "gauge", "in_spd", osd_pg_client_rep_sync.in_bytes_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_rep_sync", "gauge", "out_spd", osd_pg_client_rep_sync.out_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_pg_rep_sync", "gauge", "latency", osd_pg_client_rep_sync.lat_per_op, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_pg_rep_sync", "gauge", "process_latency", osd_pg_client_rep_sync.proce_lat_per_op, curr_time)
	
	// dump osd_pg_rep_sync (osd_pg_cluster_write)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_cluster_write", "gauge", "op_ps", osd_pg_cluster_w.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "osd_pg_cluster_write", "gauge", "in_spd", osd_pg_cluster_w.out_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "osd_pg_cluster_write", "gauge", "latency", osd_pg_cluster_w.lat_per_op, curr_time)
	
	// Filestore_Leveldb_struct
	filestore_leveldb := cal_filestore_leveldb_status(old_perf.osd_perf_dump.Leveldb, curr_perf.osd_perf_dump.Leveldb, time_delta)
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_leveldb", "gauge", "op_ps", filestore_leveldb.op_ps, curr_time)
	
	filestore_wb := cal_filestore_WBThrottle(old_perf.osd_perf_dump.WBThrottle, curr_perf.osd_perf_dump.WBThrottle, time_delta)
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_wb", "gauge", "op_ps", filestore_wb.op_ps, curr_time)
	
	filestore_op_queue, filestore_journal_queue, filestore_journal, filestore_commit, filestore_apply := 
	      cal_filestore_ops_status(old_perf.osd_perf_dump.Filestore, curr_perf.osd_perf_dump.Filestore, time_delta)
	
	// dump filestore_op_queue
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_op_queue", "gauge", "op_ps", filestore_op_queue.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_op_queue", "gauge", "in_spd", filestore_op_queue.in_bytes_ps, curr_time)
	
	// dump filestore_journal_queue
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_journal_queue", "gauge", "op_ps", filestore_journal_queue.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_journal_queue", "gauge", "in_spd", filestore_journal_queue.in_bytes_ps, curr_time)
	
	// dump filestore_journal 
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_journal", "gauge", "op_ps", filestore_journal.op_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_journal", "gauge", "in_spd", filestore_journal.in_bytes_ps, curr_time)
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_journal", "gauge", "out_spd", filestore_journal.out_bytes_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "filestore_journal", "gauge", "latency", filestore_journal.lat_per_op, curr_time)
	
	// dump filestore_commit
	stat += get_submit_int_stat_str(procname, "throttle", "filestore_commit", "gauge", "op_ps", filestore_commit.op_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "filestore_commit", "gauge", "latency", filestore_commit.lat_per_op, curr_time)
	
	// dump filestore_apply
	//stat += get_submit_int_stat_str(procname, "throttle", "filestore_apply", "gauge", "op_ps", filestore_apply.op_ps, curr_time)
	stat += get_submit_float_stat_str(procname, "throttle", "filestore_apply", "gauge", "latency", filestore_apply.lat_per_op, curr_time)
	
	f.Write([]byte(stat))
	f.Flush()
	
}

func get_process_osd_throttle_perf(name string) {
	var cmd string
	var str_json string
	var curr_osd_perf osd_throttle_perf_extend
	var old_osd_perf osd_throttle_perf_extend
	var num int 
	var err error
	
	
	num, err = fmt.Sscanf(name, "ceph-osd.%d.asok", &curr_osd_perf.osd_id)
	if (num <= 0) {
		return
	}
	
	cmd = "daemon /var/run/ceph/" + name + " perf dump"
	str_json, err = exec_command("ceph", cmd)
	if err != nil {
		return
	}
	
	err = json.Unmarshal([]byte(str_json), &curr_osd_perf.osd_perf_dump)
	if err != nil {
		return
	}
	
	curr_osd_perf.time = time.Now().Unix()
	old_osd_perf = osd_perf_map[curr_osd_perf.osd_id]
	if (old_osd_perf.time > 0 && curr_osd_perf.osd_id == old_osd_perf.osd_id) {
		dump_osd_throttle_perf(&old_osd_perf, &curr_osd_perf, curr_osd_perf.time)
	} 
	
	osd_perf_map[curr_osd_perf.osd_id] = curr_osd_perf
}

func get_process_throttle_perf() {
	var fi []os.FileInfo
	var err error
	
	fi, err = ioutil.ReadDir("/var/run/ceph/")
	if err != nil {
		return
	}
	
	for i := range fi {
		if fi[i].IsDir() == false {
			if (strings.Contains(fi[i].Name(), "ceph-osd")) {
				get_process_osd_throttle_perf(fi[i].Name())
			}
		}
	}
	
}

func main() {
	
	//osd_throttle_perf_array = make([]osd_throttle_perf_extend, 0)
	osd_perf_map =  make(map[int]osd_throttle_perf_extend, 20)
	//fmt.Println("Hello World!")
	//s := "ceph-osd.1000.asok"
	//var id int
	//fmt.Sscanf(s, "ceph-osd.%d.asok", &id)
	//fmt.Println(id)
	
	ticker := time.NewTicker(time.Second * time.Duration(sleepTime))

	go func() {
		for t := range ticker.C {
			if debug {
				fmt.Println("DEBUG", time.Now(), " - ", t)
			}
			get_process_throttle_perf()
		}
	}()
	// run for a year - as collectd will restart it
	time.Sleep(time.Second * 86400 * 365 * 100)
	ticker.Stop()
	fmt.Println("Ticker stopped")
}
