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
/**
 * 十二平均律中使用的音名常量。
 *
 * `CS`、`DS` 等名称分别表示升 C、升 D；值是对外的显示名称。
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
// 调式标识。Mock 用于占位或测试，其余值对应项目支持的大小调、教会调式、
// 五声音阶与布鲁斯音阶。
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
/**
 * 将 MIDI 音高编号解析为音名、八度、半音序号和频率。
 *
 * @param {number} number MIDI 音高编号；会截断并限制在 0..127。
 * @param {object|null} [destination] 可复用的结果对象。
 * @returns {{Number:number, Name:string, Group:number, Tone:number, Frequency:number}}
 *   解析后的音高信息；A4（69）的频率为 440Hz。
 */
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
/**
 * 创建拍号对象。
 *
 * @param {number} upper 每小节拍数。
 * @param {number} lower 以何种音符为一拍，例如四分音符为 4。
 * @returns {{Upper:number, Lower:number}} 新的拍号对象。
 */
export function NewTempo(upper, lower) {
    return { Upper: upper, Lower: lower };
}
/**
 * 基础音乐配置，保存调式、BPM、拍号和主音。
 */
export class SimpleProfile {
    scale = ScaleModeMajor;
    bpm = 100;
    tempo = NewTempo(4, 4);
    tone = ToneName.C;
    /** @returns {string} 当前调式。 */
    GetScale() {
        return this.scale;
    }
    /** @returns {number} 当前速度（BPM）。 */
    GetSpeed() {
        return this.bpm;
    }
    /** @returns {{Upper:number, Lower:number}} 当前拍号。 */
    GetTempo() {
        return this.tempo;
    }
    /** @returns {string} 当前主音名称。 */
    GetTone() {
        return this.tone;
    }
    /**
     * 设置调式。
     * @param {string} scale 调式标识。
     * @returns {SimpleProfile} 当前实例，便于链式调用。
     */
    SetScale(scale) {
        this.scale = scale;
        return this;
    }
    /**
     * 设置速度。
     * @param {number} bpm 每分钟拍数。
     * @returns {SimpleProfile} 当前实例，便于链式调用。
     */
    SetSpeed(bpm) {
        this.bpm = bpm;
        return this;
    }
    /**
     * 设置拍号。
     * @param {{Upper:number, Lower:number}} tempo 拍号对象。
     * @returns {SimpleProfile} 当前实例，便于链式调用。
     */
    SetTempo(tempo) {
        this.tempo = tempo;
        return this;
    }
    /**
     * 设置主音。
     * @param {string} tone 主音名称。
     * @returns {SimpleProfile} 当前实例，便于链式调用。
     */
    SetTone(tone) {
        this.tone = tone;
        return this;
    }
    /**
     * 恢复默认配置：100 BPM、C 大调和 4/4 拍。
     * @returns {SimpleProfile} 当前实例，便于链式调用。
     */
    Init() {
        this.bpm = 100;
        this.scale = ScaleModeMajor;
        this.tone = ToneName.C;
        this.tempo = NewTempo(4, 4);
        return this;
    }
    /**
     * 从另一个 profile 复制配置，或在未提供时恢复默认值。
     * @param {{GetSpeed:Function, GetScale:Function, GetTone:Function, GetTempo:Function}|null} [base]
     *   要复制的音乐配置。
     * @returns {SimpleProfile} 当前实例，便于链式调用。
     */
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
/**
 * 控制器可执行的动作类型。
 *
 * `Min`/`Max` 是范围哨兵；其余值覆盖引擎、走带、音符和主音量控制。
 */
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
// 与枚举成员同名的常量别名，方便不使用枚举对象语法的调用方。
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
/**
 * 用类似 Go `fmt` 的最小占位符集合格式化日志。
 *
 * 支持 `%s`、`%d`、`%v`、`%f` 和 `%g`；参数不足时保留原占位符。
 *
 * @param {string} format 日志格式字符串。
 * @param {unknown[]} args 依次替换占位符的参数。
 * @returns {string} 格式化后的文本。
 */
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
/**
 * 基于浏览器或 Node 全局 `console` 的日志实现。
 */
export class ConsoleLogger {
    /**
     * 输出普通信息。
     * @param {string} format 日志格式字符串。
     * @param {...unknown} args 格式化参数。
     */
    Info(format, ...args) {
        console.info(formatLog(format, args));
    }
    /**
     * 输出错误信息。
     * @param {string} format 日志格式字符串。
     * @param {...unknown} args 格式化参数。
     */
    Error(format, ...args) {
        console.error(formatLog(format, args));
    }
    /**
     * 输出跟踪信息；当前映射到 `console.debug`。
     * @param {string} format 日志格式字符串。
     * @param {...unknown} args 格式化参数。
     */
    Trace(format, ...args) {
        console.debug(formatLog(format, args));
    }
    /**
     * 输出调试信息。
     * @param {string} format 日志格式字符串。
     * @param {...unknown} args 格式化参数。
     */
    Debug(format, ...args) {
        console.debug(formatLog(format, args));
    }
    /**
     * 输出警告信息。
     * @param {string} format 日志格式字符串。
     * @param {...unknown} args 格式化参数。
     */
    Warn(format, ...args) {
        console.warn(formatLog(format, args));
    }
}
const defaultLogger = new ConsoleLogger();
/**
 * 返回模块共享的默认日志器。
 * @returns {ConsoleLogger} 默认日志器。
 */
export function Log() {
    return defaultLogger;
}
/**
 * 将异步入口捕获到的错误写入默认日志器。
 *
 * @param {unknown} error 要记录的异常；`null` 和 `undefined` 会被忽略。
 * @returns {void}
 */
export function handleError(error) {
    // Timer、时钟监听器等异步入口不能向调用方抛错，因此统一交给默认 Logger。
    if (error === null || error === undefined) {
        return;
    }
    const message = error instanceof Error ? error.message : String(error);
    defaultLogger.Error("%s", message);
}
/**
 * Web Audio 振荡器可使用的波形类型。
 */
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
/**
 * 音频上下文生命周期状态。
 */
export var AudioContextState;
(function (AudioContextState) {
    AudioContextState["Suspended"] = "suspended";
    AudioContextState["Running"] = "running";
    AudioContextState["Closed"] = "closed";
})(AudioContextState || (AudioContextState = {}));
export const AudioContextStateSuspended = AudioContextState.Suspended;
export const AudioContextStateRunning = AudioContextState.Running;
export const AudioContextStateClosed = AudioContextState.Closed;
/**
 * 将 Web Audio 调度时间规范化为非负有限数。
 *
 * @param {unknown} value 待检查的时间值。
 * @returns {number} 合法时间，或负数、NaN、Infinity 对应的 0。
 */
function normalizeAudioTime(value) {
    // Web Audio 不接受负数或 NaN 的调度时间。
    return Number.isFinite(value) && value >= 0 ? value : 0;
}
////////////////////////////////////////////////////////////////////////////////
// AudioParam implementation
// AudioParam 内部调度事件种类，仅用于当前文件中的内存实现。
var AudioParamEventKind;
(function (AudioParamEventKind) {
    AudioParamEventKind[AudioParamEventKind["Set"] = 0] = "Set";
    AudioParamEventKind[AudioParamEventKind["LinearRamp"] = 1] = "LinearRamp";
    AudioParamEventKind[AudioParamEventKind["ExponentialRamp"] = 2] = "ExponentialRamp";
    AudioParamEventKind[AudioParamEventKind["Target"] = 3] = "Target";
})(AudioParamEventKind || (AudioParamEventKind = {}));
/**
 * 可独立测试的 AudioParam 实现。
 *
 * 支持即时赋值、线性/指数渐变和目标值逼近；所有时间单位均为秒。
 */
export class DefaultAudioParam {
    now;
    // 事件按时间排序。ramp 事件在到达其时间点前会从当前值插值，
    // SetTargetAtTime 则从起始时刻开始指数逼近目标值。
    baseValue;
    baseTime;
    events = [];
    /**
     * @param {number} value 初始值。
     * @param {number} baseTime 初始值对应的时间。
     * @param {() => number} [now] 返回当前音频时间的函数。
     */
    constructor(value, baseTime, now = () => 0) {
        this.now = now;
        this.baseValue = value;
        this.baseTime = baseTime;
    }
    /** @returns {number} 当前时刻的参数值。 */
    GetValue() {
        return this.valueAt(this.now());
    }
    /**
     * 立即设置参数值并清除已有调度事件。
     * @param {number} value 新值。
     * @returns {void}
     */
    SetValue(value) {
        const now = this.now();
        this.baseValue = value;
        this.baseTime = now;
        this.events = [];
    }
    /**
     * 在指定时间把参数设为固定值。
     * @param {number} value 目标值。
     * @param {number} startTime 生效时间（秒）。
     * @returns {void}
     */
    SetValueAtTime(value, startTime) {
        this.addEvent({
            Kind: AudioParamEventKind.Set,
            Value: value,
            Time: normalizeAudioTime(startTime),
            TimeConstant: 0,
        });
    }
    /**
     * 从上一事件值线性渐变到指定时间的目标值。
     * @param {number} value 目标值。
     * @param {number} endTime 渐变结束时间（秒）。
     * @returns {void}
     */
    LinearRampToValueAtTime(value, endTime) {
        this.addEvent({
            Kind: AudioParamEventKind.LinearRamp,
            Value: value,
            Time: normalizeAudioTime(endTime),
            TimeConstant: 0,
        });
    }
    /**
     * 从上一事件值指数渐变到指定时间的目标值。
     * @param {number} value 目标值。
     * @param {number} endTime 渐变结束时间（秒）。
     * @returns {void}
     */
    ExponentialRampToValueAtTime(value, endTime) {
        this.addEvent({
            Kind: AudioParamEventKind.ExponentialRamp,
            Value: value,
            Time: normalizeAudioTime(endTime),
            TimeConstant: 0,
        });
    }
    /**
     * 从起始时间开始以指数方式逼近目标值。
     * @param {number} target 目标值。
     * @param {number} startTime 起始时间（秒）。
     * @param {number} timeConstant 时间常数（秒）。
     * @returns {void}
     */
    SetTargetAtTime(target, startTime, timeConstant) {
        this.addEvent({
            Kind: AudioParamEventKind.Target,
            Value: target,
            Time: normalizeAudioTime(startTime),
            TimeConstant: normalizeAudioTime(timeConstant),
        });
    }
    /**
     * 取消指定时间之后的调度事件，并固化取消前的当前值。
     * @param {number} startTime 取消起点（秒）。
     * @returns {void}
     */
    CancelScheduledValues(startTime) {
        // 先把当前值固化到 now，避免删除事件后参数值瞬间跳回旧基准。
        const time = normalizeAudioTime(startTime);
        const now = this.now();
        this.baseValue = this.valueAt(now);
        this.baseTime = now;
        this.events = this.events.filter((event) => event.Time > now && event.Time < time);
    }
    /**
     * 按事件时间稳定地插入事件。
     * @param {{Kind:number, Value:number, Time:number, TimeConstant:number}} event
     *   要添加的调度事件。
     * @returns {void}
     */
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
    /**
     * 计算参数在任意时间点的值。
     * @param {number} time 查询时间（秒）。
     * @returns {number} 该时间点的参数值。
     */
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
/**
 * 计算普通区间或目标逼近区间的参数值。
 *
 * @param {number} startTime 区间起点。
 * @param {number} startValue 区间起点值。
 * @param {number} time 查询时间。
 * @param {number} mode 当前区间的事件模式。
 * @param {number} target 目标值。
 * @param {number} timeConstant 指数逼近时间常数。
 * @returns {number} 查询时间对应的值。
 */
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
/**
 * 在起止值之间做线性插值。
 *
 * @param {number} startTime 起点时间。
 * @param {number} startValue 起点值。
 * @param {number} endTime 终点时间。
 * @param {number} endValue 终点值。
 * @param {number} time 查询时间。
 * @returns {number} 插值结果。
 */
function interpolateValue(startTime, startValue, endTime, endValue, time) {
    if (endTime <= startTime) {
        return endValue;
    }
    const ratio = Math.max(0, Math.min(1, (time - startTime) / (endTime - startTime)));
    return startValue + (endValue - startValue) * ratio;
}
/**
 * 在正数起止值之间做指数插值；输入不合法时退化为线性插值。
 *
 * @param {number} startTime 起点时间。
 * @param {number} startValue 起点值。
 * @param {number} endTime 终点时间。
 * @param {number} endValue 终点值。
 * @param {number} time 查询时间。
 * @returns {number} 插值结果。
 */
function interpolateExponentialValue(startTime, startValue, endTime, endValue, time) {
    if (startValue <= 0 || endValue <= 0 || endTime <= startTime) {
        return interpolateValue(startTime, startValue, endTime, endValue, time);
    }
    const ratio = Math.max(0, Math.min(1, (time - startTime) / (endTime - startTime)));
    return startValue * Math.pow(endValue / startValue, ratio);
}
////////////////////////////////////////////////////////////////////////////////
// In-memory audio graph
// 内存音频后端的默认参数，保证无音频设备环境也能获得稳定的测试结果。
const MOCK_DEFAULT_SAMPLE_RATE = 44100;
const MOCK_DEFAULT_CHANNEL_COUNT = 2;
const MOCK_DEFAULT_FREQUENCY = 440;
const MOCK_DEFAULT_GAIN = 1;
const MOCK_DEFAULT_PLAYBACK_RATE = 1;
/**
 * 内存音频节点基类，仅记录连接关系而不渲染 PCM。
 */
export class MockAudioNode {
    // 记录连接关系便于测试断言；Mock 后端不实际渲染 PCM。
    Connections = [];
    /**
     * 连接目标节点。
     * @param {MockAudioNode} destination 目标音频节点。
     * @returns {void}
     */
    Connect(destination) {
        if (destination && !this.Connections.includes(destination)) {
            this.Connections.push(destination);
        }
    }
    /** 清除该节点的所有连接。 */
    Disconnect() {
        this.Connections.length = 0;
    }
}
/**
 * 支持起止调度的内存音频源节点基类。
 */
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
    /**
     * @param {MockAudioContext} Context 所属音频上下文。
     */
    constructor(Context) {
        super();
        this.Context = Context;
    }
    /**
     * 计划启动源节点。
     * @param {number} [when] 绝对音频时间（秒）。
     * @returns {void}
     */
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
    /**
     * 计划停止源节点。
     * @param {number} [when] 绝对音频时间（秒）。
     * @returns {void}
     */
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
    /**
     * 注册结束回调；若节点已经结束则立即调用。
     * @param {(() => void)|null} handler 回调函数。
     * @returns {void}
     */
    SetOnEnded(handler) {
        this.onEnded = handler;
        if (this.ended && handler) {
            handler();
        }
    }
    /** @returns {boolean} 源节点是否已经结束。 */
    IsEnded() {
        return this.ended;
    }
    /**
     * 推动源节点状态，供 Mock 上下文模拟时钟。
     * @param {number} now 当前音频时间（秒）。
     * @returns {void}
     */
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
    /** 标记源节点结束并从上下文注销。 */
    internalFinish() {
        if (this.ended || !this.started) {
            return;
        }
        this.ended = true;
        this.Context.internalUnregisterSource(this);
        this.onEnded?.();
    }
}
/**
 * 带频率、失谐和波形参数的内存振荡器节点。
 */
export class MockOscillatorNode extends MockAudioScheduledSourceNode {
    Frequency;
    Detune;
    Type = OscillatorType.Sine;
    /**
     * @param {MockAudioContext} context 所属音频上下文。
     */
    constructor(context) {
        super(context);
        const now = context.GetCurrentTime();
        this.Frequency = new DefaultAudioParam(MOCK_DEFAULT_FREQUENCY, now, () => context.GetCurrentTime());
        this.Detune = new DefaultAudioParam(0, now, () => context.GetCurrentTime());
    }
    /** @returns {DefaultAudioParam} 频率参数（Hz）。 */
    GetFrequency() {
        return this.Frequency;
    }
    /** @returns {DefaultAudioParam} 失谐参数（音分）。 */
    GetDetune() {
        return this.Detune;
    }
    /** @returns {string} 当前振荡器波形。 */
    GetType() {
        return this.Type;
    }
    /**
     * 设置振荡器波形。
     * @param {string} nodeType 波形类型。
     * @returns {void}
     */
    SetType(nodeType) {
        this.Type = nodeType;
    }
}
/**
 * 带增益参数的内存音频节点。
 */
export class MockGainNode extends MockAudioNode {
    Gain;
    /**
     * @param {MockAudioContext} context 所属音频上下文。
     */
    constructor(context) {
        super();
        this.Gain = new DefaultAudioParam(MOCK_DEFAULT_GAIN, context.GetCurrentTime(), () => context.GetCurrentTime());
    }
    /** @returns {DefaultAudioParam} 增益参数。 */
    GetGain() {
        return this.Gain;
    }
}
/**
 * 内存音频图的输出终点。
 */
export class MockAudioDestinationNode extends MockAudioNode {
    MaxChannels;
    /**
     * @param {number} [MaxChannels] 支持的最大声道数。
     */
    constructor(MaxChannels = MOCK_DEFAULT_CHANNEL_COUNT) {
        super();
        this.MaxChannels = MaxChannels;
    }
    /** @returns {number} 最大声道数。 */
    GetMaxChannelCount() {
        return this.MaxChannels;
    }
}
/**
 * 由普通 JavaScript 数组承载 PCM 数据的内存音频缓冲区。
 */
export class MockAudioBuffer {
    SampleRate;
    Channels;
    /**
     * @param {number} SampleRate 采样率（Hz）。
     * @param {number[][]} Channels 每个声道的采样数组。
     */
    constructor(SampleRate, Channels) {
        this.SampleRate = SampleRate;
        this.Channels = Channels;
    }
    /** @returns {number} 采样率（Hz）。 */
    GetSampleRate() {
        return this.SampleRate;
    }
    /** @returns {number} 每个声道的采样帧数。 */
    GetLength() {
        return this.Channels.length === 0 ? 0 : this.Channels[0].length;
    }
    /** @returns {number} 声道数。 */
    GetNumberOfChannels() {
        return this.Channels.length;
    }
    /**
     * 获取指定声道的采样数据。
     * @param {number} channel 声道索引。
     * @returns {number[]|null} 声道数据；越界时返回 `null`。
     */
    GetChannelData(channel) {
        if (channel < 0 || channel >= this.Channels.length) {
            return null;
        }
        return this.Channels[channel];
    }
}
/**
 * 播放内存音频缓冲区的可调度源节点。
 */
export class MockAudioBufferSourceNode extends MockAudioScheduledSourceNode {
    PlaybackRate;
    Detune;
    Buffer = null;
    /**
     * @param {MockAudioContext} context 所属音频上下文。
     */
    constructor(context) {
        super(context);
        const now = context.GetCurrentTime();
        this.PlaybackRate = new DefaultAudioParam(MOCK_DEFAULT_PLAYBACK_RATE, now, () => context.GetCurrentTime());
        this.Detune = new DefaultAudioParam(0, now, () => context.GetCurrentTime());
    }
    /** @returns {MockAudioBuffer|null} 当前缓冲区。 */
    GetBuffer() {
        return this.Buffer;
    }
    /**
     * 设置要播放的缓冲区。
     * @param {MockAudioBuffer|null} buffer 新缓冲区。
     * @returns {void}
     */
    SetBuffer(buffer) {
        this.Buffer = buffer;
    }
    /** @returns {DefaultAudioParam} 播放速率参数。 */
    GetPlaybackRate() {
        return this.PlaybackRate;
    }
    /** @returns {DefaultAudioParam} 失谐参数。 */
    GetDetune() {
        return this.Detune;
    }
}
/**
 * 创建 Mock 音频上下文的工厂。
 */
export class MockAudioContextFactory {
    SampleRate;
    /**
     * @param {number} [SampleRate] 新上下文的采样率。
     */
    constructor(SampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
        this.SampleRate = SampleRate;
    }
    /** @returns {MockAudioContext} 新的内存音频上下文。 */
    NewAudioContext() {
        return new MockAudioContext(this.SampleRate);
    }
}
/**
 * 可手动推进、暂停和关闭的内存 AudioContext。
 */
export class MockAudioContext {
    sampleRate;
    state = AudioContextState.Suspended;
    destination;
    sources = new Set();
    elapsed = 0;
    runningSince = 0;
    sourceCheck = null;
    /**
     * @param {number} [sampleRate] 上下文采样率。
     */
    constructor(sampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
        this.sampleRate = sampleRate;
        this.destination = new MockAudioDestinationNode(MOCK_DEFAULT_CHANNEL_COUNT);
    }
    /** 关闭上下文、结束所有源并释放定时器。 */
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
    /** 恢复时钟和音频源运行。 */
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
    /** 暂停时钟；已累计的音频时间会保留。 */
    Suspend() {
        if (this.state === AudioContextState.Running) {
            this.freezeClock();
        }
        if (this.state !== AudioContextState.Closed) {
            this.state = AudioContextState.Suspended;
        }
    }
    /** @returns {MockOscillatorNode} 新的振荡器节点。 */
    CreateOscillator() {
        return new MockOscillatorNode(this);
    }
    /** @returns {MockGainNode} 新的增益节点。 */
    CreateGain() {
        return new MockGainNode(this);
    }
    /**
     * 创建一个由零填充的内存音频缓冲区。
     * @param {number} numberOfChannels 声道数。
     * @param {number} length 每个声道的采样帧数。
     * @param {number} [sampleRate] 采样率；缺省时使用上下文采样率。
     * @returns {MockAudioBuffer} 新缓冲区。
     */
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
    /** @returns {MockAudioBufferSourceNode} 新的缓冲区源节点。 */
    CreateBufferSource() {
        return new MockAudioBufferSourceNode(this);
    }
    /**
     * 获取当前音频时间，并顺带推动源节点完成状态。
     * @returns {number} 从上下文创建起累计的秒数。
     */
    GetCurrentTime() {
        this.runSources();
        return this.currentTime();
    }
    /** @returns {number} 当前采样率。 */
    GetSampleRate() {
        return this.sampleRate;
    }
    /** @returns {MockAudioDestinationNode} 音频输出终点。 */
    GetDestination() {
        return this.destination;
    }
    /** @returns {string} 当前上下文状态。 */
    GetState() {
        return this.state;
    }
    /**
     * 注册待检查的音频源。
     * @param {MockAudioScheduledSourceNode} source 音频源。
     * @returns {void}
     */
    internalRegisterSource(source) {
        if (this.state !== AudioContextState.Closed) {
            this.sources.add(source);
            this.scheduleSourceCheck();
        }
    }
    /**
     * 注销已经完成的音频源。
     * @param {MockAudioScheduledSourceNode} source 音频源。
     * @returns {void}
     */
    internalUnregisterSource(source) {
        this.sources.delete(source);
    }
    /** @returns {number} 根据运行状态计算出的音频时间。 */
    currentTime() {
        if (this.state !== AudioContextState.Running) {
            return this.elapsed;
        }
        return this.elapsed + (Date.now() - this.runningSince) / 1000;
    }
    /** 将运行中的时间增量固化到 `elapsed`。 */
    freezeClock() {
        if (this.state !== AudioContextState.Running) {
            return;
        }
        this.elapsed += (Date.now() - this.runningSince) / 1000;
        this.runningSince = 0;
    }
    /** 用当前音频时间推动所有已注册源节点。 */
    runSources() {
        if (this.state !== AudioContextState.Running) {
            return;
        }
        const now = this.currentTime();
        for (const source of [...this.sources]) {
            source.internalTick(now);
        }
    }
    /** 安排低频检查，确保异步结束的源最终触发回调。 */
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
/**
 * 使用 Mock 音频上下文的驱动，适合测试和无音频设备环境。
 */
export class MockDriver {
    sampleRate;
    /**
     * @param {number} [sampleRate] 音频上下文采样率。
     */
    constructor(sampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
        this.sampleRate = sampleRate;
    }
    /** @returns {MockAudioContextFactory} 当前驱动对应的上下文工厂。 */
    GetAudioContextFactory() {
        return new MockAudioContextFactory(this.sampleRate);
    }
}
/**
 * 创建 Mock 音频驱动。
 * @param {number} [sampleRate] 音频上下文采样率。
 * @returns {MockDriver} 新驱动实例。
 */
export function NewMockDriver(sampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
    return new MockDriver(sampleRate);
}
/**
 * 将浏览器原生 `AudioParam` 转换为项目统一的参数接口。
 */
class NativeAudioParamAdapter {
    param;
    /**
     * @param {AudioParam} param 原生音频参数。
     */
    constructor(param) {
        this.param = param;
    }
    /** @returns {number} 原生参数当前值。 */
    GetValue() {
        return Number(this.param.value);
    }
    /**
     * 立即设置原生参数值。
     * @param {number} value 新值。
     * @returns {void}
     */
    SetValue(value) {
        this.param.value = value;
    }
    /**
     * 转发 `setValueAtTime`。
     * @param {number} value 目标值。
     * @param {number} startTime 生效时间。
     * @returns {void}
     */
    SetValueAtTime(value, startTime) {
        this.param.setValueAtTime(value, startTime);
    }
    /**
     * 转发 `linearRampToValueAtTime`。
     * @param {number} value 目标值。
     * @param {number} endTime 渐变结束时间。
     * @returns {void}
     */
    LinearRampToValueAtTime(value, endTime) {
        this.param.linearRampToValueAtTime(value, endTime);
    }
    /**
     * 转发 `exponentialRampToValueAtTime`。
     * @param {number} value 目标值。
     * @param {number} endTime 渐变结束时间。
     * @returns {void}
     */
    ExponentialRampToValueAtTime(value, endTime) {
        this.param.exponentialRampToValueAtTime(value, endTime);
    }
    /**
     * 转发 `setTargetAtTime`。
     * @param {number} target 目标值。
     * @param {number} startTime 起始时间。
     * @param {number} timeConstant 时间常数。
     * @returns {void}
     */
    SetTargetAtTime(target, startTime, timeConstant) {
        this.param.setTargetAtTime(target, startTime, timeConstant);
    }
    /**
     * 转发 `cancelScheduledValues`。
     * @param {number} startTime 取消起点。
     * @returns {void}
     */
    CancelScheduledValues(startTime) {
        this.param.cancelScheduledValues(startTime);
    }
}
/**
 * 将浏览器原生 `AudioNode` 转换为项目统一的节点接口。
 */
class NativeAudioNodeAdapter {
    node;
    /**
     * @param {AudioNode} node 原生音频节点。
     */
    constructor(node) {
        this.node = node;
    }
    /**
     * 连接到另一个适配节点。
     * @param {NativeAudioNodeAdapter|{NativeNode:AudioNode}} destination 目标节点。
     * @returns {void}
     */
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
    /** 断开该原生节点的全部连接。 */
    Disconnect() {
        this.node.disconnect();
    }
    /** @returns {AudioNode} 被包装的原生节点。 */
    nativeNode() {
        return this.node;
    }
}
/**
 * 原生可调度音频源的公共适配器。
 */
class NativeAudioScheduledSourceAdapter extends NativeAudioNodeAdapter {
    /**
     * 在指定时间启动原生源。
     * @param {number} [when] 绝对音频时间。
     * @returns {void}
     */
    Start(when) {
        this.node.start(normalizeAudioTime(when));
    }
    /**
     * 在指定时间停止原生源。
     * @param {number} [when] 绝对音频时间。
     * @returns {void}
     */
    Stop(when) {
        this.node.stop(normalizeAudioTime(when));
    }
    /**
     * 设置原生 `onended` 回调。
     * @param {(() => void)|null} handler 结束回调。
     * @returns {void}
     */
    SetOnEnded(handler) {
        this.node.onended = handler;
    }
}
/**
 * 原生振荡器节点适配器。
 */
class NativeOscillatorAdapter extends NativeAudioScheduledSourceAdapter {
    /** @returns {NativeAudioParamAdapter} 频率参数。 */
    GetFrequency() {
        return new NativeAudioParamAdapter(this.node.frequency);
    }
    /** @returns {NativeAudioParamAdapter} 失谐参数。 */
    GetDetune() {
        return new NativeAudioParamAdapter(this.node.detune);
    }
    /** @returns {string} 原生振荡器波形。 */
    GetType() {
        return this.node.type;
    }
    /**
     * 设置原生振荡器波形。
     * @param {string} nodeType 波形类型。
     * @returns {void}
     */
    SetType(nodeType) {
        this.node.type = nodeType;
    }
}
/**
 * 原生增益节点适配器。
 */
class NativeGainAdapter extends NativeAudioNodeAdapter {
    /** @returns {NativeAudioParamAdapter} 增益参数。 */
    GetGain() {
        return new NativeAudioParamAdapter(this.node.gain);
    }
}
/**
 * 原生音频输出终点适配器。
 */
class NativeDestinationAdapter extends NativeAudioNodeAdapter {
    /** @returns {number} 原生输出支持的最大声道数。 */
    GetMaxChannelCount() {
        return Number(this.node.maxChannelCount ?? 2);
    }
}
/**
 * 原生 `AudioBuffer` 适配器。
 */
class NativeAudioBufferAdapter {
    buffer;
    /**
     * @param {AudioBuffer} buffer 原生音频缓冲区。
     */
    constructor(buffer) {
        this.buffer = buffer;
    }
    /** @returns {number} 缓冲区采样率。 */
    GetSampleRate() {
        return Number(this.buffer.sampleRate);
    }
    /** @returns {number} 每个声道的采样帧数。 */
    GetLength() {
        return Number(this.buffer.length);
    }
    /** @returns {number} 声道数。 */
    GetNumberOfChannels() {
        return Number(this.buffer.numberOfChannels);
    }
    /**
     * 获取原生声道数据。
     * @param {number} channel 声道索引。
     * @returns {Float32Array|null} 声道数据；越界时返回 `null`。
     */
    GetChannelData(channel) {
        if (channel < 0 || channel >= this.GetNumberOfChannels()) {
            return null;
        }
        return this.buffer.getChannelData(channel);
    }
}
/**
 * 原生缓冲区源节点适配器。
 */
class NativeBufferSourceAdapter extends NativeAudioScheduledSourceAdapter {
    /** @returns {NativeAudioBufferAdapter|null} 当前缓冲区适配器。 */
    GetBuffer() {
        return this.node.buffer ? new NativeAudioBufferAdapter(this.node.buffer) : null;
    }
    /**
     * 设置缓冲区，并在后端类型不同时复制声道数据。
     * @param {NativeAudioBufferAdapter|{GetNumberOfChannels:Function, GetLength:Function, GetSampleRate:Function, GetChannelData:Function}|null} buffer
     *   新缓冲区。
     * @returns {void}
     */
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
    /** @returns {NativeAudioParamAdapter} 播放速率参数。 */
    GetPlaybackRate() {
        return new NativeAudioParamAdapter(this.node.playbackRate);
    }
    /** @returns {NativeAudioParamAdapter} 失谐参数。 */
    GetDetune() {
        return new NativeAudioParamAdapter(this.node.detune);
    }
}
/**
 * 浏览器原生 `AudioContext` 的接口适配层。
 */
export class WebAudioContext {
    context;
    // 把本文件的 Web Audio 抽象转发到浏览器原生 AudioContext。
    destination;
    /**
     * @param {AudioContext} context 原生音频上下文。
     */
    constructor(context) {
        this.context = context;
        this.destination = new NativeDestinationAdapter(context.destination);
    }
    /** 请求关闭原生音频上下文。 */
    Close() {
        void this.context.close();
    }
    /** @returns {Promise<void>} 原生上下文的恢复操作。 */
    Resume() {
        return this.context.resume();
    }
    /** @returns {Promise<void>} 原生上下文的暂停操作。 */
    Suspend() {
        return this.context.suspend();
    }
    /** @returns {NativeOscillatorAdapter} 原生振荡器适配器。 */
    CreateOscillator() {
        return new NativeOscillatorAdapter(this.context.createOscillator());
    }
    /** @returns {NativeGainAdapter} 原生增益节点适配器。 */
    CreateGain() {
        return new NativeGainAdapter(this.context.createGain());
    }
    /**
     * 创建原生音频缓冲区。
     * @param {number} numberOfChannels 声道数。
     * @param {number} length 每声道采样帧数。
     * @param {number} [sampleRate] 采样率。
     * @returns {NativeAudioBufferAdapter} 缓冲区适配器。
     */
    CreateBuffer(numberOfChannels, length, sampleRate) {
        const rate = sampleRate || this.context.sampleRate;
        return new NativeAudioBufferAdapter(this.context.createBuffer(numberOfChannels, length, rate));
    }
    /** @returns {NativeBufferSourceAdapter} 原生缓冲区源适配器。 */
    CreateBufferSource() {
        return new NativeBufferSourceAdapter(this.context.createBufferSource());
    }
    /** @returns {number} 原生上下文当前音频时间。 */
    GetCurrentTime() {
        return this.context.currentTime;
    }
    /** @returns {number} 原生上下文采样率。 */
    GetSampleRate() {
        return this.context.sampleRate;
    }
    /** @returns {NativeDestinationAdapter} 原生输出终点适配器。 */
    GetDestination() {
        return this.destination;
    }
    /** @returns {string} 原生上下文状态。 */
    GetState() {
        return this.context.state;
    }
}
/**
 * 浏览器 Web Audio 上下文工厂。
 */
export class WebAudioContextFactory {
    /**
     * 创建 Web Audio 上下文。
     * @returns {WebAudioContext} 上下文适配器。
     * @throws {Error} 当前环境不提供 Web Audio API 时抛出。
     */
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
/**
 * 浏览器 Web Audio 驱动。
 */
export class WebAudioDriver {
    // 浏览器推荐使用该 Driver；无 Web Audio 的环境可改用 MockDriver。
    /** @returns {WebAudioContextFactory} Web Audio 上下文工厂。 */
    GetAudioContextFactory() {
        return new WebAudioContextFactory();
    }
}
////////////////////////////////////////////////////////////////////////////////
// Timeline and positions
/**
 * 时间线层级类型。`Root` 表示绝对时间根，`Work` 表示可重定位的工作时间线。
 */
export var TimelineType;
(function (TimelineType) {
    TimelineType[TimelineType["Min"] = 0] = "Min";
    TimelineType[TimelineType["Root"] = 1] = "Root";
    TimelineType[TimelineType["Work"] = 2] = "Work";
    TimelineType[TimelineType["Clip"] = 3] = "Clip";
    TimelineType[TimelineType["Max"] = 4] = "Max";
})(TimelineType || (TimelineType = {}));
/**
 * 可嵌套的时间线节点。
 *
 * 相对位置通过父链上的 `Offset` 累加为绝对时间；`Profile` 保存该时间线的速度和拍号。
 */
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
    /**
     * 在必要字段缺失时填充默认值。
     * @returns {Timeline} 当前实例，便于链式调用。
     */
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
    /**
     * 判断时间线类型。
     * @param {number} type 要匹配的类型。
     * @returns {boolean} 类型是否一致。
     */
    IsTypeOf(type) {
        return this.Type === type;
    }
    /**
     * 判断两个时间线是否属于同一个根时间线。
     * @param {Timeline|null} other 另一个时间线。
     * @returns {boolean} 根节点是否相同。
     */
    HasSameRoot(other) {
        return this.GetRoot().Same(other?.GetRoot() ?? null);
    }
    /**
     * 判断是否为同一个时间线实例。
     * @param {Timeline|null} other 另一个时间线。
     * @returns {boolean} 实例是否相同。
     */
    Same(other) {
        return this === other;
    }
    /**
     * 递归比较时间线标签、偏移、长度、循环标记和父链。
     * @param {Timeline|null} other 另一个时间线。
     * @returns {boolean} 内容是否相等。
     */
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
    /** @returns {Timeline} 沿父链找到的根时间线。 */
    GetRoot() {
        let current = this;
        for (let ttl = 32; ttl > 0 && current.Parent; ttl--) {
            current = current.Parent;
        }
        return current;
    }
    /** @returns {number} 当前时间线及其所有祖先的累计 `Offset`。 */
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
/**
 * 标识输入时间值是绝对音频时间。
 *
 * 该函数不改变数值，仅让调用处明确单位/坐标系语义。
 *
 * @param {number} value 绝对时间（秒）。
 * @returns {number} 原值。
 */
export function AbsolutePosition(value) {
    return value;
}
/**
 * 将已经是绝对坐标的位置原样返回。
 *
 * @param {number} position 绝对位置（秒）。
 * @returns {number} 原值。
 */
export function RelativePositionToAbsolute(position) {
    return position;
}
/**
 * 相对于某条时间线的位置。
 */
export class RelativePosition {
    Ref;
    Pos;
    /**
     * @param {Timeline|null} [Ref] 参照时间线。
     * @param {number} [Pos] 相对于参照时间线的秒数。
     */
    constructor(Ref = null, Pos = 0) {
        this.Ref = Ref;
        this.Pos = Pos;
    }
    /** @returns {number} 换算后的绝对时间（秒）。 */
    Absolute() {
        return (this.Ref?.GetTotalOffset() ?? 0) + this.Pos;
    }
    /** @returns {RelativePosition} 保持相同参照和位置的新对象。 */
    Clone() {
        return new RelativePosition(this.Ref, this.Pos);
    }
}
/**
 * 将绝对时间转换为相对于指定时间线的位置。
 *
 * @param {number} absolute 绝对时间（秒）。
 * @param {Timeline|null} [base] 目标参照时间线。
 * @param {RelativePosition} [destination] 可复用的结果对象。
 * @returns {RelativePosition} 相对位置。
 */
export function AbsolutePositionRelative(absolute, base, destination) {
    // 公式：relative = absolute - base.GetTotalOffset()。
    // 每次返回新 RelativePosition；调用方主动传入 destination 可复用对象。
    const dst = destination ?? new RelativePosition();
    dst.Ref = base ?? new Timeline();
    dst.Pos = absolute - dst.Ref.GetTotalOffset();
    return dst;
}
/**
 * 基于 `setTimeout` 的重复定时器。
 */
export class Timer {
    // Timer 使用 setTimeout 而非阻塞循环。Start/Stop 通过 workerID
    // 让已排队的旧回调自动失效。
    Delay = 0;
    Interval = 0;
    Handler = null;
    workerID = 0;
    timeout = null;
    /** 从 `Delay` 开始按 `Interval` 重复调用 `Handler`。 */
    Start() {
        this.Stop();
        const worker = this.workerID;
        const handler = this.Handler;
        if (!handler) {
            return;
        }
        this.work(worker, handler, this.Delay);
    }
    /** 停止定时器，并让已排队的旧回调失效。 */
    Stop() {
        this.workerID++;
        if (this.timeout !== null) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }
    }
    /**
     * 调度一次定时器回调，并在成功后继续下一轮。
     * @param {number} worker 当前定时器代次。
     * @param {(timer: Timer) => void} handler 回调。
     * @param {number} delay 延迟秒数。
     * @returns {void}
     */
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
    /**
     * 将秒转换为 `setTimeout` 使用的毫秒，并保证至少 1ms。
     * @param {number} seconds 秒数。
     * @returns {number} 毫秒数。
     */
    sec2ms(seconds) {
        return Math.max(1, Math.trunc(seconds * 1000));
    }
}
////////////////////////////////////////////////////////////////////////////////
// Event dispatch
/**
 * 事件分发系统支持的循环边界事件类型。
 */
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
/**
 * 事件缓冲区和回调注册表的共享上下文。
 */
export class EventContext {
    // 事件先进入 buffer，再由 Runner 按 FIFO 取出。回调中 Dispatch 的新事件
    // 会追加到队尾，因此 Run 仍能在同一次调用中处理它。
    dispatcherInstance = null;
    runnerInstance = null;
    callbacks = [];
    buffer = [];
    /** 创建默认分发器和运行器。 */
    constructor() {
        this.dispatcherInstance = new DefaultEventDispatcher(this);
        this.runnerInstance = new DefaultEventRunner(this);
    }
    /** @returns {DefaultEventDispatcher} 默认事件分发器。 */
    GetDispatcher() {
        return this.dispatcherInstance;
    }
    /** @returns {DefaultEventRunner} 默认事件运行器。 */
    GetRunner() {
        return this.runnerInstance;
    }
    /**
     * 将事件副本追加到队列。
     * @param {{Type:number, ID:number}} event 待分发事件。
     * @returns {void}
     */
    internalAdd(event) {
        this.buffer.push({ ...event });
    }
    /**
     * 注册事件回调。
     * @param {{Type:number, ID:number, Fn:Function}} callback 回调描述。
     * @returns {void}
     */
    internalAddCallback(callback) {
        this.callbacks.push(callback);
    }
    /**
     * 注销事件回调。
     * @param {{Type:number, ID:number, Fn:Function}} callback 回调描述。
     * @returns {boolean} 是否找到并移除。
     */
    internalRemoveCallback(callback) {
        const index = this.callbacks.indexOf(callback);
        if (index >= 0) {
            this.callbacks.splice(index, 1);
            return true;
        }
        return false;
    }
    /** @returns {{Type:number, ID:number}|undefined} 队首事件。 */
    internalTake() {
        // 先取出再执行回调；如果回调抛错，当前事件已被消费，后续事件仍保留。
        return this.buffer.shift();
    }
    /**
     * 查找与事件类型和 ID 匹配的可调用回调。
     * @param {{Type:number, ID:number}} event 当前事件。
     * @returns {Array<{Type:number, ID:number, Fn:Function}>} 匹配的回调列表。
     */
    internalMatching(event) {
        return this.callbacks.filter((callback) => callback.Type === event.Type &&
            callback.ID === event.ID &&
            typeof callback.Fn === "function");
    }
    /** @returns {number} 尚未处理的事件数量。 */
    BufferLength() {
        return this.buffer.length;
    }
}
/**
 * 默认事件分发器，负责把事件和回调登记到 `EventContext`。
 */
export class DefaultEventDispatcher {
    context;
    /**
     * @param {EventContext} context 事件上下文。
     */
    constructor(context) {
        this.context = context;
    }
    /**
     * 提交一个事件；空事件会被忽略。
     * @param {{Type:number, ID:number}|null} event 待处理事件。
     * @returns {void}
     */
    Dispatch(event) {
        if (!event) {
            return;
        }
        if (!this.context) {
            throw new Error("event dispatcher context is nil");
        }
        this.context.internalAdd(event);
    }
    /**
     * 注册事件回调。
     * @param {{Type:number, ID:number, Fn:Function}} callback 回调描述。
     * @returns {void}
     */
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
    /**
     * 注销已注册的事件回调。
     * @param {{Type:number, ID:number, Fn:Function}} callback 回调描述。
     * @returns {void}
     */
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
/**
 * 按 FIFO 顺序同步处理事件队列的默认运行器。
 */
export class DefaultEventRunner {
    context;
    /**
     * @param {EventContext} context 事件上下文。
     */
    constructor(context) {
        this.context = context;
    }
    /** 处理默认上下文中的所有当前及新追加事件。 */
    Run() {
        if (!this.context) {
            throw new Error("event runner context is nil");
        }
        this.RunContext(this.context);
    }
    /**
     * 处理指定上下文中的全部事件。
     * @param {EventContext} context 事件上下文。
     * @returns {void}
     */
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
/**
 * 创建默认事件上下文。
 * @returns {EventContext} 新事件上下文。
 */
export function NewEventContext() {
    return new EventContext();
}
/**
 * 组件生命周期回调集合。
 *
 * 除 `OnInit` 接收引擎上下文外，其余阶段回调均可同步或异步执行。
 */
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
/**
 * 生命周期管理器内部共享的运行状态。
 */
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
/**
 * 管理和按顺序驱动已注册组件的生命周期。
 */
export class DefaultLifeManager {
    state = new LifeCycleContext();
    mainLife;
    /** 创建主生命周期并绑定内部状态。 */
    constructor() {
        this.mainLife = new MainLife(this.state);
        this.state.main = this.mainLife;
    }
    /** @returns {MainLife} 驱动所有已注册组件的主生命周期。 */
    GetMain() {
        return this.mainLife;
    }
    /**
     * 注册组件；若管理器已经进入某个阶段，会立即回放当前阶段。
     * @param {{GetLifeCycle:Function}|null} life 生命周期组件。
     * @returns {void}
     */
    Add(life) {
        if (!life) {
            return;
        }
        this.state.all.push(life);
        this.state.cache = null;
        replayCurrentLifeCycle(life, this.state);
    }
    /**
     * 注销组件。
     * @param {{GetLifeCycle:Function}|null} life 生命周期组件。
     * @returns {void}
     */
    Remove(life) {
        if (!life) {
            return;
        }
        this.state.all = this.state.all.filter((item) => item !== life);
        this.state.cache = null;
    }
}
/**
 * 把所有组件的生命周期聚合为一个对外的主生命周期。
 */
class MainLife {
    state;
    /**
     * @param {LifeCycleContext} state 生命周期管理器状态。
     */
    constructor(state) {
        this.state = state;
    }
    /** @returns {LifeCycle} 聚合后的生命周期回调集合。 */
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
    /**
     * 初始化上下文并调用所有组件的 `OnInit`。
     * @param {EngineContext} context 引擎上下文。
     * @returns {void}
     */
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
    /**
     * 按注册顺序调用指定阶段；清理阶段可反向执行。
     * @param {string} stage 生命周期回调名。
     * @param {boolean} reverse 是否反向遍历。
     * @returns {Promise<void>}
     */
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
    /** @returns {LifeCycle[]} 按 `Order` 排序的生命周期缓存。 */
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
/**
 * 将已经发生的生命周期阶段补发给后注册的组件。
 *
 * @param {{GetLifeCycle:Function}} life 新注册组件。
 * @param {LifeCycleContext} state 当前生命周期阶段状态。
 * @returns {void}
 */
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
/**
 * 创建默认生命周期管理器。
 * @returns {DefaultLifeManager} 新管理器。
 */
export function NewLifeManager() {
    return new DefaultLifeManager();
}
/**
 * 手动按完整阶段顺序运行一个生命周期。
 */
export class LifeCycleRunner {
    // 主生命周期驱动引擎：OnRun 通常会等待 Engine.Join()，
    // 所以后续 Pause/Stop/Destroy 必须等它结束后才能执行。
    context = null;
    /**
     * 绑定生命周期初始化上下文。
     * @param {EngineContext} context 引擎上下文。
     * @returns {LifeCycleRunner} 当前实例，便于链式调用。
     */
    Init(context) {
        this.context = context;
        return this;
    }
    /**
     * 依次执行初始化、创建、启动、运行及清理阶段。
     * @param {LifeCycle|null} cycle 生命周期回调集合。
     * @returns {Promise<void>}
     */
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
    /**
     * 运行组件的生命周期。
     * @param {{GetLifeCycle:Function}|null} life 生命周期组件。
     * @returns {Promise<void>}
     */
    RunL(life) {
        return this.RunLC(life?.GetLifeCycle() ?? null);
    }
    /**
     * 运行生命周期管理器的主生命周期。
     * @param {{GetMain:Function}} manager 生命周期管理器。
     * @returns {Promise<void>}
     */
    RunLM(manager) {
        return this.RunL(manager.GetMain());
    }
}
/**
 * 记录每个生命周期阶段调用顺序的诊断组件。
 */
export class LifeCycleStageLogger {
    // 可作为配置中的 Part 加入，用于观察每个生命周期阶段的调用顺序。
    logger = Log();
    /** @returns {PartInfo} 用于注册到 Part 管理器的组件信息。 */
    GetInfo() {
        return new PartInfo({ Name: "LifeCycleStageLogger", Part: this });
    }
    /** @returns {LifeCycle} 记录所有阶段的生命周期回调。 */
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
    /**
     * 记录一个生命周期阶段。
     * @param {string} stage 阶段名称。
     * @returns {void}
     */
    log(stage) {
        this.logger.Info("engine.lifecycle.%s()", stage);
    }
}
/**
 * Part 注册信息的基础数据结构。
 *
 * 同一个对象可表示已有实例、工厂或注册记录；具体由 `Part` 和 `Factory` 字段决定。
 */
export class PartMeta {
    ID = 0;
    Name = "";
    Classes = [];
    Context = null;
    Factory = null;
    Part = null;
    /**
     * @param {Partial<PartMeta>} [init] 初始字段。
     */
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
/**
 * Part 实例向注册表提供的描述信息。
 */
export class PartInfo extends PartMeta {
}
/**
 * 注册表内部保存的 Part 注册记录。
 */
export class PartRegistration extends PartMeta {
}
/**
 * Part 注册表、管理器及其共享状态的容器。
 */
export class PartContext {
    EC;
    Registry;
    Manager;
    /**
     * @param {EngineContext|null} [context] 引擎上下文。
     */
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
/**
 * `PartInfo` 的链式构建器。
 */
export class PartInfoBuilder {
    info = new PartInfo();
    /** @returns {PartInfo} 当前字段的副本。 */
    Build() {
        return new PartInfo(this.info);
    }
    /**
     * 设置 Part 实例。
     * @param {object} part Part 实例。
     * @returns {PartInfoBuilder} 当前构建器。
     */
    SetPart(part) {
        this.info.Part = part;
        return this;
    }
    /**
     * 设置 Part 工厂。
     * @param {object} factory Part 工厂。
     * @returns {PartInfoBuilder} 当前构建器。
     */
    SetFactory(factory) {
        this.info.Factory = factory;
        return this;
    }
    /**
     * 设置 Part 名称。
     * @param {string} name 唯一名称。
     * @returns {PartInfoBuilder} 当前构建器。
     */
    SetName(name) {
        this.info.Name = name;
        return this;
    }
    /**
     * 设置 Part ID。
     * @param {number} id 唯一 ID。
     * @returns {PartInfoBuilder} 当前构建器。
     */
    SetID(id) {
        this.info.ID = id;
        return this;
    }
    /**
     * 设置 Part 所属上下文。
     * @param {PartContext|null} context Part 上下文。
     * @returns {PartInfoBuilder} 当前构建器。
     */
    SetContext(context) {
        this.info.Context = context;
        return this;
    }
    /**
     * 设置 Part 的能力标签。
     * @param {...string} classes 能力标签。
     * @returns {PartInfoBuilder} 当前构建器。
     */
    SetClasses(...classes) {
        this.info.Classes = classes;
        return this;
    }
}
/**
 * 按唯一 ID 和名称登记 Part 实例或工厂。
 */
export class DefaultPartRegistry {
    state;
    /**
     * @param {object} state Part 上下文的共享状态。
     */
    constructor(state) {
        this.state = state;
    }
    /**
     * 注册工厂提供的 Part 描述。
     * @param {{GetRegistration:Function}} factory Part 工厂。
     * @returns {void}
     */
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
    /**
     * 注册已有的 Part 实例。
     * @param {{GetInfo:Function}} part Part 实例。
     * @returns {void}
     */
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
    /**
     * 校验并登记一条注册记录。
     * @param {PartRegistration} registration 注册记录。
     * @returns {void}
     */
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
    /** @returns {string[]} 按字典序排列的已注册名称。 */
    ListNames() {
        return [...this.state.byName.keys()].sort();
    }
    /**
     * 按 ID 查找注册记录。
     * @param {number} id Part ID。
     * @returns {PartRegistration|null} 注册记录。
     */
    FindByID(id) {
        return this.state.byID.get(id) ?? null;
    }
    /**
     * 按名称查找注册记录。
     * @param {string} name Part 名称。
     * @returns {PartRegistration|null} 注册记录。
     */
    FindByName(name) {
        return this.state.byName.get(name) ?? null;
    }
    /**
     * 列出具有指定能力标签的注册记录。
     * @param {string} partClass 能力标签。
     * @returns {PartRegistration[]} 匹配的注册记录。
     */
    ListWithClass(partClass) {
        return [...this.state.byID.values()]
            .filter((item) => item.Classes.includes(partClass))
            .sort((a, b) => a.ID === b.ID ? a.Name.localeCompare(b.Name) : a.ID - b.ID);
    }
    /** @returns {number} 自动生成的唯一 Part ID。 */
    makeNextID() {
        if (this.state.nextID === 0) {
            this.state.nextID = 0x123456789;
        }
        this.state.nextID++;
        return this.state.nextID;
    }
}
/**
 * 按注册信息延迟创建 Part，并缓存每个注册项对应的实例。
 */
export class DefaultPartManager {
    state;
    // Manager 按需调用 Factory.CreatePart，并缓存创建结果，
    // 因此相同 ID/Name 的后续查询会返回同一个实例。
    /**
     * @param {object} state Part 上下文的共享状态。
     */
    constructor(state) {
        this.state = state;
    }
    /** @returns {string[]} 已注册 Part 的名称。 */
    ListNames() {
        return [...this.state.byName.keys()].sort();
    }
    /**
     * 按 ID 获取或创建 Part。
     * @param {number} id Part ID。
     * @returns {object|null} Part 实例；未注册时为 `null`。
     */
    FindPartByID(id) {
        const registration = this.state.byID.get(id);
        return registration ? this.partFromRegistration(registration) : null;
    }
    /**
     * 按名称获取或创建 Part。
     * @param {string} name Part 名称。
     * @returns {object|null} Part 实例；未注册时为 `null`。
     */
    FindPartByName(name) {
        const registration = this.state.byName.get(name);
        return registration ? this.partFromRegistration(registration) : null;
    }
    /**
     * 获取所有具有指定能力标签的 Part。
     * @param {string} partClass 能力标签。
     * @returns {object[]} Part 实例列表。
     */
    ListPartsWithClass(partClass) {
        return [...this.state.byID.values()]
            .filter((item) => item.Classes.includes(partClass))
            .sort((a, b) => a.ID === b.ID ? a.Name.localeCompare(b.Name) : a.ID - b.ID)
            .map((item) => this.partFromRegistration(item))
            .filter((part) => part !== null);
    }
    /**
     * 返回注册项对应的缓存实例，必要时调用工厂创建。
     * @param {PartRegistration} registration 注册记录。
     * @returns {object} Part 实例。
     */
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
    /**
     * 按 ID 和名称缓存 Part 实例。
     * @param {PartRegistration} registration 注册记录。
     * @param {object} part Part 实例。
     * @returns {void}
     */
    cache(registration, part) {
        this.state.managerByID.set(registration.ID, part);
        this.state.managerByName.set(registration.Name, part);
    }
}
/**
 * 创建 Part 上下文及其注册表和管理器。
 * @param {EngineContext|null} [context] 引擎上下文。
 * @returns {PartContext} 新 Part 上下文。
 */
export function NewPartContext(context = null) {
    return new PartContext(context);
}
////////////////////////////////////////////////////////////////////////////////
// MIDI and instruments
/**
 * MIDI 音符事件类型。
 */
export var MidiEventType;
(function (MidiEventType) {
    // 首尾哨兵沿用 Go iota 结构，业务事件从 NoteOn 开始。
    MidiEventType[MidiEventType["Min"] = 0] = "Min";
    MidiEventType[MidiEventType["NoteOn"] = 1] = "NoteOn";
    MidiEventType[MidiEventType["NoteOff"] = 2] = "NoteOff";
    MidiEventType[MidiEventType["Max"] = 3] = "Max";
})(MidiEventType || (MidiEventType = {}));
/**
 * 单个 MIDI 音符事件的调度数据。
 */
export class MidiNote {
    // Position 是绝对音频时间。小于 1ms 的值由乐器解释为“立即播放”。
    EventType = MidiEventType.NoteOn;
    Position = 0;
    Channel = 0;
    Key = 0;
    Velocity = 0;
    /**
     * @param {Partial<MidiNote>} [init] 初始字段。
     */
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
/**
 * 同时承载 NoteOn/NoteOff 的 MIDI 调度消息。
 */
export class MidiMessage {
    // 一个调度消息通常同时携带 NoteOn 和 NoteOff，
    // SimpleSynth 会据此预先安排同一个 voice 的起止时间。
    NoteOn = null;
    NoteOff = null;
    /**
     * @param {Partial<MidiMessage>} [init] 初始字段。
     */
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
/**
 * 乐器名称、显示标签和预设组成的描述信息。
 */
export class InstrumentInfo {
    Name = "";
    Label = "";
    Preset = "";
    /**
     * @param {Partial<InstrumentInfo>} [init] 初始字段。
     */
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
/**
 * 乐器描述及其创建工厂的注册记录。
 */
export class InstrumentRegistration extends InstrumentInfo {
    Factory;
    /**
     * @param {Partial<InstrumentRegistration>} [init] 初始字段。
     */
    constructor(init = {}) {
        super(init);
        this.Factory = init.Factory ?? null;
    }
}
/**
 * 按乐器名称选择工厂，并支持一个默认工厂兜底。
 */
export class InstrumentFactoryRegistry {
    // 按乐器 Name 选择工厂；未注册名称回退到 SetDefault 设置的工厂。
    table = new Map();
    factoryDefault = null;
    /** @returns {PartInfo} 注册表自身的 Part 信息。 */
    GetInfo() {
        return new PartInfo({
            Name: "InstrumentFactoryRegistry",
            Classes: ["InstrumentFactoryRegistry"],
            Part: this,
        });
    }
    /** @returns {LifeCycle} 空生命周期，使注册表可作为 Part 注册。 */
    GetLifeCycle() {
        return new LifeCycle();
    }
    /**
     * 根据乐器名称创建实例。
     * @param {EngineContext} context 引擎上下文。
     * @param {InstrumentInfo} info 乐器描述。
     * @returns {object} 乐器实例。
     */
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
    /**
     * 设置未注册名称使用的默认工厂。
     * @param {object|null} factory 默认乐器工厂。
     * @returns {void}
     */
    SetDefault(factory) {
        if (factory) {
            this.factoryDefault = factory;
        }
    }
    /** @returns {InstrumentInfo[]} 排序后的已注册乐器描述。 */
    ListInfo() {
        return [...this.table.values()]
            .map((item) => new InstrumentInfo(item))
            .sort(compareInstrumentInfo);
    }
    /**
     * 注册工厂声明提供的所有乐器。
     * @param {{ListInfo:Function}} factory 乐器工厂。
     * @returns {void}
     */
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
    /**
     * 添加一条注册记录。
     * @param {InstrumentRegistration} registration 乐器注册记录。
     * @returns {void}
     */
    Add(registration) {
        this.add(registration);
    }
    /**
     * 内部登记入口；不完整的记录会被忽略。
     * @param {InstrumentRegistration} registration 乐器注册记录。
     * @returns {void}
     */
    add(registration) {
        if (!registration ||
            !registration.Factory ||
            !registration.Name) {
            return;
        }
        this.table.set(registration.Name, registration);
    }
}
/**
 * 按名称、标签、预设稳定排序乐器描述。
 * @param {InstrumentInfo} a 左值。
 * @param {InstrumentInfo} b 右值。
 * @returns {number} 比较结果。
 */
function compareInstrumentInfo(a, b) {
    return (a.Name.localeCompare(b.Name) ||
        a.Label.localeCompare(b.Label) ||
        a.Preset.localeCompare(b.Preset));
}
/**
 * 获取或延迟创建引擎当前主乐器。
 */
export class DefaultInstrumentHolder {
    /**
     * 返回当前主乐器；不存在时使用默认三角波预设创建并缓存。
     * @param {EngineContext} context 引擎上下文。
     * @returns {object} 当前乐器。
     */
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
/**
 * 汇总所有 InstrumentProvider Part 提供的乐器工厂。
 */
export class InstrumentRegistry {
    // 收集实现了 InstrumentProvider 的 Part，并建立 Name -> Factory 索引。
    // 该注册表在 InstrumentManager 的 OnInit 阶段填充。
    table = new Map();
    /**
     * 扫描所有 Part 并注册其中声明的乐器。
     * @param {EngineContext} context 引擎上下文。
     * @returns {void}
     */
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
    /**
     * 注册单个 Part 提供的乐器。
     * @param {object} part 可能实现 `ListRegistrations` 的 Part。
     * @returns {void}
     */
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
/**
 * 乐器管理器 Part，负责查询工厂和描述信息。
 */
export class InstrumentManagerImpl {
    Registry = new InstrumentRegistry();
    /**
     * 按乐器名称查找工厂。
     * @param {InstrumentInfo} info 乐器描述。
     * @returns {object} 对应的乐器工厂。
     */
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
    /** @returns {InstrumentInfo[]} 排序后的可用乐器列表。 */
    ListInfo() {
        return [...this.Registry.table.values()]
            .map((item) => new InstrumentInfo(item))
            .sort(compareInstrumentInfo);
    }
    /** @returns {PartInfo} 管理器自身的 Part 信息。 */
    GetInfo() {
        return new PartInfoBuilder().SetName("InstrumentManager").Build();
    }
    /** @returns {LifeCycle} 初始化时刷新乐器注册表的生命周期。 */
    GetLifeCycle() {
        const cycle = new LifeCycle();
        cycle.OnInit = (context) => this.Registry.RegisterAllInstruments(context);
        return cycle;
    }
}
/**
 * 创建默认乐器管理器。
 * @returns {InstrumentManagerImpl} 新管理器。
 */
export function NewInstrumentManager() {
    return new InstrumentManagerImpl();
}
/**
 * MIDI 输入转发器，将消息发送给当前主乐器。
 */
export class InnerInputInstrument {
    context;
    // 输入乐器是一个间接层：外部 MIDI 总是转发到当前主乐器，
    // 因此切换主乐器时不需要重新连接输入设备。
    /**
     * @param {EngineContext} context 引擎上下文。
     */
    constructor(context) {
        this.context = context;
    }
    /**
     * 将 MIDI 消息转发给当前主乐器。
     * @param {MidiMessage} message MIDI 调度消息。
     * @returns {void}
     */
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
    /** @returns {InstrumentInfo} 输入转发器的乐器描述。 */
    GetInfo() {
        return new InstrumentInfo({
            Name: "InputInstrument",
            Label: "Input-Instrument",
            Preset: "default",
        });
    }
}
/**
 * 按需创建并缓存 MIDI 输入乐器。
 */
export class AutoInputInstrumentHolder {
    /**
     * 获取或创建输入乐器。
     * @param {EngineContext} context 引擎上下文。
     * @returns {InnerInputInstrument} 输入乐器。
     */
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
/**
 * 音序格中保存的音符数据。
 */
export class Note {
    Key = 0;
    Velocity = 0;
    /**
     * @param {Partial<Note>} [init] 初始字段。
     */
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
/**
 * SimpleSynth 分配给单个音符的源节点和增益节点。
 */
export class SimpleSynthVoice {
    Source;
    Gain;
    /**
     * @param {object} Source 可调度音频源。
     * @param {object} Gain 增益节点。
     */
    constructor(Source, Gain) {
        this.Source = Source;
        this.Gain = Gain;
    }
}
/**
 * 创建 SimpleSynth 的默认乐器工厂。
 */
export class SimpleSynthFactory {
    /** @returns {InstrumentInfo[]} 工厂提供的默认乐器描述。 */
    ListInfo() {
        return [
            new InstrumentInfo({
                Name: "default",
                Label: "Simple Synth",
                Preset: "default",
            }),
        ];
    }
    /**
     * 创建连接到音频输出终点的 SimpleSynth。
     * @param {EngineContext} context 引擎上下文。
     * @param {InstrumentInfo} [info] 乐器描述。
     * @returns {SimpleSynth} 新乐器实例。
     */
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
/**
 * 基于 Web Audio 抽象的轻量多音混合成器。
 *
 * 每个 MIDI key 同时只保留一个 voice，不同 key 可以并行发声；可选用
 * 振荡器或预生成噪声缓冲区作为声源。
 */
export class SimpleSynth {
    audio;
    destination;
    info;
    // 每个 MIDI key 同时只保留一个 voice。重复 NoteOn 会先停止旧 voice，
    // 避免同一个音符叠加造成非预期增益。
    waveform;
    voices = new Map();
    /**
     * @param {object} audio 音频上下文抽象。
     * @param {object} destination 音频输出节点。
     * @param {InstrumentInfo} info 乐器描述，其 `Preset` 决定波形。
     */
    constructor(audio, destination, info) {
        this.audio = audio;
        this.destination = destination;
        this.info = info;
        this.waveform = simpleSynthWaveform(info.Preset);
    }
    /**
     * 执行 MIDI 消息中的 NoteOn/NoteOff。
     * @param {MidiMessage|null} message MIDI 调度消息。
     * @returns {void}
     */
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
    /** @returns {InstrumentInfo} 当前乐器描述副本。 */
    GetInfo() {
        return new InstrumentInfo(this.info);
    }
    /**
     * 启动音符；同 key 的旧 voice 会先被停止。
     * @param {MidiNote} note NoteOn 事件。
     * @returns {void}
     */
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
    /**
     * 为音符创建源节点、增益节点及连接。
     * @param {MidiNote} note 音符事件。
     * @param {number} now 当前音频时间。
     * @returns {SimpleSynthVoice} 新建 voice。
     */
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
    /**
     * 停止并清理指定 key 的 voice。
     * @param {MidiNote} note NoteOff 事件。
     * @returns {void}
     */
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
    /**
     * 处理音频源结束事件并释放节点。
     * @param {number} key MIDI 音高。
     * @param {SimpleSynthVoice} voice 已结束的 voice。
     * @returns {void}
     */
    voiceEnded(key, voice) {
        if (this.voices.get(key) === voice) {
            this.voices.delete(key);
        }
        voice.Source.Disconnect();
        voice.Gain.Disconnect();
    }
    /**
     * 将接近 0 的音符位置解释为立即播放，否则使用绝对调度时间。
     * @param {MidiNote} note 音符事件。
     * @returns {number} 音频时间（秒）。
     */
    noteTime(note) {
        // 绝对位置接近 0 时按即时消息处理；否则直接作为 Web Audio 调度时间。
        return note.Position < 0.001 ? this.audio.GetCurrentTime() : note.Position;
    }
    /**
     * 生成一秒确定性白噪声并写入缓冲区源。
     * @param {object} source 缓冲区源节点。
     * @returns {void}
     */
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
/**
 * 将预设名称归一化为合成器内部波形名。
 * @param {string} preset 预设名称。
 * @returns {"triangle"|"sawtooth"|"square"|"noise"|"sine"} 波形名。
 */
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
/**
 * 将合成器波形名映射为 Web Audio 振荡器类型。
 * @param {string} waveform 合成器波形名。
 * @returns {string} 振荡器类型。
 */
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
/**
 * 立方体单元格坐标 `{bank, step, track}`。
 */
export class Cell {
    BankIndex = 0;
    StepIndex = 0;
    TrackIndex = 0;
    /**
     * @param {Partial<Cell>} [init] 初始坐标。
     */
    constructor(init = {}) {
        Object.assign(this, init);
    }
}
/**
 * 简单的可排序单元格列表。
 */
export class CellCache {
    List = [];
    /**
     * 追加非空单元格。
     * @param {SimpleCell|null} cell 单元格。
     * @returns {void}
     */
    Add(cell) {
        if (cell) {
            this.List.push(cell);
        }
    }
    /** 清空单元格列表。 */
    Reset() {
        this.List.length = 0;
    }
    /**
     * 按调用方提供的比较函数排序。
     * @param {(a:SimpleCell, b:SimpleCell) => number} compare 比较函数。
     * @returns {void}
     */
    Sort(compare) {
        this.List.sort(compare);
    }
}
/**
 * 普通音符格，也可作为多格 `ComplexCell` 的首发格。
 */
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
    /**
     * 在当前 step 触发音符；复杂音符仅由首发格负责派发。
     * @param {StepTickContext} context step 调度上下文。
     * @returns {void}
     */
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
    /**
     * 为当前 step 构造并派发配对的一步长 NoteOn/NoteOff。
     * @param {StepTickContext} context step 调度上下文。
     * @param {Note} note 音符数据。
     * @returns {void}
     */
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
/**
 * 跨越多个 SimpleCell 槽位的复杂音符格。
 */
export class ComplexCell extends Cell {
    // ComplexCell 占用多个 SimpleCell 槽位。这里把 Notes 同时交给首发格，
    // 保持“多音”语义；更复杂时值拆分可由上层 Note 数据表达。
    First = null;
    SimpleCells = [];
    Notes = [];
    /**
     * 在首发格派发该复杂音符包含的所有音符。
     * @param {StepTickContext} context step 调度上下文。
     * @returns {void}
     */
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
/**
 * 按 `{bank, step, track}` 坐标管理共享的 SimpleCell 实例。
 */
export class CellManager {
    // 立方体中的每个 {bank, step, track} 组合只创建一个 SimpleCell；
    // 不同 Bank/Step/Track 通过 owner 引用共享这个对象。
    table = new Map();
    /**
     * 查找单元格，必要时创建它。
     * @param {Cell} want 目标坐标。
     * @param {boolean} [autoMake] 不存在时是否自动创建。
     * @returns {SimpleCell} 对应单元格。
     */
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
    /**
     * 获取单元格的内部 Map 键。
     * @param {Cell} cell 单元格坐标。
     * @returns {string} 坐标键。
     */
    KeyForCell(cell) {
        return this.keyForCell(cell);
    }
    /**
     * 生成单元格坐标键。
     * @param {Cell|null} cell 单元格坐标。
     * @returns {string} 坐标键；空坐标返回空字符串。
     */
    keyForCell(cell) {
        return cell
            ? `b${cell.BankIndex}_s${cell.StepIndex}_t${cell.TrackIndex}`
            : "";
    }
}
/**
 * Bank 对象池，保留容量与当前数据对象。
 */
export class BankArray extends Array {
    /** @returns {number} 对象池容量。 */
    Capacity() {
        return this.length;
    }
    /**
     * 按容量补齐 Bank，并同步每个 Bank 的索引。
     * @param {number} capacity 至少包含的 Bank 数。
     * @returns {BankArray} 初始化后的新数组。
     */
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
    /**
     * 获取前 `length` 个 Bank 的切片。
     * @param {number} length 有效长度。
     * @returns {BankSlice} Bank 切片。
     */
    GetSlice(length) {
        return new BankSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
    }
    /**
     * 安全获取 Bank。
     * @param {number} index Bank 索引。
     * @returns {Bank|null} Bank 实例。
     */
    GetItem(index) {
        return this[index] ?? null;
    }
}
/**
 * 对当前有效 Bank 集合的类型化切片。
 */
export class BankSlice extends Array {
    /** @returns {number} 当前有效 Bank 数。 */
    Length() {
        return this.length;
    }
    /**
     * 从数组创建指定长度的切片。
     * @param {Bank[]} array 源 Bank 数组。
     * @param {number} length 切片长度。
     * @returns {BankSlice} 新切片。
     */
    Init(array, length) {
        return new BankSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
    }
    /**
     * 安全获取 Bank。
     * @param {number} index Bank 索引。
     * @returns {Bank|null} Bank 实例。
     */
    Get(index) {
        return this[index] ?? null;
    }
}
/**
 * Step 对象池。
 */
export class StepArray extends Array {
    /** @returns {number} 对象池容量。 */
    Capacity() {
        return this.length;
    }
    /**
     * 按容量补齐 Step，并同步索引和默认权重。
     * @param {number} capacity 至少包含的 Step 数。
     * @returns {StepArray} 初始化后的新数组。
     */
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
    /**
     * 获取前 `length` 个 Step 的切片。
     * @param {number} length 有效长度。
     * @returns {StepSlice} Step 切片。
     */
    GetSlice(length) {
        return new StepSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
    }
    /**
     * 安全获取 Step。
     * @param {number} index Step 索引。
     * @returns {Step|null} Step 实例。
     */
    GetItem(index) {
        return this[index] ?? null;
    }
}
/**
 * 对当前有效 Step 集合的类型化切片。
 */
export class StepSlice extends Array {
    /** @returns {number} 当前有效 Step 数。 */
    Length() {
        return this.length;
    }
    /**
     * 从数组创建指定长度的切片。
     * @param {Step[]} array 源 Step 数组。
     * @param {number} length 切片长度。
     * @returns {StepSlice} 新切片。
     */
    Init(array, length) {
        return new StepSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
    }
    /**
     * 安全获取 Step。
     * @param {number} index Step 索引。
     * @returns {Step|null} Step 实例。
     */
    Get(index) {
        return this[index] ?? null;
    }
}
/**
 * Track 对象池。
 */
export class TrackArray extends Array {
    /** @returns {number} 对象池容量。 */
    Capacity() {
        return this.length;
    }
    /**
     * 按容量补齐 Track，并同步每个 Track 的索引。
     * @param {number} capacity 至少包含的 Track 数。
     * @returns {TrackArray} 初始化后的新数组。
     */
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
    /**
     * 获取前 `length` 个 Track 的切片。
     * @param {number} length 有效长度。
     * @returns {TrackSlice} Track 切片。
     */
    GetSlice(length) {
        return new TrackSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
    }
    /**
     * 安全获取 Track。
     * @param {number} index Track 索引。
     * @returns {Track|null} Track 实例。
     */
    GetItem(index) {
        return this[index] ?? null;
    }
}
/**
 * 对当前有效 Track 集合的类型化切片。
 */
export class TrackSlice extends Array {
    /** @returns {number} 当前有效 Track 数。 */
    Length() {
        return this.length;
    }
    /**
     * 从数组创建指定长度的切片。
     * @param {Track[]} array 源 Track 数组。
     * @param {number} length 切片长度。
     * @returns {TrackSlice} 新切片。
     */
    Init(array, length) {
        return new TrackSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
    }
    /**
     * 安全获取 Track。
     * @param {number} index Track 索引。
     * @returns {Track|null} Track 实例。
     */
    Get(index) {
        return this[index] ?? null;
    }
    /**
     * 过滤 Track，并保持返回值为 `TrackSlice`。
     * @param {(track:Track, index:number, array:Track[]) => boolean} predicate 过滤函数。
     * @returns {TrackSlice} 过滤后的切片。
     */
    Filter(predicate) {
        return new TrackSlice(...Array.prototype.filter.call(this, predicate));
    }
}
/**
 * step 在时间线上的位置及循环内索引。
 */
export class StepPosition extends RelativePosition {
    Index = 0;
    Location = 0;
}
/**
 * 音序中的一个 step，保存该 step 上的所有单元格。
 */
export class Step {
    Context = null;
    Index = 0;
    Weight = 60;
    Cells = new CellCache();
    /**
     * 依次触发此 step 上的所有单元格。
     * @param {StepTickContext} context step 调度上下文。
     * @returns {void}
     */
    OnStep(context) {
        if (!context) {
            throw new Error("StepTickContext is nil");
        }
        for (const cell of this.Cells.List) {
            cell.OnStep(context);
        }
    }
}
/**
 * 所有 bank 共享的轨道元数据。
 */
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
/**
 * 音序轨道，保存元数据和轨道维度的单元格缓存。
 */
export class Track {
    Context = null;
    Meta = null;
    Index = 0;
    Cells = new CellCache();
}
/**
 * Cube 使用的 Bank、Step、Track 最大容量对象池。
 */
export class CubeBuffers {
    BankBuffer = new BankArray();
    StepBuffer = new StepArray();
    TrackBuffer = new TrackArray();
}
/**
 * `{bank, step, track}` 三维音序数据容器。
 */
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
    /**
     * 根据配置初始化对象池、有效切片和所有单元格关联。
     * @param {SequencerContext|null} context 音序器上下文。
     * @param {Configuration|null} [configuration] 引擎配置。
     * @returns {void}
     */
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
    /**
     * 根据最新有效长度刷新一个或所有 Bank 的切片。
     * @param {number} bankIndex 要刷新的 Bank 索引。
     * @param {boolean} all 是否刷新全部 Bank。
     * @returns {void}
     */
    Reload(bankIndex, all) {
        if (all) {
            for (const bank of this.Buffers.BankBuffer) {
                bank?.Reload();
            }
            return;
        }
        this.Buffers.BankBuffer.GetItem(bankIndex)?.Reload();
    }
    /**
     * 为 Bank 建立 Step/Track 视图，并绑定每一格的共享单元格。
     * @param {Bank} bank 要初始化的 Bank。
     * @returns {void}
     */
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
/**
 * 一组独立的 Step/Track 视图；轨道元数据与 Cube 共享。
 */
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
    /**
     * 按当前循环索引触发 Bank 中对应的 Step。
     * @param {StepTickContext} context step 调度上下文。
     * @returns {void}
     */
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
    /**
     * 根据 Cube 的对象池和有效长度初始化当前 Bank 的视图。
     * @param {Cube} cube 所属 Cube。
     * @returns {void}
     */
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
    /** 根据 Cube 当前有效长度刷新 Step/Track 切片。 */
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
/**
 * 音序器速度、拍号和预加载配置。
 */
export class SequencerInfo {
    // BPM 和 Tempo 共同决定 SecondsPerBeat/SecondsPerStep；
    // PreloadSeconds 表示调度器希望提前触发多少秒的音符。
    Profile = null;
    SecondsPerStep = 0;
    SecondsPerBeat = 0;
    StepsPerBeat = 1;
    StepsPerCycle = 16;
    PreloadSeconds = 1;
    /**
     * 修正非法时间参数并返回规范化配置。
     * @param {boolean} [rewrite] 是否就地修改当前对象。
     * @returns {SequencerInfo} 规范化后的配置。
     */
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
/**
 * 音序器在某个时间点的位置、全局 step 序号和循环内索引。
 */
export class StepTickPos {
    Position = new RelativePosition();
    Index = 0;
    Location = 0;
    /**
     * 计算循环内 step 索引。
     * @param {SequencerInfo} info 音序器配置。
     * @param {boolean} [rewrite] 是否写回 `Index`。
     * @returns {number} 循环内索引。
     */
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
/**
 * 单次 step 调度使用的上下文快照。
 */
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
/**
 * 音序器的运行时状态和 Cube 数据容器。
 */
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
    /**
     * 重新计算 step 位置；可选地将位置对齐到 step 起点。
     * @param {StepTickPos} position 要规范化的位置。
     * @param {boolean} align 是否对齐到工作时间线。
     * @returns {void}
     */
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
    /**
     * 将任意时间线位置换算为工作时间线上的全局 step 序号。
     * @param {RelativePosition} position 输入位置。
     * @returns {number} 非负 step 序号。
     */
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
    /**
     * 计算位置对应的循环内 step 索引。
     * @param {RelativePosition} position 输入位置。
     * @returns {number} 循环内索引。
     */
    ComputeStepIndex(position) {
        const location = this.ComputeStepLocation(position);
        const stepsPerCycle = this.Info.StepsPerCycle;
        return stepsPerCycle < 1 ? 0 : location % stepsPerCycle;
    }
    /**
     * 创建并初始化一个 Bank。
     * @param {number} index Bank 索引。
     * @returns {Bank} 新 Bank。
     */
    NewBank(index) {
        const bank = new Bank();
        bank.Init(this.Cube);
        bank.Index = index;
        return bank;
    }
    /**
     * 创建并绑定一个 Step。
     * @param {number} index Step 索引。
     * @returns {Step} 新 Step。
     */
    NewStep(index) {
        const step = new Step();
        step.Index = index;
        step.Context = this.Cube;
        return step;
    }
    /**
     * 创建并绑定一个 Track。
     * @param {number} index Track 索引。
     * @returns {Track} 新 Track。
     */
    NewTrack(index) {
        const track = new Track();
        track.Index = index;
        track.Context = this.Cube;
        return track;
    }
}
////////////////////////////////////////////////////////////////////////////////
// Clock and real-time positions
/**
 * 一次时钟 tick 携带的时间、时钟和引擎上下文。
 */
export class ClockEvent {
    Clock = null;
    Time = 0;
    Context = null;
}
/**
 * 引擎共享的实时位置快照。
 *
 * `Precision` 为连续工作时间，`Normal` 对齐到 step 起点，`Raw` 为根时间坐标。
 */
export class RTC {
    Precision = new StepTickPos();
    Normal = new StepTickPos();
    Raw = new StepTickPos();
}
/**
 * 读取音频时间、刷新 RTC 并同步通知所有监听器。
 */
export class SimpleClock {
    // Clock 只负责读取 AudioContext 时间、刷新 RTC，然后通知监听器。
    // 音序器和监控组件因此看到同一个“当前时间”快照。
    listeners = [];
    /**
     * 添加去重后的时钟监听器。
     * @param {{OnClock:Function}} listener 监听器。
     * @returns {void}
     */
    AddListener(listener) {
        if (listener && !this.listeners.includes(listener)) {
            this.listeners.push(listener);
        }
    }
    /**
     * 移除时钟监听器。
     * @param {{OnClock:Function}} listener 监听器。
     * @returns {void}
     */
    RemoveListener(listener) {
        if (!listener) {
            return;
        }
        this.listeners = this.listeners.filter((item) => item !== listener);
    }
    /**
     * 刷新 RTC 并调用所有监听器；多个监听器错误会合并抛出。
     * @param {ClockEvent} event 时钟事件。
     * @returns {void}
     */
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
    /**
     * 根据音频上下文当前时间刷新 Root、Work 和 step 位置。
     * @param {ClockEvent} event 时钟事件。
     * @returns {void}
     */
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
/**
 * 按需选择并缓存音序器当前 Bank。
 */
class CurrentBankHolder {
    context;
    /**
     * @param {SequencerContext} context 音序器上下文。
     */
    constructor(context) {
        this.context = context;
    }
    /**
     * 返回当前 Bank；未选择时回退到索引 0。
     * @returns {Bank} 当前 Bank。
     */
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
/**
 * 在配置变化后重算速度和当前 Bank。
 */
class SequencerUpdater {
    context;
    // profile、速度或 bank 配置变化后，由 Updating 标记触发一次重算。
    /**
     * @param {SequencerContext} context 音序器上下文。
     */
    constructor(context) {
        this.context = context;
    }
    /** 执行一次速度与 Bank 配置刷新。 */
    Update() {
        this.context.Updating = false;
        Log().Info("innerSequencer.Update()");
        this.computeSpeed();
        this.loadCurrentBank();
    }
    /** 根据 BPM 和每拍 step 数计算 step 时长。 */
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
    /** 校验当前 Bank 索引并加载对应 Bank。 */
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
/**
 * 音序器内部实现，向引擎暴露 Cube 数据和 step 调度入口。
 */
export class InnerSequencer {
    context;
    /**
     * @param {SequencerContext} context 音序器上下文。
     */
    constructor(context) {
        this.context = context;
    }
    /** @returns {Bank[]} 所有有效 Bank。 */
    GetBanks() {
        return [...this.context.Cube.Banks];
    }
    /** @returns {SequencerContext} 音序器运行时上下文。 */
    GetContext() {
        return this.context;
    }
    /** @returns {Step[]} 所有有效 Step。 */
    GetSteps() {
        return [...this.context.Cube.Steps];
    }
    /** @returns {Track[]} 所有有效 Track。 */
    GetTracks() {
        return [...this.context.Cube.Tracks];
    }
    /**
     * 将 step 调度转发给当前 Bank。
     * @param {StepTickContext} context step 调度上下文。
     * @returns {void}
     */
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
    /**
     * 当上下文标记为需要更新时刷新速度和 Bank。
     * @param {SequencerContext} context 音序器上下文。
     * @returns {void}
     */
    Update(context) {
        if (context.Updating) {
            new SequencerUpdater(this.context).Update();
        }
    }
}
/**
 * 时钟监听器，推进音序器已处理位置并预调度未来 step。
 */
export class SequencerTrigger {
    context;
    // 每次时钟 tick 都从已处理的 Back 推进到预加载前沿 Front。
    // 仅当 Transport.Walking() 为 true 时实际派发 step。
    /**
     * @param {SequencerContext} context 音序器上下文。
     */
    constructor(context) {
        this.context = context;
    }
    /**
     * 处理一次时钟 tick；仅走带运行时派发新 step。
     * @param {ClockEvent} _event 时钟事件。
     * @returns {void}
     */
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
    /** @returns {boolean} 走带是否正在推进。 */
    isWalking() {
        return this.context.EngineContext?.TransportContext.Walking() ?? false;
    }
    /** 将当前实时位置同步到走带状态。 */
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
    /**
     * 准备 step 调度上下文。
     * @param {StepTickContext|null} source 可复用的上下文。
     * @returns {StepTickContext} 已绑定音序器上下文的调度上下文。
     */
    prepare(source) {
        const stc = source ?? new StepTickContext();
        stc.Context = this.context;
        stc.WorkTimeline = this.context.EngineContext?.WorkTileline ?? null;
        return stc;
    }
    /**
     * 校验调度上下文和音序器参数，并绑定工作时间线引用。
     * @param {StepTickContext} stc step 调度上下文。
     * @returns {void}
     */
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
    /**
     * 根据预加载秒数计算未来调度前沿。
     * @param {StepTickContext} stc step 调度上下文。
     * @returns {void}
     */
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
    /**
     * 从已处理位置逐个推进到前沿，并调用音序器派发 step。
     * @param {StepTickContext} stc step 调度上下文。
     * @returns {void}
     */
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
/**
 * 从 step 上下文取出音序器。
 * @param {StepTickContext} stc step 调度上下文。
 * @returns {InnerSequencer} 音序器。
 */
function contextSequencer(stc) {
    const sequencer = stc.Context?.Sequencer;
    if (!sequencer) {
        throw new Error("stc.context.Sequencer is nil");
    }
    return sequencer;
}
/**
 * 深复制 step 位置，避免推进时修改共享的 RTC 快照。
 * @param {StepTickPos} source 源位置。
 * @returns {StepTickPos} 独立副本。
 */
function cloneTickPos(source) {
    const destination = new StepTickPos();
    destination.Position = source.Position.Clone();
    destination.Index = source.Index;
    destination.Location = source.Location;
    return destination;
}
/**
 * 装配音序器、时钟触发器和默认 Cube 配置。
 */
export class InnerSequencerFactory {
    /**
     * 创建并初始化音序器及其运行时上下文。
     * @param {EngineContext} engineContext 引擎上下文。
     * @returns {InnerSequencer} 已接入时钟的音序器。
     */
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
/**
 * 监控组件的连接和日志开关配置。
 */
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
    /**
     * @param {Partial<MonitorConfig>} [init] 初始配置。
     */
    constructor(init = {}) {
        Object.assign(this, init);
    }
    /**
     * 从 URL 拆出协议、主机和端口，或补齐默认本地地址。
     * @returns {MonitorConfig} 当前实例，便于链式调用。
     */
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
/**
 * 创建引擎所需的驱动、乐器工厂、时间配置和 Part 列表。
 */
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
/**
 * 创建可直接用于 Mock 驱动测试的默认引擎配置。
 * @returns {Configuration} 默认配置；调用方仍需设置乐器工厂。
 */
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
/**
 * 引擎启动和停止过程的状态标志。
 */
export class EngineState {
    Starting = false;
    Stopping = false;
    Stopped = false;
    Started = false;
    /** @returns {string} 便于日志阅读的状态摘要。 */
    String() {
        return `engine.[state starting:${this.Starting} started:${this.Started} stopping:${this.Stopping} stopped:${this.Stopped}]`;
    }
    /** @returns {string} 与 `String()` 一致的状态摘要。 */
    toString() {
        return this.String();
    }
}
/**
 * 播放、暂停、录制以及当前/目标位置等走带状态。
 */
export class TransportContext {
    // StopAt 表示暂停/停止后的目标位置；Resume/Play 会据此重定位工作时间线。
    paused = false;
    playing = false;
    recording = false;
    Current = new StepTickPos();
    StopAt = null;
    /** @returns {boolean} 是否正在播放且未暂停。 */
    Walking() {
        return this.playing && !this.paused;
    }
    /** @returns {boolean} 是否已暂停。 */
    IsPaused() {
        return this.paused;
    }
    /** @returns {boolean} 是否处于播放模式。 */
    IsPlaying() {
        return this.playing;
    }
    /** @returns {boolean} 是否处于录制模式。 */
    IsRecording() {
        return this.recording;
    }
    /**
     * 设置暂停标志。
     * @param {boolean} value 新状态。
     * @returns {void}
     */
    internalSetPaused(value) {
        this.paused = value;
    }
    /**
     * 设置播放标志。
     * @param {boolean} value 新状态。
     * @returns {void}
     */
    internalSetPlaying(value) {
        this.playing = value;
    }
    /**
     * 设置录制标志。
     * @param {boolean} value 新状态。
     * @returns {void}
     */
    internalSetRecording(value) {
        this.recording = value;
    }
}
/**
 * 引擎各子系统共享的依赖容器和运行快照。
 */
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
    /**
     * 将位置转换到工作时间线；已是 Work 坐标时直接返回。
     * @param {RelativePosition} [source] 输入位置。
     * @returns {RelativePosition} 工作时间线上的位置。
     */
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
    /**
     * 获取停机目标位置，没有目标时返回当前精确位置。
     * @param {RelativePosition} [destination] 可复用的结果对象。
     * @returns {RelativePosition} 当前位置。
     */
    GetCurrentPosition(destination) {
        const dst = destination ?? new RelativePosition();
        const position = this.TransportContext.StopAt ?? this.RTC.Precision.Position;
        dst.Ref = position.Ref;
        dst.Pos = position.Pos;
        return dst;
    }
    /**
     * 调整 Work 时间线偏移，使目标位置保持相同的绝对时间。
     * @param {RelativePosition|null} to 新的工作位置。
     * @returns {void}
     */
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
/**
 * 通过 `GetInfo` 能力判断对象能否作为 Part。
 * @param {unknown} value 待检查对象。
 * @returns {boolean} 是否满足 Part 能力。
 */
function isPart(value) {
    return (typeof value === "object" &&
        value !== null &&
        typeof value.GetInfo === "function");
}
/**
 * 通过工厂方法能力判断对象能否创建 Part。
 * @param {unknown} value 待检查对象。
 * @returns {boolean} 是否满足 PartFactory 能力。
 */
function isPartFactory(value) {
    return (typeof value === "object" &&
        value !== null &&
        typeof value.CreatePart === "function" &&
        typeof value.GetRegistration === "function");
}
/**
 * 通过 `GetLifeCycle` 能力判断对象是否可加入生命周期。
 * @param {unknown} value 待检查对象。
 * @returns {boolean} 是否满足 Life 能力。
 */
function isLife(value) {
    return (typeof value === "object" &&
        value !== null &&
        typeof value.GetLifeCycle === "function");
}
/**
 * 创建并装配完整引擎的默认工厂。
 */
export class InnerEngineFactory {
    /**
     * 根据配置建立音频、生命周期、Part、乐器、时钟和音序器子系统。
     * @param {Configuration} config 引擎配置。
     * @returns {InnerEngine} 已装配的引擎。
     */
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
    /**
     * 注册配置中的 Part 实例、工厂和注册记录。
     * @param {EngineContext} context 引擎上下文。
     * @returns {void}
     */
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
    /**
     * 将具有生命周期能力的 Part 加入管理器。
     * @param {EngineContext} context 引擎上下文。
     * @returns {void}
     */
    initLifeCycles(context) {
        for (const name of context.PartManager.ListNames()) {
            const part = context.PartManager.FindPartByName(name);
            if (isLife(part)) {
                context.LifeManager.Add(part);
            }
        }
    }
    /**
     * 创建默认乐器，并绑定到尚未指定乐器的轨道。
     * @param {EngineContext} context 引擎上下文。
     * @returns {void}
     */
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
/**
 * 创建默认引擎工厂。
 * @returns {InnerEngineFactory} 新工厂实例。
 */
export function DefaultEngineFactory() {
    return new InnerEngineFactory();
}
/**
 * 引擎控制对象，负责创建、启动、等待和销毁运行时资源。
 */
export class InnerEngine {
    // Join 不阻塞 JS 线程，而是返回 Promise，由生命周期运行器等待。
    Context;
    /** @returns {EngineContext} 引擎共享上下文。 */
    GetContext() {
        return this.Context;
    }
    /**
     * 等待引擎停止，不阻塞 JavaScript 线程。
     * @returns {Promise<void>} 引擎进入停止过程后完成。
     */
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
    /** 配置主定时器及其时钟回调。 */
    Create() {
        // 主 Timer 约每 10ms 产生一次时钟事件，实际精度取决于宿主调度器。
        const context = this.Context;
        const timer = context.Timer ?? new Timer();
        context.Timer = timer;
        timer.Delay = 0.01;
        timer.Interval = 0.01;
        timer.Handler = context.TimerHandler;
    }
    /** 释放主定时器并标记引擎已停止。 */
    Destroy() {
        this.Context.Timer = new Timer();
        this.Context.State.Stopped = true;
    }
    /** 恢复音频、启动主定时器，并按配置启动监控。 */
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
    /** 停止主定时器并进入引擎停止过程。 */
    Stop() {
        this.Context.State.Stopping = true;
        this.Context.Timer.Stop();
    }
}
/**
 * 走带控制器，维护播放状态、暂停位置及恢复时的重定位。
 */
export class InnerTransport {
    // Transport 只维护播放状态和目标位置，实际时间推进由 Clock 完成。
    Context;
    TransportContext;
    /** @returns {boolean} 走带是否正在推进。 */
    Walking() {
        return this.TransportContext.Walking();
    }
    /** 记录当前位置并暂停走带。 */
    Pause() {
        // 仅在真正 Walking 时记录位置；重复 Pause 不会覆盖 StopAt。
        if (this.TransportContext.Walking()) {
            this.TransportContext.StopAt =
                this.Context.RTC.Precision.Position.Clone();
            this.TransportContext.internalSetPaused(true);
        }
    }
    /** @returns {boolean} 是否已暂停。 */
    Paused() {
        return this.TransportContext.IsPaused();
    }
    /** 进入播放模式并从停靠位置恢复。 */
    Play() {
        this.TransportContext.internalSetPlaying(true);
        this.tryRebaseToStopAt();
    }
    /** @returns {boolean} 是否处于播放模式。 */
    Playing() {
        return this.TransportContext.IsPlaying();
    }
    /** 进入录制和播放模式，并从停靠位置恢复。 */
    Record() {
        this.TransportContext.internalSetRecording(true);
        this.TransportContext.internalSetPlaying(true);
        this.tryRebaseToStopAt();
    }
    /** @returns {boolean} 是否处于录制模式。 */
    Recording() {
        return this.TransportContext.IsRecording();
    }
    /** 解除暂停并从停靠位置继续。 */
    Resume() {
        this.TransportContext.internalSetPaused(false);
        this.tryRebaseToStopAt();
    }
    /** 清除播放/录制状态，并将目标位置重置到工作起点。 */
    Stop() {
        const context = this.Context;
        this.TransportContext.internalSetRecording(false);
        this.TransportContext.internalSetPlaying(false);
        this.TransportContext.internalSetPaused(false);
        this.TransportContext.StopAt = new RelativePosition(context.WorkTileline, 0);
    }
    /**
     * 主定时器回调：刷新音序器配置并分发一次时钟事件。
     * @param {Timer} _timer 触发该回调的定时器。
     * @returns {void}
     */
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
    /** 根据停靠位置重定位工作时间线和音序器已处理位置。 */
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
/**
 * 将引擎的创建、启动、等待、停止和销毁接入通用生命周期。
 */
export class AutoEngineController {
    // 把 Engine 的 Create/Start/Join/Stop/Destroy 接入通用生命周期。
    // OnRun 等待 Join 完成，确保后续停止阶段不会提前执行。
    context = null;
    /** @returns {PartInfo} 控制器注册信息。 */
    GetInfo() {
        return new PartInfo({ Name: "AutoEngineController", Part: this });
    }
    /** @returns {LifeCycle} 映射到引擎控制方法的生命周期。 */
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
/**
 * 引擎启动入口，负责创建引擎并运行主生命周期。
 */
export class Boot {
    // 创建引擎并运行主生命周期。若配置包含 AutoEngineController，
    // Run 返回的 Promise 会保持到引擎被停止。
    /**
     * 使用指定配置创建并运行引擎。
     * @param {Configuration|null} [config] 引擎配置；缺省时使用默认 Mock 配置。
     * @returns {Promise<void>} 主生命周期结束时完成。
     */
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
/**
 * 创建并运行引擎的便捷函数。
 * @param {Configuration|null} [config] 引擎配置。
 * @returns {Promise<void>} 主生命周期结束时完成。
 */
export function Run(config = null) {
    return new Boot().Run(config);
}
/**
 * 将控制动作映射到走带操作或输入乐器 MIDI 操作。
 */
export class InnerActionController {
    context;
    // 将控制动作映射为 Transport 或 MIDI 操作。Pause/Resume 共用一个
    // 切换逻辑：Walking 时暂停，否则恢复。
    handlers = new Map();
    /**
     * 建立动作处理表。
     * @param {EngineContext} context 引擎上下文。
     */
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
    /**
     * 控制器由调用方驱动，因此该方法当前不执行初始化。
     * @returns {void}
     */
    Start() {
        // The action controller is driven by its caller.
    }
    /**
     * 执行控制事件对应的处理函数。
     * @param {{Action:number, NoteNumber?:number, NoteVelocity?:number}} event 控制事件。
     * @returns {void}
     */
    Execute(event) {
        const handler = this.handlers.get(event.Action);
        if (!handler) {
            throw new Error(`no handler with action [${event.Action}]`);
        }
        handler(this.context, event);
    }
}
/**
 * 创建动作控制器。
 * @param {EngineContext} context 引擎上下文。
 * @returns {InnerActionController} 新控制器。
 */
export function NewActionController(context) {
    return new InnerActionController(context);
}
/**
 * 周期性输出引擎、走带、音序器和输入状态的日志监控器。
 */
export class Monitor {
    // 低频状态监控器，只负责日志输出，不参与音频调度。
    context = null;
    timer = null;
    InputKeyCount = 0;
    InputKeyCode = 0;
    /**
     * 绑定要监控的引擎上下文。
     * @param {EngineContext} context 引擎上下文。
     * @returns {void}
     */
    Init(context) {
        this.context = context;
    }
    /** 创建并启动低频状态日志定时器。 */
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
    /** 停止并释放监控定时器。 */
    Stop() {
        this.timer?.Stop();
        this.timer = null;
    }
    /** 输出一次当前引擎状态快照，并在引擎停止时自行注销。 */
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
