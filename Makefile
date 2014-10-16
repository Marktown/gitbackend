GIT2GO = github.com/libgit2/git2go
GOPATH1 = $(firstword $(subst :, ,${GOPATH}))
prepare:
	go get -d -u ${GIT2GO}
	cd ${GOPATH1}/src/${GIT2GO} && git submodule update --init
	cd ${GOPATH1}/src/${GIT2GO} && make install

test:
	go test

ci: test
