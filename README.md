### eventual design
- service directs requests to assigned proxy node
- node handles buckets based on capacity
- service-to-node and node-to-target uses REST
- node-to-node communicates uses RPC
#### with central node
- central node handles distribution of buckets to nodes
- when request to a node does not match any of bucket handled by it, node will:
	- check if it has info of assigned node, if not
	- forward the request to central node and requests for bucket handler info
	- if bucket is assigned, central node will forward the request to it
	- if bucket is not assigned, central node will look for one with capacity/spawn a new one
	- central node will then disseminate this info to all alive nodes
- in case of central node failures ...
- service -> assigned proxy node -> (central node) -> bucket handler node -> ...
- ...
#### without central node
- when request to a node does not match any of bucket handled by it, node will:
	- check if it has info of assigned node, if not
	- communicate with other alive nodes to find existing handler 
	- other alive nodes would respond with buckets handled (capacity can be inferred/stated)
	- if bucket is assigned, forward the request to handler node
	- if bucket is not assigned, receiver node will assign to one with capacity/propose spawning a new one
- service -> assigned node -> bucket-handler node -> ...
- ...

### TODOs
- [x] single node proxy/rate limiter
- [x] global rate limiter
- [x] logs
- [ ] query stale requests -> currently from log
- [ ] metrics -> expose to collector?
- [ ] kube setup
- [ ] handle herd requests
- [ ] communication and consensus in distributed fashion
- [ ] custom rate-limited requests treatment (throttle, prioritized retry/stale)
- [ ] redis storage
- [ ] bucket-to-node assignment/distribution
- [ ] node create/failure handling
- [ ] actual integration
- [ ] benchmark

### references
- [google](https://cloud.google.com/solutions/rate-limiting-strategies-techniques)
- [line](https://engineering.linecorp.com/en/blog/high-throughput-distributed-rate-limiter)
- [Mailgun's Gubernator](https://github.com/mailgun/gubernator)
- [Discord API library in Go](https://github.com/bwmarrin/discordgo) -> some copying
- [Discord API library in Rust](https://github.com/serenity-rs/serenity)
- ...
