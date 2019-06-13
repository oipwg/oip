#!/bin/bash
FLOCORE_DATADIR=<ADD DIRECTORY>/FLOCore

# Cleanup the environment for a fresh test
## Stop FLOCore if it is running
if pgrep -f flod; 
then
  echo 'Stopping FLOCore to remove regtest files';
  kill -2 $(pgrep -f flod);
  sleep 5
fi

## Remove the FLOCore regtest blockchain
echo "Cleaning FLOCore regtest directory"
mv $FLOCORE_DATADIR/flo.conf ./flo.conf
rm -rf $FLOCORE_DATADIR/regtest
mv ./flo.conf $FLOCORE_DATADIR/flo.conf

# Startup new environment
## Stop the docker container and rebuild
echo "Building & Running new OIP docker container"
./start-dev.sh
echo "Starting FLOCore"
flod -datadir=$FLOCORE_DATADIR -daemon

echo "Waiting for OIP daemon to come online"
sleep 75

flo-cli -conf="$FLOCORE_DATADIR/flo.conf" importprivkey "cVrcv3NnTJEfQ3ZpT7yUoSPgMsrngsMWZ7uBBBjR3vWgi6nGbUKr"

# Run Reorg Commands
mine_block () {
  # dumpprivkey oMGVCJ68Q54woRwqq8uVcSM1x7CCxiLnpe
  # cVrcv3NnTJEfQ3ZpT7yUoSPgMsrngsMWZ7uBBBjR3vWgi6nGbUKr
  BLOCK_HASH=$(flo-cli -conf="$FLOCORE_DATADIR/flo.conf" generatetoaddress 1 "oMGVCJ68Q54woRwqq8uVcSM1x7CCxiLnpe" | jq --raw-output '.[0]')
  docker exec oip floctl -C /data/flod/floctl.conf submitblock $(flo-cli -conf="$FLOCORE_DATADIR/flo.conf" getblock $BLOCK_HASH 0)
  echo "Mined Block #$BLOCK_NUMBER: $BLOCK_HASH"
}

for i in {1..200}
do
  BLOCK_NUMBER=$i
  mine_block

  if [ $i == 101 ]; then
    REORG_BLOCK_HASH="$BLOCK_HASH"
    echo "Reorg Block Hash: $REORG_BLOCK_HASH"
  fi
done

flo-cli -conf="$FLOCORE_DATADIR/flo.conf" invalidateblock "$REORG_BLOCK_HASH"
echo "Invalidated Block #101, causing blocks 101-150 to become invalidated"

sleep 1

echo "Now mining reorged blocks starting with #101"
for i in {1..1000}
do
  BLOCK_NUMBER="$((i + 100))"
  mine_block
done