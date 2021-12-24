#!/usr/bin/env bash
# 
globo=0
gcp=0
#### Bloco principal
for run in $(seq 1 100);do
		result=$(host s3.aws.cloud.globo|awk '{print $NF}')
	if [ $result == "186.192.90.3" ];then
		globo=$(expr $globo + 1)
	elif [ $result == "34.149.183.254" ];then
		gcp=$(expr $gcp + 1)
	fi
done
echo GCP:$gcp GLOBO:$globo
##################
