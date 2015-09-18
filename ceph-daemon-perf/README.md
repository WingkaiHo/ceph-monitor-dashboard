This plugin is use for dump ceph osd daemon perf for all throttle of osd daemon. It is can be execute by collectd EXEC plugin.

It will dump perf for follow throttle

osd_client_messenger
osd_dispatch_client
osd_dispatch_cluster
osd_pg
osd_pg_client_w (client write)
osd_pg_client_r (client read)
osd_pg_client_rw (client upload pg)
osd_pg_cluster_w 
filestore_wb (file store write back)
filestore_journal
filestore_op_queue
filestore_leveldb
filestore_commit

attribute

ops/s
in_bytes/s
out_bytes/s
latency

Install

1. This plugin must use root to run. Because the visit the  so you must fix the collect exec plugin can run plugin by root.
collectd-5.5.0/src/exec.c

comment follow code
424 /*
425   if (uid == 0)
426   {
427     ERROR ("exec plugin: Cowardly refusing to exec program as root.");
428     return (-1);
429   }
430 */

rebuild the exec so file.

2. Copy the collectd exec plugin and ceph-deamon-perf plugin to directory /lib64/collectd/

3. Configure the file in /etc/collectd.conf

1) Enable the exec plugin
   LoadPlugin exec

2) configure to start ceph-deamon-perf plugin
   <Plugin exec>
    Exec "root" "/usr/lib64/collectd/ceph-deamon-perf"
  </Plugin>

3) restart collectd
  service collectd resart

