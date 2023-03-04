# Lightpicture

Author: Kevin Böckler

Contact: dev@kevinboeckler.de

## Intention

This software works as a proxy service which you can deploy one the same machine or network as your webdav server, e.g.
Nextcloud. The goal is to reduce network traffic by not requesting images from the webdav server itself but from this
proxy server, which will handle image reduction and compression.

## Usage

Once installed, a request is simply made by calling the resource of an imagefile like so:

```curl
curl -X GET \
 -H "Authorization: Basic [[basicHash]]" \
 "http://localhost:8080/{file}?width=1280&height=720"
```

See [http://localhost:8080/openapi]() or file `openapi.yaml` for documentation of the REST endpoint.

This proxy forwards a base authentication header to the webdav endpoint. Make sure you have atleast a SSL encrypted
connection.

## Installation

### Get the binary

Download or compile yourself by running

`env GOOS=linux GOARCH=amd64 go build`

replacing the target runtime as you like.

### Install the binary

You can simply put the binary `lightpicture` (or `lightpicture.exe`) into an arbitray folder on your server. For the
following documentation, I assume it is the same as your Webdav instance.

You could e.g. create a directory `/opt/lightpicture`.

### Configuration

A sample configuration file looks as follows:

```json
{
  "hostname": "my-nextcloud.com",
  "port": 8080,
  "certFile": "/etc/ssl/certs/my-nextcloud.come/fullchain.pem",
  "keyFile": "/etc/ssl/certs/my-nextcloud.com/privkey.pem",
  "baseUrl": "https://my-nextcloud.com",
  "homePath": "/remote.php/dav"
}
```

| Field    | Description                                                                                                                                                                                                                  | Required |
|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| hostname | the hostname of _lightpicture_. Other services will perform requests against it. In production, this will be usually your domain name.                                                                                       | yes      |
| port     | the port of _lightpicture_. In production, this should be an SSL port like 443.                                                                                                                                              | yes      |
| certFile | the path to your public cert file. When provided, the service will start with SSL, also requiring a _keyFile_.                                                                                                               | no       |
| keyFile  | the path to your private key file. When provided, the service will start with SSL, also requiring a _certFile_.                                                                                                              | no       |
| baseUrl  | the base url of your webdav server, which you want to create a proxy for. This is only the base url of the server. In production, this will be usually your domain name. Most likely same to the hostname of _lightpicture_. | yes      |
| homePath | the relative url after your _baseUrl_. This points to the endpoint of a webdav servlet. In a Nextcloud instance, this is `remote.php/dav`.                                                                                   | yes      |

### Security

Since _lightpicture_ takes a Base Authentication header and passes it on to the webdav server, it is recommended to use
SSL in production.

I also recommend creating service user in your webdav instance which has only limited priviliges so even if an attacker
can get your Base Authentication Header, you do not lose any data or expose your system to external threats.

### Installing as service

A service file could be looking as follows:

```
    [Unit]
    Description=Lightpicture endpoint for picture access of nextcloud
    After=network.target remote-fs.target nss-lookup.target
    
    [Service]
    Type=forking
    ExecStart=/opt/lightpicture/start.sh
    ExecStop=/opt/lightpicture/stop.sh
    ExecReload=/opt/lightpicture/restart.sh
    Restart=on-abort
    
    [Install]
    WantedBy=multi-user.target
```

Using additional scripts like so:

start.sh

```shell
    #!/bin/bash
    cd /opt/lightpicture
    ./lightpicture_linux &
```

stop.sh

```shell
    #!/bin/bash
    kill $(pidof lightpicture_linux)
```

restart.sh

```shell
    #!/bin/bash
    cd /opt/lightpicture
    ./stop.sh
    ./start.sh
```

# Contributing

Feel free to make changes or create Feature requests. It is however not my first priority to maintain this project.

# Known Issues

- Most of errors will only be visible by looking into `log.json` since they are treated as HTTP 500 publicly.
