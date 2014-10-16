git-backend
===========

git backend using git2go

# Install

This project needs libgit2, which is written in C so we need to build that as well. In order to build libgit2, you need `cmake`, `pkg-config` and a C compiler. You will also need the development packages for OpenSSL and LibSSH2 if you want to use HTTPS and SSH respectively.

    make prepare
    make test
