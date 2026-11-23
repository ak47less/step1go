# REST API

1. Web 前端与后端通过基于 HTTP 的API通信;
2. 请求和响应的实体数据使用 JSON 编码, 是一个 JSON 对象;
3. 通过执行名为 "status" 的action, 查询引擎的各种状态;
4. 通过执行名为 "start_engine" 的action, 启动引擎;
5. 通过执行名为 "stop_engine" 的action, 停止引擎;

## 请求 JSON 结构

    {
        "actions":[{
            "name":"start_engine",
            "params":{

            }
        },{
            "name":"status",
            "all":false,
            "keys":["a","b","c"]
        }]
    }

## 响应 JSON 结构

    {
        "status":{
            "a":"1",
            "b":"2",
            "c":"3",
        }
    }

## Action 表

| name         | 功能                     | 说明        |
| ------------ | ------------------------ | ----------- |
| start_engine | 启动引擎                 |             |
| stop_engine  | 停止引擎                 |             |
| play         | 开始播放                 |             |
| record       | 开始录制                 |             |
| pause        | 暂停                     |             |
| stop         | 停止播放                 |             |
| seek         | 把当前位置设置为指定时间 | 时间线@文档 |

## Status 键值表

| name                | 值  | 说明             |
| ------------------- | --- | ---------------- |
| transport.playing   | 0/1 | 是否正在播放     |
| transport.paused    | 0/1 | 是否处于暂停状态 |
| transport.recording | 0/1 | 是否正在录制     |
