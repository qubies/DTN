#!/bin/sh
function send {
    tmux send -t display "$1"
    tmux send-keys -t display Enter
}

function wait_u {
    read -p "Press Enter to continue"
}

function end {
    tmux kill-session -t display
}

function list {
    send "./client -l"
}

function splitLine {
    echo "--------------------------------------------------------------------------------"
}

function displyFile {
    splitLine
    head -1 "$1"
    splitLine
}


function safeFile {
    splitLine
    xxd -l 60 "$1"
    splitLine
}

function dynamicTrue {
    echo "Turning on dynamic hashing"
    send "sed -i -e 's/DYNAMIC=false/DYNAMIC=true/g' .env"
}


function dynamicFalse {
    echo "Turning off dynamic hashing"
    send "sed -i -e 's/DYNAMIC=true/DYNAMIC=false/g' .env"
}

#refresh
make clean
make

#setup the server
scp server cybera:~/
tmux new-session -d -s display
send "ssh cybera"
send "rm -rf ~/.DTN"
send "./server"
tmux split-window -h
sleep 2
dynamicFalse
# attach to the session in new tab
thisTab=$( guake -g )
guake -n " " -r "Display" -e "tmux a -t display"
guake --select-tab="$thisTab" 2>/dev/null

# send some commands and wait.
splitLine
send 'echo "Uploading a file with a single character"'
echo "The file:"
displyFile testFiles/testfile4
echo "Sending..."
send "./client -u testFiles/testfile4 && ./client -l"
wait_u

splitLine
echo "Sending the same file again...."
echo "Notice that the file does not send a second time (Cache hit %):"
send "./client -u testFiles/testfile4 && ./client -l"
wait_u

splitLine
echo "Lets upload another couple files"
echo "First one with some repetition:"
send "./client -u testFiles/testfile3 && ./client -l"
displyFile testFiles/testfile3
splitLine
wait_u

splitLine
echo "this one is bigger, and is all random noise"
safeFile testFiles/testfile2
send "./client -u testFiles/testfile2 && ./client -l"
wait_u

splitLine
echo "now if we upload it again, it will be lightning fast because the hashign is quick, and all of the files are cached"
send "./client -u testFiles/testfile2 && ./client -l"
wait_u
splitLine

echo "The problem is that if I insert a single thing at the start of the file... say 'welcome to the demo!', the whole file will need to be reuploaded because the breakpoints are inconsistent. "
sed -i '1s/^/welcome to the demo!\n/' testFiles/testfile2
safeFile testFiles/testfile2
send "./client -u testFiles/testfile2 && ./client -l"
wait_u

echo "Now if we turn on dynmaic hasing, the blocks will be split according to a fingerprint on a rolling hash if they do not exceed a maximum length"
dynamicTrue
echo "the block size will now change as the fingerprint markers are encountered"
echo "lets remove the first line, to reset the file to the original... "
sed -i -e "1d" testFiles/testfile2
safeFile testFiles/testfile2
echo "So if we resend the file again, our hitrate should be near 0% the first time..."
send "./client -u testFiles/testfile2 && ./client -l"
wait_u

echo "but the cool thing here is that now, when we insert to the start of the file, we maintain a high cache hitrate because the block segments get re-aligned by the rolling hash"
echo "so we add the line back into the file:"
sed -i '1s/^/welcome to the demo!\n/' testFiles/testfile2
safeFile testFiles/testfile2
echo "and we send it again, it is slower, but notice the hitrate."
send "./client -u testFiles/testfile2 && ./client -l"
wait_u
sed -i -e "1d" testFiles/testfile2

splitLine
echo "the server also has remove functionality, and it automatically cleans and removes unreferenced blocks to best effort."
echo "a full clean of partial references is completed on startup, but because of concurrency issues, can only be completed during downtime."
list
send "./client -r testfile3 && ./client -l"
wait_u

echo "we can also, ofcourse, download the files that we have uploaded"
send "./client -d testfile4"
splitLine
echo "The file has a .rebuilt extension for demonstration purposes, so that we can diff it with its original:"
diff testfile4.rebuilt testFiles/testfile4
wait_u
echo "Notice that the design is not dependant on the dynamic block encoding (as testfile4 was uploaded with the original fixed blocks)"
echo "if we upload and download again, note that the file will be changed"
send "./client -u testFiles/testfile4 && ./client -l"
echo "and we delete our local cache so the file actually collects from the server:"
wait_u
rm -rf ~/.DTN
send "./client -d testfile4 && diff testfile4.rebuilt testFiles/testfile4"
wait_u
echo "and we can also re download the testfile2, larger file for a good diff:"
send "./client -d testfile2 && diff testfile2.rebuilt testFiles/testfile2"
wait_u
end
