## Device APIs

设备管理相关的 API 说明

----

### Pre Requests:
在创建设备前需要创建一个租户，因为所有请求 `header` 都需要带上这个字段作为认证信息。

### 说明
创建设备之后， `tkeel` 会返回设备token, 用户需要保存这个token信息， 并使用该 token 作为 `MQTT` 接入时的 密码。

>注意
下面所有的请求都 header 都需要带上 `Authorization` 租户 token。

### 示例
-
### 创建设备 Entity
- Method: **POST**
- Request:

 ```sh
curl --location --request POST '192.168.123.12:30707/apis/tkeel-device/v1/devices' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2NDE3ODEwNTEsInN1YiI6InVzci00LTlhNGRmOWNlNjA0ZTgwNDRmZmYwZDM2MTUzOTQ3NDVmIn0.wg_uUwCV3iZHFjMV4SJlu8mrlxFIm9vq7vrcRli4ouTFq643uGc7SmwTLn3LVqWcTTLDupes9LODPDD3kBQEZQ' \
--header 'Content-Type: application/json' \
--data-raw '{
  "name": "dev1",
  "desc": "description info",
  "group": "default",
  "ext": {
      "address": "http://xxx.yyy.com",
      "alias": "dev"
  }
}'
```
- Response:
```json
{
    "dev": {
        "name": "dev1",
        "desc": "description info",
        "group": "default",
        "ext": {
            "address": "http://xxx.yyy.com",
            "alias": "dev"
        }
    },
    "sysField": {
        "_id": "4e901bc2-927b-4d4f-8a0e-25fa32a66ada",
        "_createdAt": 1638277082172443100,
        "_updatedAt": 1638277082172443100,
        "_enable": true,
        "_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbnRpdHlfaWQiOiI0ZTkwMWJjMi05MjdiLTRkNGYtOGEwZS0yNWZhMzJhNjZhZGEiLCJlbnRpdHlfdHlwZSI6ImRldmljZSIsImV4cCI6MTY2OTgxMzA4Miwib3duZXIiOiJ1c3ItNC05YTRkZjljZTYwNGU4MDQ0ZmZmMGQzNjE1Mzk0NzQ1ZiJ9.bzwafQsQFJbCYdfvlRuGJUciamsyre86-MQnyxOi75PmZfOqGEX-MDdcXkowpOL0SG2Xoc851oxotgqWGneotA"
    }
}
```
**Params：**

| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| name | string | true | body | 设备名称。 | 
| desc | string | false | body | 设备描述信息。|
| group | string | false | body | 设备组名称。|
| ext | string | false | body | 设备扩展信息。|
| _id | string | | body | 设备ID。|
| _createdAt | int64 | | body | 设备创建时间。|
| _updatedAt | int64 | | body | 设备最后修改时间。|
| _enable | bool | | body | 设备使能。|
| _token | string | | body | 设备token。|


### 获取设备 Entity 详情
- Method: **GET**
- Request:

 ```sh
curl --location --request GET '192.168.123.12:30707/apis/tkeel-device/v1/devices/4e901bc2-927b-4d4f-8a0e-25fa32a66ada' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2NDE3ODEwNTEsInN1YiI6InVzci00LTlhNGRmOWNlNjA0ZTgwNDRmZmYwZDM2MTUzOTQ3NDVmIn0.wg_uUwCV3iZHFjMV4SJlu8mrlxFIm9vq7vrcRli4ouTFq643uGc7SmwTLn3LVqWcTTLDupes9LODPDD3kBQEZQ'
```
- Response:
```json
{
  "dev": {
    "name": "dev1",
    "desc": "description info",
    "group": "default",
    "ext": {
      "address": "http://xxx.yyy.com",
      "alias": "dev"
    }
  },
  "sysField": {
    "_id": "4e901bc2-927b-4d4f-8a0e-25fa32a66ada",
    "_createdAt": 1638277082172443100,
    "_updatedAt": 1638277082172443100,
    "_enable": true,
    "_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbnRpdHlfaWQiOiI0ZTkwMWJjMi05MjdiLTRkNGYtOGEwZS0yNWZhMzJhNjZhZGEiLCJlbnRpdHlfdHlwZSI6ImRldmljZSIsImV4cCI6MTY2OTgxMzA4Miwib3duZXIiOiJ1c3ItNC05YTRkZjljZTYwNGU4MDQ0ZmZmMGQzNjE1Mzk0NzQ1ZiJ9.bzwafQsQFJbCYdfvlRuGJUciamsyre86-MQnyxOi75PmZfOqGEX-MDdcXkowpOL0SG2Xoc851oxotgqWGneotA"
  }
}
```


### 更新设备 Entity
- Method: **PUT**
- Request:
 ```sh
 curl --location --request PUT '192.168.123.12:30707/apis/tkeel-device/v1/devices/4e901bc2-927b-4d4f-8a0e-25fa32a66ada' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2NDE3ODEwNTEsInN1YiI6InVzci00LTlhNGRmOWNlNjA0ZTgwNDRmZmYwZDM2MTUzOTQ3NDVmIn0.wg_uUwCV3iZHFjMV4SJlu8mrlxFIm9vq7vrcRli4ouTFq643uGc7SmwTLn3LVqWcTTLDupes9LODPDD3kBQEZQ' \
--header 'Content-Type: application/json' \
--data-raw '{
  "name": "dev1",
  "desc": "description info changed",
  "group": "default",
  "ext": {
      "address": "http://xxx.yyy.com",
      "alias": "dev2"
  }
}'
```
- Response:
```json
{
  "dev": {
    "name": "dev1",
    "desc": "description info changed",
    "group": "default",
    "ext": {
      "address": "http://xxx.yyy.com",
      "alias": "dev2"
    }
  },
  "sysField": {
    "_updatedAt": 1638279110297178000
  }
}
```


### 获取设备 Entity 列表
- Method: **POST**
- Request:
 ```sh
curl --location --request POST '192.168.123.12:30707/apis/tkeel-device/v1/devices/search' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2NDE3ODEwNTEsInN1YiI6InVzci00LTlhNGRmOWNlNjA0ZTgwNDRmZmYwZDM2MTUzOTQ3NDVmIn0.wg_uUwCV3iZHFjMV4SJlu8mrlxFIm9vq7vrcRli4ouTFq643uGc7SmwTLn3LVqWcTTLDupes9LODPDD3kBQEZQ' \
--header 'Content-Type: application/json' \
--data-raw '{
    "page": {
        "offset": 1,
        "limit": 3,
        "sort": "name",
        "reverse": false
    }
}'
```
- Response:
```json
{
    "result": {
        "items": [
            {
                "id": "7d04ed27-fd08-4585-8631-ef99f6d37717",
                "plugin": "device",
                "properties": {
                    "dev": "{}",
                    "id": "7d04ed27-fd08-4585-8631-ef99f6d37717",
                    "last_time": 1638181884820,
                    "name": "dddd",
                    "owner": "",
                    "source": "",
                    "type": "",
                    "version": 1
                }
            },
            {
                "id": "a61a3540-f6fd-4a0e-ad7a-519b8139436c",
                "plugin": "device",
                "properties": {
                    "dev": "{}",
                    "id": "a61a3540-f6fd-4a0e-ad7a-519b8139436c",
                    "last_time": 1638181918578,
                    "name": "dddd",
                    "owner": "",
                    "source": "",
                    "type": "",
                    "version": 1
                }
            },
            {
                "id": "27422200-6188-4480-b596-91308ee09882",
                "plugin": "device",
                "properties": {
                    "dev": "{}",
                    "id": "27422200-6188-4480-b596-91308ee09882",
                    "last_time": 1638181955486,
                    "name": "dddd",
                    "owner": "",
                    "source": "",
                    "type": "",
                    "version": 1
                }
            }
        ],
        "limit": 3,
        "total": 17
    }
}
```


### 删除设备 Entity
- Method: **PUT**
- Request:
 ```sh
curl --location --request POST '192.168.123.12:30707/apis/tkeel-device/v1/devices/delete' \
--header 'Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIwMDAwMDAiLCJleHAiOjE2NDE3ODEwNTEsInN1YiI6InVzci00LTlhNGRmOWNlNjA0ZTgwNDRmZmYwZDM2MTUzOTQ3NDVmIn0.wg_uUwCV3iZHFjMV4SJlu8mrlxFIm9vq7vrcRli4ouTFq643uGc7SmwTLn3LVqWcTTLDupes9LODPDD3kBQEZQ' \
--header 'Content-Type: application/json' \
--data-raw '{
    "ids": ["4e901bc2-927b-4d4f-8a0e-25fa32a66ada"]
}'
```
- Response:
```json
{"result":"ok"}
```
