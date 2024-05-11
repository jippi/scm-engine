# Installation

## Docker

```shell
docker run --rm ghcr.io/jippi/scm-engine
```

## homebrew tap

```shell
brew install jippi/tap/scm-engine
```

## apt

```shell
echo 'deb [trusted=yes] https://pkg.jippi.dev/apt/ * *' | sudo tee /etc/apt/sources.list.d/scm-engine.list
sudo apt update
sudo apt install scm-engine
```

## yum

```shell
echo '[scm-engine]
name=scm-engine
baseurl=https://pkg.jippi.dev/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/scm-engine.repo
sudo yum install scm-engine
```

## snapcraft

```shell
sudo snap install scm-engine
```

## scoop

```shell
scoop bucket add scm-engine https://github.com/jippi/scoop-bucket.git
scoop install scm-engine
```

## aur

```shell
yay -S scm-engine-bin
```

## deb, rpm and apk packages

Download the `.deb`, `.rpm` or `.apk` packages from the [releases page](https://github.com/jippi/scm-engine/releases) and install them with the appropriate tools.

## go install

```shell
go install github.com/jippi/scm-engine/cmd@latest
```
