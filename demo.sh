#!/bin/sh
function send {
    tmux send -t display "$1"
    tmux send-keys -t display Enter
}

function sendServer {
    tmux send -t server "$1"
    tmux send-keys -t server Enter
}

function wait_u {
    read -p "Press Enter to continue"
}

function end {
    tmux kill-session -t display
    tmux kill-session -t server
    guake -s $SERVERSHELL -e 'exit'
}

function list {
    send "./DTNclient -l"
}

function splitLine {
    echo "--------------------------------------------------------------------------------"
}

function displyFile {
    splitLine
    echo -e "\t$1"
    splitLine
    head -1 "$1"
    splitLine
}


function safeFile {
    splitLine
    echo -e "\t$1"
    splitLine
    xxd -l 60 "$1"
    splitLine
}

function dynamicTrue {
    echo "Turning on dynamic hashing"
    send "sed -i -e 's/DYNAMIC=false/DYNAMIC=true/g' .env"
    send "cat .env"
}

function setBlockSize {
    echo "set blocksize to $1"
    send "sed -i -e \"s/MAXIMUM_BLOCK_SIZE=.*/MAXIMUM_BLOCK_SIZE=$1/\" .env"
}

function dynamicFalse {
    echo "Turning off dynamic hashing"
    send "sed -i -e 's/DYNAMIC=true/DYNAMIC=false/g' .env"
    send "cat .env"
}

function handleint {
    end
    exit 0
}

function sendAndList {
    splitLine
    echo $2
    safeFile "$1"
    send "./DTNclient -u $1 && ./DTNclient -l"
}

function restartServer {
    echo "----------RESTARTING SERVER, REMOVING ALL FILES----------"
    guake -s $SERVERSHELL -e 'exit'
    startServer
}

function startServer {
    echo "Removing DTN Files..."
    rm -rf ~/.DTN
    echo "Done"
    pkill DTNserver
    tmux new-session -d -s server
    guake -n " " -r "server" -e "tmux a -t server"
    SERVERSHELL=$( guake -g )
}

trap 'handleint' SIGINT

#refresh
make clean
make

#setup the server
DEMOSHELL=$( guake -g )
startServer
guake -s $DEMOSHELL
tmux new-session -d -s display
sleep 1
dynamicTrue
sendServer "./DTNserver"
sleep 1
terminator -e "tmux a -t display" &
# guake -n " " -r "Display" -e "tmux a -t display"

# send some commands and wait.
sendAndList "testFiles/testfile4" "Uploading a file with a single character"
wait_u

sendAndList "testFiles/testfile4" "Notice that when we send it again, it does not send"
wait_u

sendAndList "testFiles/modified/testfile4" "Now if we make a modification to the file, it will upload again, but the initial version will remain". 
wait_u

echo "Now we look at an image...."
wait_u
feh -F "testFiles/memeTemplate.tif"
sendAndList "testFiles/memeTemplate.tif" "And we send it... Note that the image is in UNCOMPRESSED format"
wait_u

echo "Now if we add a meme caption to the file:"
wait_u
feh -F "testFiles/modified/memeTemplate.tif"
sendAndList "testFiles/modified/memeTemplate.tif" "And we send it... "
echo "You can see that the server again, maintains both versions of the image, and that the second image has a reasonable hitrate."
wait_u

## DOWNLOAD DEMO
echo "so if we want to download the image, we get the most recent version by default, the one with the caption:"
send "./DTNclient -d memeTemplate.tif"
wait_u
feh -F "memeTemplate.tif.rebuilt"

echo "If we want the original one, we now specify the hash signature of the file we want:"
send "./DTNclient -d memeTemplate.tif -h 91d1a15119ceb2000c72c225c936b600bc377cd750be13aa715208eb44f5d581"
wait_u
feh -F "memeTemplate.tif.rebuilt"

echo "This is the end of the demo. Thanks!!!"
wait_u
end
