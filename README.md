<p align="center">
  <a href="https://github.com/gsh-lan/unwindia" target="blank"><img src="https://raw.githubusercontent.com/GSH-LAN/Unwindia/main/.resources/images/logo.png" height="128" alt="unwindia logo">
  <a href="https://github.com/gsh-lan/unwindia" target="blank"><img src="https://raw.githubusercontent.com/GSH-LAN/Unwindia/main/.resources/images/header.svg" height="128" alt="unwindia header" /></a>
</p>

# UNWINDIA DotLAN MatchService
> The DotLAN MatchService for [Unwindia](https://github.com/GSH-LAN/Unwindia)
---

This service integrates UNWINDIA with the tourney system of [DotLAN](https://dotlan.net). 



## Configuration

### Environment Variables


### Dynamic Configuration 
The dynamic configuration is typically loaded from Config-Service but also supports simple JSON-File


## Open Tasks
* Close all clients and threads on main context cancellation
* Cancel main context on syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM
* Subscriber for new server info 
* Add templated text to contest field internal on server info
* REST-API for contest status and server info
* Tests
* API-Docs
* Support ARM (aarch64) binaries and images
