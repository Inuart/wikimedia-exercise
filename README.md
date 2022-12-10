# Wikimedia Exercise: Short Description API

This API takes the name of a person and returns a short description of them, similar to how you would see on the right-hand side of a Google search result. The description is extracted from the person's English Wikipedia page using the MediaWiki API.

It also integrates an in-memory LRU cache that can be configured using the `CACHE_SIZE` and `CACHED_RESULT_TTL` environment variables.

The package is structured in a way that permits it's use as a client package or as a standalone server through http. The provided `main.go` can be seen as an example of creating a `shortdescription` client and then binding it to an http server.

The repo consists of only one `shortdescription` package mostly because I don't feel like any of its features need to be isolated into its own package at this (early) stage. If the code would need to keep growing I would start separating it into different packages according to their responsibilities inside an `internal` folder to prevent them from being imported by a user of the root package.

## Input and Output Schema
I've implemented the only REST API endpoint in the root path.

To use the API, send a GET request to the `/` endpoint, passing the person's name as a query parameter. For example, if you want to get a description of Yoshua Bengio, you would send a request like this:

> GET http://localhost:8080?person=Yoshua+Bengio

The API will respond with a JSON object containing the person's description. For example, if you make the request above, you will receive the following response:

```json
{
    "person": "Yoshua Bengio",
    "description": "Canadian computer scientist"
}
```

Where `person` is the name of the person to get a short description for and `description` is the short description of the person, as extracted from their English Wikipedia page.



## Running the API Server Locally
You'll need `Go 1.19` installed. Having `Make` makes things easier to run.

From inside the `wikimedia-exercise` folder, run `make run-demo` to execute the API server on `localhost:8080`, a `CACHE_SIZE` of 10 entries and a Time To Live for cache results of 1h.

To build a binary from source, run `make build`. The resulting binary reads the following environment variables:

- `ADDR`: A `host:port` format string. It will choose a free port by default.
- `CONTACT_INFO`: **Required**. You need to provide an your contact info. See https://meta.wikimedia.org/wiki/User-Agent_policy.
- `CACHE_SIZE`: The maximum amount of results the cache should hold.
- `CACHED_RESULT_TTL`: The Time To Live for each cached result before it is considered outdated.


## Limitations and theoretical future work

### Error handling
Given more time, I would have implemented a better error handling API by providing the error message inside the JSON response, like this:

```json
{
    "person": "Yoshua Bengio",
    "error": "Short description not found"
}
```

In that case, `description` would not be present and `error` would contain a description of the error that occurred without the frontend having to deal with API errors having a different content type.

### Just one description at a time
The wikimedia API is capable of returning more than one description per query. However the exercise does not mention this needs to be supported, so I'm keeping things simple.

A query like

> http://localhost:8080?person=France|Yoshua+Bengio

will return only the description of the first parameter and ignore the rest:

```json
{
    "person": "France",
    "description": "Country in Western Europe"
}
```

### Just people..?
The exercise clearly states that the API should provide a short description of a **person**. Well, I haven't found a way to filter for persons in the query. As a consequence `France` will happily return a result even if it's not a person.

### Not using an existing library
I'm aware of the existence of some libraries written in Go that I could have used to interact with the MediaWiki API. However, I had the feeling that using them might defeat the purpose of the exercise a bit.

### Not using a retry strategy
Currently there's no retry/exponential backoff algorithm implemented. Time constraints let to it being left out and also the fact that the package is structured so it can be used as a client as well as an http server handler. In the client use case, the client user will be able to provide their own retry algorithm or strategy.

### TLS?
Right now, `main.go` creates a server that does not use TLS certificates. This was left out on purpose because adding support for TLS in Go only involves changing the call to `http.Serve()` to `http.ServeTLS()` and then providing the necessary certificate file and key. Since this is only an exercise I figured it would be mostly a distraction.


# Exercise questions

## What if there's no short description?

> How will the schema take into consideration if the person being provided is not on English wikipedia? What if "short description" in content is missing?

If the person is not on English Wikipedia or their page does not contain a short description, the response will be a `404 Not Found` with a `text/plain` message:

`short description not found`

See [Future work on error handling](#error-handling).

## Keeping the API Service Highly Available and Reliable

> Consider this hypothetical scenario: Your API is going to be deployed and made available to the public for use. What things could you do to keep this API service highly available and reliable? (Think of as many issues as come to your mind and propose your potential solutions. No code is required for this)

Right now the API is provided by a binary in someone's computer or server, so what I would personally do is upload it to some cloud-based web application hosting service.

I have experience using `Google App Engine` so that's what I would use.

> Google App Engine is a cloud computing platform as a service for developing and hosting web applications in Google-managed data centers. Applications are sandboxed and run across multiple servers.

I would deploy it in there mostly as is, but also making it take advantage of the multi-instance cache.

*However*, since this is a hypothetical scenario, I can imagine I have multiple data centers at my disposal.

So let's start by estimating the initial SLO requirements taking into account that Google and other search engines might want to call our API every time they want to show a search result:

- Queries should complete in less than 1 second.
- The data should be less than 1 day old.

According to Google, they handle 99.000 searches per second while Wikipedia gets 2.500 views per second. That could potentially be a problem for the Wikipedia servers so caching results will become essential.

Lets assume a person's name is a 100 bytes string, ignoring the url it comes wrapped into. Let's start by looking at the necessary network speed:

>  (10^5 searches/s) * (2 * 10^2 bytes/search) = 20 MB/s

Seems manageable even in a single machine if the storage can write fast enough. Now let's see how many data that is in a day:

> (20 MB/s) * (8.64 * 10^4 s/day) = 17.28 * 10^11 = 1.73 TB/day

Our system could potentially try to cache 1.73 TB every day. Unfortunately I could not find data on how often Google is able to serve their search results from a cache. By the nature of the internet I'll assume either Google or our system are able to cache at least 50% of their search results. That leaves our system with 865 GB/day.

*However*, according to https://en.wikipedia.org/wiki/Wikipedia:Size_of_Wikipedia, there are currently `6.584.955` articles on wikipedia in English. That limits the amount of different entries significantly. Let's say each short description can be as long as 500 bytes.

> (7 * 10^6 descriptions) * (5 * 10^2 bytes/description) = 3.5 GB

My current implementation only makes use of an in-memory LRU cache and, surprisingly, it looks like we can manage to make it work only using RAM.

At this point I'll focus on what can go down. We obviously cannot use a single server because then a single restart will bring down our API.

We need a load balancer that directly distributes the load over as many physical servers as we want, or mount a Kubernetes cluster in our little cluster to automate maintenance and deployment tasks.

I don't think we need a data replication strategy because what we are storing is actually just a cache and unavailable data will just be fetched from the WikiMedia API.

Quite the opposite actually. We are replicating a lot of data inside every server in their own cache and that is costly for our servers and for Wikipedia.

To avoid hitting Wikipedia directly as much as possible we can use a shared cache. Something like a Redis could work as a slower cache for all server instances. Every time an instance has to reach Wikipedia for new content, it will also update the shared cache so that other instances can get the updated data.

Another possibility would be to shard the instances from the data center so that each one deals with specific queries by calculating a hash number from the requested entry "modulo" the number of instances in the data center. This work would be done by the load balancer in each data center.

In either case, we can set up this system across all our data centers and, since they can work independently of each other, we know they will make at most 7 * 10^6 api calls/h (168 * 10^6 api calls/day) each using the WikiMedia API. If one datacenter goes down, the others are still able to serve requests.

Now, if I could **use the WikiMedia API to subscribe to receive changes made to any short description** in real time, the potential efficiency improvement could be enormous, as I don't expect them to change that often, or at all.
