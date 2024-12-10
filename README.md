# Dirk

Leverage DaggerCI and GameCI to create local and remote Unity3d builds

## Opiniated

Dirk is opiniated in regards to params but gives you several ways to apply variables to ensure your needs are met.

It does this parsing variables in this order:

- System Environment Variables
- [A local dotenv file](#local-dotenv-files)
- CLI argumenets

## [Local dotenv files]

Dirk will check in the root of your project for 4 files:

- `unity.env`
- `unity_secrets.env`
- `unity_test.env`
- `unity_test_secrets.env`

It is recommended that you store what variables you can as code to be commited, excluding the secret dotenvs.

THOSE SHOULD NEVER BE COMMITED

There is nothing stopping you from placing your secrets in the regular dotenv but if you they will appear as plain text in the logs. The secrets dotenv leverages Dagger Secrets to inject a secret into the build container and scrub it from the logs.

## Build

`--gameSrc` is the only "required" param. If no params are set, Dirk will assume that these values have been set via the dotenv or as an environment variable.

```
dagger call build \
    --game-src="./example/game" \
    --build-name="demo" \
    --build-target="StandaloneOSX|StandaloneWindows|iOS|Android|StandaloneWindows64|WebGL|StandaloneLinux64|tvOS" \
    --gameci-version="3.1.0" \
    --pass="DONT-PASS-IT-PLAIN-TEXT-BUT-YOU-CAN" \
    --platform="android|webgl|windows-mono|ios|mac-mono|linux-il2cpp|base" \
    --serial="DONT-PASS-IT-PLAIN-TEXT-BUT-YOU-CAN" \
    --service-config="./services-config.json" \
    --target-os="ubuntu|windows" \
    --ulf="./Unity_v6000.x.ulf" \
    --unity-version="6000.0.29f1" \
    --user="email@address.com" \
    export --path=./builds
```

## Test

```
dagger call test
    --game-src="./example/game" \
    --build-name="demo" \
    --build-target="StandaloneOSX|StandaloneWindows|iOS|Android|StandaloneWindows64|WebGL|StandaloneLinux64|tvOS" \
    --gameci-version="3.1.0" \
    --junit-transform="./nunit-transforms/nunit3-junit.xslt" \
    --pass="DONT-PASS-IT-PLAIN-TEXT-BUT-YOU-CAN" \
    --platform="android|webgl|windows-mono|ios|mac-mono|linux-il2cpp|base" \
    --serial="DONT-PASS-IT-PLAIN-TEXT-BUT-YOU-CAN" \
    --service-config="./services-config.json" \
    --target-os="ubuntu|windows" \
    --testinging-platform="editor|play" \
    --ulf="./Unity_v6000.x.ulf" \
    --unity-version="6000.0.29f1" \
    --user="email@address.com" \
    export --path=./tests
```

## Setup

**ULF**
Refer to GameCI [documentation](https://game.ci/docs/gitlab/activation#b-locally) on how to active a personal licesnce. This is how you will retreive your ULF.
