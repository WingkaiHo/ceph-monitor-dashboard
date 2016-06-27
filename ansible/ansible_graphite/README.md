##1. Install ansible env

ansible-server# yum install ftp://195.220.108.108/linux/fedora/linux/updates/24/i386/a/ansible-2.1.0.0-1.fc24.noarch.rpm 

##2. Install grahite env

1) copy the ansbile_graphite ansible roles
   ansible-server# cp -r ansbile_graphite /etc/ansible/roles/ 

2) fix the /etc/ansible/hosts 

   add group to install grahite database

   for examle the list of host to install graphite database

   [graphite-server]
   192.168.0.10          

   ansible-server# vim  /etc/ansible/hosts

3) setting the variable 

   container_instant_path: The path to store the container instance data
   graphite_web_port: The port to visit the graphite web database

##3 Exectue the script

   ansible-server# cd /etc/ansible/roles/ansible_graphite/
   ansible-server# ansible-playbook main.yml -k

##4 Start or stop graphite web docker container

   Start: 
   graphite-server# service graphite start
   
   Stop:
   graphite-server# service graphite stop   
