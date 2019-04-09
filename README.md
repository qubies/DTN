## Build
running make inside $GOPATH will collect dependancies, and build the project. 

## Run
Modify the .env file to configure both the client and the server. 

Client options:
```
Usage: DTNclient [-l] [-d value] [-h value] [--help] [-r value] [-u value] [parameters ...]
 -d, --download=value
             The file you wish to download
 -h, --HASH Value=value
             The specific hash value you wish to target (Remove and
             Download)
     --help  Help
 -l, --list  Get a list of the files on the server by name
 -r, --remove=value
             The file you wish to remove
 -u, --upload=value
             The file you wish to upload
```
