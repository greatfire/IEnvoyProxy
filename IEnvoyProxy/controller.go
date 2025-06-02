package IEnvoyProxy

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log"
	"net"
	"net/url"
	"os"
	"path"

	"fmt"
	"strconv"
	"sync"
	"time"

	hysteria2 "github.com/apernet/hysteria/app/v2/cmd"
	ndns "github.com/ncruces/go-dns"
	v2ray "github.com/v2fly/v2ray-core/v5/envoy"
	"gitlab.com/stevenmcdonald/tubesocks"
	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	ptlog "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/common/log"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports/base"
	sfversion "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/version"
	"golang.org/x/net/proxy"
)

// LogFileName - the filename of the log residing in `StateDir`.
const LogFileName = "iep.log"

//goland:noinspection GoUnusedConst
const (
	// ScrambleSuit - DEPRECATED transport implemented in Lyrebird.
	ScrambleSuit = "scramblesuit"

	// Obfs2 - DEPRECATED transport implemented in Lyrebird.
	Obfs2 = "obfs2"

	// Obfs3 - DEPRECATED transport implemented in Lyrebird.
	Obfs3 = "obfs3"

	// Obfs4 - Transport implemented in Lyrebird.
	Obfs4 = "obfs4"

	// MeekLite - Transport implemented in Lyrebird.
	MeekLite = "meek_lite"

	// Webtunnel - Transport implemented in Lyrebird.
	Webtunnel = "webtunnel"

	// Snowflake - Transport implemented in Snowflake.
	Snowflake = "snowflake"

	// Obfs4TubeSocks - Obfs4 transport using TubeSocks to configure Obfs4
	//
	// This is probably not the ideal way to do things, but it's expedient.
	// We've been unable to configure cronet to use a socks proxy that requires
	// auth info. TubeSocks bridges that gap by running a second socks proxy.
	Obfs4TubeSocks = "obfs4_tubesocks"

	// MeekLiteTubeSocks - Meek Lite transport using TubeSocks to configure Meek Lite
	//
	// This is probably not the ideal way to do things, but it's expedient.
	// We've been unable to configure cronet to use a socks proxy that requires
	// auth info. TubeSocks bridges that gap by running a second socks proxy.
	MeekLiteTubeSocks = "meek_tubesocks"

	// V2RayWs - V2Ray Proxy via WebSocket
	V2RayWs = "v2ray_ws"

	// V2RaySrtp - V2Ray Proxy via SRTP
	V2RaySrtp = "v2ray_srtp"

	// V2RayWechat - V2Ray Proxy via WeChat
	V2RayWechat = "v2ray_wechat"

	// Hysteria2 - Hysteria 2 Proxy
	Hysteria2 = "hysteria2"

	// EnvoyEch - Envoy ECH proxy
	EnvoyEch = "envoy_ech"

	// Masque - Envoy MASQUE Proxy
	Masque = "masque"
)

var (
	transportsInitOnce sync.Once
)

// OnTransportStopped - Interface to get notified when a transport stopped again.
type OnTransportStopped interface {
	Stopped(name string, error error)
}

// Controller - Class to start and stop transports.
type Controller struct {

	// SnowflakeIceServers is a comma-separated list of ICE server addresses.
	SnowflakeIceServers string

	// SnowflakeBrokerUrl - URL of signaling broker.
	SnowflakeBrokerUrl string

	// SnowflakeFrontDomains is a comma-separated list of domains for either
	// the domain fronting or AMP cache rendezvous methods.
	SnowflakeFrontDomains string

	// SnowflakeAmpCacheUrl - URL of AMP cache to use as a proxy for signaling.
	// Only needed when you want to do the rendezvous over AMP instead of a domain fronted server.
	SnowflakeAmpCacheUrl string

	// SnowflakeSqsUrl - URL of SQS Queue to use as a proxy for signaling.
	SnowflakeSqsUrl string

	// SnowflakeSqsCreds - Credentials to access SQS Queue.
	SnowflakeSqsCreds string

	// SnowflakeMaxPeers - Capacity for number of multiplexed WebRTC peers. DEFAULTs to 1 if less than that.
	SnowflakeMaxPeers int

	// Obfs4TubeSocksUser - Username which TubeSocks should use to start Obfs4 with.
	Obfs4TubeSocksUser string

	// Obfs4TubeSocksPassword - Password which TubeSocks should use to start Obfs4 with.
	Obfs4TubeSocksPassword string

	// MeekLiteTubeSocksUser - Username which TubeSocks should use to start MeekLite with.
	MeekLiteTubeSocksUser string

	// MeekLiteTubeSocksPassword - Password which TubeSocks should use to start MeekLite with.
	MeekLiteTubeSocksPassword string

	// V2RayServerAddress - Hostname of WS web server proxy
	V2RayServerAddress string

	// V2RayServerPort - Port of the WS listener (probably 443)
	V2RayServerPort string

	// V2RayWsPath - path to the websocket (V2RayWs only!)
	V2RayWsPath string

	// V2RayId - V2Ray UUID for auth
	V2RayId string

	// V2RayAllowInsecure - If true, V2Ray allows insecure connection at TLS client
	V2RayAllowInsecure bool

	// V2RayServerName - Server name used for TLS authentication.
	V2RayServerName string

	// Hysteria2Server - A Hysteria2 server URL https://v2.hysteria.network/docs/developers/URI-Scheme/
	Hysteria2Server string

	// ECH
	// URL to test, e.g. https://www.google.com/generate_204
	EchTestTarget		  string
	// expected response code, e.g. 204
	EchTestResponse		  int
	// Salt to use for cache busting param
	Salt 				  string
	// upstream Envoy server URL
	EchEnvoyUrl     	  string
	// hostname of the upstream URL (for convenience)
	EchEnvoyHost	      string
	// ECH config list data for the Envoy host, likely fetched from DNS
	EchEnvoyEchConfigList []byte
	// The Proxy's Envoy URL
	EchProxyUrl			  string
	// instace
	echProxy 			  *EchProxy

	// upstream host and port
	MasqueHost			  string
	MasquePort			  int
	// instance
	masqueProxy			  *EnvoyMasqueProxy

	stateDir         string
	transportStopped OnTransportStopped
	listeners        map[string]*pt.SocksListener
	shutdown         map[string]chan struct{}

	v2rayWsRunning     bool
	v2raySrtpRunning   bool
	v2rayWechatRunning bool
	hysteria2Running   bool
	echProxyRunning    bool
	masqueRunning      bool

	obf4TubeSocksPort     int
	meekLiteTubeSocksPort int
	v2rayWsPort           int
	v2raySrtpPort         int
	v2rayWechatPort       int
	hysteria2Port         int
	echProxyPort          int
	masquePort			  int
}

// NewController - Create a new Controller object.
//
// @param enableLogging Log to StateDir/ipt.log.
//
// @param unsafeLogging Disable the address scrubber.
//
// @param logLevel Log level (ERROR/WARN/INFO/DEBUG). Defaults to ERROR if empty string.
//
// @param transportStopped A delegate, which is called, when the started transport stopped again.
// Will be called on its own thread! You will need to switch to your own UI thread,
// if you want to do UI stuff!
//
//goland:noinspection GoUnusedExportedFunction
func NewController(stateDir string, enableLogging, unsafeLogging bool, logLevel string, transportStopped OnTransportStopped) *Controller {
	c := &Controller{
		stateDir:         stateDir,
		transportStopped: transportStopped,
		v2raySrtpPort:    47600,
		v2rayWechatPort:  47700,
		v2rayWsPort:      47800,
		hysteria2Port:    48000,

		echProxyPort:     49000,
		EchTestTarget:    "https://www.google.com/generate_204",
		EchTestResponse:  204,

		masquePort:		  50000,
	}

	if logLevel == "" {
		logLevel = "ERROR"
	}

	if err := createStateDir(c.stateDir); err != nil {
		log.Printf("Failed to set up state directory: %s", err)
		return nil
	}
	if err := ptlog.Init(enableLogging,
		path.Join(c.stateDir, LogFileName), unsafeLogging); err != nil {
		log.Printf("Failed to set initialize log: %s", err.Error())
		return nil
	}
	if err := ptlog.SetLogLevel(logLevel); err != nil {
		log.Printf("Failed to set log level: %s", err.Error())
		ptlog.Warnf("Failed to set log level: %s", err.Error())
	}

	// This should only ever be called once, even when new `Controller` instances are created.
	var err error
	transportsInitOnce.Do(func() {
		err = transports.Init()
	})

	if err != nil {
		ptlog.Warnf("Failed to initialize transports: %s", err.Error())
		return nil
	}

	c.listeners = make(map[string]*pt.SocksListener)
	c.shutdown = make(map[string]chan struct{})

	return c
}

// StateDir - The StateDir set in the constructor.
//
// @returns the directory you set in the constructor, where transports store their state and where the log file resides.
func (c *Controller) StateDir() string {
	return c.stateDir
}

// addExtraArgs adds the args in extraArgs to the connection args
func addExtraArgs(args *pt.Args, extraArgs *pt.Args) {
	if extraArgs == nil {
		return
	}

	for name := range *extraArgs {
		// Only add if extra arg doesn't already exist, and is not empty.
		if value, ok := args.Get(name); !ok || value == "" {
			if value, ok := extraArgs.Get(name); ok && value != "" {
				args.Add(name, value)
			}
		}
	}
}

func acceptLoop(f base.ClientFactory, ln *pt.SocksListener, proxyURL *url.URL,
	extraArgs *pt.Args, shutdown chan struct{}, methodName string, transportStopped OnTransportStopped) {

	defer func(ln *pt.SocksListener) {
		_ = ln.Close()
	}(ln)

	for {
		conn, err := ln.AcceptSocks()
		if err != nil {
			var e net.Error
			if errors.As(err, &e) && !e.Temporary() {
				return
			}

			continue
		}

		go clientHandler(f, conn, proxyURL, extraArgs, shutdown, methodName, transportStopped)
	}
}

func clientHandler(f base.ClientFactory, conn *pt.SocksConn, proxyURL *url.URL,
	extraArgs *pt.Args, shutdown chan struct{}, methodName string, transportStopped OnTransportStopped) {

	defer func(conn *pt.SocksConn) {
		_ = conn.Close()
	}(conn)

	addExtraArgs(&conn.Req.Args, extraArgs)
	args, err := f.ParseArgs(&conn.Req.Args)
	if err != nil {
		ptlog.Errorf("Error parsing PT args: %s", err.Error())
		_ = conn.Reject()

		if transportStopped != nil {
			transportStopped.Stopped(methodName, err)
		}

		return
	}

	dialFn := proxy.Direct.Dial
	if proxyURL != nil {
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			ptlog.Errorf("Error getting proxy dialer: %s", err.Error())
			_ = conn.Reject()

			if transportStopped != nil {
				transportStopped.Stopped(methodName, err)
			}

			return
		}
		dialFn = dialer.Dial
	}

	remote, err := f.Dial("tcp", conn.Req.Target, dialFn, args)
	if err != nil {
		ptlog.Errorf("Error dialing PT: %s", err.Error())

		if transportStopped != nil {
			transportStopped.Stopped(methodName, err)
		}

		return
	}

	err = conn.Grant(&net.TCPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		ptlog.Errorf("conn.Grant error: %s", err)

		if transportStopped != nil {
			transportStopped.Stopped(methodName, err)
		}

		return
	}

	defer func(remote net.Conn) {
		_ = remote.Close()
	}(remote)

	done := make(chan struct{}, 2)
	go copyLoop(conn, remote, done)

	// wait for copy loop to finish or for shutdown signal
	select {
	case <-shutdown:
	case <-done:
		ptlog.Noticef("copy loop ended")
	}

	if transportStopped != nil {
		ptlog.Noticef("call transportStopped")
		transportStopped.Stopped(methodName, nil)
	}
}

// Exchanges bytes between two ReadWriters.
// (In this case, between a SOCKS connection and a pt conn)
func copyLoop(socks, sfconn io.ReadWriter, done chan struct{}) {
	go func() {
		if _, err := io.Copy(socks, sfconn); err != nil {
			ptlog.Errorf("copying transport to SOCKS resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
	go func() {
		if _, err := io.Copy(sfconn, socks); err != nil {
			ptlog.Errorf("copying SOCKS to transport resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
}

// LocalAddress - Address of the given transport.
//
// @param methodName one of the constants `ScrambleSuit` (deprecated), `Obfs2` (deprecated), `Obfs3` (deprecated),
// `Obfs4`, `MeekLite`, `Webtunnel` or `Snowflake`.
//
// @return address string containing host and port where the given transport listens.
func (c *Controller) LocalAddress(methodName string) string {
	switch methodName {
	case V2RayWs:
		if c.v2rayWsRunning {
			return net.JoinHostPort("127.0.0.1", strconv.Itoa(c.v2rayWsPort))
		}
		return ""

	case V2RaySrtp:
		if c.v2raySrtpRunning {
			return net.JoinHostPort("127.0.0.1", strconv.Itoa(c.v2raySrtpPort))
		}
		return ""

	case V2RayWechat:
		if c.v2rayWechatRunning {
			return net.JoinHostPort("127.0.0.1", strconv.Itoa(c.v2rayWechatPort))
		}
		return ""

	case Hysteria2:
		if c.hysteria2Running {
			return net.JoinHostPort("127.0.0.1", strconv.Itoa(c.hysteria2Port))
		}
		return ""

	case EnvoyEch:
		if c.echProxyRunning {
			return net.JoinHostPort("127.0.0.1", strconv.Itoa(c.echProxyPort))
		}
		return ""

	case Masque:
		if c.masqueRunning {
			return net.JoinHostPort("127.0.0.1", strconv.Itoa(c.masquePort))
		}
		return ""

	default:
		if ln, ok := c.listeners[methodName]; ok {
			return ln.Addr().String()
		}
		return ""
	}
}

// Port - Port of the given transport.
//
// @param methodName one of the constants `ScrambleSuit` (deprecated), `Obfs2` (deprecated), `Obfs3` (deprecated),
// `Obfs4`, `MeekLite`, `Webtunnel` or `Snowflake`.
//
// @return port number on localhost where the given transport listens.
func (c *Controller) Port(methodName string) int {
	switch methodName {
	case Obfs4TubeSocks:
		return c.obf4TubeSocksPort

	case MeekLiteTubeSocks:
		return c.meekLiteTubeSocksPort

	case V2RayWs:
		if c.v2rayWsRunning {
			return c.v2rayWsPort
		}
		return 0

	case V2RaySrtp:
		if c.v2raySrtpRunning {
			return c.v2raySrtpPort
		}
		return 0

	case V2RayWechat:
		if c.v2rayWechatRunning {
			return c.v2rayWechatPort
		}
		return 0

	case Hysteria2:
		if c.hysteria2Running {
			return c.hysteria2Port
		}
		return 0

	case EnvoyEch:
		if c.echProxyRunning {
			return c.echProxyPort
		}
		return 0

	case Masque:
		if c.masqueRunning {
			return c.masquePort
		}
		return 0

	default:
		if ln, ok := c.listeners[methodName]; ok {
			return int(ln.Addr().(*net.TCPAddr).AddrPort().Port())
		}
		return 0
	}
}

func createStateDir(path string) error {
	info, err := os.Stat(path)

	// If dir does not exist, try to create it.
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(path, 0700)

		if err == nil {
			info, err = os.Stat(path)
		}
	}

	// If it is not a dir, return error
	if err == nil && !info.IsDir() {
		err = fs.ErrInvalid
		return err
	}

	// Create a file within dir to test writability.
	tempFile := path + "/.iptproxy-writetest"
	var file *os.File
	file, err = os.Create(tempFile)

	// Remove the test file again.
	if err == nil {
		_ = file.Close()

		err = os.Remove(tempFile)
	}
	return err
}

// Start - Start given transport.
//
// @param methodName one of the constants `ScrambleSuit` (deprecated), `Obfs2` (deprecated), `Obfs3` (deprecated),
// `Obfs4`, `MeekLite`, `Webtunnel` or `Snowflake`.
//
// @param proxy HTTP, SOCKS4 or SOCKS5 proxy to be used behind Lyrebird. E.g. "socks5://127.0.0.1:12345"
//
// @throws if the proxy URL cannot be parsed, if the given `methodName` cannot be found, if the transport cannot
// be initialized or if it couldn't bind a port for listening.
func (c *Controller) Start(methodName string, proxy string) error {
	var proxyURL *url.URL
	var err error

	if proxy != "" {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			ptlog.Errorf("Failed to parse proxy address: %s", err.Error())
			return err
		}
	}

	switch methodName {
	case Obfs4TubeSocks:
		if c.Port(Obfs4) < 1 {
			err := c.Start(Obfs4, proxy)
			if err != nil {
				return err
			}
		}

		c.obf4TubeSocksPort = findPort(47350)

		tubesocks.Start(
			c.Obfs4TubeSocksUser,
			c.Obfs4TubeSocksPassword,
			net.JoinHostPort("127.0.0.1", strconv.Itoa(c.Port(Obfs4))),
			c.obf4TubeSocksPort)

	case MeekLiteTubeSocks:
		if c.Port(MeekLite) < 1 {
			err := c.Start(MeekLite, proxy)
			if err != nil {
				return err
			}
		}

		c.meekLiteTubeSocksPort = findPort(47360)

		tubesocks.Start(
			c.MeekLiteTubeSocksUser,
			c.MeekLiteTubeSocksPassword,
			net.JoinHostPort("127.0.0.1", strconv.Itoa(c.Port(MeekLite))),
			c.meekLiteTubeSocksPort)

	case V2RayWs:
		if !c.v2rayWsRunning {
			c.v2rayWsPort = findPort(c.v2rayWsPort)

			err := v2ray.StartWs(c.v2rayWsPort, c.V2RayServerAddress, c.V2RayServerPort, c.V2RayWsPath, c.V2RayId, v2ray.WsConfigOptional{
				AllowInsecure: c.V2RayAllowInsecure,
				ServerName:    c.V2RayServerName,
			})
			if err != nil {
				ptlog.Errorf("Failed to initialize %s: %s", methodName, err)
				return err
			}

			c.v2rayWsRunning = true
		}

	case V2RaySrtp:
		if !c.v2raySrtpRunning {
			c.v2raySrtpPort = findPort(c.v2raySrtpPort)

			err := v2ray.StartSrtp(c.v2raySrtpPort, c.V2RayServerAddress, c.V2RayServerPort, c.V2RayId)
			if err != nil {
				ptlog.Errorf("Failed to initialize %s: %s", methodName, err)
				return err
			}

			c.v2raySrtpRunning = true
		}

	case V2RayWechat:
		if !c.v2rayWechatRunning {
			c.v2rayWechatPort = findPort(c.v2rayWechatPort)

			err := v2ray.StartWechat(c.v2rayWechatPort, c.V2RayServerAddress, c.V2RayServerPort, c.V2RayId)
			if err != nil {
				ptlog.Errorf("Failed to initialize %s: %s", methodName, err)
				return err
			}

			c.v2rayWechatRunning = true
		}

	case Hysteria2:
		if !c.hysteria2Running {
			c.hysteria2Port = findPort(c.hysteria2Port)

			configFile := fmt.Sprintf("%s/hysteria.yaml", c.stateDir)

			err = os.WriteFile(configFile,
				[]byte(fmt.Sprintf("server: %s\n\nsocks5:\n  listen: 127.0.0.1:%d\n", c.Hysteria2Server, c.hysteria2Port)),
				0644)

			if err != nil {
				ptlog.Errorf("Could not write config file: %s\n", err.Error())
				return err
			}

			c.hysteria2Running = true

			go hysteria2.Start(configFile)

			// Need to sleep a little here, to give Hysteria2 a chance to start.
			// Otherwise, Hysteria2 wouldn't be listening
			// on that configured SOCKS5 port, yet and connections would fail.
			time.Sleep(time.Second)
		}

	case Snowflake:
		extraArgs := &pt.Args{}
		extraArgs.Add("fronts", c.SnowflakeFrontDomains)
		extraArgs.Add("ice", c.SnowflakeIceServers)
		extraArgs.Add("max", strconv.Itoa(max(1, c.SnowflakeMaxPeers)))
		extraArgs.Add("url", c.SnowflakeBrokerUrl)
		extraArgs.Add("ampcache", c.SnowflakeAmpCacheUrl)
		extraArgs.Add("sqsqueue", c.SnowflakeSqsUrl)
		extraArgs.Add("sqscreds", c.SnowflakeSqsCreds)
		extraArgs.Add("proxy", proxy)

		t := transports.Get(methodName)
		if t == nil {
			ptlog.Errorf("Failed to initialize %s: no such method", methodName)
			return fmt.Errorf("failed to initialize %s: no such method", methodName)
		}
		f, err := t.ClientFactory(c.stateDir)
		if err != nil {
			ptlog.Errorf("Failed to initialize %s: %s", methodName, err.Error())
			return err
		}
		ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
		if err != nil {
			ptlog.Errorf("Failed to initialize %s: %s", methodName, err.Error())
			return err
		}

		c.shutdown[methodName] = make(chan struct{})
		c.listeners[methodName] = ln

		go acceptLoop(f, ln, nil, extraArgs, c.shutdown[methodName], methodName, c.transportStopped)

	case EnvoyEch:
		if !c.echProxyRunning {
			c.echProxyPort = findPort(c.echProxyPort)

			c.echProxy = &EchProxy{
				TestTarget: c.EchTestTarget,
				TargetResponse: c.EchTestResponse,
				ProxyListen: c.LocalAddress(EnvoyEch),
				EnvoyUrl: c.EchEnvoyUrl,
				EnvoyHost: c.EchEnvoyHost,
				EchConfigList: c.EchEnvoyEchConfigList,
			}

			c.echProxyRunning = true
			go c.echProxy.startProxy()

			// wait for it to start
			isItUpYet(c.LocalAddress(EnvoyEch))

			// test HTTP/2 and HTTP/3, selecting the one that responds first
			c.EchProxyUrl = c.echProxy.testHttps()
		}

	case Masque:
		if !c.masqueRunning {
			c.masquePort = findPort(c.masquePort)

			c.masqueProxy = &EnvoyMasqueProxy{
				UpstreamServer: c.MasqueHost,
				UpstreamPort: c.MasquePort,
				ListenPort: c.masquePort,

				insecure: false,
				token: "NEEDED FOR invisv-privacy/masque",
			}

			c.masqueRunning = true
			go c.masqueProxy.Start()
		}

	default:
		// at the moment, everything else is in lyrebird
		t := transports.Get(methodName)
		if t == nil {
			ptlog.Errorf("Failed to initialize %s: no such method", methodName)
			return fmt.Errorf("failed to initialize %s: no such method", methodName)
		}

		f, err := t.ClientFactory(c.stateDir)
		if err != nil {
			ptlog.Errorf("Failed to initialize %s: %s", methodName, err.Error())
			return err
		}

		ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
		if err != nil {
			ptlog.Errorf("Failed to initialize %s: %s", methodName, err.Error())
			return err
		}

		c.listeners[methodName] = ln
		c.shutdown[methodName] = make(chan struct{})

		go acceptLoop(f, ln, proxyURL, nil, c.shutdown[methodName], methodName, c.transportStopped)
	}

	ptlog.Noticef("Launched transport: %v", methodName)

	return nil
}

// Stop - Stop given transport.
//
// @param methodName one of the constants `ScrambleSuit` (deprecated), `Obfs2` (deprecated), `Obfs3` (deprecated),
// `Obfs4`, `MeekLite`, `Webtunnel` or `Snowflake`.
func (c *Controller) Stop(methodName string) {
	switch methodName {
	case Obfs4TubeSocks:
		c.Stop(Obfs4)
		c.obf4TubeSocksPort = 0

	case MeekLiteTubeSocks:
		c.Stop(MeekLite)
		c.meekLiteTubeSocksPort = 0

	case V2RayWs:
		if c.v2rayWsRunning {
			ptlog.Noticef("Shutting down %s", methodName)
			go v2ray.StopWs()
			c.v2rayWsRunning = false
		} else {
			ptlog.Warnf("No listener for %s", methodName)
		}

	case V2RaySrtp:
		if c.v2raySrtpRunning {
			ptlog.Noticef("Shutting down %s", methodName)
			go v2ray.StopSrtp()
			c.v2raySrtpRunning = false
		} else {
			ptlog.Warnf("No listener for %s", methodName)
		}

	case V2RayWechat:
		if c.v2rayWechatRunning {
			ptlog.Noticef("Shutting down %s", methodName)
			go v2ray.StopWechat()
			c.v2rayWechatRunning = false
		} else {
			ptlog.Warnf("No listener for %s", methodName)
		}

	case Hysteria2:
		if c.hysteria2Running {
			ptlog.Noticef("Shutting down %s", methodName)
			go hysteria2.Stop()
			_ = os.Remove(fmt.Sprintf("%s/hysteria.yaml", c.stateDir))
			c.hysteria2Running = false
		} else {
			ptlog.Warnf("No listener for %s", methodName)
		}

	case EnvoyEch:
		if c.echProxyRunning {
			ptlog.Warnf("TODO %s can't stop!", methodName)
		} else {
			ptlog.Warnf("No listener for %s", methodName)
		}

	case Masque:
		if c.echProxyRunning {
			ptlog.Warnf("TODO %s can't stop!", methodName)
		} else {
			ptlog.Warnf("No listener for %s", methodName)
		}

	default:
		if ln, ok := c.listeners[methodName]; ok {
			_ = ln.Close()

			ptlog.Noticef("Shutting down %s", methodName)

			close(c.shutdown[methodName])
			delete(c.shutdown, methodName)
			delete(c.listeners, methodName)
		} else {
			ptlog.Warnf("No listener for %s", methodName)
		}
	}
}

// SnowflakeVersion - The version of Snowflake bundled with IPtProxy.
//
//goland:noinspection GoUnusedExportedFunction
func SnowflakeVersion() string {
	return sfversion.GetVersion()
}

// LyrebirdVersion - The version of Lyrebird bundled with IPtProxy.
//
//goland:noinspection GoUnusedExportedFunction
func LyrebirdVersion() string {
	return "lyrebird-0.6.0"
}

// SetDOHServer - set the default Go resolver to use DNS over HTTPS with our
// working server
func (c *Controller) SetDOHServer(dohServer string) {
	log.Printf("Setting default Go DNS resolver to use DOH: %s", dohServer)
	doh_url := "https://" + dohServer + "/dns-query{?dns}"
	resolver, r_err := ndns.NewDoHResolver(doh_url)
	if r_err != nil {
		log.Fatalf("Failed to make a resolver: %s", r_err)
	}
	net.DefaultResolver = resolver
}

func findPort(port int) int {
	temp := port

	for !isPortAvailable(temp) {
		temp++
	}

	return temp
}

// isPortAvailable - Checks to see if a given port is not in use.
//
// @param port The port to check.
func isPortAvailable(port int) bool {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)

	if err != nil {
		return true
	}

	_ = conn.Close()

	return false
}

///
// Attempt a basic TCP connection to see if a service is up yet
// blocks and polls until the connection is made
//
// XXX this probably should have some kind of timeout for the service
// failing to start
//
func isItUpYet(addr string) (bool, error) {

	var d net.Dialer

	// poll every 2 seconds until the service is listening
	up := false
	for !up {
		ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
		conn, err := d.DialContext(ctx, "tcp", addr)
		// This can throw timeout and connection refused... probably more
		// just ignore it all ;-)
		log.Printf("&&&& isItUpYet %s", err)
		if err == nil {
			conn.Close()
			cancel()
			return true, nil
		}

		time.Sleep(2 * time.Second)
	}

	return false, nil
}