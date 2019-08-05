#!/bin/bash
# Trap Sigterm and graceful shutdown.
trap '{ \
  kill $1
  if ! [ -f /data/flod/flod.pid ]; then kill "$(cat /data/flod/flod.pid)"; fi;\
  if ! [ -f /data/oipd/oipd.pid ]; then kill "$(cat /data/oipd/oipd.pid)"; fi;\
  service elasticsearch stop;
  service kibana stop;
  exit 0; \
}' SIGTERM

# Set Env Variables to Defaults if unset
if [ -z "$NETWORK" ]
then
	NETWORK="mainnet"
fi

if [ -z "$RPC_USER" ]
then
	RPC_USER="oipd"
fi

if [ -z "$RPC_PASSWORD" ]
then
	# Generate random RPC Password if unset
	RPC_PASSWORD="$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 13 ; echo '')"
fi

if [ -z "$HTTP_USER" ]
then
	HTTP_USER="oipd"
fi

if [ -z "$HTTP_PASSWORD" ]
then
	echo "HTTP_PASSWORD is a REQUIRED environment variable! Please set it and restart!"
	exit 1
fi

if [ -z "$CUSTOM_BLACKLIST_FILTER" ]
then
	CUSTOM_BLACKLIST_FILTER=""
fi

# Startup Nginx (since docker has it stopped at startup)
mkdir -p /data/nginx
service nginx start

# Setup NGINX passwords for elasticsearch and kibana
if ! [ -f /data/nginx/http.passwd ];
then
	htpasswd -b -c /data/nginx/http.passwd $HTTP_USER $HTTP_PASSWORD
fi

# Create Flod config
mkdir -p /data/flod
echo -e "\
	datadir=/data/flod \n\
	txindex=1 \n\
	addrindex=1 \n\
	listen=0.0.0.0" > /data/flod/flod.conf

## Write network settings
if [ "$NETWORK" == "testnet" ]
then
	echo testnet=1 >> /data/flod/flod.conf
else
	if [ "$NETWORK" == "regtest" ]
	then
		echo regtest=1 >> /data/flod/flod.conf
	fi
fi

## Add Seednode config for Flod if needed
if [ ! -z "$ADDNODE" ]
then
	echo addpeer="$ADDNODE" >> /data/flod/flod.conf
fi

## Add RPC settings to flo config
echo rpclisten=:8334 >> /data/flod/flod.conf
echo rpcuser="$RPC_USER" > /data/flod/floctl.conf
echo rpcpass="$RPC_PASSWORD" >> /data/flod/floctl.conf
echo rpccert=/data/flod/rpc.cert >> /data/flod/floctl.conf
cat /data/flod/floctl.conf >> /data/flod/flod.conf
echo rpckey=/data/flod/rpc.key >> /data/flod/flod.conf

if ! [ -f /data/flod/rpc.cert ];
then
	echo "Generating flod RPC certificates"
	gencerts -d /data/flod/
fi


# Startup Flod in daemon mode
echo 'Starting Flo Blockchain (bitspill/flod)'
flod --configfile=/data/flod/flod.conf &> /data/flod/flod.log &
# Store PID for later
echo $! > /data/flod/flod.pid
sleep 1

# Wait for flod to be fully synced.
floctl --configfile=/data/flod/floctl.conf getblocktemplate &> /data/flod/syncStatus.log
echo "Initial log of floctl: $(cat /data/flod/syncStatus.log) (having an error on this line is ok usually!)"
while [[ "$(cat /data/flod/syncStatus.log)" == *"downloading blocks"* ]] || [[ "$(cat /data/flod/syncStatus.log)" == *"not connected"* ]] || [[ "$(cat /data/flod/syncStatus.log)" == *"Failed"* ]] || [[ "$(cat /data/flod/syncStatus.log)" == *"refused"* ]];
do
	sleep 5
	mv /data/flod/syncStatus.log /data/flod/oldSyncStatus.log
	floctl --configfile=/data/flod/floctl.conf getblocktemplate &> /data/flod/syncStatus.log
	echo "$(floctl --configfile=/data/flod/floctl.conf getblockchaininfo | grep -Po 'blocks": \K[0-9]+')/$(floctl --configfile=/data/flod/floctl.conf getblockchaininfo | grep -Po 'headers": \K[0-9]+') blocks downloaded ($(date -d "@$(floctl --configfile=/data/flod/floctl.conf getblockchaininfo | grep -Po 'mediantime": \K[0-9]+')"))"
done

echo 'Flo Blockchain Sync Complete'

if [ -z "$ELASTIC_RAM_SIZE" ]
then
	ELASTIC_RAM_SIZE="$(grep MemTotal /proc/meminfo | awk '{print int($2 / 1024 / 4)}')m"
	echo "Setting ELASTIC_RAM_SIZE to 1/4 of available system ram: $ELASTIC_RAM_SIZE"
fi

echo "Setting ElasticSearch RAM to $ELASTIC_RAM_SIZE"
sed -i "s/^-Xms1g/-Xms$ELASTIC_RAM_SIZE/" /etc/elasticsearch/jvm.options
sed -i "s/^-Xmx1g/-Xmx$ELASTIC_RAM_SIZE/" /etc/elasticsearch/jvm.options

# Startup ElasticSearch and Kibana
echo 'Starting ElasticSearch & Kibana...'
mkdir -p /data/elasticsearch
chmod 777 /data/elasticsearch
service elasticsearch start
service kibana start
sleep 5

# Wait for ElasticSearch to startup and not have a status of Red
echo 'Waiting for ElasticSearch to startup and have a status of Yellow or Green...'
touch /data/elasticsearch/last.log
while ! [ -s /data/elasticsearch/last.log ] || [[ "$(cat /data/elasticsearch/last.log)" == *'"status" : "red"'* ]]
do
	sleep 1
	curl -s http://127.0.0.1:9201/_cluster/health?pretty=true &> /data/elasticsearch/last.log
	cat /data/elasticsearch/last.log | grep "status"
done
rm /data/elasticsearch/last.log

# Startup OIP daemon
echo 'Starting up OIP daemon'
## Link Config
mkdir -p /data/oipd
cp /oip/config.yml /data/oipd/config.yml
## Edit config to use Env vars
sed -i "s/NETWORK_SELECTION/$NETWORK/g" /data/oipd/config.yml
sed -i "s/RPC_USER/$RPC_USER/g" /data/oipd/config.yml
sed -i "s/RPC_PASS/$RPC_PASSWORD/g" /data/oipd/config.yml
sed -i "s/CUSTOM_BLACKLIST_FILTER/$CUSTOM_BLACKLIST_FILTER/g" /data/oipd/config.yml
## Startup first time
env LOG=* oipd > /data/oipd/latest.log &
# Store PID for later
echo $! > /data/oipd/oipd.pid

## Wait five minutes (create elasticsearch indexes), then restart after creating kibana indicies
timeout 5m tail -n 1000 -f /data/oipd/latest.log
kill $(cat /data/oipd/oipd.pid)

## Create Kibana Indexies & set default. Set to silent to stop log spam, but you might want to unsilence them if you are having issues...
curl -s -f -XPOST -H 'Content-Type: application/json' -H 'kbn-xsrf: anything' 'http://localhost:5602/api/saved_objects/index-pattern/*blocks' -d '{"attributes":{"title":"*blocks","timeFieldName":"block.time"}}'
curl -s -f -XPOST -H 'Content-Type: application/json' -H 'kbn-xsrf: anything' 'http://localhost:5602/api/saved_objects/index-pattern/*transactions' -d '{"attributes":{"title":"*transactions","timeFieldName":"tx.time"}}'
curl -s -f -XPOST -H 'Content-Type: application/json' -H 'kbn-xsrf: anything' 'http://localhost:5602/api/saved_objects/index-pattern/*oip-multipart*' -d '{"attributes":{"title":"*oip-multipart*","timeFieldName":"meta.time"}}'
curl -s -f -XPOST -H 'Content-Type: application/json' -H 'kbn-xsrf: anything' 'http://localhost:5602/api/saved_objects/index-pattern/*publisher' -d '{"attributes":{"title":"*publisher","timeFieldName":""}}'
curl -s -f -XPOST -H 'Content-Type: application/json' -H 'kbn-xsrf: anything' 'http://localhost:5602/api/saved_objects/index-pattern/*artifact' -d '{"attributes":{"title":"*artifact","timeFieldName":"meta.time"}}'
curl -s -f -XPOST -H 'Content-Type: application/json' -H 'kbn-xsrf: anything' 'http://localhost:5602/api/saved_objects/index-pattern/*historian*' -d '{"attributes":{"title":"*historian*","timeFieldName":"meta.time"}}'
curl -s -f -XPOST -H 'Content-Type: application/json' -H 'kbn-xsrf: anything' 'http://localhost:5602/api/saved_objects/index-pattern/*edit' -d '{"attributes":{"title":"*edit","timeFieldName":"meta.time"}}'
curl -s -f -XPOST -H 'Content-Type: application/json' -H 'kbn-xsrf: anything' 'http://localhost:5602/api/kibana/settings/defaultIndex' -d '{"value": "*artifact"}'

# Final startup of oipd
env LOG=* oipd