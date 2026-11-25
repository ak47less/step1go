# STEP1 Lite2 REST API（草案）

本文档定义 Web 前端与 STEP1 Lite2 后端之间的 HTTP/JSON 接口约定，接口范围来自当前
`lib` 中的引擎、走带、状态、Cube（bank/step/track/cell）和 MIDI 相关能力。

> 实现状态：`ui/restful` 已按本文提供基于标准库 `net/http` 的 HTTP 服务，服务器
> 由 `WebServerConfig` 控制启用、主机和端口。引擎与走带命令最终由
> `ActionController` 的后端实现执行。

## 1. 通用约定

### 1.1 基础信息

| 项目 | 约定 |
| --- | --- |
| Base URL | `http://localhost:8080/api/v1` |
| 数据格式 | 请求和响应实体均为 JSON 对象，编码为 UTF-8 |
| 请求头 | 写操作使用 `Content-Type: application/json` |
| 资源标识 | bank、step、track 使用从 `0` 开始的整数索引 |
| 时间单位 | 秒，使用 JSON 数字；允许小数 |
| 认证 | 草案暂不定义认证；如需远程访问，应在部署层增加认证和访问控制 |

不带请求体的操作也应发送空对象：

```json
{}
```

### 1.2 路径与状态码

- `GET` 用于读取资源，不应改变引擎状态。
- `POST` 用于执行不可自然映射为资源替换的操作，例如启动引擎或开始播放。
- `PATCH` 用于部分更新资源。
- `PUT` 用于完整设置可替换的状态，例如设置走带位置。
- 成功读取返回 `200 OK`；成功执行命令返回 `200 OK`。
- 参数错误返回 `400 Bad Request`；资源不存在返回 `404 Not Found`；当前状态不允许
  该操作返回 `409 Conflict`；音频驱动或引擎不可用返回 `503 Service Unavailable`。

### 1.3 错误响应

所有非 `2xx` 响应使用统一结构：

```json
{
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "position must be greater than or equal to 0",
    "details": {
      "field": "position"
    }
  }
}
```

`code` 是供程序判断的稳定字符串，`message` 供展示或日志使用，`details` 可以省略。

## 2. 接口总览

| 方法 | 路径 | 功能 |
| --- | --- | --- |
| `GET` | `/health` | 检查 HTTP 服务是否可用 |
| `GET` | `/status` | 查询引擎、走带和音频上下文状态 |
| `GET` | `/engine` | 查询引擎状态 |
| `POST` | `/engine/start` | 启动引擎 |
| `POST` | `/engine/stop` | 停止引擎 |
| `POST` | `/engine/restart` | 重启引擎 |
| `GET` | `/transport` | 查询走带状态和当前位置 |
| `POST` | `/transport/play` | 开始播放 |
| `POST` | `/transport/record` | 开始录制 |
| `POST` | `/transport/pause` | 暂停 |
| `POST` | `/transport/resume` | 从暂停状态恢复 |
| `POST` | `/transport/stop` | 停止播放或录制 |
| `PUT` | `/transport/position` | 设置播放位置 |
| `GET` | `/cube` | 查询 Cube 的规模和当前时间线信息 |
| `GET` | `/banks` | 列出 bank |
| `GET` | `/banks/{bankIndex}` | 查询单个 bank |
| `PATCH` | `/banks/{bankIndex}` | 更新 bank 的显示信息 |
| `GET` | `/steps` | 列出 step |
| `GET` | `/steps/{stepIndex}` | 查询单个 step |
| `GET` | `/tracks` | 列出 track |
| `GET` | `/tracks/{trackIndex}` | 查询单个 track |
| `PATCH` | `/tracks/{trackIndex}` | 更新 track 的混音和显示参数 |
| `GET` | `/cells` | 列出或筛选 cell |
| `GET` | `/cells/{cellName}` | 按名称查询 cell |
| `PATCH` | `/cells/{cellName}` | 设置或清除 cell 的 MIDI note |
| `POST` | `/midi/events` | 批量发送 MIDI 事件到后端 |
| `POST` | `/actions` | 批量执行兼容 action |

## 3. 服务状态

### 3.1 健康检查

`GET /health`

```json
{
  "ok": true,
  "version": "0.0.6"
}
```

### 3.2 状态查询

`GET /status`

可选查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `keys` | string | 逗号分隔的状态键；省略或为空时返回全部状态 |

示例：

```http
GET /api/v1/status?keys=engine.started,transport.playing
```

```json
{
  "status": {
    "engine.started": 1,
    "transport.playing": 0
  }
}
```

可用状态键：

| 状态键 | 值 | 说明 |
| --- | --- | --- |
| `engine.starting` | `0/1` | 是否正在启动 |
| `engine.started` | `0/1` | 是否已经启动 |
| `engine.stopping` | `0/1` | 是否正在停止 |
| `engine.stopped` | `0/1` | 是否已经停止 |
| `audio.state` | string | 音频上下文状态，例如 `running`、`suspended` |
| `transport.playing` | `0/1` | 是否正在播放 |
| `transport.paused` | `0/1` | 是否处于暂停状态 |
| `transport.recording` | `0/1` | 是否正在录制 |
| `transport.position` | number | 当前播放位置，单位为秒 |

客户端应按 JSON 数字解析 `0/1`。只请求部分键时，响应只包含实际存在的键；未知键
不返回，调用方应根据缺失项或接口返回的错误决定是否降级。

## 4. 引擎控制

### 4.1 查询引擎

`GET /engine`

```json
{
  "state": "started",
  "starting": false,
  "started": true,
  "stopping": false,
  "stopped": false,
  "audio": {
    "state": "running"
  }
}
```

`state` 的取值为 `stopped`、`starting`、`started`、`stopping` 之一。

### 4.2 启动引擎

`POST /engine/start`

请求体可以为空对象。重复启动已启动的引擎返回 `409 Conflict`。

```json
{}
```

响应返回新的引擎对象，结构同 `GET /engine`。

### 4.3 停止引擎

`POST /engine/stop`

请求体可以为空对象。该操作会先停止走带，再释放音频资源。

```json
{}
```

### 4.4 重启引擎

`POST /engine/restart`

请求体可以为空对象。服务按“停止（如正在运行）→ 启动”的顺序执行；走带位置恢复
到 `0`，播放、暂停和录制状态清零。

```json
{}
```

## 5. 走带控制

### 5.1 查询走带

`GET /transport`

```json
{
  "state": "playing",
  "playing": true,
  "paused": false,
  "recording": false,
  "position": 12.5,
  "timeline": "@document"
}
```

`state` 的取值为 `stopped`、`playing`、`recording`、`paused` 之一。

### 5.2 播放

`POST /transport/play`

从当前位置开始播放。引擎未启动时返回 `409 Conflict`。请求体可以为空对象。

### 5.3 录制

`POST /transport/record`

从当前位置开始录制。录制状态同时包含走带推进；引擎未启动时返回 `409 Conflict`。
请求体可以为空对象。

### 5.4 暂停

`POST /transport/pause`

仅在播放或录制时有效，保留当前位置。其他状态返回 `409 Conflict`。请求体可以为空
对象。

### 5.5 恢复

`POST /transport/resume`

从暂停状态继续原来的播放或录制。非暂停状态返回 `409 Conflict`。请求体可以为空
对象。

### 5.6 停止

`POST /transport/stop`

停止播放和录制，走带位置保持不变。请求体可以为空对象。

### 5.7 设置播放位置

`PUT /transport/position`

```json
{
  "position": 12.5,
  "timeline": "@document"
}
```

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `position` | number | 是 | 相对于 `timeline` 的位置，单位为秒，必须大于等于 `0` |
| `timeline` | string | 否 | 时间线名称，默认 `@document` |

成功响应返回新的走带对象，结构同 `GET /transport`。位置超出文档长度或时间线不存在
时返回 `422 Unprocessable Entity`。

## 6. Cube 与轨道

Cube 由 `bank × step × track` 组成。当前配置中的可用长度为 4 个 bank、16 个 step
和 6 个 track；容量分别为 8、32 和 12。容量在启动时固定，本版 API 不提供创建或
删除 bank、step、track 的接口。

### 6.1 查询 Cube

`GET /cube`

```json
{
  "banks": {
    "length": 4,
    "capacity": 8
  },
  "steps": {
    "length": 16,
    "capacity": 32
  },
  "tracks": {
    "length": 6,
    "capacity": 12
  },
  "timelines": {
    "document": "@document",
    "transport": "@transport",
    "step": "@step",
    "frame": "@frame"
  }
}
```

### 6.2 Bank

`GET /banks`

```json
{
  "items": [
    {
      "index": 0,
      "label": "bank_0",
      "description": ""
    }
  ]
}
```

`GET /banks/{bankIndex}` 返回单个 bank 对象。

`PATCH /banks/{bankIndex}` 更新显示信息：

```json
{
  "label": "Verse",
  "description": "First verse"
}
```

至少提供一个字段。成功响应返回更新后的 bank 对象。

### 6.3 Step

`GET /steps` 返回当前可用的 step：

```json
{
  "items": [
    {
      "index": 0
    },
    {
      "index": 1
    }
  ]
}
```

`GET /steps/{stepIndex}` 返回单个 step。当前模型没有可修改的 step 字段。

### 6.4 Track

`GET /tracks`

```json
{
  "items": [
    {
      "index": 0,
      "name": "",
      "label": "track_0",
      "description": "",
      "order": 0,
      "enabled": false,
      "solo": false,
      "mute": false,
      "gainDb": 0,
      "channel": 0
    }
  ]
}
```

`GET /tracks/{trackIndex}` 返回单个 track 对象。

`PATCH /tracks/{trackIndex}` 支持以下字段的部分更新：

| 字段 | 类型 | 约束 |
| --- | --- | --- |
| `name` | string | 内部名称 |
| `label` | string | 界面显示名称 |
| `description` | string | 描述 |
| `order` | integer | 显示顺序 |
| `enabled` | boolean | 是否参与处理 |
| `solo` | boolean | 是否独奏 |
| `mute` | boolean | 是否静音 |
| `gainDb` | number | 增益，单位为 dB |
| `channel` | integer | MIDI 通道；`-1` 表示 omni |

示例：

```json
{
  "label": "Bass",
  "enabled": true,
  "gainDb": -6
}
```

成功响应返回更新后的完整 track 对象。字段类型错误返回 `400 Bad Request`，track
索引越界返回 `404 Not Found`。

### 6.5 Cell

cell 的默认名称格式为 `b{bankIndex}_s{stepIndex}_t{trackIndex}`，例如
`b0_s3_t2`。

`GET /cells`

可选筛选参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `bankIndex` | integer | bank 索引 |
| `stepIndex` | integer | step 索引 |
| `trackIndex` | integer | track 索引 |

```http
GET /api/v1/cells?bankIndex=0&stepIndex=3
```

```json
{
  "items": [
    {
      "name": "b0_s3_t0",
      "bankIndex": 0,
      "stepIndex": 3,
      "trackIndex": 0,
      "note": null
    },
    {
      "name": "b0_s3_t1",
      "bankIndex": 0,
      "stepIndex": 3,
      "trackIndex": 1,
      "note": {
        "number": 60,
        "velocity": 100
      }
    }
  ]
}
```

`GET /cells/{cellName}` 返回单个 cell：

```json
{
  "name": "b0_s3_t2",
  "bankIndex": 0,
  "stepIndex": 3,
  "trackIndex": 2,
  "note": {
    "number": 60,
    "velocity": 100
  }
}
```

`note` 为 `null` 表示该 cell 尚未设置音符；非空时是 MIDI note 对象：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `number` | integer | MIDI 音高编号，范围为 `0..127` |
| `velocity` | integer | MIDI 力度，范围为 `0..127` |

`PATCH /cells/{cellName}` 用于设置或清除 `cell.note`：

```json
{
  "note": {
    "number": 60,
    "velocity": 100
  }
}
```

`note` 字段必须提供；`number` 和 `velocity` 均为必填整数，缺失或超出范围时返回
`400 Bad Request`。提交 note 对象会整体替换当前值。传 `null` 可清除 note：

```json
{
  "note": null
}
```

成功响应返回更新后的完整 cell 对象，结构同 `GET /cells/{cellName}`；cell 名称不存在
时返回 `404 Not Found`。当前接口仅支持编辑 `note`，不接受其他 cell 内容字段。

## 7. MIDI 事件接口

### 7.1 批量发送 MIDI 事件

`POST /midi/events`

批量发送 MIDI 事件到当前引擎的输入乐器。请求中的事件按数组顺序处理；负的
`time` 表示立即触发，非负 `time` 表示在时间轴上的绝对时刻触发，单位为秒。
建议使用 `-1` 表示立即触发。

```json
{
  "events": [
    {
      "type": "note_on",
      "time": -1,
      "channel": 0,
      "note": {
        "number": 60,
        "velocity": 100
      }
    },
    {
      "type": "note_off",
      "time": 0.5,
      "channel": 0,
      "note": {
        "number": 60,
        "velocity": 0
      }
    },
    {
      "type": "cc",
      "time": -1,
      "channel": 1,
      "cc": {
        "key": 74,
        "value": 127
      }
    }
  ]
}
```

请求字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `events` | array | 是 | 事件数组，至少包含一个事件 |
| `events[].type` | string | 是 | `note_on`、`note_off` 或 `cc` |
| `events[].time` | number | 否 | 触发时刻，单位为秒；省略或为负数时立即触发 |
| `events[].channel` | integer | 是 | MIDI 通道，范围为 `0..15` |
| `events[].note` | object | 条件必填 | `note_on` 和 `note_off` 事件必须提供 |
| `events[].cc` | object | 条件必填 | `cc` 事件必须提供 |

`note` 对象：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `number` | integer | 是 | MIDI 音高编号，范围为 `0..127` |
| `velocity` | integer | 是 | MIDI 力度，范围为 `0..127` |

`cc` 对象：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `key` | integer | 是 | MIDI 控制器编号，范围为 `0..127` |
| `value` | integer | 是 | 控制器值，范围为 `0..127` |

每个事件只能提供与其 `type` 对应的数据：`note_on` 和 `note_off` 提供 `note`，
`cc` 提供 `cc`。事件缺少必要数据、字段类型错误、数值越界或包含未知字段时返回
`400 Bad Request`。

服务会先校验整批请求，校验通过后再按顺序交给乐器。引擎未启动或当前状态不允许
接收 MIDI 事件时返回 `409 Conflict`；乐器不可用返回 `503 Service Unavailable`；
乐器拒绝事件返回 `422 Unprocessable Entity`，错误码为 `MIDI_DISPATCH_FAILED`。

成功响应：

```json
{
  "accepted": 3,
  "results": [
    {
      "index": 0,
      "type": "note_on",
      "ok": true
    },
    {
      "index": 1,
      "type": "note_off",
      "ok": true
    },
    {
      "index": 2,
      "type": "cc",
      "ok": true
    }
  ]
}
```

`accepted` 为成功提交给乐器的事件数量；`results` 与请求中的 `events` 一一对应。
调用方可以一次发送成对的 `note_on`/`note_off`，但服务不会自动生成 `note_off`
事件，也不会在批次结束后停止音符。

## 8. 批量 Action 接口（兼容协议）

批量 action 沿用现有的 `actions` 数组和 action 名称，适合前端在一次交互中同时执行
多个操作和状态查询。本草案补充统一的 `results` 字段，并保留顶层 `status` 对象。

### 8.1 请求

`POST /actions`

```json
{
  "actions": [
    {
      "name": "start_engine",
      "params": {}
    },
    {
      "name": "status",
      "params": {
        "all": false,
        "keys": [
          "engine.started",
          "transport.playing"
        ]
      }
    }
  ]
}
```

`actions` 按数组顺序串行执行。遇到第一个失败后不再执行后续 action；已完成 action
不会回滚。批量请求体结构错误返回 `400 Bad Request`。

为兼容旧版文档，`status` action 仍允许把 `all` 和 `keys` 直接放在 action 对象上；
该写法与放在 `params` 中完全等价，新客户端应优先使用 `params`：

```json
{
  "name": "status",
  "all": false,
  "keys": [
    "engine.started"
  ]
}
```

### 8.2 Action 表

| `name` | `params` | 功能 |
| --- | --- | --- |
| `start_engine` | `{}` | 启动引擎 |
| `stop_engine` | `{}` | 停止引擎 |
| `restart_engine` | `{}` | 重启引擎 |
| `play` | `{}` | 开始播放 |
| `record` | `{}` | 开始录制 |
| `pause` | `{}` | 暂停 |
| `resume` | `{}` | 从暂停恢复 |
| `stop` | `{}` | 停止播放或录制 |
| `seek` | `{"position": number, "timeline": string}` | 设置播放位置 |
| `status` | `{"all": boolean, "keys": string[]}` | 查询状态 |

示例 seek action：

```json
{
  "name": "seek",
  "params": {
    "position": 12.5,
    "timeline": "@document"
  }
}
```

### 8.3 响应

```json
{
  "results": [
    {
      "index": 0,
      "name": "start_engine",
      "ok": true
    },
    {
      "index": 1,
      "name": "status",
      "ok": true
    }
  ],
  "status": {
    "engine.started": 1,
    "transport.playing": 0
  }
}
```

`status` 仅在请求中包含 `status` action 时返回。action 执行失败时，响应使用对应的
HTTP 状态码，并在 `results` 末尾包含失败项：

```json
{
  "error": {
    "code": "INVALID_STATE",
    "message": "engine must be started before playing"
  },
  "results": [
    {
      "index": 0,
      "name": "start_engine",
      "ok": true
    },
    {
      "index": 1,
      "name": "play",
      "ok": false,
      "error": {
        "code": "INVALID_STATE",
        "message": "engine must be started before playing"
      }
    }
  ]
}
```

## 9. 请求示例

启动引擎并开始播放：

```bash
curl -X POST http://localhost:8080/api/v1/engine/start \
  -H 'Content-Type: application/json' \
  -d '{}'

curl -X POST http://localhost:8080/api/v1/transport/play \
  -H 'Content-Type: application/json' \
  -d '{}'
```

查询状态：

```bash
curl 'http://localhost:8080/api/v1/status?keys=engine.started,transport.playing'
```

更新轨道：

```bash
curl -X PATCH http://localhost:8080/api/v1/tracks/0 \
  -H 'Content-Type: application/json' \
  -d '{"label":"Bass","enabled":true,"gainDb":-6}'
```

设置 cell 的 MIDI note：

```bash
curl -X PATCH http://localhost:8080/api/v1/cells/b0_s3_t2 \
  -H 'Content-Type: application/json' \
  -d '{"note":{"number":60,"velocity":100}}'
```

批量发送 MIDI 事件：

```bash
curl -X POST http://localhost:8080/api/v1/midi/events \
  -H 'Content-Type: application/json' \
  -d '{
    "events": [
      {
        "type": "note_on",
        "time": -1,
        "channel": 0,
        "note": {"number": 60, "velocity": 100}
      },
      {
        "type": "note_off",
        "time": -1,
        "channel": 0,
        "note": {"number": 60, "velocity": 0}
      }
    ]
  }'
```

## 10. 后续扩展

以下能力已有底层抽象，但当前实现尚未完整提供，建议作为后续 API 版本补充：

- 乐器注册表查询：`GET /instruments`。
- 为 track 绑定、替换或卸载乐器。
- cell 的 CC 和事件时间读写。
- 工程文件、配置和会话的导入导出。
- 基于 WebSocket 或 Server-Sent Events 的状态推送，避免前端频繁轮询 `/status`。
