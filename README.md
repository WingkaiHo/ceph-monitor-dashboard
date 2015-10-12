#Install enviroment 

This document is base for centos7.1, collected-5.5.0, grafana-2.02-1, graphite-web-0.9.12-8 


##1. Install graphite-web and set configure

   The graphite-web only need to install into one machine of ceph cluster.

  1) Install graphite-web and mysql backend:
    
     ...
     #yum install http://yum.puppetlabs.com/puppetlabs-release-el-7.noarch.rpm 
     #yum install graphite-web mariadb-server.x86_64  MySQL-python 
     ...

  2) Enable the mysql start when system start:
     
     ...
     #systemctl enable mariadb.service

     #systemctl  mariadb start
     ...

  3) Setting default mysql password:
     
     ...
     #mysql_secure_installation
     ...

  4) Create the "graphite" database and set user which can visit this database:
     
     ...
     #mysql -e "CREATE DATABASE graphite;" -u root -p

     #mysql -e "GRANT ALL PRIVILEGES ON graphite.* TO 'graphite'@'localhost'IDENTIFIED BY 'graphitePW01Vxzsigavms';" -u root -p

     #mysql -e 'FLUSH PRIVILEGES;' -u root -p
     ...

  5) Fix the graphite web configure file:
     vi /etc/graphite-web/local_settings.py
     
      DATABASES = { 
      'default': {
      'NAME': 'graphite',
      'ENGINE': 'django.db.backends.mysql',
      'USER': 'graphite',
      'PASSWORD': 'graphitePW01Vxzsigavms',
      }
     }
     
  6) Init the "graphite" database:
     
     ...
     #/usr/lib/python2.7/site-packages/graphite/manage.py syncdb
     ...

  7) Install Carbon and Whisper:
     
     ...
     #yum install python-carbon python-whisper
     ...

  8) Enable carbon service start when system start:
     
     ...
     #systemctl enable carbon-cache.service

     #systemctl start  carbon-cache.service
     ...

  9) Fix the /etc/httpd/conf.d/graphite-web.conf to solve the bug AH01630: client denied by server configuration in appach. You can copy the file to overwrite /etc/httpd/conf.d/graphite-web.conf:
     
     ...
     #cp ./graphite-web/graphite-web.conf /etc/httpd/conf.d/
     ...

  10) restart httpd service:
      
      ...
      #systemctl start httpd
 
      #systemctl enable httpd.service
      ...

  11) graphite data store in the directory /var/lib/carbon/whisper/

  12) Configure the storage schema of carbon cache for graphite:
      
      ...
      vi /etc/carbon/storage-schemas.conf

      [collectd]
      pattern = ^collectd\.

      retentions = 10s:1d,1m:7d,10m:1y
      ...

  13) Test if graphite_web deploy successful. Please use firfox or chrome:
      
      ...
      http://ip:8080
       
      If you can vist the web, your deploy successful. Good luck!!!
      ...

##2. Ceph monitor plugin

     This project have two ceph monitor plugin. First is use to get the perf of whole ceph cluster and every host cpu and disk information. The second use to get the perf of all osd deamon. Two plugin will be install in all the osd and monitor machine. Two plugin design as collectd exec plugin. It will be run by collectd deamon.

     1) File first plugin and code in directory ceph-cluser-perf/

     2) Second plugin in directory directory in ceph-daemon-perf/

     3) You can compile yourself or get brinary in thease directory.

##3. Collectd install and configure.

    Collectd is used to collect the information of the machine, it will be install at all the ceph cluster machine.
    
   1)Install collectd 
     
     ...
     #yum install collectd
  
   2) Copy the two plugin to /usr/lib64/collectd/
  
   3)Fix the cofigure file /etc/collectd.conf
     
     vim /etc/collectd.conf

     ...
     #Enable follow plugin
     LoadPlugin exec
     LoadPlugin interface
	 LoadPlugin load
     LoadPlugin memory
     LoadPlugin write_graphite
     #UnEnable follow plugin
     LoadPlugin cpu
 
     
     <Plugin exec>
      Exec "root" "/usr/lib64/collectd/ceph-cluser-perf"
      Exec "root" "/usr/lib64/collectd/ceph-daemon-perf"
    </Plugin>

    <Plugin write_graphite>
    <Node "graphing">
        #The ip of graphite-web
        Host "192.168.0.1"
        Port "2003"
        Protocol "tcp"
        LogSendErrors true
        Prefix "collectd."
        Postfix ""
        StoreRates true
        AlwaysAppendDS false
        EscapeCharacter "_"
    </Node>
   </Plugin>

##4. Grafana install and configure. 

1)Install grafana
  
  ...
  #yum install https://grafanarel.s3.amazonaws.com/builds/grafana-2.1.3-1.x86_64.rpm

2) Configure the port of grafana

   ...
   vim /etc/grafana/grafana.ini and fix:

   http_port=xxxx

3) Enable grafana service start when system start: 

   ...
   #systemctl enable grafana-server.service
   #service grafana-server start

4) Configure the data for grafana

   ...
   I.  Use the firefox/chrome visit the web of grafana
       
       http://grafana_web_ip:http_port

   II. Default username: admin  password: admin
      
   III.Click "Data sources" left top.

   IV. Select "Add New"

   V. The follow setting or reference PNG(./grafana/SettingDataSource.png)

      Data Source：
      Name： ceph-mon    Default： true
      Type： Graphite

      Http settings：
      Url：http://graphite_ip_addr:port # port可以通过/etc/httpd/conf.d/graphite-web.conf获取
      Acces：proxy

5) Import the grafana dashborad:

   ...
   I.  Click the manu "Dashboard"
   II. Click th button "Choose file" reference PNG (grafana/ChooseDashboardFile.png)
   III. json in directory "./grafana/json_page/"
  
   Base dashboard  (data recv by plugin ceph-cluser-perf, and interface plugin of collectd)
   Ceph_Cluster_Home
   Ceph_OSD_Information
   Ceph_Performance
   Ceph_Pool_Information
   Host_Disk
   Host_Load_CPU_Memory
   Host_Network

   Advance dashbord (data recv by plugin ceph-daemon-perf)
   Ceph_OSDs_throttles_IOPS
   Ceph_OSD_Throttle_Information
   
