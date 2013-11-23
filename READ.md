Register a name with SkyDNS
===========================

Changes from original README.md

1.  Allow for (HTTP) callbacks/notifies when a name changes;
2.  Make the registration "simpler";
3.  Do away with host names, and directly use IP addresses. Intermediate
    hostnames will be generated and will be resolved in the DNS reply
    within the additional section. These hostnames will be named
    UUID.skydns.local.
4.  Killed version, don't now if this is wise, but if I were to use
    version, I would be interested in < 1.0 and stuff like that, and you
    can not do this with the DNS.

Each client generates a UUID which they use to update names, this is
just to ensure that clients registering the same name (if they operate
as a cluster) to be distinguished.

Each name is setup like: <name>.<environment>.<region>.skydns.local.
Each registered names should be fully qualified.

Names cannot contain dots. Each part of the name roughly means the
following:

-   name: a name of the service;
-   region: where does it live;
-   environment: can be something like: production, test, canary, etc.;
-   and then .skydns.local as the local domain name.

Registering names can be done by a PUT request with a small bit of json
data, as follows:

    PUT /skydns/services/uuid/<name.region.environment.skydns.local>/ -d '{"TTL":400,"Address":192.168.1.1,"Port":6000}'

A few examples:

    // register testservice with UUID 1001 (East Region)
    curl -X PUT -L http://localhost:8080/skydns/services/1001 -d '{"Name":"testservice","Region":"east","Environment":"production","Address":"192.168.1.2","Port":80,"TTL":4000}'

    // testservice with UUID 1002 (East Region)
    curl -X PUT -L http://localhost:8080/skydns/services/1002/testservice.east.production.skydns.local/ -d '{"Address":192.168.1.3,"Port":8080,"TTL":4000}'

    // testservice with UUID 1003 (West Region)
    curl -X PUT -L http://localhost:8080/skydns/services/1003/testservice.west.production.skydns.local/ -d '{"Address":172.16.1.1,"Port":80,"TTL":4000}'

    // testservice with UUID 1004 (West Region)
    curl -X PUT -L http://localhost:8080/skydns/services/1004/testservice.west.production.skydns.local/ -d '{"Address":172.16.1.2,"Port":80,"TTL":4000}'

Callbacks
---------

When registering a callback, the service doing so registers the name it
is interested in, maybe with wildcards, and registers, the IP address,
port and TTL (how long should this callback live), when the name of
interest changes (new address, new port or it is deleted), the callback
is executed. If the remote server does answer the callback, the callback
is deleted. When the TTL of the callback is reached the callback is
deleted. Registering a callback is thus fairly similar to registering a
name:

    curl -X PUT -L http://localhost:8080/skydns/callback/testservice.east.production.skydns.local/ -d '{"Address":192.168.1.2,"Port":81,"TTL":4000}'

If an address for a name is updated and this address matches an address
set in a callback, the callback is updated with the new address.

Now we can try some of our example DNS lookups:

All services in the Production Environment
------------------------------------------

dig @localhost production.skydns.local SRV, none specified fields to the
left will be taken as wildcards.

    ;; ANSWER SECTION:
    production.skydns.local.    3979    IN  SRV 10 20 80   1001.skydns.local.   ;; made up name from uuid and skydns.local
    production.skydns.local.    3979    IN  SRV 10 20 8080 1002.skydns.local.
    production.skydns.local.    3601    IN  SRV 10 20 80   1003.skydns.local.
    production.skydns.local.    3985    IN  SRV 10 20 80   1004.skydns.local.

    ;; ADDITIONAL SECTION:
    1001.skydns.local.          3979    IN  A   192.168.1.2
    1002.skydns.local.          3979    IN  A   192.168.1.3
    1003.skydns.local.          3601    IN  A   172.16.1.1
    1004.skydns.local.          3985    IN  A   172.16.1.2

All TestService instances in Production Environment
---------------------------------------------------

dig @localhost testservice.*.production.skydns.local SRV, we misuse DNS
wildcards here (actually those aren't actually DNS wildcards):

    ;; QUESTION SECTION:
    ;testservice.*.production.skydns.local.     IN  SRV

    ;; ANSWER SECTION:
    testservice.east.production.skydns.local.    3979    IN SRV 10 20 80   1001.skydns.local.
    testservice.east.production.skydns.local.    3979    IN SRV 10 20 8080 1002.skydns.local.
    testservice.west.production.skydns.local.    3601    IN SRV 10 20 80   1003.skydns.local.
    testservice.west.production.skydns.local.    3985    IN SRV 10 20 80   1004.skydns.local.

    ;; ADDITIONAL SECTION:
    1001.skydns.local.          3979    IN  A   192.168.1.2
    1002.skydns.local.          3979    IN  A   192.168.1.3
    1003.skydns.local.          3601    IN  A   172.16.1.1
    1004.skydns.local.          3985    IN  A   172.16.1.2

All TestService Instances within the East region
------------------------------------------------

dig @localhost testservice.east.production.skydns.local SRV

This is where we've changed things up a bit. We've supplied an explicit
region that we're looking for we get that as the highest priority within
the SRV record, with the weight being distributed evenly. Then all of
our West instances still show up for fail-over, but with a lower
priority.

    ;; QUESTION SECTION:
    ;testservice.east.production.skydns.local. IN   SRV

    ;; ANSWER SECTION:
    testservice.east.production.skydns.local.    3979    IN SRV 10 20 80   1001.skydns.local.
    testservice.east.production.skydns.local.    3979    IN SRV 10 20 8080 1002.skydns.local.
    testservice.west.production.skydns.local.    3601    IN SRV 20 20 80   1003.skydns.local.
    testservice.west.production.skydns.local.    3985    IN SRV 20 20 80   1004.skydns.local.

    ;; ADDITIONAL SECTION:
    1001.skydns.local.          3979    IN  A   192.168.1.2
    1002.skydns.local.          3979    IN  A   192.168.1.3
    1003.skydns.local.          3601    IN  A   172.16.1.1
    1004.skydns.local.          3985    IN  A   172.16.1.2
