# Local Game CI

Leverage DaggerCI and GameCI to create local and remote builds


## Build

**Personal License**
```
 dagger call build --src="./example/game" \
    --ulf="./Unity_v6000.x.ulf" \
    --build-target="StandaloneWindows64" \
    --build-name="demo" \
    --platform="windows-mono" \
    --os="ubuntu"
    --user=env:USERNAME \
    --pass=env:PASS \
    export --path=./Builds
```

**Serial License**
```
 dagger call build --src="./example/game" \
    --build-target="StandaloneWindows64" \
    --build-name="demo" \
    --platform="windows-mono" \
    --os="ubuntu"
    --user=env:USERNAME \
    --pass=env:PASS \
    --serial=env:SERIAL \
    export --path=./Builds
```

## Setup

**ULF**
Refer to GameCI [documentation](https://game.ci/docs/gitlab/activation#b-locally) on how to active a personal licesnce.  This is how you will retreive your ULF. 

**Build Targets**

```
    StandaloneOSX	Build a macOS standalone (Intel 64-bit).
    StandaloneWindows	Build a Windows 32-bit standalone.
    iOS	Build an iOS player.
    Android	Build an Android .apk standalone app.
    StandaloneWindows64	Build a Windows 64-bit standalone.
    WebGL	Build to WebGL platform.
    WSAPlayer	Build an Windows Store Apps player.
    StandaloneLinux64	Build a Linux 64-bit standalone.
    PS4	Build a PS4 Standalone.
    XboxOne	Build a Xbox One Standalone.
    tvOS	Build to Apple's tvOS platform.
    Switch	Build a Nintendo Switch player.
    LinuxHeadlessSimulation	Build a LinuxHeadlessSimulation standalone.
```

**OS/Platform**

windows:
```
    android
    windows-il2cpp
    universal-windows-platform
    appletv
    base
```

ubuntu:
```
    android
    webgl
    windows-mono
    ios
    mac-mono
    linux-il2cpp
    base
```