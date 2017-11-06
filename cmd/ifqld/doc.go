/*
IFQLD is a basic HTTP server that exposes a sinle endpoint
for processing IFQL queries to 1 or more InfluxDB servers.
It can return data in either line protocol or the InfluxDB
1.x JSON query response format. Requests go here:

http://localhost:8080/query?q=...&verbose=true&trace=true&format=line|json

q is the IFQL query string. Format specifies what the response
format should be. JSON is the default. verbose and trace are optional
parameters that will make the server output additional log
information.
*/
package main
