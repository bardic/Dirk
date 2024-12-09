# Local Game CI

Leverage DaggerCI and GameCI to create local and remote builds


## Build

```
// Build unity project with a personal license targeting Windows Mono on Ubuntu
dagger call test --src="./example/game" \
    --ulf="./Unity_v6000.x.ulf" \
    --build-target="StandaloneWindows64" \
    --build-name="demo" \
    --platform="windows-mono" \
    --os="ubuntu" \
    --user=env:USER \
    --pass=env:PASS \
    export ./builds
```
```
// Build unity project with a User and Serail targeting Windows Mono on Ubuntu
dagger call test --src="./example/game" \
    --build-target="StandaloneWindows64" \
    --build-name="demo" \
    --platform="windows-mono" \
    --os="ubuntu" \
    --user=env:USER \
    --pass=env:PASS \
    --serial=env:SERIAL \
    export ./builds
```

```
// Build unity project with Service Config (float license) targeting Windows Mono on Ubuntu
dagger call test --src="./example/game" \
    --build-target="StandaloneWindows64" \
    --build-name="demo" \
    --platform="windows-mono" \
    --os="ubuntu" \
    --user=env:USER \
    --pass=env:PASS \
    --service-config="./service-config.json" \
    export ./builds
```

## Test
```
// Test unity project with personal license targeting Windows Mono on Ubuntu
dagger call test \
    --src="./example/game" \
    --user=env:USER \
    --platform="windows-mono" \
    --build-target="StandaloneWindows64" \
    --os="ubuntu" \
    --build-name="demo" \
    --testinging-platform="editor" \
    --pass=env:PASS \
    --junitTransform="/nunit-transforms/nunit3-junit.xslt" \
    --ulf="./Unity_v6000.x.ulf" \
    export ./results
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