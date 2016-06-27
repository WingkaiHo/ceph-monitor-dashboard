docker run -d \
	--name graphite \
	-p 8080:80 \
	-p 2003:2003 \
	-v /root/hyj/graphite/etc/nginx/.htpasswd:/etc/nginx/.htpasswd \
	-v /root/hyj/graphite/storage/whisper:/opt/graphite/storage/whisper \
	-v /root/hyj/graphite/storage/whisper:/opt/graphite/storage/whisper \
	sitespeedio/graphite
