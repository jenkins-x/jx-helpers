### Linux

```shell
curl -L https://github.com/jenkins-x/jx-helpers/v3/releases/download/v{{.Version}}/jx-helpers-linux-amd64.tar.gz | tar xzv 
sudo mv jx-helpers /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-helpers/v3/releases/download/v{{.Version}}/jx-helpers-darwin-amd64.tar.gz | tar xzv
sudo mv jx-helpers /usr/local/bin
```

