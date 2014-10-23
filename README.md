git-backend
===========

git backend using [git2go](https://github.com/libgit2/git2go)

[![Build Status](https://travis-ci.org/Marktown/gitbackend.svg?branch=master)](https://travis-ci.org/Marktown/gitbackend)

# Getting Started

This project needs libgit2, which is written in C so we need to build that as well. In order to build libgit2, you need `cmake`, `pkg-config` and a C compiler. You will also need the development packages for OpenSSL and LibSSH2 if you want to use HTTPS and SSH respectively.

## Install

    make prepare
    
## Run the tests

    make test
