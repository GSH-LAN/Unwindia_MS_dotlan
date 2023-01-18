<p align="center">
  <a href="https://github.com/gsh-lan/unwindia" target="blank"><img src="https://raw.githubusercontent.com/GSH-LAN/Unwindia/main/.resources/images/logo.png" height="128" alt="unwindia logo">
  <a href="https://github.com/gsh-lan/unwindia" target="blank"><img src="https://raw.githubusercontent.com/GSH-LAN/Unwindia/main/.resources/images/header.svg" height="128" alt="unwindia header" /></a>
</p>

# UNWINDIA DotLAN MatchService
> The DotLAN MatchService for [Unwindia](https://github.com/GSH-LAN/Unwindia)
---

[![Codacy grade](https://img.shields.io/codacy/grade/efb2b55cfdc1472d98c3913fb098b657?style=for-the-badge)](https://www.codacy.com/gh/GSH-LAN/Unwindia_MS_dotlan/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=GSH-LAN/Unwindia_ms_dotlan&amp;utm_campaign=Badge_Grade)
[![Codecov](https://img.shields.io/codecov/c/gh/GSH-LAN/Unwindia_MS_dotlan?style=for-the-badge&token=D3ME50U8KT)](https://codecov.io/gh/GSH-LAN/Unwindia_MS_dotlan)


This service integrates UNWINDIA with the tourney system of [DotLAN](https://dotlan.net). 

## Configuration

### Environment Variables

### Dynamic Configuration 
The dynamic configuration is typically loaded from Config-Service but also supports simple JSON-File


## Open Tasks
  * Close all clients and threads on main context cancellation
  * Cancel main context on syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM
  * Add http server for metrics, tracing (jaeger) and health 
  * Subscriber for new server info 
  * Add templated text to contest field internal on server info
  * REST-API for contest status
  * Tests
  * API-Docs
