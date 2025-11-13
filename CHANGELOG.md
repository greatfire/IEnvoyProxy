# IEnvoyProxy Changlog

## 3.5.0
- Added ECH and MASQUE support.
- Added Xray support.
- Added Hysteria2 client connection settings.
- Added V2Ray-HTTP and uTLS/Hostname support.
- Added V2Ray HTTP/2 flavor support.
- Added V2Ray domain2ip map support.
- Added TLS settings for V2Ray Websocket.
- Updated V2Ray to version 5.40.0.

## 3.4.0
- Updated V2Ray to version 5.30.0.
- Fixed Hysteria submodule to version 2.6.1. Was 2.6.0 before.
- Fixed Lyrebird version string.

## 3.3.0
- Updated Snowflake to version 2.11.0.
- Updated Lyrebird to version 0.6.0.
- Updated V2Ray to version 5.29.2.
- Updated Hysteria to version 2.6.1.
- Removed V2Ray submodule, since we don't patch it anymore. Use directly instead.

## 3.2.0
- Updated V2Ray to latest version 5.24.0.
- Brought back own Hysteria2 support, as V2Ray dropped theirs again.

## 3.1.0
- Updated V2Ray to latest version 5.23.0.
- V2Ray now incorporates Hysteria2, so dropped own Hysteria2 for V2Ray version.

## 3.0.0
- Took over complete rewrite of IPtProxy:
  - Got rid of patches and the goptlib interface.
  - Instead, have our own unified code which creates transports using Lyrebird and Snowflake as dependencies.
  - Structured with classes now instead of global functions.
  - Improved interface:
    - When `#start` returns, it's now safe to use the transport.
    - `#start` will throw errors if something's wrong.
    - Callback for when transport stopped. (PTs only!)
  - Added TubeSocks, V2Ray and Hysteria 2 support to `Controller`.
- Updated Snowflake to v2.10.1.
- Updated Lyrebird to v0.5.0.

## 2.0.1
- Fixed weird mixup of obfs4Port and obfs4TubesocksPort. Properly expose all correctly.

## 2.0.0
- Updated Lyrebird to version 0.2.0.
- Added Webtunnel support.
- Updated Snowflake to version 2.9.2.
- Updated V2Ray to version 5.15.1.
- Replaced Hysteria 1 (discontinued) with Hysteria 2.4.1 (incompatible!).
- Now builds for macOS and iOS.

## 1.5.0
 - Bring Lyrebird (forked from obfs4proxy) back(ish) over from IPtProxy

## 1.4.2
 - Default UTLs to random masquerading by default for snowflake

## 1.4.1
 - Update v2ray to v4.45.2 for security fixes
 - Update hysteria to v1.1.0 (the latest our patches apply cleanly to)

## 1.4.0
 - Remove DNSTT (for binary size, currently unused)
 - Add Snowflake client (back) from IPtProxy
 - Some build script security updates from IPtProxy

## 1.3.1
 - Remove the PT environment variables, nothing seems to be using them
 - Set v2ray QUIC protocols to log at "error" level instead of "debug"
 - copy/paste error in the V2Ray patch, the SRTP and WeChat services were using the Websocket channel
 - obfs4proxy updates from @bitmold

## 1.3.0
 - Disable obfs4 support, we're not using is and it makes the bianary bigger. Maybe be removed in the future, but it's easy enough to re-enable for now
 - Use more secure temp file storage

## 1.2.1
 - Fix a crash in Hysteria when connection to the server fails


## 1.2.0
 - rework V2Ray support so Websocket, RSTP, and Wechat can be started separately

## 1.1.1
 - Bug fix: hysteria code would exit embedding app on error

## 1.1.0
 - V2ray support
 - remove binaries from the repo
 - remove snowflake

## 1.0.2
 - utlsDistribution doesn't need to be passed in
 - add default utlsDistribution to pass in to DNSTT

## 1.0.1
 - Ability to specify Hysteria protocol and ALPN

## 1.0.0
 - Forked from IPtProxy
 - DNSTT support added
 - Hysteria support added

# IPtProxy Changelog

## 1.6.0
- Update Snowflake to latest version 2.2.0.
- Added `IPtProxyObfs4ProxyVersion` returning the version of the used Obfs4proxy.
- Use latest Android NDK v24.0 which raises the minimally supported Android API level to 19.
- Added support for MacOS.

## 1.5.1
- Update Snowflake to latest main. Contains a crash fix.
- Added `IsSnowflakeProxyRunning` method to easily check,
  if the Snowflake Proxy is running.
- Exposed `IsPortAvailable` so consumers don't need to 
  implement this themselves, if they happen to do something similar.

## 1.5.0
- Updated Obfs4proxy to latest version 0.0.13.
- Updated Snowflake to latest version 2.1.0.
- Fixed bug when stopping Snowflake proxy. (Thanks bitmold!)

## 1.4.0
- Updated Obfs4proxy to latest 0.0.13-dev which fixes a bug which made prior 
  versions distinguishable.
- Fixed minor documentation issues.

## 1.3.0
- Updated Snowflake to version 2.0.1.
- Added Snowflake AMP support.
- Switched to newer DTLS library, which improves fingerprinting resistance for Snowflake.
- Added callback to `StartSnowflakeProxy`, to allow counting of connected clients.
- Fixed iOS warnings about wrong iOS SDK.

## 1.2.0
- Added explicit support for a proxy behind Obfs4proxy.

## 1.1.0
- Updated Snowflake to latest master. Fixes multiple minor issues.
- Registers Snowflake proxy with type "iptproxy" to improve statistics.

## 1.0.0
- Updated Snowflake to latest master. Fixes multiple minor issues.
- Updated Obfs4proxy to latest master. Contains a minor fix for Meek.
- Added port test mechanism to avoid port collisions when started multiple times.
- Improved documentation.

## 0.6.0
- Updated Obfs4proxy to latest master. Fixes support for unsafe logging.
- Added `StopSnowflakeProxy` from feature branch.
- Updated Snowflake to latest master. Fixes multiple minor issues.

## 0.5.2
- Updated Obfs4proxy to fix broken meek_lite due to Microsoft Azure certificate
  changes. NOTE: If you still experience HPKP issues, you can use 
  "disableHPKP=true" in the meek_lite configuration.

## 0.5.1

- Base on latest Snowflake master which contains a lot of patches we previously
  had to provide ourselves.

## 0.5.0

- Added `StopSnowflake` function.

## 0.4.0

- Added `StopObfs4Proxy` function.
- Updated Snowflake to latest master.

## 0.3.0

- Added Snowflake Proxy support, so contributors can run proxies on their 
  mobile devices.
- Updated Snowflake to latest master.
- Fixed doc to resemble proper Objective-C documentation.

## 0.2.0

- Improved Android support.
- Improved documentation.
- Updated Snowflake to latest master.

## 0.1.0

Initial version
