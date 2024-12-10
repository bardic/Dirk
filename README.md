# Dirk

Leverage DaggerCI and GameCI to create local and remote Unity3d builds

## Opinionated

Dirk is opinionated in regards to params but gives you several ways to apply variables to ensure your needs are met.

It does this parsing variables in this order:

- System Environment Variables
- [A local dotenv file](#local-dotenv-files)
- CLI arguments

## [Local dotenv files]

Dirk will check in the root of your project for 4 files:

- `unity.env`
- `unity_secrets.env`
- `unity_test.env`
- `unity_test_secrets.env`

It is recommended that you store what variables you can as code to be committed, excluding the secret dotenvs.

THOSE SHOULD NEVER BE COMMITTED

There is nothing stopping you from placing your secrets in the regular dotenv but if you they will appear as plain text in the logs. The secrets dotenv leverages Dagger Secrets to inject a secret into the build container and scrub it from the logs.

### Example

**unity.env**
```
DIRK_BUILD_NAME=demo
DIRK_BUILD_TARGET=StandaloneWindows64
DIRK_GAMECI_VERSION=3.1.0
DIRK_PLATFORM=windows-mono
DIRK_OS=ubuntu
DIRK_ULF=./License/Unity_v6000.x.ulf
```

**unity_test.env**
```
DIRK_GAMECI_VERSION=3.1.0
DIRK_PLATFORM=windows-mono
DIRK_TESTING_PLATFORM=editor
DIRK_OS=ubuntu
DIRK_ULF=./License/Unity_v6000.x.ulf
DIRK_UNITY_VERSION=6000.0.29f1
DIRK_JUNIT_TRANSFORM=./nunit-transforms/nunit3-junit.xslt
```

**unity_secrets.env**
```
DIRK_PASS=passw0rd
DIRK_USER=me@there.com
```

## Build

`--gameSrc` is the only "required" param. If no params are set, Dirk will assume that these values have been set via the dotenv or as an environment variable.

### dotenv usage
```
dagger call build --game-src=./example/game export --path=./builds
```

### cli arg usage
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

### dotenv usage
```
dagger call test --game-src=./example/game export --path=./tests 
```

### cli arg usage
```
dagger call test
    --game-src="./example/game" \
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
Refer to GameCI [documentation](https://game.ci/docs/gitlab/activation#b-locally) on how to activate a personal licence. This is how you will retrieve your ULF.
