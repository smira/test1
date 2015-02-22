## Publish-Subscribe Service

This is example service, please ignore it :)

## Running

Go is required:

    make
    test1


## Tests

With service running (on default localhost:3000), tests from
https://github.com/smira/test2 could be run:

    make test

This performs simple integration testing.

## Some notes

Simple lock covering whole service is used as all operations are fast
and non-blocking (only memory access).

Some things left out:

 * topics and users are never deleted, even if they are not used anymore
 * some kind of resource limits: currently it's easy to eat all memory simply by
   touching some resources (topics, users)
 * limits on mailbox length (how to handle users never reading their subscriptions)
 * scalability (doesn't scale beyond one instance)
 * persistence (hasn't been required)
 * authentication (?)