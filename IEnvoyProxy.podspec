#
# Be sure to run `pod lib lint IPtProxy.podspec' to ensure this is a
# valid spec before submitting.
#
# Any lines starting with a # are optional, but their use is encouraged
# To learn more about a Podspec see https://guides.cocoapods.org/syntax/podspec.html
#

Pod::Spec.new do |s|
  s.name             = 'IEnvoyProxy'
  s.version          = '3.5.0'
  s.summary          = 'Lyrebird/Obfs4proxy, Snowflake and V2Ray for iOS and macOS'

  s.description      = <<-DESC
    All contained libraries are written in Go, which
    is a little annoying to use on iOS and Android.
    This project encapsulates all the machinations to make it work and provides an
    easy to install binary including a wrapper around all.

    Problems solved in particular are:

    - One cannot compile `main` packages with `gomobile`. All libs are patched
      to avoid this.
    - All libs are gathered under one roof here, since you cannot have two
      `gomobile` frameworks as dependencies, as there are some common Go
      runtime functions exported, which will create a name clash.
    - Environment variable changes during runtime will not be recognized by
      `goptlib` when done from within Swift/Objective-C. Therefore, sensible
      values are hardcoded in the Go wrapper.
    - The ports where the libs will listen on are hardcoded, since communicating
      the used ports back to the app would be quite some work (e.g. trying to
      read it from STDOUT) for very little benefit.
    - All libs are patched to accept all configuration parameters
      directly.

    Contained transport versions:

    | Transport |        Version        |
    |-----------|----------------------:|
    | Lyrebird  |                 0.6.0 |
    | Snowflake |                2.11.0 |
    | V2Ray     |                5.40.0 |
    | Hysteria2 |                 2.6.1 |
    | Xray      |               25.9.11 |


  DESC

  s.homepage         = 'https://github.com/greatfire/IEnvoyProxy'
  s.license          = { :type => 'MIT', :file => 'LICENSE' }
  s.author           = { 'Benjamin Erhart' => 'berhart@netzarchitekten.com' }
  s.source           = { :http => "https://github.com/greatfire/IEnvoyProxy/releases/download/e#{s.version}/IEnvoyProxy.xcframework.zip" }
  s.social_media_url = 'https://chaos.social/@tla'

  s.ios.deployment_target = '12.0'
  s.osx.deployment_target = '11'

  s.vendored_frameworks = 'IEnvoyProxy.xcframework'

  s.libraries = 'resolv'

end
