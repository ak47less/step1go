package common

type SampleValue float32 //  采样值, 标准取值范围是 [-1,1] (0dB) , 实际值可能大于 0dB

type SampleCount uint64 // 采样点计数, 通常用于表示位置(position)或长度(length)

type SampleRate uint32 // 采样率, 表示每秒的采样点数
