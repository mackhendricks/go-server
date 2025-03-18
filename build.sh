set -x
rm -rf binary && mkdir binary
for GOOS in linux darwin windows; do
    for GOARCH in amd64 386; do
        mkdir -p binary/$GOOS/$GOARCH
        GOOS=$GOOS GOARCH=$GOARCH go build -buildvcs=false -o binary/$GOOS/$GOARCH/
        #if [ -f binary/$GOOS/$GOARCH/dsiprouter-cli ]; then
		#zip -r binary/$GOOS/$GOARCH/server.zip build/$GOOS/$GOARCH/server
        #	rm binary/$GOOS/$GOARCH/server
	    #fi
    done
done