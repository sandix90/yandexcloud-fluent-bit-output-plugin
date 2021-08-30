#Description
This plugin is used as an output fluent-bit plugin. The plugin should be compiled as a `.so` file with following command.
```shell
go build -buildmode=c-shared -o out_yandexcloud.so yandex_cloud_logging.go
```

## Run in Docker
If you want to test this plugin in docker here is example you can use
```shell
FROM golang:1.16 as gobuilder

WORKDIR /root

ENV GOOS=linux\
    GOARCH=amd64

COPY / /root

RUN go mod download & make all

FROM fluent/fluent-bit:1.4

COPY --from=gobuilder /root/bin/out_yandexcloud.so /fluent-bit/bin/
#COPY --from=gobuilder /root/external/fluent-bit.conf /fluent-bit/etc/
#COPY --from=gobuilder /root/external/plugins.conf /fluent-bit/etc/


EXPOSE 2020

CMD ["/fluent-bit/bin/fluent-bit", "--config", "/fluent-bit/etc/fluent-bit.conf"]
```

You might want to start docker container and pass your folder to the container. Use this command:
```shell
docker run --rm -v *path_to_configs_folder*:/fluent-bit/etc --name fluent_yandex_cloud fluent-bit-yandex-cloud
```
### Note
If you want to have your configs copied into the container, please uncomment both `COPY` commands in the `Dockerfile`


## Plugin options

* `endpoint_url` - `(optional)` `string` yandex url to write logs. Leave empty to use `default` - `https://logging.api.cloud.yandex.net/logging/v1/write`
* `log_group_id` - `(optional)` `string` id of yandex log group
* `folder_id` - `(optional)` `string` id of folder id
* `resource_id` - `(optional)` `string` field for yandex logging record
* `resource_type` - `(optional)` `string` field for yandex logging record
* `key_id` - `(required)` `string` id of the key for getting iam-token
* `service_account_id` - `(required)` `string` id of the yandex service account 
* `private_key_file_path` - `(required)` `string` private ket path of the yandex auth key
* `log_level_key` - `(optional)` `string` name of the level log field. `default` - `level`

### Note
Either folder_id or log_group_id should have been created and properly configured.


How to generate protoc in case you need it:
```shell
protoc -I ./third_party/googleapis -I . --go_out=paths=source_relative:. yandex/cloud/logging/v1/*.proto 
```