# IEnvoyProxy

This is a fork of IPtProxy (https://github.com/tladesignz/IPtProxy) modified to include censorship evading proxies used by Envoy (https://github.com/greatfire/envoy)

currently this includes:

* Lyrebird(https://gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird)
* Snowflake(https://gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake)
* [V2ray](https://github.com/v2fly/v2ray-core)
* [XRay](https://github.com/XTLS/Xray-core)
* [Hysteria2](https://github.com/apernet/hysteria.git)

We have previously supported (available in git history):

* [obfs4proxy](https://github.com/Yawning/obfs4)
* [DNSTT](https://www.bamsoftware.com/software/dnstt/)
* [Hysteria](https://github.com/HyNetwork/hysteria)


While this library was made for use with Envoy, it does not depend on Envoy, and may be useful for other situations. 
Some of the choices made are specific to our needs, but if others want to use this, we can look in to making it more flexible.

In all cases there is the `Controller.Start()` and `Controller.Stop()` function to start a service. 
There is also `Controller.Port()` and `Controller.LocalAddres()` to get the port each service is listening on.
If the respective service is not started, yet, these functions will return `0` resp. an empty string.

IEnvoyProxy is still a work in progress. Feel free to open issues in this repo if you have questions or comments.

Problems solved in particular are:

- One cannot compile `main` packages with `gomobile`. Both PTs are patched
  to avoid this.
- Proxies are gathered under one roof here, since you cannot have two
  `gomobile` frameworks as dependencies, since there are some common Go
  runtime functions exported, which would create a name clash.
- Free ports to be used are automatically found by this library and can be fetched by the
  consuming app. 

## iOS/macOS

### Installation

IEnvoyProxy is available through [CocoaPods](https://cocoapods.org). To install
it, add the following line to your `Podfile`:

```ruby
pod 'IEnvoyProxy', :podspec => 'https://raw.githubusercontent.com/greatfire/IEnvoyProxy/refs/tags/e3.5.0/IEnvoyProxy.podspec'
```

### Getting Started

Before using IEnvoyProxy you need to specify a place on disk for the transports
to store their state information and log files.

You will need to provide a writable `StateLocation` on construction:

```swift
let fm = FileManager.default

// Good choice for apps where IEnvoyProxy runs inside an extension:

guard let ptDir = fm
    .containerURL(forSecurityApplicationGroupIdentifier: "group.com.example.app")? 
    .appendingPathComponent("pt_state")?
    .path
else {
    return
}

let ptc = IEnvoyProxyController(ptDir, enableLogging: true, unsafeLogging: false, logLevel: "INFO")

// For normal apps which run IEnvoyProxy inline:

guard let ptDir = fm.urls(for: .documentDirectory, in: .userDomainMask)
    .first?
    .appendingPathComponent("pt_state")
    .path 
else {
    return
}

let ptc = IEnvoyProxyController(ptDir, enableLogging: true, unsafeLogging: false, logLevel: "INFO")
```

## Android 

### Installation

IEnvoyProxy is available through [JitPack](https://jitpack.io). To install
it, simply add the following line to your `build.gradle` file:

```groovy
implementation 'org.greatfire:IEnvoyProxy:3.5.0'
```

And this to your root `build.gradle` at the end of repositories:

```groovy
allprojects {
	repositories {
		// ...
		maven { url 'https://jitpack.io' }
	}
}
```

For newer Android Studio projects created in 
[Android Studio Bumblebee | 2021.1.1](https://developer.android.com/studio/preview/features?hl=hu#settings-gradle) 
or newer</a>, the JitPack repository needs to be added into the root level file `settings.gradle` 
instead of `build.gradle`:

```groovy
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
	  // ...
        maven { url 'https://jitpack.io' }
    }
}
```

Precomiled binaries are also available on the [releases page](https://github.com/stevenmcdonald/IEnvoyProxy/releases)

### Getting Started

If you are building a new Android application be sure to declare that it uses the
`INTERNET` permission in your Android Manifest:

```xml
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="my.test.app">

    <uses-permission android:name="android.permission.INTERNET"/>
    <application ...

```

Before using IEnvoyProxy you need to specify a place on disk for it to store its state
information. We recommend the path returned by `Context#getCacheDir()`:

```java
import IEnvoyProxy

File ptDir = new File(getCacheDir(), "pt_state");

Controller ptc = Controller(ptDir, true, false, "INFO");
```


## Build

### Requirements

You'll need Go 1.25 as a prerequisite.

You will also need Xcode installed when compiling for iOS and an Android NDK
when compiling for Android.

The build script needs the gomobile binary and will install it, if not available, yet.
However, you'll still need to make it accessible in your `$PATH`.

So, if it's not already, add `$GOPATH/bin` to `$PATH`. The default location 
for `$GOPATH` is `$HOME/go`: 

```bash
export PATH=$HOME/go/bin/:$PATH` 
```

### iOS

Make sure Xcode and Xcode's command line tools are installed. Then run

```bash
rm -rf IEnvoyProxy.xcframework && ./build.sh
```

This will create an `IEnvoyProxy.xcframework`, which you can directly drop in your app,
if you don't want to rely on CocoaPods.

### Android

Make sure that `javac` is in your `$PATH`. If you do not have a JDK instance, on Debian systems you can install it with: 

```bash
apt install default-jdk 
````

If they aren't already, make sure the `$ANDROID_HOME` and `$ANDROID_NDK_HOME` 
environment variables are set:

```bash
export ANDROID_HOME=~/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/$NDK_VERSION

rm -rf IEnvoyProxy.aar IEnvoyProxy-sources.jar && ./build.sh android
```

This will create an `IEnvoyProxy.aar` file, which you can directly drop in your app, 
if you don't want to rely on JitPack.

On certain CPU architectures `gobind` might fail with this error due to setting
a flag that is no longer supported by Go 1.16:

```
go tool compile: exit status 1
unsupported setting GO386=387. Consider using GO386=softfloat instead.
gomobile: go build -v -buildmode=c-shared -o=/tmp/gomobile-work-855414073/android/src/main/jniLibs/x86/libgojni.so ./gobind failed: exit status 1
```

If this is the case, you will need to set this flag to build IEnvoyProxy:

```bash
export GO386=sse2
``` 

## Development

### To update dependencies and clean up:
```bash
cd IEnvoyProxy
go mod tidy
go get golang.org/x/mobile/cmd/gomobile@latest
go get google.golang.org/genproto@latest
```

## Authors

- Steven McDonald, scm@eds.org
- Benjamin Erhart, berhart@netzarchitekten.com

for GreatFire https://en.greatfire.org/

### IPtProxy Authors:

- Benjamin Erhart, berhart@netzarchitekten.com
- Nathan Freitas
- Annette (formerly Bim)
- cohosh

for the Guardian Project https://guardianproject.info

## License

IEnvoyProxy is available under the MIT license. See the [LICENSE](LICENSE) file for more info.
