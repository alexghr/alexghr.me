# {{ pageTitle }}

A couple of weekends ago I had a problem: I was bored. I needed to find 
something to tinker with, so I eyed the Raspberry Pi in the corner of our flat 
and SSH'd in. 

## The Raspberry Pi
I use this Raspberry Pi to run the Unifi Controller for the single
UAP-AC-LITE access point I use to blast WiFi all over the flat. I basically set
it up once, a (long) time ago and forgot about it, which also meant no one updated
it in a long while. The Pi was still running Raspbian Jessie, which came 
installed on the official starter SD Cards at the time. I decided it was time to
upgrade to Stretch: step 1: upgrade everything that could be upgraded, step 2: 
switch over the APT sources from `jessie` to `stretch` and upgrade again.

## The failed update
So I promptly ran `sudo apt update && sudo apt dist-upgrade` and went to make a
cup of coffee. When I got back, the Pi was offline.

"Odd.. maybe it's restarting to install updates?" I thought. No Alex, this isn't 
Windows. A hard reboot later and I'm back in an SSH session... when I notice (quite 
surprisingly) that `bash` couldn't do command or file autocompletion anymore. 
The update had filled the tiny 16GB SD Card the Pi came in the box with.

"No problem", I think, "I'll just delete some of the .deb files I downloaded 
ages ago". `rm`-ing long filenames without file autocompletion is... a good way
to train one's patience. I knew the problem was bigger than first thought, when
after (successfully) deleting 100s of megabytes of files, `df -h` still showed
the partition as being full. 

`fsck` and a round of googling later: it looks like some SD cards go into
read-only mode when they fail (see this super helpful [StackExchange answer]).
The weirdest thing about it is that the SD Card's firmware still accepts writes,
it just commits them to disk.

## Moral of the story

SD Cards will fail, at some point. Write-heavy workloads (like `dist-upgrade`) 
are more likely to kill your Card.

I ended up buying a new SD Card and flashing Ubuntu 18.04 to it by following
[the official guide]. The read-only mode on the old SD Card was a life-saver in
the end as it allowed me to copy any old files I hadn't backed up (read: all of
them!).

[StackExchange answer]: https://unix.stackexchange.com/a/386886/339790
[the official guide]: https://wiki.ubuntu.com/ARM/RaspberryPi