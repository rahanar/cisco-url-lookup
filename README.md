# cisco-url-lookup
A URL lookup service that responds whether a URL contains a malware or not

This is a simple golang HTTP server that responds only to HTTP GET requests on `/urlinfo/1/{hostname_and_port}/{original_path_and_query_string}` resource. Currently, the service uses in-memory URL database to verify whether the URL, that's a part of the HTTP GET request resource path, is malicious or not. The in-memory database is populated from a local `url-database.json` file. The service uses the `hostname` of the incoming URL to look it up in the database. The current implementation only checks whether the URL is in the database. If it is, then it would also check whether `malicious` field is set to `True` or `False`. If the URL is marked as malicious, the service would return `StatusForbidden` with HTTP code 403. If the URL is not in the database or if it is not marked as malicious, the service would return `StatusOK` with HTTP code 200.

The service can be further enhanced to do a more thorough and granular heuristic to verify whether a URL is malicious or not. For example, it can be enhanced to verify the port of the incoming URL as well as the original path.

## Requirements
To run this application locally, you would either need:

1. golang version - `1.15.6`
2. docker version - `19.03.12`

## Instructions

To run this application locally, you can simply run the following commands:

1. `make docker-build`
2. `make docker-run`

The first command will create a local docker image with a `url-lookup:latest` tag for the service. The second command will create a new container based on the image created in the previous command. The container will be exposed on port `8000`.

If you want to run the app without running it inside the container, you can simply run `make run`. This command will start-up the app and will listen on port `8000`.

To run the unit-tests, simply run `make test`.


# Thought exercise responses

* The current implementation of the service holds information about URLs in memory in a key/value paradigm. I use hostname as a key to check if an incoming URL is in the database or not and act on that information inside the service.  

  In order to support a much higher scale (that's beyond its memory capacity), I would introduce a separate data layer that the service would inquire to find information about an incoming URL. I would use a NOSQL key value data store, since this service is primarily read heavy and would benefit in high availability and would tolerate eventual consistency of new URLs being added to the data store.

  To support an infinite amount of URLs and to improve reliability of the service, the URLs would have to be sharded among a cluster of data stores. The service would have to be modified to access a correct shard to find information about an inquired URL.

  The reliability of the service would be improved from the fact that if one shard goes down, the service can still validate incoming URLs that are not affected by the shard. For improved security, the service can automatically reject requests that would be dependent on the downed shard (until it comes back).

* In order to support a number of requests that are beyond the capacity of a single system, I would horizontally scale the application. To horizontally scale this service, I would introduce a scalable Layer 7 loadbalancer that would route requests to each server that's running the service, starting with with a "Round Robin" algorithm to distribute the request to the servers. This would work well in a homegeneous environment (where all servers are running on the same hardware specification); however, if I have to account for some servers that are slower than the rest, I would use an algorithm that takes into account request latencies (such as EWMA - exponentially weighted moving average). 

  Furthermore, if I can't infinitely scale the number of servers that run the service, I would implement a server-side rate limiting mechanism that would put an upper bound on the number of requests each instance of service can handle. This would require the service's clients to implement a retry mechanism to retry failed requests.

  I would also use a platform that supports auto-scaling of servers that run the service. This would help with cases when the current fleet of servers cannot serve requests within the expected latencies range for the service. Auto-scaling would add more servers to the fleet so as to improve latencies and replace servers that are unhealthy or down.

  To support a global scale of the service, I would run separate instances of the service per region. For example, I would have a separate set of servers with their own load balancer and datastore in North America, Europe, and Asia. This would help to isolate regions, so if the service in one region is down, it won't affect the other regions.

* The current implementation of updating URLs in an in-memory "database" is very primitive. The "database" is populated from a file when the application starts up. This is neither scalable nor maintainable for a production service.

  The first thing I would do is to introduce a POST API that would allow the service to either add new URLs or update exisiting ones. The new API would allow any one (with an appropriate authorization) to add or update URLs to the data store without directly interacting with it. This would also allow the service to do required work on incoming URLs (such as sanitazation and creating required internal abstraction). Furthermore, the API can be used as a part of a pipeline that automatically updates the service's data stores whenever it is required without or minimal human intervention.

  If the service needs to support as much as 5,000 URLs per day, with updates arriving every 10 mins, I don't think there is anything else to be done. 5,000 URLs per day that are delivered across updates every 10 mins would translate into 144 updates with ~34.7 requests every update. Provided that there will be multiple servers running the service behind a loadbalancer, the requests will be distributed among those servers.

  However, if the service has to support even more write requests, I would consider deploying a separate set of servers to run the service, but have them be solely dedicated to the POST API. I would set up the layer 7 loadbalancer to distribute requests based on HTTP request path and method. This way I can separate write and read requests for the service.

  If the data store layer is the bottleneck for the write operations, I would consider introducing a message queue between the service application and the data store. The application can write to the message queue the new URLs and they would be asynchronously inserted or updated to the data store. This would potentially require another application to read from the message queue and insert to the data store. This approach would increase the complexity of running and maintaining the service in production, but would alleviate pressure on the data store during high loads.

* If I am woken up up at 3AM, I would be expected to be woken up by one (or several) alerts that are defined for the service. First, I would try to figure out where the current issue(s) might be occuring. I would check application metrics to determine whether requests are reaching it (whether the loadbalancer, DNS or networking infrastructure is the issue). If request are not reaching the service, I would focus on determining the root cause at that layer.

  If requests are reaching the service, I would next check whether the application is able to serve them and if it can do it within expected latency ranges. If there is a high error rate, then either the application is not able to talk to the data store (either because there is an issue with communicating with it or the data store is down) or there is an application bug that got triggered due to some unhandled event. Using application metrics and logs, I would narrow down where the issue might be occuring. If the issue is with the data store, I would check its metrics and application logs to figure out what the issue might be.

  If the issue is occuring due to a bug that is introduced due to a recent deployment, I would first figure out whether it is safe to revert back the deployment. Restoring the service is the highest priority, so I would focus on that first. If I can't, I would debug the issue and try to recreate the issue either locally in another non-production environment. If I can confirm the bug via unit-tests, I would fix the bug and relaese a new version of the service.

  If the issue is occuring outside the application layer, I would focus my efforts on figuring out the issue at the infrastrucurre and the platform layer. I would use metrics and logging provided for those layers to determine the root cause.

* The current implementation of the service does not output any metrics. Before deploying to a production environment, I would add support for metrics to track the number of requests, errors, and duration. I would also introduce metrics for the data store and the message queue (if there is one). Based on those metrics, I would create alerts to be triggered when the defined threshholds for the service are violated.

  I would also introduce integration tests between the application and the data store. If the service utilizes a message queue, I would also have integration tests for that. Furthermore, I would have system tests defined for all my APIs.

  Finally, I would ensure that this application has a robust CI pipeline, so all changes to the application are automatically tested and if there is a new release of the application, the pipeline would create all required artifacts.

* If I had to deploy a new version of this application, I would first ensure that all its tests are passing. I would then increment its version and verify the CI pipeline has executed all defined stages to release the new version of the application, such as running tests and linters, and creating artifacts. I would then deploy the application to a development environment first and verify the new version is working. If there is a staging environment, I would target that next. If that's successful, I would deploy the new version to production while verifying that all metrics are within expected ranges.
