#Quick start

##Linux (CentOS 6)
* go get github.com/pharmacy72/gobak
* go install github.com/pharmacy72/gobak
* copy config.json from $GOPATH/src/github.com/pharmacy72/gobak to $GOPATH/bin/
* $GOPATH/bin/gobak svc install
* configure config.json
* service gobak start

##Windows
**WARNING** golang must be i386 
* go get github.com/pharmacy72/gobak
* go install github.com/pharmacy72/gobak
* copy config.json from $GOPATH\src\github.com\pharmacy72\gobak to $GOPATH\bin\
* configure config.json
* %GOPATH%\bin\gobak.exe svc install
* in Services.msc start incremental go backup service
