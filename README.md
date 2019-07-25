### imgsrvr
temporary git for my web/image server app thing.. yeah




This should work on any version of go but please try latest, if latest doesnt work make a issue.


#To configure, do something like this:

1. Setup nginx to listen on port 80 for /app/ (or whatever its set to in a config) and fastcgi_pass all of it to 9001. Apache should work but has not been tested and is not reccomended, if you want to test it go ahead.
1. git clone, edit the script/config file with your domain name and such
1. Make a mysql DB with name `ImgSrvr` and command `create table files ( hash char(6), user VARCHAR(255), filename VARCHAR(255), ip VARCHAR(45));`, `CREATE TABLE users ( hash varchar(1000), salt char(40), user varchar(100));` and `CREATE TABLE sessions ( token char(129), expiration varchar(100), user varchar(255));` on user a user, it can be root but thats insecure but its not my issue. Just please make sure your user has access to ONLY the ImgSrvr db.
1. Run, if you encounter errors, google it if its a issue with a config or look at it if its a issue about my script, then add it as a issue.
1. Dont be evil.


#How to modular
1. Fork the repo, it'll make your life easier when I update the software
1. add your software as a go package under server/selector/modules in its OWN folder. This will allow you to copy and paste git repos. You may need to make that directory.
1. add your software to the case statement in server/selector/seletor.go with the same format as the others

   Copyright 2019 gidoBOSSftw5731



