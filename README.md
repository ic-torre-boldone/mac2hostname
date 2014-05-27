mac2hostname
============

A really simple web app which translates MAC addresses into unique hostnames.

This app has been created to solve a really stupid problem: all the machines
provisioned via kickstart had the same hostname. This caused a bit of confusion
when they got registered against our salt server.

The purpose of this app is simple: given a MAC address the app will return a
unique hostname.

All the generated hostnames are stored inside of a sqlite3 database.
The hostname is built starting from a base string which can even be specified
when invoking the remote API.


Build
=====

`mac2hostname` is written in Go. Ensure a go compiler is installed:

```
go get
go build
```

This will produce a single statically linked binary which can be copied to the
final machine (given the build host architecture matches with the one of the
final server).

Usage
=====

The web app can be started by doing: `./mac2hostname [options]`.

More details can be obtaining by running `mac2hostname --help`.


API
===

Right now there's just a single API:

```
GET /mac2hostname
```

The API takes the following parameters:

  * **mac:** mac address of the machine. This parameter is mandatory.
     `_` characters are automatically replaced with `:` symbols.
     **Note well:** right now no validation is made against the specified MAC
     address.
  * **hostname_base:** the string used to compose the final hostname. This param
    is optional.

The API returns the final hostname to be used by the client.

Example of usage with curl:

```
curl http://localhost:3000/mac2hostname\?mac\=4D_19_0E_E4_9C_EC\&hostname_base\=lab
```
