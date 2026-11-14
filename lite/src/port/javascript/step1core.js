/**
 * TypeScript port of github.com/ak47less/step1go/lite.
 *
 * The Go API is intentionally kept in PascalCase so the two implementations
 * remain easy to compare. Native malgo is platform-bound and is represented by
 * the AudioContext abstraction instead. This file includes an in-memory
 * implementation for tests and a Web Audio adapter for browsers.
 *
 * No runtime dependency is required.
 *
 * 阅读顺序：
 *   1. Common / Web Audio：值类型、音频后端抽象。
 *   2. Timeline / Event：时间换算与事件分发。
 *   3. LifeCycle / Part：引擎生命周期和可注册组件。
 *   4. Instrument / Cube：乐器与 {bank, step, track} 音序数据。
 *   5. Sequencer / Engine：实时时钟、走带和引擎装配。
 */
export const ToneName = {
    C: "C",
    CS: "C#",
    D: "D",
    DS: "D#",
    E: "E",
    F: "F",
    FS: "F#",
    G: "G",
    GS: "G#",
    A: "A",
    AS: "A#",
    B: "B",
};
export const ScaleModeMock = "mock";
export const ScaleModeMajor = "major";
export const ScaleModeNaturalMajor = "natural_major";
export const ScaleModeMinor = "minor";
export const ScaleModeNaturalMinor = "natural_minor";
export const ScaleModeHarmonicMinor = "harmonic_minor";
export const ScaleModeMelodicMinor = "melodic_minor";
export const ScaleModeIonian = "ionian";
export const ScaleModeDorian = "dorian";
export const ScaleModePhrygian = "phrygian";
export const ScaleModeLydian = "lydian";
export const ScaleModeMixolydian = "mixolydian";
export const ScaleModeAeolian = "aeolian";
export const ScaleModeLocrian = "locrian";
export const ScaleModePentatonic = "Pentatonic";
export const ScaleModeMajorPentatonic = "MajorPentatonic";
export const ScaleModeMinorPentatonic = "MinorPentatonic";
export const ScaleModeBlues = "blues";
export const ScaleModeMajorBlues = "major_blues";
const toneNames = [
    ToneName.C,
    ToneName.CS,
    ToneName.D,
    ToneName.DS,
    ToneName.E,
    ToneName.F,
    ToneName.FS,
    ToneName.G,
    ToneName.GS,
    ToneName.A,
    ToneName.AS,
    ToneName.B,
];
export function GetPitchInfo(number, destination) {
    // 与原实现一致：输入先夹到合法 MIDI 范围，再按十二平均律计算频率。
    const dst = destination ?? {};
    const midiNumber = Math.max(0, Math.min(127, Math.trunc(number)));
    const group = Math.floor(midiNumber / 12) - 1;
    const tone = midiNumber % 12;
    dst.Number = midiNumber;
    dst.Name = `${toneNames[tone]}${group}`;
    dst.Group = group;
    dst.Tone = tone;
    dst.Frequency = 440 * Math.pow(2, (midiNumber - 69) / 12);
    return dst;
}
export function NewTempo(upper, lower) {
    return { Upper: upper, Lower: lower };
}
export class SimpleProfile {
    scale = ScaleModeMajor;
    bpm = 100;
    tempo = NewTempo(4, 4);
    tone = ToneName.C;
    GetScale() {
        return this.scale;
    }
    GetSpeed() {
        return this.bpm;
    }
    GetTempo() {
        return this.tempo;
    }
    GetTone() {
        return this.tone;
    }
    SetScale(scale) {
        this.scale = scale;
        return this;
    }
    SetSpeed(bpm) {
        this.bpm = bpm;
        return this;
    }
    SetTempo(tempo) {
        this.tempo = tempo;
        return this;
    }
    SetTone(tone) {
        this.tone = tone;
        return this;
    }
    Init() {
        this.bpm = 100;
        this.scale = ScaleModeMajor;
        this.tone = ToneName.C;
        this.tempo = NewTempo(4, 4);
        return this;
    }
    InitWith(base) {
        // 传入 base 时复制其配置；否则恢复默认的 100 BPM、C 大调、4/4 拍。
        if (base) {
            this.bpm = base.GetSpeed();
            this.scale = base.GetScale();
            this.tone = base.GetTone();
            this.tempo = base.GetTempo();
        }
        else {
            this.Init();
        }
        return this;
    }
}
export var Action;
(function (Action) {
    Action[Action["Min"] = 0] = "Min";
    Action[Action["StartEngine"] = 1] = "StartEngine";
    Action[Action["StopEngine"] = 2] = "StopEngine";
    Action[Action["Play"] = 3] = "Play";
    Action[Action["Loop"] = 4] = "Loop";
    Action[Action["Stop"] = 5] = "Stop";
    Action[Action["Pause"] = 6] = "Pause";
    Action[Action["Resume"] = 7] = "Resume";
    Action[Action["Record"] = 8] = "Record";
    Action[Action["NoteOn"] = 9] = "NoteOn";
    Action[Action["NoteOff"] = 10] = "NoteOff";
    Action[Action["MasterVolumeUp"] = 11] = "MasterVolumeUp";
    Action[Action["MasterVolumeDown"] = 12] = "MasterVolumeDown";
    Action[Action["MasterVolumeMute"] = 13] = "MasterVolumeMute";
    Action[Action["KeyUp"] = 14] = "KeyUp";
    Action[Action["KeyDown"] = 15] = "KeyDown";
    Action[Action["Max"] = 16] = "Max";
})(Action || (Action = {}));
export const ActionMin = Action.Min;
export const ActionStartEngine = Action.StartEngine;
export const ActionStopEngine = Action.StopEngine;
export const ActionPlay = Action.Play;
export const ActionLoop = Action.Loop;
export const ActionStop = Action.Stop;
export const ActionPause = Action.Pause;
export const ActionResume = Action.Resume;
export const ActionRecord = Action.Record;
export const ActionNoteOn = Action.NoteOn;
export const ActionNoteOff = Action.NoteOff;
export const ActionMasterVolumeUp = Action.MasterVolumeUp;
export const ActionMasterVolumeDown = Action.MasterVolumeDown;
export const ActionMasterVolumeMute = Action.MasterVolumeMute;
export const ActionKeyUp = Action.KeyUp;
export const ActionKeyDown = Action.KeyDown;
export const ActionMax = Action.Max;
function formatLog(format, args) {
    let next = 0;
    return format.replace(/%[sdvfg]/g, (token) => {
        if (next >= args.length) {
            return token;
        }
        const value = args[next++];
        if (token === "%d") {
            return String(Math.trunc(Number(value)));
        }
        if (token === "%f") {
            return String(Number(value));
        }
        return String(value);
    });
}
export class ConsoleLogger {
    Info(format, ...args) {
        console.info(formatLog(format, args));
    }
    Error(format, ...args) {
        console.error(formatLog(format, args));
    }
    Trace(format, ...args) {
        console.debug(formatLog(format, args));
    }
    Debug(format, ...args) {
        console.debug(formatLog(format, args));
    }
    Warn(format, ...args) {
        console.warn(formatLog(format, args));
    }
}
const defaultLogger = new ConsoleLogger();
export function Log() {
    return defaultLogger;
}
export function handleError(error) {
    // Timer、时钟监听器等异步入口不能向调用方抛错，因此统一交给默认 Logger。
    if (error === null || error === undefined) {
        return;
    }
    const message = error instanceof Error ? error.message : String(error);
    defaultLogger.Error("%s", message);
}
export var OscillatorType;
(function (OscillatorType) {
    OscillatorType["Sine"] = "sine";
    OscillatorType["Square"] = "square";
    OscillatorType["Sawtooth"] = "sawtooth";
    OscillatorType["Triangle"] = "triangle";
    OscillatorType["Custom"] = "custom";
})(OscillatorType || (OscillatorType = {}));
export const OscillatorTypeSine = OscillatorType.Sine;
export const OscillatorTypeSquare = OscillatorType.Square;
export const OscillatorTypeSawtooth = OscillatorType.Sawtooth;
export const OscillatorTypeTriangle = OscillatorType.Triangle;
export const OscillatorTypeCustom = OscillatorType.Custom;
export var AudioContextState;
(function (AudioContextState) {
    AudioContextState["Suspended"] = "suspended";
    AudioContextState["Running"] = "running";
    AudioContextState["Closed"] = "closed";
})(AudioContextState || (AudioContextState = {}));
export const AudioContextStateSuspended = AudioContextState.Suspended;
export const AudioContextStateRunning = AudioContextState.Running;
export const AudioContextStateClosed = AudioContextState.Closed;
function normalizeAudioTime(value) {
    // Web Audio 不接受负数或 NaN 的调度时间。
    return Number.isFinite(value) && value >= 0 ? value : 0;
}
////////////////////////////////////////////////////////////////////////////////
// AudioParam implementation
var AudioParamEventKind;
(function (AudioParamEventKind) {
    AudioParamEventKind[AudioParamEventKind["Set"] = 0] = "Set";
    AudioParamEventKind[AudioParamEventKind["LinearRamp"] = 1] = "LinearRamp";
    AudioParamEventKind[AudioParamEventKind["ExponentialRamp"] = 2] = "ExponentialRamp";
    AudioParamEventKind[AudioParamEventKind["Target"] = 3] = "Target";
})(AudioParamEventKind || (AudioParamEventKind = {}));
export class DefaultAudioParam {
    now;
    // 事件按时间排序。ramp 事件在到达其时间点前会从当前值插值，
    // SetTargetAtTime 则从起始时刻开始指数逼近目标值。
    baseValue;
    baseTime;
    events = [];
    constructor(value, baseTime, now = () => 0) {
        this.now = now;
        this.baseValue = value;
        this.baseTime = baseTime;
    }
    GetValue() {
        return this.valueAt(this.now());
    }
    SetValue(value) {
        const now = this.now();
        this.baseValue = value;
        this.baseTime = now;
        this.events = [];
    }
    SetValueAtTime(value, startTime) {
        this.addEvent({
            Kind: AudioParamEventKind.Set,
            Value: value,
            Time: normalizeAudioTime(startTime),
            TimeConstant: 0,
        });
    }
    LinearRampToValueAtTime(value, endTime) {
        this.addEvent({
            Kind: AudioParamEventKind.LinearRamp,
            Value: value,
            Time: normalizeAudioTime(endTime),
            TimeConstant: 0,
        });
    }
    ExponentialRampToValueAtTime(value, endTime) {
        this.addEvent({
            Kind: AudioParamEventKind.ExponentialRamp,
            Value: value,
            Time: normalizeAudioTime(endTime),
            TimeConstant: 0,
        });
    }
    SetTargetAtTime(target, startTime, timeConstant) {
        this.addEvent({
            Kind: AudioParamEventKind.Target,
            Value: target,
            Time: normalizeAudioTime(startTime),
            TimeConstant: normalizeAudioTime(timeConstant),
        });
    }
    CancelScheduledValues(startTime) {
        // 先把当前值固化到 now，避免删除事件后参数值瞬间跳回旧基准。
        const time = normalizeAudioTime(startTime);
        const now = this.now();
        this.baseValue = this.valueAt(now);
        this.baseTime = now;
        this.events = this.events.filter((event) => event.Time > now && event.Time < time);
    }
    addEvent(event) {
        let index = this.events.length;
        for (let i = 0; i < this.events.length; i++) {
            if (this.events[i].Time > event.Time) {
                index = i;
                break;
            }
        }
        this.events.splice(index, 0, event);
    }
    valueAt(time) {
        // 顺序扫描已排序事件：游标表示上一事件结束后的状态，
        // 当前区间按事件类型求值。
        let value = this.baseValue;
        let cursor = this.baseTime;
        let mode = AudioParamEventKind.Set;
        let target = value;
        let timeConstant = 0;
        for (const event of this.events) {
            if (time < event.Time) {
                if (event.Kind === AudioParamEventKind.LinearRamp) {
                    return interpolateValue(cursor, value, event.Time, event.Value, time);
                }
                if (event.Kind === AudioParamEventKind.ExponentialRamp) {
                    return interpolateExponentialValue(cursor, value, event.Time, event.Value, time);
                }
                return evaluateParam(cursor, value, time, mode, target, timeConstant);
            }
            value = evaluateParam(cursor, value, event.Time, mode, target, timeConstant);
            cursor = event.Time;
            if (event.Kind === AudioParamEventKind.Target) {
                mode = AudioParamEventKind.Target;
                target = event.Value;
                timeConstant = event.TimeConstant;
            }
            else {
                mode = AudioParamEventKind.Set;
                value = event.Value;
            }
        }
        return evaluateParam(cursor, value, time, mode, target, timeConstant);
    }
}
function evaluateParam(startTime, startValue, time, mode, target, timeConstant) {
    if (time <= startTime || mode !== AudioParamEventKind.Target) {
        return startValue;
    }
    if (timeConstant <= 0) {
        return target;
    }
    const factor = 1 - Math.exp(-(time - startTime) / timeConstant);
    return startValue + (target - startValue) * factor;
}
function interpolateValue(startTime, startValue, endTime, endValue, time) {
    if (endTime <= startTime) {
        return endValue;
    }
    const ratio = Math.max(0, Math.min(1, (time - startTime) / (endTime - startTime)));
    return startValue + (endValue - startValue) * ratio;
}
function interpolateExponentialValue(startTime, startValue, endTime, endValue, time) {
    if (startValue <= 0 || endValue <= 0 || endTime <= startTime) {
        return interpolateValue(startTime, startValue, endTime, endValue, time);
    }
    const ratio = Math.max(0, Math.min(1, (time - startTime) / (endTime - startTime)));
    return startValue * Math.pow(endValue / startValue, ratio);
}
////////////////////////////////////////////////////////////////////////////////
// In-memory audio graph
const MOCK_DEFAULT_SAMPLE_RATE = 44100;
const MOCK_DEFAULT_CHANNEL_COUNT = 2;
const MOCK_DEFAULT_FREQUENCY = 440;
const MOCK_DEFAULT_GAIN = 1;
const MOCK_DEFAULT_PLAYBACK_RATE = 1;
export class MockAudioNode {
    // 记录连接关系便于测试断言；Mock 后端不实际渲染 PCM。
    Connections = [];
    Connect(destination) {
        if (destination && !this.Connections.includes(destination)) {
            this.Connections.push(destination);
        }
    }
    Disconnect() {
        this.Connections.length = 0;
    }
}
export class MockAudioScheduledSourceNode extends MockAudioNode {
    Context;
    // Starts/Stops 保留所有调用记录，包括被忽略的重复调度，
    // 便于验证音序器生成的 MIDI 时间。
    Starts = [];
    Stops = [];
    startScheduled = false;
    startTime = 0;
    started = false;
    stopScheduled = false;
    stopTime = 0;
    ended = false;
    onEnded = null;
    constructor(Context) {
        super();
        this.Context = Context;
    }
    Start(when) {
        const time = normalizeAudioTime(when);
        this.Starts.push(time);
        // 与 Web Audio 一致：同一个 source 只能真正启动一次。
        if (this.startScheduled || this.ended) {
            return;
        }
        this.startScheduled = true;
        this.startTime = time;
        this.Context.internalRegisterSource(this);
    }
    Stop(when) {
        const time = normalizeAudioTime(when);
        this.Stops.push(time);
        // Stop 可以晚于 Start 调用；真正结束后再次 Stop 没有效果。
        if (this.ended) {
            return;
        }
        this.stopScheduled = true;
        this.stopTime = time;
        this.Context.internalRegisterSource(this);
    }
    SetOnEnded(handler) {
        this.onEnded = handler;
        if (this.ended && handler) {
            handler();
        }
    }
    IsEnded() {
        return this.ended;
    }
    internalTick(now) {
        // Mock 上下文在 GetCurrentTime 和短定时器中被驱动，
        // 到达结束时间时补发 onended。
        if (this.ended || !this.startScheduled) {
            return;
        }
        if (!this.started && now >= this.startTime) {
            this.started = true;
        }
        const finished = this.started && this.stopScheduled && now >= this.stopTime;
        if (finished) {
            this.internalFinish();
        }
    }
    internalFinish() {
        if (this.ended || !this.started) {
            return;
        }
        this.ended = true;
        this.Context.internalUnregisterSource(this);
        this.onEnded?.();
    }
}
export class MockOscillatorNode extends MockAudioScheduledSourceNode {
    Frequency;
    Detune;
    Type = OscillatorType.Sine;
    constructor(context) {
        super(context);
        const now = context.GetCurrentTime();
        this.Frequency = new DefaultAudioParam(MOCK_DEFAULT_FREQUENCY, now, () => context.GetCurrentTime());
        this.Detune = new DefaultAudioParam(0, now, () => context.GetCurrentTime());
    }
    GetFrequency() {
        return this.Frequency;
    }
    GetDetune() {
        return this.Detune;
    }
    GetType() {
        return this.Type;
    }
    SetType(nodeType) {
        this.Type = nodeType;
    }
}
export class MockGainNode extends MockAudioNode {
    Gain;
    constructor(context) {
        super();
        this.Gain = new DefaultAudioParam(MOCK_DEFAULT_GAIN, context.GetCurrentTime(), () => context.GetCurrentTime());
    }
    GetGain() {
        return this.Gain;
    }
}
export class MockAudioDestinationNode extends MockAudioNode {
    MaxChannels;
    constructor(MaxChannels = MOCK_DEFAULT_CHANNEL_COUNT) {
        super();
        this.MaxChannels = MaxChannels;
    }
    GetMaxChannelCount() {
        return this.MaxChannels;
    }
}
export class MockAudioBuffer {
    SampleRate;
    Channels;
    constructor(SampleRate, Channels) {
        this.SampleRate = SampleRate;
        this.Channels = Channels;
    }
    GetSampleRate() {
        return this.SampleRate;
    }
    GetLength() {
        return this.Channels.length === 0 ? 0 : this.Channels[0].length;
    }
    GetNumberOfChannels() {
        return this.Channels.length;
    }
    GetChannelData(channel) {
        if (channel < 0 || channel >= this.Channels.length) {
            return null;
        }
        return this.Channels[channel];
    }
}
export class MockAudioBufferSourceNode extends MockAudioScheduledSourceNode {
    PlaybackRate;
    Detune;
    Buffer = null;
    constructor(context) {
        super(context);
        const now = context.GetCurrentTime();
        this.PlaybackRate = new DefaultAudioParam(MOCK_DEFAULT_PLAYBACK_RATE, now, () => context.GetCurrentTime());
        this.Detune = new DefaultAudioParam(0, now, () => context.GetCurrentTime());
    }
    GetBuffer() {
        return this.Buffer;
    }
    SetBuffer(buffer) {
        this.Buffer = buffer;
    }
    GetPlaybackRate() {
        return this.PlaybackRate;
    }
    GetDetune() {
        return this.Detune;
    }
}
export class MockAudioContextFactory {
    SampleRate;
    constructor(SampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
        this.SampleRate = SampleRate;
    }
    NewAudioContext() {
        return new MockAudioContext(this.SampleRate);
    }
}
export class MockAudioContext {
    sampleRate;
    state = AudioContextState.Suspended;
    destination;
    sources = new Set();
    elapsed = 0;
    runningSince = 0;
    sourceCheck = null;
    constructor(sampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
        this.sampleRate = sampleRate;
        this.destination = new MockAudioDestinationNode(MOCK_DEFAULT_CHANNEL_COUNT);
    }
    Close() {
        if (this.state === AudioContextState.Closed) {
            return;
        }
        this.freezeClock();
        this.state = AudioContextState.Closed;
        const sources = [...this.sources];
        this.sources.clear();
        for (const source of sources) {
            source.internalFinish();
        }
        if (this.sourceCheck !== null) {
            clearTimeout(this.sourceCheck);
            this.sourceCheck = null;
        }
    }
    Resume() {
        if (this.state === AudioContextState.Closed ||
            this.state === AudioContextState.Running) {
            return;
        }
        // 暂停期间累计 elapsed；恢复后 currentTime 从该点继续增长。
        this.state = AudioContextState.Running;
        this.runningSince = Date.now();
        this.runSources();
    }
    Suspend() {
        if (this.state === AudioContextState.Running) {
            this.freezeClock();
        }
        if (this.state !== AudioContextState.Closed) {
            this.state = AudioContextState.Suspended;
        }
    }
    CreateOscillator() {
        return new MockOscillatorNode(this);
    }
    CreateGain() {
        return new MockGainNode(this);
    }
    CreateBuffer(numberOfChannels, length, sampleRate) {
        // 与原 malgo/Mock 实现一致：负声道数按 0 处理，0 采样率回退到上下文采样率。
        const channelsCount = Math.max(0, Math.trunc(numberOfChannels));
        const sampleCount = Number.isFinite(length) && length >= 0 ? Math.trunc(length) : 0;
        const rate = sampleRate || this.sampleRate || MOCK_DEFAULT_SAMPLE_RATE;
        const channels = [];
        for (let channel = 0; channel < channelsCount; channel++) {
            channels.push(new Array(sampleCount).fill(0));
        }
        return new MockAudioBuffer(rate, channels);
    }
    CreateBufferSource() {
        return new MockAudioBufferSourceNode(this);
    }
    GetCurrentTime() {
        this.runSources();
        return this.currentTime();
    }
    GetSampleRate() {
        return this.sampleRate;
    }
    GetDestination() {
        return this.destination;
    }
    GetState() {
        return this.state;
    }
    internalRegisterSource(source) {
        if (this.state !== AudioContextState.Closed) {
            this.sources.add(source);
            this.scheduleSourceCheck();
        }
    }
    internalUnregisterSource(source) {
        this.sources.delete(source);
    }
    currentTime() {
        if (this.state !== AudioContextState.Running) {
            return this.elapsed;
        }
        return this.elapsed + (Date.now() - this.runningSince) / 1000;
    }
    freezeClock() {
        if (this.state !== AudioContextState.Running) {
            return;
        }
        this.elapsed += (Date.now() - this.runningSince) / 1000;
        this.runningSince = 0;
    }
    runSources() {
        if (this.state !== AudioContextState.Running) {
            return;
        }
        const now = this.currentTime();
        for (const source of [...this.sources]) {
            source.internalTick(now);
        }
    }
    scheduleSourceCheck() {
        // 用低频轮询推动 source 完成状态；Node 下 unref，避免测试进程被定时器挂住。
        if (this.sourceCheck !== null || this.sources.size === 0) {
            return;
        }
        this.sourceCheck = setTimeout(() => {
            this.sourceCheck = null;
            this.runSources();
            this.scheduleSourceCheck();
        }, 10);
        const handle = this.sourceCheck;
        handle.unref?.();
    }
}
export class MockDriver {
    sampleRate;
    constructor(sampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
        this.sampleRate = sampleRate;
    }
    GetAudioContextFactory() {
        return new MockAudioContextFactory(this.sampleRate);
    }
}
export function NewMockDriver(sampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
    return new MockDriver(sampleRate);
}
class NativeAudioParamAdapter {
    param;
    constructor(param) {
        this.param = param;
    }
    GetValue() {
        return Number(this.param.value);
    }
    SetValue(value) {
        this.param.value = value;
    }
    SetValueAtTime(value, startTime) {
        this.param.setValueAtTime(value, startTime);
    }
    LinearRampToValueAtTime(value, endTime) {
        this.param.linearRampToValueAtTime(value, endTime);
    }
    ExponentialRampToValueAtTime(value, endTime) {
        this.param.exponentialRampToValueAtTime(value, endTime);
    }
    SetTargetAtTime(target, startTime, timeConstant) {
        this.param.setTargetAtTime(target, startTime, timeConstant);
    }
    CancelScheduledValues(startTime) {
        this.param.cancelScheduledValues(startTime);
    }
}
class NativeAudioNodeAdapter {
    node;
    constructor(node) {
        this.node = node;
    }
    Connect(destination) {
        // 只有适配器包装的节点才能可靠转换为原生 AudioNode。
        if (destination instanceof NativeAudioNodeAdapter) {
            this.node.connect(destination.nativeNode());
            return;
        }
        const nativeDestination = destination.NativeNode;
        if (nativeDestination) {
            this.node.connect(nativeDestination);
        }
    }
    Disconnect() {
        this.node.disconnect();
    }
    nativeNode() {
        return this.node;
    }
}
class NativeAudioScheduledSourceAdapter extends NativeAudioNodeAdapter {
    Start(when) {
        this.node.start(normalizeAudioTime(when));
    }
    Stop(when) {
        this.node.stop(normalizeAudioTime(when));
    }
    SetOnEnded(handler) {
        this.node.onended = handler;
    }
}
class NativeOscillatorAdapter extends NativeAudioScheduledSourceAdapter {
    GetFrequency() {
        return new NativeAudioParamAdapter(this.node.frequency);
    }
    GetDetune() {
        return new NativeAudioParamAdapter(this.node.detune);
    }
    GetType() {
        return this.node.type;
    }
    SetType(nodeType) {
        this.node.type = nodeType;
    }
}
class NativeGainAdapter extends NativeAudioNodeAdapter {
    GetGain() {
        return new NativeAudioParamAdapter(this.node.gain);
    }
}
class NativeDestinationAdapter extends NativeAudioNodeAdapter {
    GetMaxChannelCount() {
        return Number(this.node.maxChannelCount ?? 2);
    }
}
class NativeAudioBufferAdapter {
    buffer;
    constructor(buffer) {
        this.buffer = buffer;
    }
    GetSampleRate() {
        return Number(this.buffer.sampleRate);
    }
    GetLength() {
        return Number(this.buffer.length);
    }
    GetNumberOfChannels() {
        return Number(this.buffer.numberOfChannels);
    }
    GetChannelData(channel) {
        if (channel < 0 || channel >= this.GetNumberOfChannels()) {
            return null;
        }
        return this.buffer.getChannelData(channel);
    }
}
class NativeBufferSourceAdapter extends NativeAudioScheduledSourceAdapter {
    GetBuffer() {
        return this.node.buffer ? new NativeAudioBufferAdapter(this.node.buffer) : null;
    }
    SetBuffer(buffer) {
        // 原生 buffer 直接复用；来自 Mock 等后端的 buffer 则复制为原生 AudioBuffer。
        if (buffer instanceof NativeAudioBufferAdapter) {
            this.node.buffer = buffer.buffer;
        }
        else if (buffer === null) {
            this.node.buffer = null;
        }
        else {
            const native = this.node.context.createBuffer(buffer.GetNumberOfChannels(), buffer.GetLength(), buffer.GetSampleRate());
            for (let channel = 0; channel < buffer.GetNumberOfChannels(); channel++) {
                const data = buffer.GetChannelData(channel);
                if (data) {
                    native.copyToChannel(data, channel);
                }
            }
            this.node.buffer = native;
        }
    }
    GetPlaybackRate() {
        return new NativeAudioParamAdapter(this.node.playbackRate);
    }
    GetDetune() {
        return new NativeAudioParamAdapter(this.node.detune);
    }
}
export class WebAudioContext {
    context;
    // 把本文件的 Web Audio 抽象转发到浏览器原生 AudioContext。
    destination;
    constructor(context) {
        this.context = context;
        this.destination = new NativeDestinationAdapter(context.destination);
    }
    Close() {
        void this.context.close();
    }
    Resume() {
        return this.context.resume();
    }
    Suspend() {
        return this.context.suspend();
    }
    CreateOscillator() {
        return new NativeOscillatorAdapter(this.context.createOscillator());
    }
    CreateGain() {
        return new NativeGainAdapter(this.context.createGain());
    }
    CreateBuffer(numberOfChannels, length, sampleRate) {
        const rate = sampleRate || this.context.sampleRate;
        return new NativeAudioBufferAdapter(this.context.createBuffer(numberOfChannels, length, rate));
    }
    CreateBufferSource() {
        return new NativeBufferSourceAdapter(this.context.createBufferSource());
    }
    GetCurrentTime() {
        return this.context.currentTime;
    }
    GetSampleRate() {
        return this.context.sampleRate;
    }
    GetDestination() {
        return this.destination;
    }
    GetState() {
        return this.context.state;
    }
}
export class WebAudioContextFactory {
    NewAudioContext() {
        // Safari 的旧实例名是 webkitAudioContext。
        const scope = globalThis;
        const Constructor = scope.AudioContext ?? scope.webkitAudioContext;
        if (!Constructor) {
            throw new Error("Web Audio API is unavailable; use MockDriver or provide a Driver");
        }
        return new WebAudioContext(new Constructor());
    }
}
export class WebAudioDriver {
    // 浏览器推荐使用该 Driver；无 Web Audio 的环境可改用 MockDriver。
    GetAudioContextFactory() {
        return new WebAudioContextFactory();
    }
}
////////////////////////////////////////////////////////////////////////////////
// Timeline and positions
export var TimelineType;
(function (TimelineType) {
    TimelineType[TimelineType["Min"] = 0] = "Min";
    TimelineType[TimelineType["Root"] = 1] = "Root";
    TimelineType[TimelineType["Work"] = 2] = "Work";
    TimelineType[TimelineType["Clip"] = 3] = "Clip";
    TimelineType[TimelineType["Max"] = 4] = "Max";
})(TimelineType || (TimelineType = {}));
export class Timeline {
    // 每个 Timeline 的 Offset 相对于 Parent；绝对位置是所有祖先 Offset 之和
    // 加上当前 Pos。GetRoot 限制 32 层是为了防止错误父子环造成无限循环。
    Parent = null;
    Label = "";
    Offset = 0;
    Length = 0;
    Type = TimelineType.Min;
    Profile = null;
    Loop = false;
    Init() {
        if (!this.Profile) {
            this.Label = "default";
            this.Length = 999;
            this.Loop = false;
            this.Offset = 0;
            this.Parent = null;
            this.Profile = new SimpleProfile().Init();
        }
        return this;
    }
    IsTypeOf(type) {
        return this.Type === type;
    }
    HasSameRoot(other) {
        return this.GetRoot().Same(other?.GetRoot() ?? null);
    }
    Same(other) {
        return this === other;
    }
    Equals(other) {
        if (!other) {
            return false;
        }
        if (this === other) {
            return true;
        }
        if (this.Label !== other.Label ||
            this.Offset !== other.Offset ||
            this.Length !== other.Length ||
            this.Loop !== other.Loop) {
            return false;
        }
        if (!this.Parent || !other.Parent) {
            return this.Parent === other.Parent;
        }
        return this.Parent.Equals(other.Parent);
    }
    GetRoot() {
        let current = this;
        for (let ttl = 32; ttl > 0 && current.Parent; ttl--) {
            current = current.Parent;
        }
        return current;
    }
    GetTotalOffset() {
        let total = 0;
        let current = this;
        for (let ttl = 32; ttl > 0 && current; ttl--) {
            total += current.Offset;
            current = current.Parent;
        }
        return total;
    }
}
export function AbsolutePosition(value) {
    return value;
}
export function RelativePositionToAbsolute(position) {
    return position;
}
export class RelativePosition {
    Ref;
    Pos;
    constructor(Ref = null, Pos = 0) {
        this.Ref = Ref;
        this.Pos = Pos;
    }
    Absolute() {
        return (this.Ref?.GetTotalOffset() ?? 0) + this.Pos;
    }
    Clone() {
        return new RelativePosition(this.Ref, this.Pos);
    }
}
export function AbsolutePositionRelative(absolute, base, destination) {
    // 公式：relative = absolute - base.GetTotalOffset()。
    // 每次返回新 RelativePosition；调用方主动传入 destination 可复用对象。
    const dst = destination ?? new RelativePosition();
    dst.Ref = base ?? new Timeline();
    dst.Pos = absolute - dst.Ref.GetTotalOffset();
    return dst;
}
export class Timer {
    // Timer 使用 setTimeout 而非阻塞循环。Start/Stop 通过 workerID
    // 让已排队的旧回调自动失效。
    Delay = 0;
    Interval = 0;
    Handler = null;
    workerID = 0;
    timeout = null;
    Start() {
        this.Stop();
        const worker = this.workerID;
        const handler = this.Handler;
        if (!handler) {
            return;
        }
        this.work(worker, handler, this.Delay);
    }
    Stop() {
        this.workerID++;
        if (this.timeout !== null) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }
    }
    work(worker, handler, delay) {
        this.timeout = setTimeout(() => {
            this.timeout = null;
            if (worker !== this.workerID) {
                return;
            }
            try {
                handler(this);
            }
            catch (error) {
                handleError(error);
            }
            if (worker === this.workerID) {
                this.work(worker, handler, this.Interval);
            }
        }, this.sec2ms(delay));
    }
    sec2ms(seconds) {
        return Math.max(1, Math.trunc(seconds * 1000));
    }
}
////////////////////////////////////////////////////////////////////////////////
// Event dispatch
export var EventType;
(function (EventType) {
    EventType[EventType["Min"] = 0] = "Min";
    EventType[EventType["LoopBegin"] = 1] = "LoopBegin";
    EventType[EventType["LoopEnd"] = 2] = "LoopEnd";
    EventType[EventType["Max"] = 3] = "Max";
})(EventType || (EventType = {}));
export const EventTypeMin = EventType.Min;
export const EventTypeLoopBegin = EventType.LoopBegin;
export const EventTypeLoopEnd = EventType.LoopEnd;
export const EventTypeMax = EventType.Max;
export class EventContext {
    // 事件先进入 buffer，再由 Runner 按 FIFO 取出。回调中 Dispatch 的新事件
    // 会追加到队尾，因此 Run 仍能在同一次调用中处理它。
    dispatcherInstance = null;
    runnerInstance = null;
    callbacks = [];
    buffer = [];
    constructor() {
        this.dispatcherInstance = new DefaultEventDispatcher(this);
        this.runnerInstance = new DefaultEventRunner(this);
    }
    GetDispatcher() {
        return this.dispatcherInstance;
    }
    GetRunner() {
        return this.runnerInstance;
    }
    internalAdd(event) {
        this.buffer.push({ ...event });
    }
    internalAddCallback(callback) {
        this.callbacks.push(callback);
    }
    internalRemoveCallback(callback) {
        const index = this.callbacks.indexOf(callback);
        if (index >= 0) {
            this.callbacks.splice(index, 1);
            return true;
        }
        return false;
    }
    internalTake() {
        // 先取出再执行回调；如果回调抛错，当前事件已被消费，后续事件仍保留。
        return this.buffer.shift();
    }
    internalMatching(event) {
        return this.callbacks.filter((callback) => callback.Type === event.Type &&
            callback.ID === event.ID &&
            typeof callback.Fn === "function");
    }
    BufferLength() {
        return this.buffer.length;
    }
}
export class DefaultEventDispatcher {
    context;
    constructor(context) {
        this.context = context;
    }
    Dispatch(event) {
        if (!event) {
            return;
        }
        if (!this.context) {
            throw new Error("event dispatcher context is nil");
        }
        this.context.internalAdd(event);
    }
    AddCallback(callback) {
        if (!this.context) {
            throw new Error("event dispatcher context is nil");
        }
        if (!callback) {
            throw new Error("event callback is nil");
        }
        if (typeof callback.Fn !== "function") {
            throw new Error("event callback function is nil");
        }
        this.context.internalAddCallback(callback);
    }
    RemoveCallback(callback) {
        if (!this.context) {
            throw new Error("event dispatcher context is nil");
        }
        if (!callback) {
            throw new Error("event callback is nil");
        }
        if (!this.context.internalRemoveCallback(callback)) {
            throw new Error("event callback is not registered");
        }
    }
}
export class DefaultEventRunner {
    context;
    constructor(context) {
        this.context = context;
    }
    Run() {
        if (!this.context) {
            throw new Error("event runner context is nil");
        }
        this.RunContext(this.context);
    }
    RunContext(context) {
        if (!context) {
            throw new Error("event context is nil");
        }
        for (;;) {
            const event = context.internalTake();
            if (!event) {
                return;
            }
            for (const callback of context.internalMatching(event)) {
                try {
                    callback.Fn(event);
                }
                catch (cause) {
                    // 保留事件类型和 ID，便于定位具体分发失败点；原始错误放在 cause。
                    const detail = cause instanceof Error ? cause.message : String(cause);
                    throw new Error(`dispatch event type ${event.Type} id ${event.ID}: ${detail}`, { cause });
                }
            }
        }
    }
}
export function NewEventContext() {
    return new EventContext();
}
export class LifeCycle {
    Order = 0;
    OnInit = null;
    OnCreate = null;
    OnStart = null;
    OnResume = null;
    OnRun = null;
    OnPause = null;
    OnStop = null;
    OnDestroy = null;
}
class LifeCycleContext {
    cache = null;
    main = null;
    context = null;
    all = [];
    onInitDone = false;
    onCreateDone = false;
    onStartDone = false;
    paused = false;
}
export class DefaultLifeManager {
    state = new LifeCycleContext();
    mainLife;
    constructor() {
        this.mainLife = new MainLife(this.state);
        this.state.main = this.mainLife;
    }
    GetMain() {
        return this.mainLife;
    }
    Add(life) {
        if (!life) {
            return;
        }
        this.state.all.push(life);
        this.state.cache = null;
        replayCurrentLifeCycle(life, this.state);
    }
    Remove(life) {
        if (!life) {
            return;
        }
        this.state.all = this.state.all.filter((item) => item !== life);
        this.state.cache = null;
    }
}
class MainLife {
    state;
    constructor(state) {
        this.state = state;
    }
    GetLifeCycle() {
        const cycle = new LifeCycle();
        cycle.OnInit = (context) => this.onInit(context);
        cycle.OnCreate = () => this.onStage("OnCreate", false);
        cycle.OnStart = () => this.onStage("OnStart", false);
        cycle.OnResume = () => this.onStage("OnResume", false);
        cycle.OnRun = () => this.onStage("OnRun", false);
        cycle.OnPause = () => this.onStage("OnPause", true);
        cycle.OnStop = () => this.onStage("OnStop", true);
        cycle.OnDestroy = () => this.onStage("OnDestroy", true);
        return cycle;
    }
    onInit(context) {
        if (!context) {
            throw new Error("innerMainLife.onInit() : param 'ec:EngineContext' is nil");
        }
        this.state.context = context;
        this.state.onInitDone = true;
        for (const cycle of this.cache()) {
            cycle.OnInit?.(context);
        }
    }
    async onStage(stage, reverse) {
        if (stage === "OnCreate")
            this.state.onCreateDone = true;
        if (stage === "OnStart")
            this.state.onStartDone = true;
        if (stage === "OnPause")
            this.state.paused = true;
        if (stage === "OnResume")
            this.state.paused = false;
        const cycles = this.cache();
        const ordered = reverse ? [...cycles].reverse() : cycles;
        for (const cycle of ordered) {
            await cycle[stage]?.();
        }
    }
    cache() {
        if (!this.state.cache) {
            this.state.cache = this.state.all
                .map((life) => life.GetLifeCycle())
                .filter((cycle) => cycle !== null)
                .sort((a, b) => a.Order - b.Order);
        }
        return this.state.cache;
    }
}
function replayCurrentLifeCycle(life, state) {
    const cycle = life.GetLifeCycle();
    if (!cycle) {
        return;
    }
    const invoke = (fn) => {
        if (!fn)
            return;
        try {
            const result = fn();
            if (result instanceof Promise) {
                result.catch(handleError);
            }
        }
        catch (error) {
            handleError(error);
        }
    };
    if (state.onInitDone && state.context) {
        try {
            cycle.OnInit?.(state.context);
        }
        catch (error) {
            handleError(error);
        }
    }
    if (state.onCreateDone)
        invoke(cycle.OnCreate);
    if (state.onStartDone) {
        invoke(cycle.OnStart);
        if (!state.paused) {
            invoke(cycle.OnResume);
            invoke(cycle.OnRun);
        }
    }
}
export function NewLifeManager() {
    return new DefaultLifeManager();
}
export class LifeCycleRunner {
    // 主生命周期驱动引擎：OnRun 通常会等待 Engine.Join()，
    // 所以后续 Pause/Stop/Destroy 必须等它结束后才能执行。
    context = null;
    Init(context) {
        this.context = context;
        return this;
    }
    async RunLC(cycle) {
        if (!cycle) {
            return;
        }
        if (!this.context) {
            throw new Error("life-cycle runner is not initialized");
        }
        cycle.OnInit?.(this.context);
        await cycle.OnCreate?.();
        await cycle.OnStart?.();
        await cycle.OnResume?.();
        await Promise.resolve(cycle.OnRun?.());
        await cycle.OnPause?.();
        await cycle.OnStop?.();
        await cycle.OnDestroy?.();
    }
    RunL(life) {
        return this.RunLC(life?.GetLifeCycle() ?? null);
    }
    RunLM(manager) {
        return this.RunL(manager.GetMain());
    }
}
export class LifeCycleStageLogger {
    // 可作为配置中的 Part 加入，用于观察每个生命周期阶段的调用顺序。
    logger = Log();
    GetInfo() {
        return new PartInfo({ Name: "LifeCycleStageLogger", Part: this });
    }
    GetLifeCycle() {
        const cycle = new LifeCycle();
        cycle.OnInit = (context) => {
            this.logger = context.Logger;
            this.log("OnInit");
        };
        for (const stage of [
            "OnCreate",
            "OnStart",
            "OnResume",
            "OnRun",
            "OnPause",
            "OnStop",
            "OnDestroy",
        ]) {
            cycle[stage] = () => this.log(stage);
        }
        return cycle;
    }
    log(stage) {
        this.logger.Info("engine.lifecycle.%s()", stage);
    }
}
export class PartMeta {
    ID = 0;
    Name = "";
    Classes = [];
    Context = null;
    Factory = null;
    Part = null;
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
export class PartInfo extends PartMeta {
}
export class PartRegistration extends PartMeta {
}
export class PartContext {
    EC;
    Registry;
    Manager;
    constructor(context = null) {
        this.EC = context;
        const state = {
            context: this,
            byID: new Map(),
            byName: new Map(),
            managerByID: new Map(),
            managerByName: new Map(),
            nextID: 0,
        };
        this.Registry = new DefaultPartRegistry(state);
        this.Manager = new DefaultPartManager(state);
    }
}
export class PartInfoBuilder {
    info = new PartInfo();
    Build() {
        return new PartInfo(this.info);
    }
    SetPart(part) {
        this.info.Part = part;
        return this;
    }
    SetFactory(factory) {
        this.info.Factory = factory;
        return this;
    }
    SetName(name) {
        this.info.Name = name;
        return this;
    }
    SetID(id) {
        this.info.ID = id;
        return this;
    }
    SetContext(context) {
        this.info.Context = context;
        return this;
    }
    SetClasses(...classes) {
        this.info.Classes = classes;
        return this;
    }
}
export class DefaultPartRegistry {
    state;
    constructor(state) {
        this.state = state;
    }
    RegisterFactory(factory) {
        if (!factory) {
            throw new Error("part factory is nil");
        }
        const registration = factory.GetRegistration();
        if (!registration) {
            throw new Error("part factory registration is nil");
        }
        registration.Factory = factory;
        registration.Context = this.state.context;
        this.Register(registration);
    }
    RegisterPart(part) {
        if (!part) {
            throw new Error("part is nil");
        }
        const info = part.GetInfo();
        if (!info) {
            throw new Error("part info is nil");
        }
        const registration = new PartRegistration(info);
        registration.Context = this.state.context;
        registration.Part = part;
        this.Register(registration);
    }
    Register(registration) {
        if (!registration) {
            throw new Error("part registration is nil");
        }
        // ID/Name 未提供时自动生成；但至少要有 Part 实例或 PartFactory。
        if (registration.ID === 0) {
            registration.ID = this.makeNextID();
        }
        if (!registration.Name) {
            registration.Name = `unnamed_part_${registration.ID}`;
        }
        if (!registration.Factory && !registration.Part) {
            throw new Error(`part ${JSON.stringify(registration.Name)} has neither a factory nor an instance`);
        }
        const existingID = this.state.byID.get(registration.ID);
        if (existingID) {
            throw new Error(`part id ${registration.ID} is already registered as ${JSON.stringify(existingID.Name)}`);
        }
        const existingName = this.state.byName.get(registration.Name);
        if (existingName) {
            throw new Error(`part name ${JSON.stringify(registration.Name)} is already registered with id ${existingName.ID}`);
        }
        registration.Context ??= this.state.context;
        this.state.byID.set(registration.ID, registration);
        this.state.byName.set(registration.Name, registration);
    }
    ListNames() {
        return [...this.state.byName.keys()].sort();
    }
    FindByID(id) {
        return this.state.byID.get(id) ?? null;
    }
    FindByName(name) {
        return this.state.byName.get(name) ?? null;
    }
    ListWithClass(partClass) {
        return [...this.state.byID.values()]
            .filter((item) => item.Classes.includes(partClass))
            .sort((a, b) => a.ID === b.ID ? a.Name.localeCompare(b.Name) : a.ID - b.ID);
    }
    makeNextID() {
        if (this.state.nextID === 0) {
            this.state.nextID = 0x123456789;
        }
        this.state.nextID++;
        return this.state.nextID;
    }
}
export class DefaultPartManager {
    state;
    // Manager 按需调用 Factory.CreatePart，并缓存创建结果，
    // 因此相同 ID/Name 的后续查询会返回同一个实例。
    constructor(state) {
        this.state = state;
    }
    ListNames() {
        return [...this.state.byName.keys()].sort();
    }
    FindPartByID(id) {
        const registration = this.state.byID.get(id);
        return registration ? this.partFromRegistration(registration) : null;
    }
    FindPartByName(name) {
        const registration = this.state.byName.get(name);
        return registration ? this.partFromRegistration(registration) : null;
    }
    ListPartsWithClass(partClass) {
        return [...this.state.byID.values()]
            .filter((item) => item.Classes.includes(partClass))
            .sort((a, b) => a.ID === b.ID ? a.Name.localeCompare(b.Name) : a.ID - b.ID)
            .map((item) => this.partFromRegistration(item))
            .filter((part) => part !== null);
    }
    partFromRegistration(registration) {
        const cachedByID = this.state.managerByID.get(registration.ID);
        if (cachedByID) {
            return cachedByID;
        }
        const cachedByName = this.state.managerByName.get(registration.Name);
        if (cachedByName) {
            return cachedByName;
        }
        if (registration.Part) {
            this.cache(registration, registration.Part);
            return registration.Part;
        }
        if (!registration.Factory) {
            throw new Error(`part ${JSON.stringify(registration.Name)} has no factory`);
        }
        const info = new PartInfo(registration);
        info.Context = this.state.context;
        const part = registration.Factory.CreatePart(info);
        if (!part) {
            throw new Error(`create part ${JSON.stringify(registration.Name)}: factory returned nil`);
        }
        registration.Part = part;
        this.cache(registration, part);
        return part;
    }
    cache(registration, part) {
        this.state.managerByID.set(registration.ID, part);
        this.state.managerByName.set(registration.Name, part);
    }
}
export function NewPartContext(context = null) {
    return new PartContext(context);
}
////////////////////////////////////////////////////////////////////////////////
// MIDI and instruments
export var MidiEventType;
(function (MidiEventType) {
    // 首尾哨兵沿用 Go iota 结构，业务事件从 NoteOn 开始。
    MidiEventType[MidiEventType["Min"] = 0] = "Min";
    MidiEventType[MidiEventType["NoteOn"] = 1] = "NoteOn";
    MidiEventType[MidiEventType["NoteOff"] = 2] = "NoteOff";
    MidiEventType[MidiEventType["Max"] = 3] = "Max";
})(MidiEventType || (MidiEventType = {}));
export class MidiNote {
    // Position 是绝对音频时间。小于 1ms 的值由乐器解释为“立即播放”。
    EventType = MidiEventType.NoteOn;
    Position = 0;
    Channel = 0;
    Key = 0;
    Velocity = 0;
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
export class MidiMessage {
    // 一个调度消息通常同时携带 NoteOn 和 NoteOff，
    // SimpleSynth 会据此预先安排同一个 voice 的起止时间。
    NoteOn = null;
    NoteOff = null;
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
export class InstrumentInfo {
    Name = "";
    Label = "";
    Preset = "";
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
export class InstrumentRegistration extends InstrumentInfo {
    Factory;
    constructor(init = {}) {
        super(init);
        this.Factory = init.Factory ?? null;
    }
}
export class InstrumentFactoryRegistry {
    // 按乐器 Name 选择工厂；未注册名称回退到 SetDefault 设置的工厂。
    table = new Map();
    factoryDefault = null;
    GetInfo() {
        return new PartInfo({
            Name: "InstrumentFactoryRegistry",
            Classes: ["InstrumentFactoryRegistry"],
            Part: this,
        });
    }
    GetLifeCycle() {
        return new LifeCycle();
    }
    CreateInstrument(context, info) {
        if (!info) {
            throw new Error("instrument info is nil");
        }
        // 默认工厂只处理未注册名称，不会覆盖已注册名称对应的专用工厂。
        const registration = this.table.get(info.Name);
        const factory = registration?.Factory ?? this.factoryDefault;
        if (!factory) {
            throw new Error(`instrument ${JSON.stringify(info.Name)} is not registered and no default factory is set`);
        }
        return factory.CreateInstrument(context, info);
    }
    SetDefault(factory) {
        if (factory) {
            this.factoryDefault = factory;
        }
    }
    ListInfo() {
        return [...this.table.values()]
            .map((item) => new InstrumentInfo(item))
            .sort(compareInstrumentInfo);
    }
    Register(factory) {
        if (!factory) {
            return;
        }
        for (const info of factory.ListInfo()) {
            if (!info) {
                continue;
            }
            this.add(new InstrumentRegistration({ ...info, Factory: factory }));
        }
    }
    Add(registration) {
        this.add(registration);
    }
    add(registration) {
        if (!registration ||
            !registration.Factory ||
            !registration.Name) {
            return;
        }
        this.table.set(registration.Name, registration);
    }
}
function compareInstrumentInfo(a, b) {
    return (a.Name.localeCompare(b.Name) ||
        a.Label.localeCompare(b.Label) ||
        a.Preset.localeCompare(b.Preset));
}
export class DefaultInstrumentHolder {
    GetInstrument(context) {
        if (!context) {
            throw new Error("lite.Context is nil");
        }
        if (context.CurrentInstrument) {
            return context.CurrentInstrument;
        }
        const factory = context.MainInstrumentFactory;
        if (!factory) {
            throw new Error("context.MainInstrumentFactory is nil");
        }
        // 默认乐器使用三角波预设，创建后缓存在 EngineContext 中。
        const instrument = factory.CreateInstrument(context, new InstrumentInfo({ Name: "", Preset: "triangle" }));
        if (!instrument) {
            throw new Error("instrument factory returned nil");
        }
        context.CurrentInstrument = instrument;
        return instrument;
    }
}
export class InstrumentRegistry {
    // 收集实现了 InstrumentProvider 的 Part，并建立 Name -> Factory 索引。
    // 该注册表在 InstrumentManager 的 OnInit 阶段填充。
    table = new Map();
    RegisterAllInstruments(context) {
        if (!context) {
            throw new Error("lite.Context is nil");
        }
        if (!context.PartManager) {
            throw new Error("context.PartManager is nil");
        }
        for (const name of context.PartManager.ListNames()) {
            const part = context.PartManager.FindPartByName(name);
            if (!part) {
                throw new Error(`part ${JSON.stringify(name)} was not found`);
            }
            this.RegisterInstrument(part);
        }
    }
    RegisterInstrument(part) {
        if (!part) {
            throw new Error("instrument provider part is nil");
        }
        const provider = part;
        if (typeof provider.ListRegistrations !== "function") {
            return;
        }
        for (const registration of provider.ListRegistrations()) {
            if (!registration) {
                throw new Error("instrument registration is nil");
            }
            if (!registration.Name) {
                throw new Error("instrument registration name is empty");
            }
            if (!registration.Factory) {
                throw new Error(`instrument ${JSON.stringify(registration.Name)} has no factory`);
            }
            this.table.set(registration.Name, registration);
        }
    }
}
export class InstrumentManagerImpl {
    Registry = new InstrumentRegistry();
    GetFactory(info) {
        if (!info) {
            throw new Error("instrument info is nil");
        }
        const registration = this.Registry.table.get(info.Name);
        if (!registration) {
            throw new Error(`instrument ${JSON.stringify(info.Name)} is not registered`);
        }
        if (!registration.Factory) {
            throw new Error(`instrument ${JSON.stringify(info.Name)} has no factory`);
        }
        return registration.Factory;
    }
    ListInfo() {
        return [...this.Registry.table.values()]
            .map((item) => new InstrumentInfo(item))
            .sort(compareInstrumentInfo);
    }
    GetInfo() {
        return new PartInfoBuilder().SetName("InstrumentManager").Build();
    }
    GetLifeCycle() {
        const cycle = new LifeCycle();
        cycle.OnInit = (context) => this.Registry.RegisterAllInstruments(context);
        return cycle;
    }
}
export function NewInstrumentManager() {
    return new InstrumentManagerImpl();
}
export class InnerInputInstrument {
    context;
    // 输入乐器是一个间接层：外部 MIDI 总是转发到当前主乐器，
    // 因此切换主乐器时不需要重新连接输入设备。
    constructor(context) {
        this.context = context;
    }
    Dispatch(message) {
        if (!message) {
            throw new Error("innerInputInstrument: param 'msg' (MidiMessage) is nil");
        }
        const target = this.context.CurrentInstrumentHolder.GetInstrument(this.context);
        if (!target) {
            throw new Error("input target instrument is nil");
        }
        target.Dispatch(message);
    }
    GetInfo() {
        return new InstrumentInfo({
            Name: "InputInstrument",
            Label: "Input-Instrument",
            Preset: "default",
        });
    }
}
export class AutoInputInstrumentHolder {
    GetInstrument(context) {
        if (!context) {
            throw new Error("param: engine.context is nil");
        }
        if (!context.InputInstrument) {
            context.InputInstrument = new InnerInputInstrument(context);
        }
        return context.InputInstrument;
    }
}
export class Note {
    Key = 0;
    Velocity = 0;
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
export class SimpleSynthVoice {
    Source;
    Gain;
    constructor(Source, Gain) {
        this.Source = Source;
        this.Gain = Gain;
    }
}
export class SimpleSynthFactory {
    ListInfo() {
        return [
            new InstrumentInfo({
                Name: "default",
                Label: "Simple Synth",
                Preset: "default",
            }),
        ];
    }
    CreateInstrument(context, info) {
        if (!context) {
            throw new Error("simplesynth: context is nil");
        }
        if (!context.AudioContext) {
            throw new Error("simplesynth: audio context is nil");
        }
        const destination = context.AudioContext.GetDestination();
        if (!destination) {
            throw new Error("simplesynth: audio destination is nil");
        }
        const instrumentInfo = info ??
            new InstrumentInfo({
                Name: "default",
                Label: "Simple Synth",
                Preset: "default",
            });
        return new SimpleSynth(context.AudioContext, destination, instrumentInfo);
    }
}
export class SimpleSynth {
    audio;
    destination;
    info;
    // 每个 MIDI key 同时只保留一个 voice。重复 NoteOn 会先停止旧 voice，
    // 避免同一个音符叠加造成非预期增益。
    waveform;
    voices = new Map();
    constructor(audio, destination, info) {
        this.audio = audio;
        this.destination = destination;
        this.info = info;
        this.waveform = simpleSynthWaveform(info.Preset);
    }
    Dispatch(message) {
        if (!message) {
            return;
        }
        // 先处理 NoteOn，确保同一消息携带的 NoteOff 能停止刚创建的 voice。
        if (message.NoteOn) {
            if (message.NoteOn.Velocity === 0) {
                this.noteOff(message.NoteOn);
            }
            else {
                this.noteOn(message.NoteOn);
            }
        }
        if (message.NoteOff) {
            this.noteOff(message.NoteOff);
        }
    }
    GetInfo() {
        return new InstrumentInfo(this.info);
    }
    noteOn(note) {
        const now = this.audio.GetCurrentTime();
        const previous = this.voices.get(note.Key);
        this.voices.delete(note.Key);
        if (previous) {
            previous.Source.Stop(now);
            previous.Source.Disconnect();
            previous.Gain.Disconnect();
        }
        const voice = this.createVoice(note, now);
        this.voices.set(note.Key, voice);
        voice.Source.Start(this.noteTime(note));
    }
    createVoice(note, now) {
        const gain = this.audio.CreateGain();
        let source;
        if (this.waveform === "noise") {
            const bufferSource = this.audio.CreateBufferSource();
            this.prepareNoise(bufferSource);
            source = bufferSource;
        }
        else {
            const oscillator = this.audio.CreateOscillator();
            oscillator.SetType(simpleSynthOscillatorType(this.waveform));
            oscillator
                .GetFrequency()
                .SetValueAtTime(GetPitchInfo(note.Key).Frequency, now);
            source = oscillator;
        }
        const level = Math.max(0, Math.min(1, note.Velocity / 127));
        gain.GetGain().SetValueAtTime(level, now);
        source.Connect(gain);
        gain.Connect(this.destination);
        const voice = new SimpleSynthVoice(source, gain);
        source.SetOnEnded(() => this.voiceEnded(note.Key, voice));
        return voice;
    }
    noteOff(note) {
        const voice = this.voices.get(note.Key);
        this.voices.delete(note.Key);
        if (!voice) {
            return;
        }
        const when = this.noteTime(note);
        voice.Source.Stop(when);
        if (when <= this.audio.GetCurrentTime()) {
            voice.Source.Disconnect();
            voice.Gain.Disconnect();
        }
    }
    voiceEnded(key, voice) {
        if (this.voices.get(key) === voice) {
            this.voices.delete(key);
        }
        voice.Source.Disconnect();
        voice.Gain.Disconnect();
    }
    noteTime(note) {
        // 绝对位置接近 0 时按即时消息处理；否则直接作为 Web Audio 调度时间。
        return note.Position < 0.001 ? this.audio.GetCurrentTime() : note.Position;
    }
    prepareNoise(source) {
        const sampleRate = this.audio.GetSampleRate() || MOCK_DEFAULT_SAMPLE_RATE;
        const buffer = this.audio.CreateBuffer(1, sampleRate, sampleRate);
        let state = 0x9e3779b9;
        const data = buffer.GetChannelData(0);
        if (data) {
            for (let i = 0; i < data.length; i++) {
                state = (Math.imul(state, 1664525) + 1013904223) >>> 0;
                data[i] = (state / 0xffffffff) * 2 - 1;
            }
        }
        source.SetBuffer(buffer);
    }
}
function simpleSynthWaveform(preset) {
    switch (preset.trim().toLowerCase()) {
        case "triangle":
            return "triangle";
        case "saw":
        case "sawtooth":
            return "sawtooth";
        case "square":
            return "square";
        case "noise":
            return "noise";
        default:
            return "sine";
    }
}
function simpleSynthOscillatorType(waveform) {
    switch (waveform) {
        case "triangle":
            return OscillatorType.Triangle;
        case "sawtooth":
            return OscillatorType.Sawtooth;
        case "square":
            return OscillatorType.Square;
        default:
            return OscillatorType.Sine;
    }
}
export class Cell {
    BankIndex = 0;
    StepIndex = 0;
    TrackIndex = 0;
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
export class CellCache {
    List = [];
    Add(cell) {
        if (cell) {
            this.List.push(cell);
        }
    }
    Reset() {
        this.List.length = 0;
    }
    Sort(compare) {
        this.List.sort(compare);
    }
}
export class SimpleCell extends Cell {
    // 普通音符格。ComplexCell.First 指向其首个 SimpleCell，
    // 从而让一个多格音符只在第一个格子上触发一次。
    Context = null;
    OwnerStep = null;
    OwnerBank = null;
    OwnerTrack = null;
    Enabled = false;
    Note = null;
    Complex = null;
    OnStep(context) {
        if (!this.Enabled) {
            return;
        }
        if (!this.Complex) {
            if (this.Note) {
                this.DispatchNote(context, this.Note);
            }
            return;
        }
        if (this.Complex.First === this) {
            this.Complex.OnStep(context);
        }
    }
    DispatchNote(context, note) {
        // 当前格子的位置来自 stc.This；音符长度固定为一个 step。
        // 生成的 NoteOff 使用绝对时间，SimpleSynth 可以直接预调度。
        const track = this.OwnerTrack;
        if (!track) {
            throw new Error("SimpleCell.track_ref is nil");
        }
        const meta = track.Meta;
        if (!meta) {
            throw new Error("SimpleCell.track_meta is nil");
        }
        if (!meta.Instrument) {
            throw new Error("track.meta.Instrument is nil");
        }
        const position = context.This.Position.Absolute();
        const noteOn = new MidiNote({
            EventType: MidiEventType.NoteOn,
            Channel: meta.Channel,
            Key: note.Key,
            Velocity: note.Velocity,
            Position: position,
        });
        const noteOff = new MidiNote({
            EventType: MidiEventType.NoteOff,
            Channel: meta.Channel,
            Key: note.Key,
            Velocity: 0,
            Position: position + context.Info.SecondsPerStep,
        });
        meta.Instrument.Dispatch(new MidiMessage({ NoteOn: noteOn, NoteOff: noteOff }));
    }
}
export class ComplexCell extends Cell {
    // ComplexCell 占用多个 SimpleCell 槽位。这里把 Notes 同时交给首发格，
    // 保持“多音”语义；更复杂时值拆分可由上层 Note 数据表达。
    First = null;
    SimpleCells = [];
    Notes = [];
    OnStep(context) {
        if (!this.First) {
            return;
        }
        const notes = this.Notes.length > 0
            ? this.Notes
            : this.SimpleCells.map((cell) => cell.Note).filter((note) => note !== null);
        for (const note of notes) {
            this.First.DispatchNote(context, note);
        }
    }
}
export class CellManager {
    // 立方体中的每个 {bank, step, track} 组合只创建一个 SimpleCell；
    // 不同 Bank/Step/Track 通过 owner 引用共享这个对象。
    table = new Map();
    OpenCell(want, autoMake = false) {
        if (!want) {
            throw new Error("cell is nil");
        }
        const key = this.keyForCell(want);
        let cell = this.table.get(key);
        if (!cell) {
            if (!autoMake) {
                throw new Error(`no item:[cell bank:${want.BankIndex} step:${want.StepIndex} track:${want.TrackIndex}] in manager`);
            }
            cell = new SimpleCell({
                BankIndex: want.BankIndex,
                StepIndex: want.StepIndex,
                TrackIndex: want.TrackIndex,
            });
            this.table.set(key, cell);
        }
        return cell;
    }
    KeyForCell(cell) {
        return this.keyForCell(cell);
    }
    keyForCell(cell) {
        return cell
            ? `b${cell.BankIndex}_s${cell.StepIndex}_t${cell.TrackIndex}`
            : "";
    }
}
export class BankArray extends Array {
    Capacity() {
        return this.length;
    }
    Init(capacity) {
        const count = Math.max(1, Math.trunc(capacity));
        const output = new BankArray();
        for (let index = 0; index < count; index++) {
            const bank = this[index] ?? new Bank();
            bank.Index = index;
            output.push(bank);
        }
        return output;
    }
    GetSlice(length) {
        return new BankSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
    }
    GetItem(index) {
        return this[index] ?? null;
    }
}
export class BankSlice extends Array {
    Length() {
        return this.length;
    }
    Init(array, length) {
        return new BankSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
    }
    Get(index) {
        return this[index] ?? null;
    }
}
export class StepArray extends Array {
    Capacity() {
        return this.length;
    }
    Init(capacity) {
        const count = Math.max(1, Math.trunc(capacity));
        const output = new StepArray();
        for (let index = 0; index < count; index++) {
            const step = this[index] ?? new Step();
            step.Index = index;
            step.Weight = step.Weight || 60;
            output.push(step);
        }
        return output;
    }
    GetSlice(length) {
        return new StepSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
    }
    GetItem(index) {
        return this[index] ?? null;
    }
}
export class StepSlice extends Array {
    Length() {
        return this.length;
    }
    Init(array, length) {
        return new StepSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
    }
    Get(index) {
        return this[index] ?? null;
    }
}
export class TrackArray extends Array {
    Capacity() {
        return this.length;
    }
    Init(capacity) {
        const count = Math.max(1, Math.trunc(capacity));
        const output = new TrackArray();
        for (let index = 0; index < count; index++) {
            const track = this[index] ?? new Track();
            track.Index = index;
            output.push(track);
        }
        return output;
    }
    GetSlice(length) {
        return new TrackSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
    }
    GetItem(index) {
        return this[index] ?? null;
    }
}
export class TrackSlice extends Array {
    Length() {
        return this.length;
    }
    Init(array, length) {
        return new TrackSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
    }
    Get(index) {
        return this[index] ?? null;
    }
    Filter(predicate) {
        return new TrackSlice(...Array.prototype.filter.call(this, predicate));
    }
}
export class StepPosition extends RelativePosition {
    Index = 0;
    Location = 0;
}
export class Step {
    Context = null;
    Index = 0;
    Weight = 60;
    Cells = new CellCache();
    OnStep(context) {
        if (!context) {
            throw new Error("StepTickContext is nil");
        }
        for (const cell of this.Cells.List) {
            cell.OnStep(context);
        }
    }
}
export class TrackMeta {
    Index = 0;
    Name = "";
    Label = "";
    Description = "";
    Weight = 0;
    Channel = 0;
    Mute = false;
    Solo = false;
    Locked = false;
    Enabled = false;
    Gain = 0;
    Order = 0;
    Instrument = null;
}
export class Track {
    Context = null;
    Meta = null;
    Index = 0;
    Cells = new CellCache();
}
export class CubeBuffers {
    BankBuffer = new BankArray();
    StepBuffer = new StepArray();
    TrackBuffer = new TrackArray();
}
export class Cube {
    // 三个层级：
    //   Buffers：最大容量的对象池；
    //   Banks/Steps/Tracks：当前有效切片；
    //   Bank.Steps/Bank.Tracks：每个 bank 自己的视图。
    // 所有 bank 的 track metadata 与 Cube.Tracks 共享同一个 TrackMeta。
    Context = null;
    CellManager = new CellManager();
    Buffers = new CubeBuffers();
    Banks = new BankSlice();
    Tracks = new TrackSlice();
    Steps = new StepSlice();
    Init(context, configuration) {
        const cfg = configuration ??
            context?.EngineContext?.Configuration ??
            null;
        if (!context && !cfg) {
            throw new Error("parameter(s) are nil");
        }
        if (!cfg) {
            throw new Error("engine.configuration is nil");
        }
        this.Buffers.BankBuffer = this.Buffers.BankBuffer.Init(cfg.BankListCapacity);
        this.Buffers.StepBuffer = this.Buffers.StepBuffer.Init(cfg.StepListCapacity);
        this.Buffers.TrackBuffer = this.Buffers.TrackBuffer.Init(cfg.TrackListCapacity);
        this.Context = context;
        this.Banks = this.Buffers.BankBuffer.GetSlice(cfg.BankListSize);
        this.Steps = this.Buffers.StepBuffer.GetSlice(cfg.StepListSize);
        this.Tracks = this.Buffers.TrackBuffer.GetSlice(cfg.TrackListSize);
        for (const [index, track] of this.Buffers.TrackBuffer.entries()) {
            track.Index = index;
            track.Meta = new TrackMeta();
            track.Meta.Index = index;
            track.Context = this;
        }
        for (const [index, step] of this.Buffers.StepBuffer.entries()) {
            step.Index = index;
            step.Context = this;
        }
        for (const [index, bank] of this.Buffers.BankBuffer.entries()) {
            bank.Index = index;
            bank.Context = this;
            this.initBank(bank);
        }
    }
    Reload(bankIndex, all) {
        if (all) {
            for (const bank of this.Buffers.BankBuffer) {
                bank?.Reload();
            }
            return;
        }
        this.Buffers.BankBuffer.GetItem(bankIndex)?.Reload();
    }
    initBank(bank) {
        // 为 bank 的每个 cell 绑定所属 bank/step/track。Step 和 Track 两侧
        // 都保留同一 cell 的引用，所以调度可以从任一维度遍历。
        const stepCapacity = this.Buffers.StepBuffer.Capacity();
        const trackCapacity = this.Buffers.TrackBuffer.Capacity();
        const stepLength = this.Steps.Length();
        const trackLength = this.Tracks.Length();
        bank.StepBuffer = bank.StepBuffer.Init(stepCapacity);
        bank.TrackBuffer = bank.TrackBuffer.Init(trackCapacity);
        bank.Steps = bank.StepBuffer.GetSlice(stepLength);
        bank.Tracks = bank.TrackBuffer.GetSlice(trackLength);
        for (let trackIndex = 0; trackIndex < trackCapacity; trackIndex++) {
            const track = bank.TrackBuffer[trackIndex];
            const sharedTrack = this.Buffers.TrackBuffer[trackIndex];
            track.Context = this;
            track.Meta = sharedTrack.Meta;
            for (let stepIndex = 0; stepIndex < stepCapacity; stepIndex++) {
                const step = bank.StepBuffer[stepIndex];
                step.Context = this;
                const cell = this.CellManager.OpenCell(new Cell({
                    BankIndex: bank.Index,
                    StepIndex: stepIndex,
                    TrackIndex: trackIndex,
                }), true);
                cell.OwnerBank = bank;
                cell.OwnerStep = step;
                cell.OwnerTrack = track;
                cell.Context = this;
                track.Cells.Add(cell);
                step.Cells.Add(cell);
            }
        }
    }
}
export class Bank {
    Context = null;
    Index = 0;
    Name = "";
    Label = "";
    Description = "";
    StepBuffer = new StepArray();
    TrackBuffer = new TrackArray();
    Steps = new StepSlice();
    Tracks = new TrackSlice();
    OnStep(context) {
        if (!context) {
            throw new Error("param:STC(StepTickContext) is nil");
        }
        if (!this.Context) {
            this.Init(context.Context?.Cube ?? null);
        }
        // This.Index 是按 StepsPerCycle 取模后的物理索引。
        const step = this.Steps.Get(context.This.Index);
        step?.OnStep(context);
    }
    Init(cube) {
        if (!cube) {
            throw new Error("bank.init() : param is nil");
        }
        this.Context = cube;
        this.StepBuffer = this.StepBuffer.Init(cube.Buffers.StepBuffer.Capacity());
        this.TrackBuffer = this.TrackBuffer.Init(cube.Buffers.TrackBuffer.Capacity());
        this.Steps = this.StepBuffer.GetSlice(cube.Steps.Length());
        this.Tracks = this.TrackBuffer.GetSlice(cube.Tracks.Length());
    }
    Reload() {
        if (!this.Context) {
            return;
        }
        this.Steps = this.StepBuffer.GetSlice(this.Context.Steps.Length());
        this.Tracks = this.TrackBuffer.GetSlice(this.Context.Tracks.Length());
    }
}
////////////////////////////////////////////////////////////////////////////////
// Sequencer timing
export class SequencerInfo {
    // BPM 和 Tempo 共同决定 SecondsPerBeat/SecondsPerStep；
    // PreloadSeconds 表示调度器希望提前触发多少秒的音符。
    Profile = null;
    SecondsPerStep = 0;
    SecondsPerBeat = 0;
    StepsPerBeat = 1;
    StepsPerCycle = 16;
    PreloadSeconds = 1;
    Normalize(rewrite = false) {
        // 防止采样过疏或除零；rewrite=false 时返回独立副本。
        const dst = rewrite ? this : new SequencerInfo();
        const profile = this.Profile ?? new SimpleProfile().Init();
        const preload = Math.max(1, this.PreloadSeconds);
        const stepsPerCycle = this.StepsPerCycle >= 4 ? this.StepsPerCycle : 16;
        const stepsPerBeat = Math.max(1, this.StepsPerBeat);
        const secondsPerBeat = Math.max(0.001, this.SecondsPerBeat);
        const secondsPerStep = Math.max(0.001, this.SecondsPerStep);
        dst.PreloadSeconds = preload;
        dst.SecondsPerBeat = secondsPerBeat;
        dst.SecondsPerStep = secondsPerStep;
        dst.StepsPerBeat = stepsPerBeat;
        dst.StepsPerCycle = stepsPerCycle;
        dst.Profile = profile;
        return dst;
    }
}
export class StepTickPos {
    Position = new RelativePosition();
    Index = 0;
    Location = 0;
    ComputeIndex(info, rewrite = false) {
        if (!info) {
            return 0;
        }
        const stepsPerCycle = Math.max(1, info.StepsPerCycle);
        const index = Math.max(0, Math.trunc(this.Location)) % stepsPerCycle;
        if (rewrite) {
            this.Index = index;
        }
        return index;
    }
}
export class StepTickContext {
    // Front = 预加载前沿，Back = 已经处理到的位置，This = 当前派发位置。
    // RTC 指向引擎共享的实时位置对象。
    Context = null;
    WorkTimeline = null;
    BreakCount = 0;
    Broken = false;
    Info = new SequencerInfo();
    RTC = null;
    Front = new StepTickPos();
    Back = new StepTickPos();
    This = new StepTickPos();
}
export class SequencerContext {
    // 音序器运行时状态。CurrentBankHolder 负责在 CurrentBank 为空时
    // 自动选择并缓存当前 bank。
    Engine = null;
    EngineContext = null;
    Sequencer = null;
    CurrentBank = null;
    CurrentBankIndex = 0;
    CurrentBankHolder = null;
    Cube = new Cube();
    Info = new SequencerInfo();
    STC = null;
    Updating = false;
    NormalizeSTP(position, align) {
        const location = this.ComputeStepLocation(position.Position);
        const index = this.ComputeStepIndex(position.Position);
        if (align) {
            const timeline = this.EngineContext?.WorkTileline ?? null;
            position.Position.Pos = this.Info.SecondsPerStep * location;
            position.Position.Ref = timeline;
        }
        position.Index = index;
        position.Location = location;
    }
    ComputeStepLocation(position) {
        // 先把任意时间线的位置转成绝对时间，再换算到 WorkTimeline 的 step 序号。
        if (!position) {
            return 0;
        }
        const workTimeline = this.EngineContext?.WorkTileline;
        const relative = AbsolutePositionRelative(position.Absolute(), workTimeline);
        const secondsPerStep = Math.max(0.001, this.Info.SecondsPerStep);
        return Math.max(0, Math.trunc(relative.Pos / secondsPerStep));
    }
    ComputeStepIndex(position) {
        const location = this.ComputeStepLocation(position);
        const stepsPerCycle = this.Info.StepsPerCycle;
        return stepsPerCycle < 1 ? 0 : location % stepsPerCycle;
    }
    NewBank(index) {
        const bank = new Bank();
        bank.Init(this.Cube);
        bank.Index = index;
        return bank;
    }
    NewStep(index) {
        const step = new Step();
        step.Index = index;
        step.Context = this.Cube;
        return step;
    }
    NewTrack(index) {
        const track = new Track();
        track.Index = index;
        track.Context = this.Cube;
        return track;
    }
}
////////////////////////////////////////////////////////////////////////////////
// Clock and real-time positions
export class ClockEvent {
    Clock = null;
    Time = 0;
    Context = null;
}
export class RTC {
    Precision = new StepTickPos();
    Normal = new StepTickPos();
    Raw = new StepTickPos();
}
export class SimpleClock {
    // Clock 只负责读取 AudioContext 时间、刷新 RTC，然后通知监听器。
    // 音序器和监控组件因此看到同一个“当前时间”快照。
    listeners = [];
    AddListener(listener) {
        if (listener && !this.listeners.includes(listener)) {
            this.listeners.push(listener);
        }
    }
    RemoveListener(listener) {
        if (!listener) {
            return;
        }
        this.listeners = this.listeners.filter((item) => item !== listener);
    }
    Dispatch(event) {
        if (!event) {
            throw new Error("lite: nil clock event");
        }
        this.updateRTC(event);
        event.Clock = this;
        const errors = [];
        for (const listener of [...this.listeners]) {
            try {
                listener.OnClock(event);
            }
            catch (error) {
                errors.push(error);
            }
        }
        if (errors.length === 1) {
            throw errors[0];
        }
        if (errors.length > 1) {
            throw new Error(errors
                .map((error) => error instanceof Error ? error.message : String(error))
                .join("; "), { cause: errors });
        }
    }
    updateRTC(event) {
        const context = event.Context;
        if (!context?.AudioContext || !context.SequencerContext) {
            return;
        }
        const sequencerContext = context.SequencerContext;
        const info = sequencerContext.Info;
        const timelineRoot = context.RootTileline;
        const timelineWork = context.WorkTileline;
        const now = AbsolutePosition(context.AudioContext.GetCurrentTime());
        const rootPosition = AbsolutePositionRelative(now, timelineRoot);
        const workPosition = AbsolutePositionRelative(now, timelineWork);
        const secondsPerStep = Math.max(0.001, info.SecondsPerStep);
        const stepsPerCycle = Math.max(4, info.StepsPerCycle);
        // Precision 保留连续时间；Normal 对齐到当前 step 起点；
        // Raw 表示 RootTimeline 坐标下的原始时间。
        const stepCount = Math.max(0, Math.trunc(workPosition.Pos / secondsPerStep));
        const location = stepCount;
        const index = stepCount % stepsPerCycle;
        const raw = new StepTickPos();
        raw.Index = index;
        raw.Location = location;
        raw.Position = rootPosition;
        const precision = new StepTickPos();
        precision.Index = index;
        precision.Location = location;
        precision.Position = workPosition;
        const normal = new StepTickPos();
        normal.Index = index;
        normal.Location = location;
        normal.Position = new RelativePosition(workPosition.Ref, secondsPerStep * stepCount);
        event.Time = now;
        context.RTC.Raw = raw;
        context.RTC.Precision = precision;
        context.RTC.Normal = normal;
    }
}
////////////////////////////////////////////////////////////////////////////////
// Sequencer implementation and trigger
class CurrentBankHolder {
    context;
    constructor(context) {
        this.context = context;
    }
    GetBank() {
        if (!this.context.CurrentBank) {
            let bank = this.context.Cube.Banks.Get(this.context.CurrentBankIndex);
            bank ??= this.context.Cube.Banks.Get(0);
            if (!bank) {
                throw new Error("innerCurrentBankHolder: current bank is nil");
            }
            this.context.CurrentBankIndex = bank.Index;
            this.context.CurrentBank = bank;
        }
        return this.context.CurrentBank;
    }
}
class SequencerUpdater {
    context;
    // profile、速度或 bank 配置变化后，由 Updating 标记触发一次重算。
    constructor(context) {
        this.context = context;
    }
    Update() {
        this.context.Updating = false;
        Log().Info("innerSequencer.Update()");
        this.computeSpeed();
        this.loadCurrentBank();
    }
    computeSpeed() {
        // 一拍时长 = 60 / BPM；每步时长再除以 StepsPerBeat。
        const context = this.context;
        const timeline = context.EngineContext?.WorkTileline;
        const profile = timeline?.Profile ?? context.Info.Profile;
        if (!profile) {
            throw new Error("sequencer profile is nil");
        }
        const bpm = profile.GetSpeed();
        const stepsPerBeat = context.Info.StepsPerBeat;
        if (bpm < 5) {
            throw new Error(`the bpm is too low (bpm=${bpm})`);
        }
        if (stepsPerBeat < 1) {
            throw new Error(`bad value of steps/beat (value=${stepsPerBeat})`);
        }
        const secondsPerBeat = 60 / bpm;
        context.Info.SecondsPerBeat = secondsPerBeat;
        context.Info.SecondsPerStep = secondsPerBeat / stepsPerBeat;
    }
    loadCurrentBank() {
        const list = this.context.Cube.Banks;
        const index = this.context.CurrentBankIndex;
        if (index >= list.Length()) {
            throw new Error(`innerSequencerUpdater: bank index is out of array [index:${index} size:${list.Length()}]`);
        }
        const bank = list[index];
        if (!bank) {
            throw new Error(`innerSequencerUpdater: bank[${index}] is nil`);
        }
        this.context.CurrentBank = bank;
    }
}
export class InnerSequencer {
    context;
    constructor(context) {
        this.context = context;
    }
    GetBanks() {
        return [...this.context.Cube.Banks];
    }
    GetContext() {
        return this.context;
    }
    GetSteps() {
        return [...this.context.Cube.Steps];
    }
    GetTracks() {
        return [...this.context.Cube.Tracks];
    }
    OnStep(context) {
        const holder = this.context.CurrentBankHolder;
        if (!holder) {
            throw new Error("context.CurrentBankHolder is nil");
        }
        const bank = holder.GetBank();
        if (!bank) {
            throw new Error("current bank is nil");
        }
        bank.OnStep(context);
    }
    Update(context) {
        if (context.Updating) {
            new SequencerUpdater(this.context).Update();
        }
    }
}
export class SequencerTrigger {
    context;
    // 每次时钟 tick 都从已处理的 Back 推进到预加载前沿 Front。
    // 仅当 Transport.Walking() 为 true 时实际派发 step。
    constructor(context) {
        this.context = context;
    }
    OnClock(_event) {
        const context = this.context;
        this.updateCurrentPosition();
        if (!this.isWalking()) {
            return;
        }
        let stc = this.prepare(context.STC);
        this.checkParameters(stc);
        this.computeAheadPosition(stc);
        this.fastForward(stc);
        context.STC = stc;
    }
    isWalking() {
        return this.context.EngineContext?.TransportContext.Walking() ?? false;
    }
    updateCurrentPosition() {
        const engine = this.context.EngineContext;
        if (!engine) {
            return;
        }
        const transport = engine.TransportContext;
        const current = engine.RTC.Precision;
        if (transport.Walking()) {
            transport.Current = cloneTickPos(current);
            return;
        }
        if (transport.StopAt) {
            current.Position = transport.StopAt.Clone();
            current.Location = this.context.ComputeStepLocation(transport.StopAt);
            current.Index = this.context.ComputeStepIndex(transport.StopAt);
            transport.Current = cloneTickPos(current);
        }
    }
    prepare(source) {
        const stc = source ?? new StepTickContext();
        stc.Context = this.context;
        stc.WorkTimeline = this.context.EngineContext?.WorkTileline ?? null;
        return stc;
    }
    checkParameters(stc) {
        if (!stc.Context) {
            throw new Error("stc.context is nil");
        }
        if (!stc.Context.EngineContext) {
            throw new Error("stc.context.EngineContext is nil");
        }
        if (!stc.Context.Sequencer) {
            throw new Error("stc.context.Sequencer is nil");
        }
        if (!stc.WorkTimeline) {
            throw new Error("stc.wktline is nil");
        }
        if (!stc.Info) {
            throw new Error("stc.Info is nil");
        }
        if (!stc.RTC) {
            throw new Error("stc.RTC is nil");
        }
        if (stc.Info.StepsPerCycle < 4) {
            throw new Error(`bad stc.Info.StepsPerCycle value: ${stc.Info.StepsPerCycle}`);
        }
        if (stc.Info.PreloadSeconds < 1) {
            throw new Error(`bad stc.Info.PreloadSeconds value: ${stc.Info.PreloadSeconds}`);
        }
        if (stc.Info.SecondsPerStep < 0.001) {
            throw new Error(`bad stc.Info.SecondsPerStep value: ${stc.Info.SecondsPerStep}`);
        }
        stc.Front.Position.Ref = stc.WorkTimeline;
        stc.This.Position.Ref = stc.WorkTimeline;
        stc.Back.Position.Ref = stc.WorkTimeline;
    }
    computeAheadPosition(stc) {
        const context = stc.Context;
        const engine = context.EngineContext;
        const secondsPerStep = Math.max(0.001, stc.Info.SecondsPerStep);
        const preloadSteps = Math.max(1, Math.trunc(stc.Info.PreloadSeconds / secondsPerStep));
        const front = cloneTickPos(engine.RTC.Normal);
        front.Location += preloadSteps;
        front.Position.Pos += secondsPerStep * preloadSteps;
        front.ComputeIndex(stc.Info, true);
        stc.Front = front;
    }
    fastForward(stc) {
        // 一个时钟间隔可能跨过多个 step，因此需要补齐中间的 step。
        // gap 超过半圈时视为大幅跳转，标记 Broken 防止卡顿恢复后瞬间补播。
        const current = cloneTickPos(stc.Back);
        const target = stc.Front;
        const maxGap = Math.max(1, Math.trunc(stc.Info.StepsPerCycle / 2));
        const secondsPerStep = stc.Info.SecondsPerStep;
        while (current.Location < target.Location) {
            const gap = target.Location - current.Location;
            if (gap > maxGap) {
                stc.BreakCount++;
                stc.Broken = true;
                break;
            }
            stc.This = cloneTickPos(current);
            try {
                contextSequencer(stc).OnStep(stc);
            }
            catch (error) {
                handleError(error);
            }
            current.Position.Pos += secondsPerStep;
            current.Location++;
            current.ComputeIndex(stc.Info, true);
        }
        stc.Back = cloneTickPos(target);
    }
}
function contextSequencer(stc) {
    const sequencer = stc.Context?.Sequencer;
    if (!sequencer) {
        throw new Error("stc.context.Sequencer is nil");
    }
    return sequencer;
}
function cloneTickPos(source) {
    const destination = new StepTickPos();
    destination.Position = source.Position.Clone();
    destination.Index = source.Index;
    destination.Location = source.Location;
    return destination;
}
export class InnerSequencerFactory {
    CreateSequencer(engineContext) {
        // 默认每拍 2 step、循环长度为 StepListSize、提前 1 秒调度。
        // 具体 step 时长在第一次 Update 时根据 profile 计算。
        const sequencerContext = new SequencerContext();
        const sequencer = new InnerSequencer(sequencerContext);
        const trigger = new SequencerTrigger(sequencerContext);
        const bankHolder = new CurrentBankHolder(sequencerContext);
        const stc = new StepTickContext();
        const config = engineContext.Configuration;
        engineContext.Clock.AddListener(trigger);
        const info = sequencerContext.Info;
        info.Profile = config.Profile;
        info.PreloadSeconds = 1;
        info.StepsPerBeat = 2;
        info.StepsPerCycle = config.StepListSize;
        info.SecondsPerBeat = 0;
        info.SecondsPerStep = 0;
        stc.Context = sequencerContext;
        stc.Info = info;
        sequencerContext.STC = stc;
        sequencerContext.Sequencer = sequencer;
        sequencerContext.Engine = engineContext.Engine;
        sequencerContext.EngineContext = engineContext;
        sequencerContext.Updating = true;
        sequencerContext.CurrentBankHolder = bankHolder;
        sequencerContext.Cube.Init(sequencerContext, config);
        return sequencer;
    }
}
////////////////////////////////////////////////////////////////////////////////
// Engine configuration and transport
export class MonitorConfig {
    // URL 可直接填写 tcp://host:port；Prepare 会拆出 Scheme/Host/Port。
    URL = "";
    Scheme = "";
    Host = "";
    Port = 0;
    FlagLogSTC = false;
    FlagLogSeq = false;
    FlagLogTransport = false;
    FlagLogEngine = false;
    constructor(init = {}) {
        Object.assign(this, init);
    }
    Prepare() {
        if (this.URL) {
            try {
                const url = new URL(this.URL);
                this.Scheme = url.protocol.replace(/:$/, "");
                this.Host = url.hostname;
                this.Port = Number(url.port) || 0;
            }
            catch {
                // Keep the raw URL when the runtime URL parser rejects it.
            }
        }
        else {
            this.Scheme ||= "std";
            this.Host ||= "127.0.0.1";
            if (this.Port < 1) {
                this.Port = 33309;
            }
            this.URL = `${this.Scheme}://${this.Host}:${this.Port}`;
        }
        return this;
    }
}
export class Configuration {
    // Driver 提供底层音频上下文，InstrumentFactory 提供默认乐器。
    // Parts 可以混合 Part、PartFactory 和 PartRegistration。
    Driver = null;
    InstrumentFactory = null;
    SampleRate = 0;
    Profile = null;
    BankListSize = 0;
    StepListSize = 0;
    TrackListSize = 0;
    BankListCapacity = 0;
    StepListCapacity = 0;
    TrackListCapacity = 0;
    Parts = [];
    Monitor = null;
}
export function DefaultConfig() {
    // 默认使用 MockDriver，可在无音频设备环境创建和测试引擎。
    // 与原 Go API 一致，调用方仍需配置 InstrumentFactory。
    const config = new Configuration();
    config.Driver = NewMockDriver();
    config.SampleRate = 44100;
    config.Profile = new SimpleProfile().Init().SetSpeed(280);
    config.BankListCapacity = 8;
    config.StepListCapacity = 128;
    config.TrackListCapacity = 32;
    config.BankListSize = 4;
    config.StepListSize = 32;
    config.TrackListSize = 8;
    config.Monitor = null;
    return config;
}
export class EngineState {
    Starting = false;
    Stopping = false;
    Stopped = false;
    Started = false;
    String() {
        return `engine.[state starting:${this.Starting} started:${this.Started} stopping:${this.Stopping} stopped:${this.Stopped}]`;
    }
    toString() {
        return this.String();
    }
}
export class TransportContext {
    // StopAt 表示暂停/停止后的目标位置；Resume/Play 会据此重定位工作时间线。
    paused = false;
    playing = false;
    recording = false;
    Current = new StepTickPos();
    StopAt = null;
    Walking() {
        return this.playing && !this.paused;
    }
    IsPaused() {
        return this.paused;
    }
    IsPlaying() {
        return this.playing;
    }
    IsRecording() {
        return this.recording;
    }
    internalSetPaused(value) {
        this.paused = value;
    }
    internalSetPlaying(value) {
        this.playing = value;
    }
    internalSetRecording(value) {
        this.recording = value;
    }
}
export class EngineContext {
    // 所有子系统共享的依赖容器；Context 是对应的 Go 风格类型别名。
    Transport;
    LifeManager;
    PartManager;
    PartRegistry;
    Logger = Log();
    Engine;
    Clock;
    EventDispatcher;
    EventRunner;
    TransportContext = new TransportContext();
    SequencerContext;
    Configuration;
    AudioContext;
    AudioContextFactory;
    CurrentInstrument = null;
    CurrentInstrumentHolder;
    MainInstrumentFactory;
    InstrumentManager;
    InputInstrument = null;
    InputInstrumentHolder;
    Timer;
    TimerHandler = null;
    RootTileline;
    WorkTileline;
    RTC = new RTC();
    State = new EngineState();
    GetPosAtWork(source) {
        // 已经是 WorkTimeline 相对位置时直接返回，避免二次扣减 offset。
        const position = source ?? new RelativePosition();
        if (!position.Ref || !this.WorkTileline) {
            return position;
        }
        if (position.Ref.IsTypeOf(TimelineType.Work)) {
            return position;
        }
        return AbsolutePositionRelative(position.Absolute(), this.WorkTileline);
    }
    GetCurrentPosition(destination) {
        const dst = destination ?? new RelativePosition();
        const position = this.TransportContext.StopAt ?? this.RTC.Precision.Position;
        dst.Ref = position.Ref;
        dst.Pos = position.Pos;
        return dst;
    }
    RebaseToPosition(to) {
        // 保持绝对位置不变，只调整 WorkTimeline.Offset：
        // work.offset + workPos = currentRootPos。
        if (!to) {
            return;
        }
        const work = this.WorkTileline;
        const absolute = to.Absolute();
        const workPosition = AbsolutePositionRelative(absolute, work);
        const rootPosition = this.RTC.Raw.Position.Absolute();
        work.Offset = rootPosition - workPosition.Pos;
    }
}
function isPart(value) {
    return (typeof value === "object" &&
        value !== null &&
        typeof value.GetInfo === "function");
}
function isPartFactory(value) {
    return (typeof value === "object" &&
        value !== null &&
        typeof value.CreatePart === "function" &&
        typeof value.GetRegistration === "function");
}
function isLife(value) {
    return (typeof value === "object" &&
        value !== null &&
        typeof value.GetLifeCycle === "function");
}
export class InnerEngineFactory {
    CreateEngine(config) {
        // 装配顺序不可随意调整：先建立基础子系统，再创建 Cube，
        // 之后注册 Part、收集生命周期，最后创建乐器并绑定到轨道。
        if (!config) {
            throw new Error("configuration is nil");
        }
        if (!config.Driver) {
            throw new Error("config.Driver is nil");
        }
        if (!config.InstrumentFactory) {
            throw new Error("config.InstrumentFactory is nil");
        }
        const context = new EngineContext();
        const engine = new InnerEngine();
        const driver = config.Driver;
        const lifeManager = NewLifeManager();
        const partContext = NewPartContext(context);
        const audioFactory = driver.GetAudioContextFactory();
        const audioContext = audioFactory.NewAudioContext();
        const transportContext = new TransportContext();
        const rootTimeline = new Timeline();
        const workTimeline = new Timeline();
        const transport = new InnerTransport();
        const timer = new Timer();
        const clock = new SimpleClock();
        const sequencerFactory = new InnerSequencerFactory();
        const instrumentHolder = new DefaultInstrumentHolder();
        const instrumentManager = NewInstrumentManager();
        const eventContext = NewEventContext();
        const inputHolder = new AutoInputInstrumentHolder();
        const profile = config.Profile ?? new SimpleProfile().Init();
        config.Profile = profile;
        engine.Context = context;
        transport.Context = context;
        transport.TransportContext = transportContext;
        rootTimeline.Label = "root";
        rootTimeline.Profile = profile;
        workTimeline.Label = "work";
        workTimeline.Profile = profile;
        workTimeline.Parent = rootTimeline;
        context.LifeManager = lifeManager;
        context.PartManager = partContext.Manager;
        context.PartRegistry = partContext.Registry;
        context.Configuration = config;
        context.Engine = engine;
        context.AudioContext = audioContext;
        context.AudioContextFactory = audioFactory;
        context.Transport = transport;
        context.TransportContext = transportContext;
        context.RootTileline = rootTimeline;
        context.WorkTileline = workTimeline;
        context.Timer = timer;
        context.TimerHandler = (source) => transport.ontimer(source);
        context.Clock = clock;
        context.Logger = Log();
        context.MainInstrumentFactory = config.InstrumentFactory;
        context.CurrentInstrumentHolder = instrumentHolder;
        context.InstrumentManager = instrumentManager;
        context.InputInstrument = null;
        context.InputInstrumentHolder = inputHolder;
        context.EventDispatcher = eventContext.GetDispatcher();
        context.EventRunner = eventContext.GetRunner();
        partContext.Registry.RegisterPart(instrumentManager);
        const sequencer = sequencerFactory.CreateSequencer(context);
        context.SequencerContext = sequencer.GetContext();
        const stepContext = context.SequencerContext.STC;
        if (!stepContext) {
            throw new Error("sequencer step context is nil");
        }
        stepContext.RTC = context.RTC;
        this.initParts(context);
        this.initLifeCycles(context);
        this.initInstruments(context);
        return engine;
    }
    initParts(context) {
        // 通过运行时能力判断 Part 角色，同一对象可以同时实现多种角色。
        for (const source of context.Configuration.Parts) {
            if (source instanceof PartRegistration) {
                context.PartRegistry.Register(source);
            }
            else if (isPartFactory(source)) {
                context.PartRegistry.RegisterFactory(source);
            }
            else if (isPart(source)) {
                context.PartRegistry.RegisterPart(source);
            }
        }
    }
    initLifeCycles(context) {
        for (const name of context.PartManager.ListNames()) {
            const part = context.PartManager.FindPartByName(name);
            if (isLife(part)) {
                context.LifeManager.Add(part);
            }
        }
    }
    initInstruments(context) {
        const first = context.CurrentInstrumentHolder.GetInstrument(context);
        if (!first) {
            throw new Error("default instrument is nil");
        }
        for (const track of context.SequencerContext.Cube.Tracks) {
            const meta = track.Meta;
            if (meta && !meta.Instrument) {
                meta.Instrument = first;
            }
        }
    }
}
export function DefaultEngineFactory() {
    return new InnerEngineFactory();
}
export class InnerEngine {
    // Join 不阻塞 JS 线程，而是返回 Promise，由生命周期运行器等待。
    Context;
    GetContext() {
        return this.Context;
    }
    Join() {
        const state = this.Context.State;
        if (!state.Starting) {
            return Promise.reject(new Error("call engine.start() first"));
        }
        return new Promise((resolve) => {
            const check = () => {
                if (state.Stopping || state.Stopped) {
                    resolve();
                    return;
                }
                setTimeout(check, 50);
            };
            check();
        });
    }
    Create() {
        // 主 Timer 约每 10ms 产生一次时钟事件，实际精度取决于宿主调度器。
        const context = this.Context;
        const timer = context.Timer ?? new Timer();
        context.Timer = timer;
        timer.Delay = 0.01;
        timer.Interval = 0.01;
        timer.Handler = context.TimerHandler;
    }
    Destroy() {
        this.Context.Timer = new Timer();
        this.Context.State.Stopped = true;
    }
    Start() {
        const context = this.Context;
        context.State.Starting = true;
        const resumed = context.AudioContext.Resume();
        if (resumed instanceof Promise) {
            resumed.catch(handleError);
        }
        context.Timer.Start();
        if (context.Configuration.Monitor) {
            const monitor = new Monitor();
            monitor.Init(context);
            monitor.Start();
        }
    }
    Stop() {
        this.Context.State.Stopping = true;
        this.Context.Timer.Stop();
    }
}
export class InnerTransport {
    // Transport 只维护播放状态和目标位置，实际时间推进由 Clock 完成。
    Context;
    TransportContext;
    Walking() {
        return this.TransportContext.Walking();
    }
    Pause() {
        // 仅在真正 Walking 时记录位置；重复 Pause 不会覆盖 StopAt。
        if (this.TransportContext.Walking()) {
            this.TransportContext.StopAt =
                this.Context.RTC.Precision.Position.Clone();
            this.TransportContext.internalSetPaused(true);
        }
    }
    Paused() {
        return this.TransportContext.IsPaused();
    }
    Play() {
        this.TransportContext.internalSetPlaying(true);
        this.tryRebaseToStopAt();
    }
    Playing() {
        return this.TransportContext.IsPlaying();
    }
    Record() {
        this.TransportContext.internalSetRecording(true);
        this.TransportContext.internalSetPlaying(true);
        this.tryRebaseToStopAt();
    }
    Recording() {
        return this.TransportContext.IsRecording();
    }
    Resume() {
        this.TransportContext.internalSetPaused(false);
        this.tryRebaseToStopAt();
    }
    Stop() {
        const context = this.Context;
        this.TransportContext.internalSetRecording(false);
        this.TransportContext.internalSetPlaying(false);
        this.TransportContext.internalSetPaused(false);
        this.TransportContext.StopAt = new RelativePosition(context.WorkTileline, 0);
    }
    ontimer(_timer) {
        const context = this.Context;
        const sequencerContext = context.SequencerContext;
        if (sequencerContext.Updating && sequencerContext.Sequencer) {
            sequencerContext.Sequencer.Update(sequencerContext);
        }
        context.State.Started = true;
        const event = new ClockEvent();
        event.Clock = context.Clock;
        event.Context = context;
        event.Time = context.RTC.Precision.Position.Absolute();
        context.Clock.Dispatch(event);
    }
    tryRebaseToStopAt() {
        // 恢复播放前同步重定位工作时间和音序器 Back，
        // 防止暂停期间积压的 step 被连续补播。
        const at = this.TransportContext.StopAt;
        this.TransportContext.StopAt = null;
        if (!at) {
            return;
        }
        this.Context.RebaseToPosition(at);
        const sequencerContext = this.Context.SequencerContext;
        const stepContext = sequencerContext.STC;
        if (!stepContext) {
            return;
        }
        const back = new StepTickPos();
        back.Position = at.Clone();
        back.Index = sequencerContext.ComputeStepIndex(at);
        back.Location = sequencerContext.ComputeStepLocation(at);
        stepContext.Back = back;
    }
}
////////////////////////////////////////////////////////////////////////////////
// Automatic engine lifecycle and boot
export class AutoEngineController {
    // 把 Engine 的 Create/Start/Join/Stop/Destroy 接入通用生命周期。
    // OnRun 等待 Join 完成，确保后续停止阶段不会提前执行。
    context = null;
    GetInfo() {
        return new PartInfo({ Name: "AutoEngineController", Part: this });
    }
    GetLifeCycle() {
        const cycle = new LifeCycle();
        cycle.OnInit = (context) => {
            this.context = context;
        };
        cycle.OnCreate = () => this.context?.Engine.Create();
        cycle.OnStart = () => this.context?.Engine.Start();
        cycle.OnRun = () => this.context?.Engine.Join();
        cycle.OnStop = () => this.context?.Engine.Stop();
        cycle.OnDestroy = () => this.context?.Engine.Destroy();
        return cycle;
    }
}
export class Boot {
    // 创建引擎并运行主生命周期。若配置包含 AutoEngineController，
    // Run 返回的 Promise 会保持到引擎被停止。
    async Run(config = null) {
        let cfg = config;
        if (!cfg) {
            cfg = DefaultConfig();
            cfg.Driver = NewMockDriver();
            cfg.Profile = new SimpleProfile().Init();
            cfg.SampleRate = 44100;
        }
        const engine = DefaultEngineFactory().CreateEngine(cfg);
        const context = engine.GetContext();
        const runner = new LifeCycleRunner();
        await runner.Init(context).RunLM(context.LifeManager);
    }
}
export function Run(config = null) {
    return new Boot().Run(config);
}
export class InnerActionController {
    context;
    // 将控制动作映射为 Transport 或 MIDI 操作。Pause/Resume 共用一个
    // 切换逻辑：Walking 时暂停，否则恢复。
    handlers = new Map();
    constructor(context) {
        this.context = context;
        this.handlers.set(Action.Min, () => undefined);
        this.handlers.set(Action.Play, (context) => context.Transport.Play());
        this.handlers.set(Action.Stop, (context) => context.Transport.Stop());
        this.handlers.set(Action.Record, (context) => context.Transport.Record());
        this.handlers.set(Action.Pause, (context) => {
            if (context.Transport.Walking()) {
                context.Transport.Pause();
            }
            else {
                context.Transport.Resume();
            }
        });
        this.handlers.set(Action.Resume, (context) => {
            if (context.Transport.Walking()) {
                context.Transport.Pause();
            }
            else {
                context.Transport.Resume();
            }
        });
        this.handlers.set(Action.NoteOn, (context, event) => {
            const instrument = context.InputInstrumentHolder.GetInstrument(context);
            instrument?.Dispatch(new MidiMessage({
                NoteOn: new MidiNote({
                    Key: event.NoteNumber ?? 0,
                    Velocity: event.NoteVelocity ?? 0,
                    EventType: MidiEventType.NoteOn,
                }),
            }));
        });
        this.handlers.set(Action.NoteOff, (context, event) => {
            const instrument = context.InputInstrumentHolder.GetInstrument(context);
            instrument?.Dispatch(new MidiMessage({
                NoteOff: new MidiNote({
                    Key: event.NoteNumber ?? 0,
                    Velocity: event.NoteVelocity ?? 0,
                    EventType: MidiEventType.NoteOff,
                }),
            }));
        });
    }
    Start() {
        // The action controller is driven by its caller.
    }
    Execute(event) {
        const handler = this.handlers.get(event.Action);
        if (!handler) {
            throw new Error(`no handler with action [${event.Action}]`);
        }
        handler(this.context, event);
    }
}
export function NewActionController(context) {
    return new InnerActionController(context);
}
export class Monitor {
    // 低频状态监控器，只负责日志输出，不参与音频调度。
    context = null;
    timer = null;
    InputKeyCount = 0;
    InputKeyCode = 0;
    Init(context) {
        this.context = context;
    }
    Start() {
        if (!this.context) {
            throw new Error("monitor is not initialized");
        }
        this.timer = new Timer();
        this.timer.Delay = 1;
        this.timer.Interval = 1 / 24;
        this.timer.Handler = () => this.ontimer();
        this.timer.Start();
    }
    Stop() {
        this.timer?.Stop();
        this.timer = null;
    }
    ontimer() {
        const context = this.context;
        if (!context) {
            return;
        }
        const logger = context.Logger;
        logger.Info("--------------------------------------------------------------");
        logger.Info("%s", context.State.String());
        const transport = context.TransportContext;
        logger.Info("transport.[state playing:%s recording:%s paused:%s]", transport.IsPlaying(), transport.IsRecording(), transport.IsPaused());
        logger.Info("transport.[position raw:%f root:%f work:%f]", context.AudioContext.GetCurrentTime(), context.RTC.Raw.Position.Pos, context.RTC.Precision.Position.Pos);
        const stepContext = context.SequencerContext.STC;
        if (stepContext) {
            logger.Info("[STC(StepTickContext) location:%d index:%d pos:%f ]", stepContext.Front.Location, stepContext.Front.Index, stepContext.Front.Position.Pos);
        }
        else {
            logger.Info("STC is nil");
        }
        logger.Info("input.[key count:%d code:%d char:'%s']", this.InputKeyCount, this.InputKeyCode, String.fromCharCode(this.InputKeyCode));
        if (context.State.Stopped || context.State.Stopping) {
            this.Stop();
        }
    }
}
