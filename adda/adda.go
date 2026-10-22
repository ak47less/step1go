package main

import (
	"fmt"
	"log"
	"math"
	"time"
	"unsafe"

	"github.com/gen2brain/malgo"
)

func main() {
	Demo()
}

func Demo() {

	// malgo.

	// 1. 初始化上下文
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	// 2. 配置设备参数（低延迟关键：小周期/小缓冲区）
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = 48000
	deviceConfig.PeriodSizeInFrames = 128 // 低延迟设置，约 2.7ms
	deviceConfig.Periods = 2

	// 3. 准备音频生成参数
	const freq = 440.0 // A4 音高
	var phase float64

	// 4. 定义音频回调（在实时线程中执行，必须无阻塞、无内存分配）
	onSendFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		// 将字节切片转换为 float32 切片以便填充
		out := *(*[]float32)(unsafe.Pointer(&pOutputSample))
		// 注意：实际项目中应避免 unsafe，此处仅为展示最简写法。
		// 更安全的方式是使用 unsafe.Slice 或预先分配的缓冲区。

		phaseIncrement := 2 * math.Pi * freq / float64(deviceConfig.SampleRate)

		for i := 0; i < int(framecount); i++ {
			sample := float32(math.Sin(phase))
			// 双声道填充相同数据
			out[i*2] = sample
			out[i*2+1] = sample
			phase += phaseIncrement
			if phase > 2*math.Pi {
				phase -= 2 * math.Pi
			}
		}
	}

	// 5. 初始化设备并启动
	callbacks := malgo.DeviceCallbacks{
		Data: onSendFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, callbacks)
	if err != nil {
		log.Fatal(err)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Playing 440Hz sine wave for 3 seconds...")
	time.Sleep(3 * time.Second)
	device.Stop()
}
