# imgsrvr
temporary git for my web/image server app thing.. yeah




This is made to run on go 1.8.3, one day I may update it but syntax is frustrating.


To configure, do something like this:
1. Setup nginx to listen on port 80 for /app/ (or whatever its set to in a config) and fastcgi_pass all if it to 9001
1. git clone, edit the script/config file with your domain name and such
1. Make a mysql DB with name `files` and command `create table file ( hash char(6), user VARCHAR(255), filename VARCHAR(255), ip VARCHAR(45))` on user root 
1. Run, if you encounter errors, google it if its a issue with a config or look at it if its a issue about my script, the add it as a issue.
1. Dont be evil.
