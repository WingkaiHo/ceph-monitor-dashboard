This document is base for centos7.1, collected-5.5.0, grafana-2.02-1, graphite-web-0.9.12-8 


1. Install graphite-web and set configure

   The graphite-web only need to install into one machine of ceph cluster.

  1) Install graphite-web and mysql backend.
    yum install http://yum.puppetlabs.com/puppetlabs-release-el-7.noarch.rpm
  
    yum install graphite-web mariadb-server.x86_64  MySQL-python 

  2) Enable the mysql start when system start.
    systemctl enable mariadb.service

    systemctl  mariadb start

  3) Setting default mysql password
     mysql_secure_installation

  4) Create the "graphite" database and set user which can visit this database.
     mysql -e "CREATE DATABASE graphite;" -u root -p

     mysql -e "GRANT ALL PRIVILEGES ON graphite.* TO 'graphite'@'localhost'IDENTIFIED BY 'graphitePW01Vxzsigavms';" -u root -p

     mysql -e 'FLUSH PRIVILEGES;' -u root -p
  
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
  
  6) Init the "graphite" database.
    /usr/lib/python2.7/site-packages/graphite/manage.py syncdb

  7) Install Carbon and Whisper
    yum install python-carbon python-whisper

  8) Enable carbon service start when system start
    systemctl enable carbon-cache.service

    systemctl start  carbon-cache.service

  9) Fix the /etc/httpd/conf.d/graphite-web.conf to solve the bug AH01630: client denied by server configuration in appach.
     you can copy the file to overwrite /etc/httpd/conf.d/graphite-web.conf
     
     cp ./graphite-web/graphite-web.conf /etc/httpd/conf.d/

  10)restart httpd service
     
     systemctl start httpd
 
     systemctl enable httpd.service

  11) graphite data store in the directory /var/lib/carbon/whisper/

  12) Configure the storage schema of carbon cache for graphite 

      vi /etc/carbon/storage-schemas.conf

      [collectd]
      pattern = ^collectd\.

      retentions = 10s:1d,1m:7d,10m:1y

  13) Test if graphite_web deploy successful. Please use firfox or chrome:
      
      http://ip:8080
       
      If you can vist the web, your deploy successful. Good luck!!!

   

