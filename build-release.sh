#!/bin/bash
export GO111MODULE=on
sum="sha1sum"

if ! hash sha1sum 2>/dev/null; then
	if ! hash shasum 2>/dev/null; then
		echo "I can't see 'sha1sum' or 'shasum'"
		echo "Please install one of them!"
		exit
	fi
	sum="shasum"
fi

UPX=false
if hash upx 2>/dev/null; then
	UPX=true
fi

VERSION=`date -u +%Y%m%d`
LDFLAGS="-X main.VERSION=$VERSION -s -w"
GCFLAGS=""

# OSES=(linux darwin windows)
OSES=(linux darwin)
ARCHS=(amd64)
for os in ${OSES[@]}; do
	for arch in ${ARCHS[@]}; do
		suffix=""
		if [ "$os" == "windows" ]
		then
			suffix=".exe"
		fi
		env CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o c3_${os}_${arch}${suffix} 
		# if $UPX; then upx -9 c3_${os}_${arch}${suffix} ; fi
	done
done

# ARM
# ARMS=(7)
# for v in ${ARMS[@]}; do
# 	env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=$v go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o c3_linux_arm$v 
# done
# if $UPX; then upx -9 c3_linux_arm* ; fi

#MIPS32LE
# env CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o c3_linux_mipsle 
# if $UPX; then upx -9 c3_linux_mips* ; fi
