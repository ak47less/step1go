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

////////////////////////////////////////////////////////////////////////////////
// Common scalar and music types

// Go 中的命名标量在 TypeScript 中统一映射为 number。保留别名后，
// API 签名仍能表达单位，同时允许直接传入普通数值。
export type Seconds = number;
export type BPM = number;
export type Frequency = number;
export type ChordMode = string;
export type ScaleMode = string;
export type ToneName = string;
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
} as const;
export type MidiNoteNumber = number;
export type MidiNoteVelocity = number;
export type SampleCount = number;
export type SampleRate = number;
export type SampleValue = number;
export type AbsolutePosition = Seconds;

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

export interface Tempo {
  /** 每小节拍数，例如 4/4 拍时为 4。 */
  Upper: number;
  /** 以何种音符为一拍，例如四分音符为 4。 */
  Lower: number;
}

/** MIDI 音高的解析结果；69 对应 A4/440Hz。 */
export interface PitchInfo {
  Number: MidiNoteNumber;
  Name: string;
  Group: number;
  Tone: number;
  Frequency: Frequency;
}

export type MidiKeyInfo = PitchInfo;

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

export function GetPitchInfo(
  number: MidiNoteNumber,
  destination?: PitchInfo,
): PitchInfo {
  // 与原实现一致：输入先夹到合法 MIDI 范围，再按十二平均律计算频率。
  const dst = destination ?? ({} as PitchInfo);
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

export function NewTempo(upper: number, lower: number): Tempo {
  return { Upper: upper, Lower: lower };
}

export interface MusicProfile {
  // 音序器每轮更新时从 profile 读取速度、拍号和调式。
  GetScale(): ScaleMode;
  GetSpeed(): BPM;
  GetTempo(): Tempo;
  GetTone(): ToneName;
}

export class SimpleProfile implements MusicProfile {
  private scale: ScaleMode = ScaleModeMajor;
  private bpm: BPM = 100;
  private tempo: Tempo = NewTempo(4, 4);
  private tone: ToneName = ToneName.C;

  GetScale(): ScaleMode {
    return this.scale;
  }

  GetSpeed(): BPM {
    return this.bpm;
  }

  GetTempo(): Tempo {
    return this.tempo;
  }

  GetTone(): ToneName {
    return this.tone;
  }

  SetScale(scale: ScaleMode): this {
    this.scale = scale;
    return this;
  }

  SetSpeed(bpm: BPM): this {
    this.bpm = bpm;
    return this;
  }

  SetTempo(tempo: Tempo): this {
    this.tempo = tempo;
    return this;
  }

  SetTone(tone: ToneName): this {
    this.tone = tone;
    return this;
  }

  Init(): this {
    this.bpm = 100;
    this.scale = ScaleModeMajor;
    this.tone = ToneName.C;
    this.tempo = NewTempo(4, 4);
    return this;
  }

  InitWith(base?: MusicProfile | null): this {
    // 传入 base 时复制其配置；否则恢复默认的 100 BPM、C 大调、4/4 拍。
    if (base) {
      this.bpm = base.GetSpeed();
      this.scale = base.GetScale();
      this.tone = base.GetTone();
      this.tempo = base.GetTempo();
    } else {
      this.Init();
    }
    return this;
  }
}

export enum Action {
  Min = 0,
  StartEngine,
  StopEngine,
  Play,
  Loop,
  Stop,
  Pause,
  Resume,
  Record,
  NoteOn,
  NoteOff,
  MasterVolumeUp,
  MasterVolumeDown,
  MasterVolumeMute,
  KeyUp,
  KeyDown,
  Max,
}

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

export interface ActionEvent {
  // ActionEvent 同时服务走带控制、电脑键盘和 MIDI 输入；具体使用哪些
  // 字段由 Action 和 Flag* 字段决定。
  Action: Action;
  FlagControl?: boolean;
  FlagNote?: boolean;
  FlagNoteOn?: boolean;
  FlagNoteOff?: boolean;
  FlagCKey?: boolean;
  FlagKeyUp?: boolean;
  FlagKeyDown?: boolean;
  KeyCode?: number;
  NoteNumber?: MidiNoteNumber;
  NoteVelocity?: MidiNoteVelocity;
}

////////////////////////////////////////////////////////////////////////////////
// Logging

export interface Logger {
  // 格式化字符串沿用 Go Logger 的 %s/%d/%v/%f 占位符。
  Info(format: string, ...args: unknown[]): void;
  Error(format: string, ...args: unknown[]): void;
  Trace?(format: string, ...args: unknown[]): void;
  Debug?(format: string, ...args: unknown[]): void;
  Warn?(format: string, ...args: unknown[]): void;
}

function formatLog(format: string, args: unknown[]): string {
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

export class ConsoleLogger implements Logger {
  Info(format: string, ...args: unknown[]): void {
    console.info(formatLog(format, args));
  }

  Error(format: string, ...args: unknown[]): void {
    console.error(formatLog(format, args));
  }

  Trace(format: string, ...args: unknown[]): void {
    console.debug(formatLog(format, args));
  }

  Debug(format: string, ...args: unknown[]): void {
    console.debug(formatLog(format, args));
  }

  Warn(format: string, ...args: unknown[]): void {
    console.warn(formatLog(format, args));
  }
}

const defaultLogger: Logger = new ConsoleLogger();

export function Log(): Logger {
  return defaultLogger;
}

export function handleError(error: unknown): void {
  // Timer、时钟监听器等异步入口不能向调用方抛错，因此统一交给默认 Logger。
  if (error === null || error === undefined) {
    return;
  }
  const message = error instanceof Error ? error.message : String(error);
  defaultLogger.Error("%s", message);
}

////////////////////////////////////////////////////////////////////////////////
// Web Audio compatible interfaces

export interface AudioNode {
  // 这些接口刻意保持 Web Audio 的概念，但不继承 DOM 类型，
  // 因而同一份代码也能在 Node 和测试环境运行。
  Connect(destination: AudioNode): void;
  Disconnect(): void;
}

export interface AudioParam {
  // 参数值可以立即设置，也可以按音频上下文时间轴安排自动化事件。
  GetValue(): number;
  SetValue(value: number): void;
  SetValueAtTime(value: number, startTime: Seconds): void;
  LinearRampToValueAtTime(value: number, endTime: Seconds): void;
  ExponentialRampToValueAtTime(value: number, endTime: Seconds): void;
  SetTargetAtTime(
    target: number,
    startTime: Seconds,
    timeConstant: Seconds,
  ): void;
  CancelScheduledValues(startTime: Seconds): void;
}

export interface AudioScheduledSourceNode extends AudioNode {
  // Start/Stop 的时间均相对于当前 AudioContext 的时间零点。
  Start(when: Seconds): void;
  Stop(when: Seconds): void;
  SetOnEnded(handler: (() => void) | null): void;
}

export enum OscillatorType {
  Sine = "sine",
  Square = "square",
  Sawtooth = "sawtooth",
  Triangle = "triangle",
  Custom = "custom",
}

export const OscillatorTypeSine = OscillatorType.Sine;
export const OscillatorTypeSquare = OscillatorType.Square;
export const OscillatorTypeSawtooth = OscillatorType.Sawtooth;
export const OscillatorTypeTriangle = OscillatorType.Triangle;
export const OscillatorTypeCustom = OscillatorType.Custom;

export interface OscillatorNode extends AudioScheduledSourceNode {
  GetFrequency(): AudioParam;
  GetDetune(): AudioParam;
  GetType(): OscillatorType;
  SetType(nodeType: OscillatorType): void;
}

export interface GainNode extends AudioNode {
  GetGain(): AudioParam;
}

export interface AudioDestinationNode extends AudioNode {
  GetMaxChannelCount(): number;
}

export interface AudioBuffer {
  GetSampleRate(): SampleRate;
  GetLength(): SampleCount;
  GetNumberOfChannels(): number;
  GetChannelData(channel: number): SampleValue[] | null;
}

export interface AudioBufferSourceNode extends AudioScheduledSourceNode {
  GetBuffer(): AudioBuffer | null;
  SetBuffer(buffer: AudioBuffer | null): void;
  GetPlaybackRate(): AudioParam;
  GetDetune(): AudioParam;
}

export enum AudioContextState {
  Suspended = "suspended",
  Running = "running",
  Closed = "closed",
}

export const AudioContextStateSuspended = AudioContextState.Suspended;
export const AudioContextStateRunning = AudioContextState.Running;
export const AudioContextStateClosed = AudioContextState.Closed;

export interface AudioContext {
  // 对 Web Audio AudioContext 的最小抽象。Malgo 等原生后端只需实现此接口。
  Close(): void;
  Resume(): void | Promise<void>;
  Suspend(): void | Promise<void>;
  CreateOscillator(): OscillatorNode;
  CreateGain(): GainNode;
  CreateBuffer(
    numberOfChannels: number,
    length: SampleCount,
    sampleRate: SampleRate,
  ): AudioBuffer;
  CreateBufferSource(): AudioBufferSourceNode;
  GetCurrentTime(): Seconds;
  GetSampleRate(): SampleRate;
  GetDestination(): AudioDestinationNode;
  GetState(): AudioContextState;
}

export interface AudioContextFactory {
  NewAudioContext(): AudioContext;
}

export interface Driver {
  GetAudioContextFactory(): AudioContextFactory;
}

function normalizeAudioTime(value: Seconds): Seconds {
  // Web Audio 不接受负数或 NaN 的调度时间。
  return Number.isFinite(value) && value >= 0 ? value : 0;
}

////////////////////////////////////////////////////////////////////////////////
// AudioParam implementation

enum AudioParamEventKind {
  Set,
  LinearRamp,
  ExponentialRamp,
  Target,
}

interface AudioParamEvent {
  Kind: AudioParamEventKind;
  Value: number;
  Time: Seconds;
  TimeConstant: Seconds;
}

export class DefaultAudioParam implements AudioParam {
  // 事件按时间排序。ramp 事件在到达其时间点前会从当前值插值，
  // SetTargetAtTime 则从起始时刻开始指数逼近目标值。
  private baseValue: number;
  private baseTime: Seconds;
  private events: AudioParamEvent[] = [];

  constructor(
    value: number,
    baseTime: Seconds,
    private readonly now: () => Seconds = () => 0,
  ) {
    this.baseValue = value;
    this.baseTime = baseTime;
  }

  GetValue(): number {
    return this.valueAt(this.now());
  }

  SetValue(value: number): void {
    const now = this.now();
    this.baseValue = value;
    this.baseTime = now;
    this.events = [];
  }

  SetValueAtTime(value: number, startTime: Seconds): void {
    this.addEvent({
      Kind: AudioParamEventKind.Set,
      Value: value,
      Time: normalizeAudioTime(startTime),
      TimeConstant: 0,
    });
  }

  LinearRampToValueAtTime(value: number, endTime: Seconds): void {
    this.addEvent({
      Kind: AudioParamEventKind.LinearRamp,
      Value: value,
      Time: normalizeAudioTime(endTime),
      TimeConstant: 0,
    });
  }

  ExponentialRampToValueAtTime(value: number, endTime: Seconds): void {
    this.addEvent({
      Kind: AudioParamEventKind.ExponentialRamp,
      Value: value,
      Time: normalizeAudioTime(endTime),
      TimeConstant: 0,
    });
  }

  SetTargetAtTime(
    target: number,
    startTime: Seconds,
    timeConstant: Seconds,
  ): void {
    this.addEvent({
      Kind: AudioParamEventKind.Target,
      Value: target,
      Time: normalizeAudioTime(startTime),
      TimeConstant: normalizeAudioTime(timeConstant),
    });
  }

  CancelScheduledValues(startTime: Seconds): void {
    // 先把当前值固化到 now，避免删除事件后参数值瞬间跳回旧基准。
    const time = normalizeAudioTime(startTime);
    const now = this.now();
    this.baseValue = this.valueAt(now);
    this.baseTime = now;
    this.events = this.events.filter(
      (event) => event.Time > now && event.Time < time,
    );
  }

  private addEvent(event: AudioParamEvent): void {
    let index = this.events.length;
    for (let i = 0; i < this.events.length; i++) {
      if (this.events[i].Time > event.Time) {
        index = i;
        break;
      }
    }
    this.events.splice(index, 0, event);
  }

  private valueAt(time: Seconds): number {
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
          return interpolateExponentialValue(
            cursor,
            value,
            event.Time,
            event.Value,
            time,
          );
        }
        return evaluateParam(cursor, value, time, mode, target, timeConstant);
      }

      value = evaluateParam(
        cursor,
        value,
        event.Time,
        mode,
        target,
        timeConstant,
      );
      cursor = event.Time;
      if (event.Kind === AudioParamEventKind.Target) {
        mode = AudioParamEventKind.Target;
        target = event.Value;
        timeConstant = event.TimeConstant;
      } else {
        mode = AudioParamEventKind.Set;
        value = event.Value;
      }
    }

    return evaluateParam(cursor, value, time, mode, target, timeConstant);
  }
}

function evaluateParam(
  startTime: Seconds,
  startValue: number,
  time: Seconds,
  mode: AudioParamEventKind,
  target: number,
  timeConstant: Seconds,
): number {
  if (time <= startTime || mode !== AudioParamEventKind.Target) {
    return startValue;
  }
  if (timeConstant <= 0) {
    return target;
  }
  const factor = 1 - Math.exp(-(time - startTime) / timeConstant);
  return startValue + (target - startValue) * factor;
}

function interpolateValue(
  startTime: Seconds,
  startValue: number,
  endTime: Seconds,
  endValue: number,
  time: Seconds,
): number {
  if (endTime <= startTime) {
    return endValue;
  }
  const ratio = Math.max(0, Math.min(1, (time - startTime) / (endTime - startTime)));
  return startValue + (endValue - startValue) * ratio;
}

function interpolateExponentialValue(
  startTime: Seconds,
  startValue: number,
  endTime: Seconds,
  endValue: number,
  time: Seconds,
): number {
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

export class MockAudioNode implements AudioNode {
  // 记录连接关系便于测试断言；Mock 后端不实际渲染 PCM。
  readonly Connections: AudioNode[] = [];

  Connect(destination: AudioNode): void {
    if (destination && !this.Connections.includes(destination)) {
      this.Connections.push(destination);
    }
  }

  Disconnect(): void {
    this.Connections.length = 0;
  }
}

export class MockAudioScheduledSourceNode
  extends MockAudioNode
  implements AudioScheduledSourceNode
{
  // Starts/Stops 保留所有调用记录，包括被忽略的重复调度，
  // 便于验证音序器生成的 MIDI 时间。
  readonly Starts: Seconds[] = [];
  readonly Stops: Seconds[] = [];

  protected startScheduled = false;
  protected startTime: Seconds = 0;
  protected started = false;
  protected stopScheduled = false;
  protected stopTime: Seconds = 0;
  protected ended = false;
  protected onEnded: (() => void) | null = null;

  constructor(protected readonly Context: MockAudioContext) {
    super();
  }

  Start(when: Seconds): void {
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

  Stop(when: Seconds): void {
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

  SetOnEnded(handler: (() => void) | null): void {
    this.onEnded = handler;
    if (this.ended && handler) {
      handler();
    }
  }

  IsEnded(): boolean {
    return this.ended;
  }

  internalTick(now: Seconds): void {
    // Mock 上下文在 GetCurrentTime 和短定时器中被驱动，
    // 到达结束时间时补发 onended。
    if (this.ended || !this.startScheduled) {
      return;
    }
    if (!this.started && now >= this.startTime) {
      this.started = true;
    }
    const finished =
      this.started && this.stopScheduled && now >= this.stopTime;
    if (finished) {
      this.internalFinish();
    }
  }

  internalFinish(): void {
    if (this.ended || !this.started) {
      return;
    }
    this.ended = true;
    this.Context.internalUnregisterSource(this);
    this.onEnded?.();
  }
}

export class MockOscillatorNode
  extends MockAudioScheduledSourceNode
  implements OscillatorNode
{
  readonly Frequency: DefaultAudioParam;
  readonly Detune: DefaultAudioParam;
  private Type: OscillatorType = OscillatorType.Sine;

  constructor(context: MockAudioContext) {
    super(context);
    const now = context.GetCurrentTime();
    this.Frequency = new DefaultAudioParam(
      MOCK_DEFAULT_FREQUENCY,
      now,
      () => context.GetCurrentTime(),
    );
    this.Detune = new DefaultAudioParam(0, now, () =>
      context.GetCurrentTime(),
    );
  }

  GetFrequency(): AudioParam {
    return this.Frequency;
  }

  GetDetune(): AudioParam {
    return this.Detune;
  }

  GetType(): OscillatorType {
    return this.Type;
  }

  SetType(nodeType: OscillatorType): void {
    this.Type = nodeType;
  }
}

export class MockGainNode extends MockAudioNode implements GainNode {
  readonly Gain: DefaultAudioParam;

  constructor(context: MockAudioContext) {
    super();
    this.Gain = new DefaultAudioParam(
      MOCK_DEFAULT_GAIN,
      context.GetCurrentTime(),
      () => context.GetCurrentTime(),
    );
  }

  GetGain(): AudioParam {
    return this.Gain;
  }
}

export class MockAudioDestinationNode
  extends MockAudioNode
  implements AudioDestinationNode
{
  constructor(private readonly MaxChannels = MOCK_DEFAULT_CHANNEL_COUNT) {
    super();
  }

  GetMaxChannelCount(): number {
    return this.MaxChannels;
  }
}

export class MockAudioBuffer implements AudioBuffer {
  constructor(
    private readonly SampleRate: SampleRate,
    private readonly Channels: SampleValue[][],
  ) {}

  GetSampleRate(): SampleRate {
    return this.SampleRate;
  }

  GetLength(): SampleCount {
    return this.Channels.length === 0 ? 0 : this.Channels[0].length;
  }

  GetNumberOfChannels(): number {
    return this.Channels.length;
  }

  GetChannelData(channel: number): SampleValue[] | null {
    if (channel < 0 || channel >= this.Channels.length) {
      return null;
    }
    return this.Channels[channel];
  }
}

export class MockAudioBufferSourceNode
  extends MockAudioScheduledSourceNode
  implements AudioBufferSourceNode
{
  readonly PlaybackRate: DefaultAudioParam;
  readonly Detune: DefaultAudioParam;
  private Buffer: AudioBuffer | null = null;

  constructor(context: MockAudioContext) {
    super(context);
    const now = context.GetCurrentTime();
    this.PlaybackRate = new DefaultAudioParam(
      MOCK_DEFAULT_PLAYBACK_RATE,
      now,
      () => context.GetCurrentTime(),
    );
    this.Detune = new DefaultAudioParam(0, now, () =>
      context.GetCurrentTime(),
    );
  }

  GetBuffer(): AudioBuffer | null {
    return this.Buffer;
  }

  SetBuffer(buffer: AudioBuffer | null): void {
    this.Buffer = buffer;
  }

  GetPlaybackRate(): AudioParam {
    return this.PlaybackRate;
  }

  GetDetune(): AudioParam {
    return this.Detune;
  }
}

export class MockAudioContextFactory implements AudioContextFactory {
  constructor(private readonly SampleRate = MOCK_DEFAULT_SAMPLE_RATE) {}

  NewAudioContext(): AudioContext {
    return new MockAudioContext(this.SampleRate);
  }
}

export class MockAudioContext implements AudioContext {
  private state: AudioContextState = AudioContextState.Suspended;
  private readonly destination: MockAudioDestinationNode;
  private readonly sources = new Set<MockAudioScheduledSourceNode>();
  private elapsed: Seconds = 0;
  private runningSince = 0;
  private sourceCheck: ReturnType<typeof setTimeout> | null = null;

  constructor(private readonly sampleRate = MOCK_DEFAULT_SAMPLE_RATE) {
    this.destination = new MockAudioDestinationNode(
      MOCK_DEFAULT_CHANNEL_COUNT,
    );
  }

  Close(): void {
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

  Resume(): void {
    if (
      this.state === AudioContextState.Closed ||
      this.state === AudioContextState.Running
    ) {
      return;
    }
    // 暂停期间累计 elapsed；恢复后 currentTime 从该点继续增长。
    this.state = AudioContextState.Running;
    this.runningSince = Date.now();
    this.runSources();
  }

  Suspend(): void {
    if (this.state === AudioContextState.Running) {
      this.freezeClock();
    }
    if (this.state !== AudioContextState.Closed) {
      this.state = AudioContextState.Suspended;
    }
  }

  CreateOscillator(): OscillatorNode {
    return new MockOscillatorNode(this);
  }

  CreateGain(): GainNode {
    return new MockGainNode(this);
  }

  CreateBuffer(
    numberOfChannels: number,
    length: SampleCount,
    sampleRate: SampleRate,
  ): AudioBuffer {
    // 与原 malgo/Mock 实现一致：负声道数按 0 处理，0 采样率回退到上下文采样率。
    const channelsCount = Math.max(0, Math.trunc(numberOfChannels));
    const sampleCount =
      Number.isFinite(length) && length >= 0 ? Math.trunc(length) : 0;
    const rate = sampleRate || this.sampleRate || MOCK_DEFAULT_SAMPLE_RATE;
    const channels: SampleValue[][] = [];
    for (let channel = 0; channel < channelsCount; channel++) {
      channels.push(new Array<SampleValue>(sampleCount).fill(0));
    }
    return new MockAudioBuffer(rate, channels);
  }

  CreateBufferSource(): AudioBufferSourceNode {
    return new MockAudioBufferSourceNode(this);
  }

  GetCurrentTime(): Seconds {
    this.runSources();
    return this.currentTime();
  }

  GetSampleRate(): SampleRate {
    return this.sampleRate;
  }

  GetDestination(): AudioDestinationNode {
    return this.destination;
  }

  GetState(): AudioContextState {
    return this.state;
  }

  internalRegisterSource(source: MockAudioScheduledSourceNode): void {
    if (this.state !== AudioContextState.Closed) {
      this.sources.add(source);
      this.scheduleSourceCheck();
    }
  }

  internalUnregisterSource(source: MockAudioScheduledSourceNode): void {
    this.sources.delete(source);
  }

  private currentTime(): Seconds {
    if (this.state !== AudioContextState.Running) {
      return this.elapsed;
    }
    return this.elapsed + (Date.now() - this.runningSince) / 1000;
  }

  private freezeClock(): void {
    if (this.state !== AudioContextState.Running) {
      return;
    }
    this.elapsed += (Date.now() - this.runningSince) / 1000;
    this.runningSince = 0;
  }

  private runSources(): void {
    if (this.state !== AudioContextState.Running) {
      return;
    }
    const now = this.currentTime();
    for (const source of [...this.sources]) {
      source.internalTick(now);
    }
  }

  private scheduleSourceCheck(): void {
    // 用低频轮询推动 source 完成状态；Node 下 unref，避免测试进程被定时器挂住。
    if (this.sourceCheck !== null || this.sources.size === 0) {
      return;
    }
    this.sourceCheck = setTimeout(() => {
      this.sourceCheck = null;
      this.runSources();
      this.scheduleSourceCheck();
    }, 10);
    const handle = this.sourceCheck as unknown as { unref?: () => void };
    handle.unref?.();
  }
}

export class MockDriver implements Driver {
  constructor(private readonly sampleRate = MOCK_DEFAULT_SAMPLE_RATE) {}

  GetAudioContextFactory(): AudioContextFactory {
    return new MockAudioContextFactory(this.sampleRate);
  }
}

export function NewMockDriver(sampleRate = MOCK_DEFAULT_SAMPLE_RATE): Driver {
  return new MockDriver(sampleRate);
}

////////////////////////////////////////////////////////////////////////////////
// Browser Web Audio adapter

// 原生对象使用 any 隔离 DOM 类型差异，公开 API 仍只暴露本文件的
// AudioContext 接口。这样既支持浏览器，也不要求消费者启用 DOM lib。
type NativeAudioContext = {
  currentTime: number;
  sampleRate: number;
  state: string;
  destination: any;
  close(): Promise<void>;
  resume(): Promise<void>;
  suspend(): Promise<void>;
  createOscillator(): any;
  createGain(): any;
  createBuffer(
    numberOfChannels: number,
    length: number,
    sampleRate: number,
  ): any;
  createBufferSource(): any;
};

class NativeAudioParamAdapter implements AudioParam {
  constructor(private readonly param: any) {}

  GetValue(): number {
    return Number(this.param.value);
  }

  SetValue(value: number): void {
    this.param.value = value;
  }

  SetValueAtTime(value: number, startTime: Seconds): void {
    this.param.setValueAtTime(value, startTime);
  }

  LinearRampToValueAtTime(value: number, endTime: Seconds): void {
    this.param.linearRampToValueAtTime(value, endTime);
  }

  ExponentialRampToValueAtTime(value: number, endTime: Seconds): void {
    this.param.exponentialRampToValueAtTime(value, endTime);
  }

  SetTargetAtTime(
    target: number,
    startTime: Seconds,
    timeConstant: Seconds,
  ): void {
    this.param.setTargetAtTime(target, startTime, timeConstant);
  }

  CancelScheduledValues(startTime: Seconds): void {
    this.param.cancelScheduledValues(startTime);
  }
}

class NativeAudioNodeAdapter implements AudioNode {
  constructor(protected readonly node: any) {}

  Connect(destination: AudioNode): void {
    // 只有适配器包装的节点才能可靠转换为原生 AudioNode。
    if (destination instanceof NativeAudioNodeAdapter) {
      this.node.connect(destination.nativeNode());
      return;
    }
    const nativeDestination = (destination as any).NativeNode;
    if (nativeDestination) {
      this.node.connect(nativeDestination);
    }
  }

  Disconnect(): void {
    this.node.disconnect();
  }

  nativeNode(): any {
    return this.node;
  }
}

class NativeAudioScheduledSourceAdapter
  extends NativeAudioNodeAdapter
  implements AudioScheduledSourceNode
{
  Start(when: Seconds): void {
    this.node.start(normalizeAudioTime(when));
  }

  Stop(when: Seconds): void {
    this.node.stop(normalizeAudioTime(when));
  }

  SetOnEnded(handler: (() => void) | null): void {
    this.node.onended = handler;
  }
}

class NativeOscillatorAdapter
  extends NativeAudioScheduledSourceAdapter
  implements OscillatorNode
{
  GetFrequency(): AudioParam {
    return new NativeAudioParamAdapter(this.node.frequency);
  }

  GetDetune(): AudioParam {
    return new NativeAudioParamAdapter(this.node.detune);
  }

  GetType(): OscillatorType {
    return this.node.type as OscillatorType;
  }

  SetType(nodeType: OscillatorType): void {
    this.node.type = nodeType;
  }
}

class NativeGainAdapter extends NativeAudioNodeAdapter implements GainNode {
  GetGain(): AudioParam {
    return new NativeAudioParamAdapter(this.node.gain);
  }
}

class NativeDestinationAdapter
  extends NativeAudioNodeAdapter
  implements AudioDestinationNode
{
  GetMaxChannelCount(): number {
    return Number(this.node.maxChannelCount ?? 2);
  }
}

class NativeAudioBufferAdapter implements AudioBuffer {
  constructor(private readonly buffer: any) {}

  GetSampleRate(): SampleRate {
    return Number(this.buffer.sampleRate);
  }

  GetLength(): SampleCount {
    return Number(this.buffer.length);
  }

  GetNumberOfChannels(): number {
    return Number(this.buffer.numberOfChannels);
  }

  GetChannelData(channel: number): SampleValue[] | null {
    if (channel < 0 || channel >= this.GetNumberOfChannels()) {
      return null;
    }
    return this.buffer.getChannelData(channel) as SampleValue[];
  }
}

class NativeBufferSourceAdapter
  extends NativeAudioScheduledSourceAdapter
  implements AudioBufferSourceNode
{
  GetBuffer(): AudioBuffer | null {
    return this.node.buffer ? new NativeAudioBufferAdapter(this.node.buffer) : null;
  }

  SetBuffer(buffer: AudioBuffer | null): void {
    // 原生 buffer 直接复用；来自 Mock 等后端的 buffer 则复制为原生 AudioBuffer。
    if (buffer instanceof NativeAudioBufferAdapter) {
      this.node.buffer = (buffer as any).buffer;
    } else if (buffer === null) {
      this.node.buffer = null;
    } else {
      const native = this.node.context.createBuffer(
        buffer.GetNumberOfChannels(),
        buffer.GetLength(),
        buffer.GetSampleRate(),
      );
      for (let channel = 0; channel < buffer.GetNumberOfChannels(); channel++) {
        const data = buffer.GetChannelData(channel);
        if (data) {
          native.copyToChannel(data, channel);
        }
      }
      this.node.buffer = native;
    }
  }

  GetPlaybackRate(): AudioParam {
    return new NativeAudioParamAdapter(this.node.playbackRate);
  }

  GetDetune(): AudioParam {
    return new NativeAudioParamAdapter(this.node.detune);
  }
}

export class WebAudioContext implements AudioContext {
  // 把本文件的 Web Audio 抽象转发到浏览器原生 AudioContext。
  private readonly destination: NativeDestinationAdapter;

  constructor(private readonly context: NativeAudioContext) {
    this.destination = new NativeDestinationAdapter(context.destination);
  }

  Close(): void {
    void this.context.close();
  }

  Resume(): Promise<void> {
    return this.context.resume();
  }

  Suspend(): Promise<void> {
    return this.context.suspend();
  }

  CreateOscillator(): OscillatorNode {
    return new NativeOscillatorAdapter(this.context.createOscillator());
  }

  CreateGain(): GainNode {
    return new NativeGainAdapter(this.context.createGain());
  }

  CreateBuffer(
    numberOfChannels: number,
    length: SampleCount,
    sampleRate: SampleRate,
  ): AudioBuffer {
    const rate = sampleRate || this.context.sampleRate;
    return new NativeAudioBufferAdapter(
      this.context.createBuffer(numberOfChannels, length, rate),
    );
  }

  CreateBufferSource(): AudioBufferSourceNode {
    return new NativeBufferSourceAdapter(this.context.createBufferSource());
  }

  GetCurrentTime(): Seconds {
    return this.context.currentTime;
  }

  GetSampleRate(): SampleRate {
    return this.context.sampleRate;
  }

  GetDestination(): AudioDestinationNode {
    return this.destination;
  }

  GetState(): AudioContextState {
    return this.context.state as AudioContextState;
  }
}

export class WebAudioContextFactory implements AudioContextFactory {
  NewAudioContext(): AudioContext {
    // Safari 的旧实例名是 webkitAudioContext。
    const scope = globalThis as any;
    const Constructor = scope.AudioContext ?? scope.webkitAudioContext;
    if (!Constructor) {
      throw new Error(
        "Web Audio API is unavailable; use MockDriver or provide a Driver",
      );
    }
    return new WebAudioContext(new Constructor() as NativeAudioContext);
  }
}

export class WebAudioDriver implements Driver {
  // 浏览器推荐使用该 Driver；无 Web Audio 的环境可改用 MockDriver。
  GetAudioContextFactory(): AudioContextFactory {
    return new WebAudioContextFactory();
  }
}

////////////////////////////////////////////////////////////////////////////////
// Timeline and positions

export enum TimelineType {
  Min = 0,
  Root,
  Work,
  Clip,
  Max,
}

export class Timeline {
  // 每个 Timeline 的 Offset 相对于 Parent；绝对位置是所有祖先 Offset 之和
  // 加上当前 Pos。GetRoot 限制 32 层是为了防止错误父子环造成无限循环。
  Parent: Timeline | null = null;
  Label = "";
  Offset: Seconds = 0;
  Length: Seconds = 0;
  Type: TimelineType = TimelineType.Min;
  Profile: MusicProfile | null = null;
  Loop = false;

  Init(): this {
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

  IsTypeOf(type: TimelineType): boolean {
    return this.Type === type;
  }

  HasSameRoot(other: Timeline | null): boolean {
    return this.GetRoot().Same(other?.GetRoot() ?? null);
  }

  Same(other: Timeline | null): boolean {
    return this === other;
  }

  Equals(other: Timeline | null): boolean {
    if (!other) {
      return false;
    }
    if (this === other) {
      return true;
    }
    if (
      this.Label !== other.Label ||
      this.Offset !== other.Offset ||
      this.Length !== other.Length ||
      this.Loop !== other.Loop
    ) {
      return false;
    }
    if (!this.Parent || !other.Parent) {
      return this.Parent === other.Parent;
    }
    return this.Parent.Equals(other.Parent);
  }

  GetRoot(): Timeline {
    let current: Timeline = this;
    for (let ttl = 32; ttl > 0 && current.Parent; ttl--) {
      current = current.Parent;
    }
    return current;
  }

  GetTotalOffset(): Seconds {
    let total = 0;
    let current: Timeline | null = this;
    for (let ttl = 32; ttl > 0 && current; ttl--) {
      total += current.Offset;
      current = current.Parent;
    }
    return total;
  }
}

export function AbsolutePosition(value: Seconds): AbsolutePosition {
  return value;
}

export function RelativePositionToAbsolute(position: AbsolutePosition): AbsolutePosition {
  return position;
}

export class RelativePosition {
  constructor(
    public Ref: Timeline | null = null,
    public Pos: Seconds = 0,
  ) {}

  Absolute(): AbsolutePosition {
    return (this.Ref?.GetTotalOffset() ?? 0) + this.Pos;
  }

  Clone(): RelativePosition {
    return new RelativePosition(this.Ref, this.Pos);
  }
}

export function AbsolutePositionRelative(
  absolute: AbsolutePosition,
  base?: Timeline | null,
  destination?: RelativePosition,
): RelativePosition {
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
  Delay: Seconds = 0;
  Interval: Seconds = 0;
  Handler: ((timer: Timer) => void) | null = null;

  private workerID = 0;
  private timeout: ReturnType<typeof setTimeout> | null = null;

  Start(): void {
    this.Stop();
    const worker = this.workerID;
    const handler = this.Handler;
    if (!handler) {
      return;
    }
    this.work(worker, handler, this.Delay);
  }

  Stop(): void {
    this.workerID++;
    if (this.timeout !== null) {
      clearTimeout(this.timeout);
      this.timeout = null;
    }
  }

  private work(
    worker: number,
    handler: (timer: Timer) => void,
    delay: Seconds,
  ): void {
    this.timeout = setTimeout(() => {
      this.timeout = null;
      if (worker !== this.workerID) {
        return;
      }
      try {
        handler(this);
      } catch (error) {
        handleError(error);
      }
      if (worker === this.workerID) {
        this.work(worker, handler, this.Interval);
      }
    }, this.sec2ms(delay));
  }

  private sec2ms(seconds: Seconds): number {
    return Math.max(1, Math.trunc(seconds * 1000));
  }
}

////////////////////////////////////////////////////////////////////////////////
// Event dispatch

export enum EventType {
  Min = 0,
  LoopBegin,
  LoopEnd,
  Max,
}

export const EventTypeMin = EventType.Min;
export const EventTypeLoopBegin = EventType.LoopBegin;
export const EventTypeLoopEnd = EventType.LoopEnd;
export const EventTypeMax = EventType.Max;

export type EventID = number;

export interface Event {
  // Type + ID 共同确定匹配的回调；Sender 是任意业务载荷。
  Type: EventType | number;
  ID: EventID;
  Sender?: unknown;
  Context?: EngineContext | null;
}

export type EventCallbackFunc = (event: Event) => void;

export interface EventCallback {
  Type: EventType | number;
  ID: EventID;
  Fn: EventCallbackFunc;
}

export interface EventDispatcher {
  Dispatch(event: Event | null): void;
  AddCallback(callback: EventCallback | null): void;
  RemoveCallback(callback: EventCallback | null): void;
}

export interface EventRunner {
  Run(): void;
  RunContext(context: EventContext | null): void;
}

export class EventContext {
  // 事件先进入 buffer，再由 Runner 按 FIFO 取出。回调中 Dispatch 的新事件
  // 会追加到队尾，因此 Run 仍能在同一次调用中处理它。
  private dispatcherInstance: EventDispatcher | null = null;
  private runnerInstance: EventRunner | null = null;
  private callbacks: EventCallback[] = [];
  private buffer: Event[] = [];

  constructor() {
    this.dispatcherInstance = new DefaultEventDispatcher(this);
    this.runnerInstance = new DefaultEventRunner(this);
  }

  GetDispatcher(): EventDispatcher | null {
    return this.dispatcherInstance;
  }

  GetRunner(): EventRunner | null {
    return this.runnerInstance;
  }

  internalAdd(event: Event): void {
    this.buffer.push({ ...event });
  }

  internalAddCallback(callback: EventCallback): void {
    this.callbacks.push(callback);
  }

  internalRemoveCallback(callback: EventCallback): boolean {
    const index = this.callbacks.indexOf(callback);
    if (index >= 0) {
      this.callbacks.splice(index, 1);
      return true;
    }
    return false;
  }

  internalTake(): Event | undefined {
    // 先取出再执行回调；如果回调抛错，当前事件已被消费，后续事件仍保留。
    return this.buffer.shift();
  }

  internalMatching(event: Event): EventCallback[] {
    return this.callbacks.filter(
      (callback) =>
        callback.Type === event.Type &&
        callback.ID === event.ID &&
        typeof callback.Fn === "function",
    );
  }

  BufferLength(): number {
    return this.buffer.length;
  }
}

export class DefaultEventDispatcher implements EventDispatcher {
  constructor(private readonly context: EventContext | null) {}

  Dispatch(event: Event | null): void {
    if (!event) {
      return;
    }
    if (!this.context) {
      throw new Error("event dispatcher context is nil");
    }
    this.context.internalAdd(event);
  }

  AddCallback(callback: EventCallback | null): void {
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

  RemoveCallback(callback: EventCallback | null): void {
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

export class DefaultEventRunner implements EventRunner {
  constructor(private readonly context: EventContext | null) {}

  Run(): void {
    if (!this.context) {
      throw new Error("event runner context is nil");
    }
    this.RunContext(this.context);
  }

  RunContext(context: EventContext | null): void {
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
        } catch (cause) {
          // 保留事件类型和 ID，便于定位具体分发失败点；原始错误放在 cause。
          const detail = cause instanceof Error ? cause.message : String(cause);
          throw new Error(
            `dispatch event type ${event.Type} id ${event.ID}: ${detail}`,
            { cause },
          );
        }
      }
    }
  }
}

export function NewEventContext(): EventContext {
  return new EventContext();
}

////////////////////////////////////////////////////////////////////////////////
// Lifecycle

export type LifeCycleStageFunc = () => void | Promise<void>;
export type LifeCycleInitFunc = (context: EngineContext) => void;

export class LifeCycle {
  Order = 0;
  OnInit: LifeCycleInitFunc | null = null;
  OnCreate: LifeCycleStageFunc | null = null;
  OnStart: LifeCycleStageFunc | null = null;
  OnResume: LifeCycleStageFunc | null = null;
  OnRun: LifeCycleStageFunc | null = null;
  OnPause: LifeCycleStageFunc | null = null;
  OnStop: LifeCycleStageFunc | null = null;
  OnDestroy: LifeCycleStageFunc | null = null;
}

export interface Life {
  GetLifeCycle(): LifeCycle | null;
}

export interface LifeManager {
  GetMain(): Life;
  Add(life: Life | null): void;
  Remove(life: Life | null): void;
}

class LifeCycleContext {
  cache: LifeCycle[] | null = null;
  main: Life | null = null;
  context: EngineContext | null = null;
  all: Life[] = [];
  onInitDone = false;
  onCreateDone = false;
  onStartDone = false;
  paused = false;
}

export class DefaultLifeManager implements LifeManager {
  private readonly state = new LifeCycleContext();
  private readonly mainLife: MainLife;

  constructor() {
    this.mainLife = new MainLife(this.state);
    this.state.main = this.mainLife;
  }

  GetMain(): Life {
    return this.mainLife;
  }

  Add(life: Life | null): void {
    if (!life) {
      return;
    }
    this.state.all.push(life);
    this.state.cache = null;
    replayCurrentLifeCycle(life, this.state);
  }

  Remove(life: Life | null): void {
    if (!life) {
      return;
    }
    this.state.all = this.state.all.filter((item) => item !== life);
    this.state.cache = null;
  }
}

class MainLife implements Life {
  constructor(private readonly state: LifeCycleContext) {}

  GetLifeCycle(): LifeCycle {
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

  private onInit(context: EngineContext): void {
    if (!context) {
      throw new Error("innerMainLife.onInit() : param 'ec:EngineContext' is nil");
    }
    this.state.context = context;
    this.state.onInitDone = true;
    for (const cycle of this.cache()) {
      cycle.OnInit?.(context);
    }
  }

  private async onStage(
    stage:
      | "OnCreate"
      | "OnStart"
      | "OnResume"
      | "OnRun"
      | "OnPause"
      | "OnStop"
      | "OnDestroy",
    reverse: boolean,
  ): Promise<void> {
    if (stage === "OnCreate") this.state.onCreateDone = true;
    if (stage === "OnStart") this.state.onStartDone = true;
    if (stage === "OnPause") this.state.paused = true;
    if (stage === "OnResume") this.state.paused = false;
    const cycles = this.cache();
    const ordered = reverse ? [...cycles].reverse() : cycles;
    for (const cycle of ordered) {
      await cycle[stage]?.();
    }
  }

  private cache(): LifeCycle[] {
    if (!this.state.cache) {
      this.state.cache = this.state.all
        .map((life) => life.GetLifeCycle())
        .filter((cycle): cycle is LifeCycle => cycle !== null)
        .sort((a, b) => a.Order - b.Order);
    }
    return this.state.cache;
  }
}

function replayCurrentLifeCycle(life: Life, state: LifeCycleContext): void {
  const cycle = life.GetLifeCycle();
  if (!cycle) {
    return;
  }
  const invoke = (fn: (() => void | Promise<void>) | null) => {
    if (!fn) return;
    try {
      const result = fn();
      if (result instanceof Promise) {
        result.catch(handleError);
      }
    } catch (error) {
      handleError(error);
    }
  };
  if (state.onInitDone && state.context) {
    try {
      cycle.OnInit?.(state.context);
    } catch (error) {
      handleError(error);
    }
  }
  if (state.onCreateDone) invoke(cycle.OnCreate);
  if (state.onStartDone) {
    invoke(cycle.OnStart);
    if (!state.paused) {
      invoke(cycle.OnResume);
      invoke(cycle.OnRun);
    }
  }
}

export function NewLifeManager(): LifeManager {
  return new DefaultLifeManager();
}

export class LifeCycleRunner {
  // 主生命周期驱动引擎：OnRun 通常会等待 Engine.Join()，
  // 所以后续 Pause/Stop/Destroy 必须等它结束后才能执行。
  private context: EngineContext | null = null;

  Init(context: EngineContext): this {
    this.context = context;
    return this;
  }

  async RunLC(cycle: LifeCycle | null): Promise<void> {
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

  RunL(life: Life | null): Promise<void> {
    return this.RunLC(life?.GetLifeCycle() ?? null);
  }

  RunLM(manager: LifeManager): Promise<void> {
    return this.RunL(manager.GetMain());
  }
}

export class LifeCycleStageLogger implements Life, Part {
  // 可作为配置中的 Part 加入，用于观察每个生命周期阶段的调用顺序。
  private logger: Logger = Log();

  GetInfo(): PartInfo {
    return new PartInfo({ Name: "LifeCycleStageLogger", Part: this });
  }

  GetLifeCycle(): LifeCycle {
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
    ] as const) {
      cycle[stage] = () => this.log(stage);
    }
    return cycle;
  }

  private log(stage: string): void {
    this.logger.Info("engine.lifecycle.%s()", stage);
  }
}

////////////////////////////////////////////////////////////////////////////////
// Parts

export type PartID = number;
export type PartName = string;
export type PartClass = string;

export interface Part {
  // Part 是可注册组件；若同时实现 Life/InstrumentProvider，
  // 引擎会自动把它纳入生命周期或乐器注册表。
  GetInfo(): PartInfo | null;
}

export interface PartFactory {
  CreatePart(info: PartInfo): Part | null;
  GetRegistration(): PartRegistration | null;
}

export interface PartManager {
  ListNames(): PartName[];
  FindPartByID(id: PartID): Part | null;
  FindPartByName(name: PartName): Part | null;
  ListPartsWithClass(partClass: PartClass): Part[];
}

export interface PartRegistry {
  Register(registration: PartRegistration | null): void;
  RegisterFactory(factory: PartFactory | null): void;
  RegisterPart(part: Part | null): void;
  ListNames(): PartName[];
  FindByID(id: PartID): PartRegistration | null;
  FindByName(name: PartName): PartRegistration | null;
  ListWithClass(partClass: PartClass): PartRegistration[];
}

export class PartMeta {
  ID: PartID = 0;
  Name: PartName = "";
  Classes: PartClass[] = [];
  Context: PartContext | null = null;
  Factory: PartFactory | null = null;
  Part: Part | null = null;

  constructor(init: Partial<PartMeta> = {}) {
    Object.assign(this, init);
  }
}

export class PartInfo extends PartMeta {}

export class PartRegistration extends PartMeta {}

export class PartContext {
  EC: EngineContext | null;
  Registry: PartRegistry;
  Manager: PartManager;

  constructor(context: EngineContext | null = null) {
    this.EC = context;
    const state: PartRegistryState = {
      context: this,
      byID: new Map<PartID, PartRegistration>(),
      byName: new Map<PartName, PartRegistration>(),
      managerByID: new Map<PartID, Part>(),
      managerByName: new Map<PartName, Part>(),
      nextID: 0,
    };
    this.Registry = new DefaultPartRegistry(state);
    this.Manager = new DefaultPartManager(state);
  }
}

interface PartRegistryState {
  context: PartContext;
  byID: Map<PartID, PartRegistration>;
  byName: Map<PartName, PartRegistration>;
  managerByID: Map<PartID, Part>;
  managerByName: Map<PartName, Part>;
  nextID: number;
}

export class PartInfoBuilder {
  private readonly info = new PartInfo();

  Build(): PartInfo {
    return new PartInfo(this.info);
  }

  SetPart(part: Part | null): this {
    this.info.Part = part;
    return this;
  }

  SetFactory(factory: PartFactory | null): this {
    this.info.Factory = factory;
    return this;
  }

  SetName(name: PartName): this {
    this.info.Name = name;
    return this;
  }

  SetID(id: PartID): this {
    this.info.ID = id;
    return this;
  }

  SetContext(context: PartContext | null): this {
    this.info.Context = context;
    return this;
  }

  SetClasses(...classes: PartClass[]): this {
    this.info.Classes = classes;
    return this;
  }
}

export class DefaultPartRegistry implements PartRegistry {
  constructor(private readonly state: PartRegistryState) {}

  RegisterFactory(factory: PartFactory | null): void {
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

  RegisterPart(part: Part | null): void {
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

  Register(registration: PartRegistration | null): void {
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
      throw new Error(
        `part ${JSON.stringify(registration.Name)} has neither a factory nor an instance`,
      );
    }
    const existingID = this.state.byID.get(registration.ID);
    if (existingID) {
      throw new Error(
        `part id ${registration.ID} is already registered as ${JSON.stringify(existingID.Name)}`,
      );
    }
    const existingName = this.state.byName.get(registration.Name);
    if (existingName) {
      throw new Error(
        `part name ${JSON.stringify(registration.Name)} is already registered with id ${existingName.ID}`,
      );
    }
    registration.Context ??= this.state.context;
    this.state.byID.set(registration.ID, registration);
    this.state.byName.set(registration.Name, registration);
  }

  ListNames(): PartName[] {
    return [...this.state.byName.keys()].sort();
  }

  FindByID(id: PartID): PartRegistration | null {
    return this.state.byID.get(id) ?? null;
  }

  FindByName(name: PartName): PartRegistration | null {
    return this.state.byName.get(name) ?? null;
  }

  ListWithClass(partClass: PartClass): PartRegistration[] {
    return [...this.state.byID.values()]
      .filter((item) => item.Classes.includes(partClass))
      .sort((a, b) =>
        a.ID === b.ID ? a.Name.localeCompare(b.Name) : a.ID - b.ID,
      );
  }

  private makeNextID(): PartID {
    if (this.state.nextID === 0) {
      this.state.nextID = 0x123456789;
    }
    this.state.nextID++;
    return this.state.nextID;
  }
}

export class DefaultPartManager implements PartManager {
  // Manager 按需调用 Factory.CreatePart，并缓存创建结果，
  // 因此相同 ID/Name 的后续查询会返回同一个实例。
  constructor(private readonly state: PartRegistryState) {}

  ListNames(): PartName[] {
    return [...this.state.byName.keys()].sort();
  }

  FindPartByID(id: PartID): Part | null {
    const registration = this.state.byID.get(id);
    return registration ? this.partFromRegistration(registration) : null;
  }

  FindPartByName(name: PartName): Part | null {
    const registration = this.state.byName.get(name);
    return registration ? this.partFromRegistration(registration) : null;
  }

  ListPartsWithClass(partClass: PartClass): Part[] {
    return [...this.state.byID.values()]
      .filter((item) => item.Classes.includes(partClass))
      .sort((a, b) =>
        a.ID === b.ID ? a.Name.localeCompare(b.Name) : a.ID - b.ID,
      )
      .map((item) => this.partFromRegistration(item))
      .filter((part): part is Part => part !== null);
  }

  private partFromRegistration(registration: PartRegistration): Part | null {
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
      throw new Error(
        `part ${JSON.stringify(registration.Name)} has no factory`,
      );
    }
    const info = new PartInfo(registration);
    info.Context = this.state.context;
    const part = registration.Factory.CreatePart(info);
    if (!part) {
      throw new Error(
        `create part ${JSON.stringify(registration.Name)}: factory returned nil`,
      );
    }
    registration.Part = part;
    this.cache(registration, part);
    return part;
  }

  private cache(registration: PartRegistration, part: Part): void {
    this.state.managerByID.set(registration.ID, part);
    this.state.managerByName.set(registration.Name, part);
  }
}

export function NewPartContext(context: EngineContext | null = null): PartContext {
  return new PartContext(context);
}

////////////////////////////////////////////////////////////////////////////////
// MIDI and instruments

export enum MidiEventType {
  // 首尾哨兵沿用 Go iota 结构，业务事件从 NoteOn 开始。
  Min = 0,
  NoteOn,
  NoteOff,
  Max,
}

export class MidiNote {
  // Position 是绝对音频时间。小于 1ms 的值由乐器解释为“立即播放”。
  EventType: MidiEventType = MidiEventType.NoteOn;
  Position: AbsolutePosition = 0;
  Channel = 0;
  Key: MidiNoteNumber = 0;
  Velocity: MidiNoteVelocity = 0;

  constructor(init: Partial<MidiNote> = {}) {
    Object.assign(this, init);
  }
}

export class MidiMessage {
  // 一个调度消息通常同时携带 NoteOn 和 NoteOff，
  // SimpleSynth 会据此预先安排同一个 voice 的起止时间。
  NoteOn: MidiNote | null = null;
  NoteOff: MidiNote | null = null;

  constructor(init: Partial<MidiMessage> = {}) {
    Object.assign(this, init);
  }
}

export interface Instrument {
  Dispatch(message: MidiMessage): void;
  GetInfo(): InstrumentInfo | null;
}

export class InstrumentInfo {
  Name = "";
  Label = "";
  Preset = "";

  constructor(init: Partial<InstrumentInfo> = {}) {
    Object.assign(this, init);
  }
}

export interface InstrumentFactory {
  ListInfo(): InstrumentInfo[];
  CreateInstrument(
    context: EngineContext,
    info: InstrumentInfo,
  ): Instrument | null;
}

export interface InstrumentManager extends Part, Life {
  ListInfo(): InstrumentInfo[];
  GetFactory(info: InstrumentInfo | null): InstrumentFactory | null;
}

export interface InstrumentHolder {
  GetInstrument(context: EngineContext): Instrument | null;
}

export interface InstrumentProvider {
  ListRegistrations(): InstrumentRegistration[];
}

export class InstrumentRegistration extends InstrumentInfo {
  Factory: InstrumentFactory | null;

  constructor(
    init: Partial<InstrumentRegistration> & {
      Factory?: InstrumentFactory | null;
    } = {},
  ) {
    super(init);
    this.Factory = init.Factory ?? null;
  }
}

export class InstrumentFactoryRegistry implements InstrumentFactory, Part, Life {
  // 按乐器 Name 选择工厂；未注册名称回退到 SetDefault 设置的工厂。
  private readonly table = new Map<string, InstrumentRegistration>();
  private factoryDefault: InstrumentFactory | null = null;

  GetInfo(): PartInfo {
    return new PartInfo({
      Name: "InstrumentFactoryRegistry",
      Classes: ["InstrumentFactoryRegistry"],
      Part: this,
    });
  }

  GetLifeCycle(): LifeCycle {
    return new LifeCycle();
  }

  CreateInstrument(
    context: EngineContext,
    info: InstrumentInfo,
  ): Instrument | null {
    if (!info) {
      throw new Error("instrument info is nil");
    }
    // 默认工厂只处理未注册名称，不会覆盖已注册名称对应的专用工厂。
    const registration = this.table.get(info.Name);
    const factory = registration?.Factory ?? this.factoryDefault;
    if (!factory) {
      throw new Error(
        `instrument ${JSON.stringify(info.Name)} is not registered and no default factory is set`,
      );
    }
    return factory.CreateInstrument(context, info);
  }

  SetDefault(factory: InstrumentFactory | null): void {
    if (factory) {
      this.factoryDefault = factory;
    }
  }

  ListInfo(): InstrumentInfo[] {
    return [...this.table.values()]
      .map((item) => new InstrumentInfo(item))
      .sort(compareInstrumentInfo);
  }

  Register(factory: InstrumentFactory | null): void {
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

  Add(registration: InstrumentRegistration | null): void {
    this.add(registration);
  }

  private add(registration: InstrumentRegistration | null): void {
    if (
      !registration ||
      !registration.Factory ||
      !registration.Name
    ) {
      return;
    }
    this.table.set(registration.Name, registration);
  }
}

function compareInstrumentInfo(a: InstrumentInfo, b: InstrumentInfo): number {
  return (
    a.Name.localeCompare(b.Name) ||
    a.Label.localeCompare(b.Label) ||
    a.Preset.localeCompare(b.Preset)
  );
}

export class DefaultInstrumentHolder implements InstrumentHolder {
  GetInstrument(context: EngineContext): Instrument | null {
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
    const instrument = factory.CreateInstrument(
      context,
      new InstrumentInfo({ Name: "", Preset: "triangle" }),
    );
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
  readonly table = new Map<string, InstrumentRegistration>();

  RegisterAllInstruments(context: EngineContext): void {
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

  RegisterInstrument(part: Part | null): void {
    if (!part) {
      throw new Error("instrument provider part is nil");
    }
    const provider = part as Part & Partial<InstrumentProvider>;
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
        throw new Error(
          `instrument ${JSON.stringify(registration.Name)} has no factory`,
        );
      }
      this.table.set(registration.Name, registration);
    }
  }
}

export class InstrumentManagerImpl implements InstrumentManager {
  readonly Registry = new InstrumentRegistry();

  GetFactory(info: InstrumentInfo | null): InstrumentFactory | null {
    if (!info) {
      throw new Error("instrument info is nil");
    }
    const registration = this.Registry.table.get(info.Name);
    if (!registration) {
      throw new Error(
        `instrument ${JSON.stringify(info.Name)} is not registered`,
      );
    }
    if (!registration.Factory) {
      throw new Error(
        `instrument ${JSON.stringify(info.Name)} has no factory`,
      );
    }
    return registration.Factory;
  }

  ListInfo(): InstrumentInfo[] {
    return [...this.Registry.table.values()]
      .map((item) => new InstrumentInfo(item))
      .sort(compareInstrumentInfo);
  }

  GetInfo(): PartInfo {
    return new PartInfoBuilder().SetName("InstrumentManager").Build();
  }

  GetLifeCycle(): LifeCycle {
    const cycle = new LifeCycle();
    cycle.OnInit = (context) => this.Registry.RegisterAllInstruments(context);
    return cycle;
  }
}

export function NewInstrumentManager(): InstrumentManager {
  return new InstrumentManagerImpl();
}

export class InnerInputInstrument implements Instrument {
  // 输入乐器是一个间接层：外部 MIDI 总是转发到当前主乐器，
  // 因此切换主乐器时不需要重新连接输入设备。
  constructor(private readonly context: EngineContext) {}

  Dispatch(message: MidiMessage): void {
    if (!message) {
      throw new Error("innerInputInstrument: param 'msg' (MidiMessage) is nil");
    }
    const target = this.context.CurrentInstrumentHolder.GetInstrument(
      this.context,
    );
    if (!target) {
      throw new Error("input target instrument is nil");
    }
    target.Dispatch(message);
  }

  GetInfo(): InstrumentInfo {
    return new InstrumentInfo({
      Name: "InputInstrument",
      Label: "Input-Instrument",
      Preset: "default",
    });
  }
}

export class AutoInputInstrumentHolder implements InstrumentHolder {
  GetInstrument(context: EngineContext): Instrument | null {
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
  Key: MidiNoteNumber = 0;
  Velocity: MidiNoteVelocity = 0;

  constructor(init: Partial<Note> = {}) {
    Object.assign(this, init);
  }
}

export class SimpleSynthVoice {
  constructor(
    readonly Source: AudioScheduledSourceNode,
    readonly Gain: GainNode,
  ) {}
}

export class SimpleSynthFactory implements InstrumentFactory {
  ListInfo(): InstrumentInfo[] {
    return [
      new InstrumentInfo({
        Name: "default",
        Label: "Simple Synth",
        Preset: "default",
      }),
    ];
  }

  CreateInstrument(
    context: EngineContext,
    info: InstrumentInfo,
  ): Instrument {
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
    const instrumentInfo =
      info ??
      new InstrumentInfo({
        Name: "default",
        Label: "Simple Synth",
        Preset: "default",
      });
    return new SimpleSynth(
      context.AudioContext,
      destination,
      instrumentInfo,
    );
  }
}

export class SimpleSynth implements Instrument {
  // 每个 MIDI key 同时只保留一个 voice。重复 NoteOn 会先停止旧 voice，
  // 避免同一个音符叠加造成非预期增益。
  private readonly waveform: string;
  private readonly voices = new Map<MidiNoteNumber, SimpleSynthVoice>();

  constructor(
    private readonly audio: AudioContext,
    private readonly destination: AudioDestinationNode,
    private readonly info: InstrumentInfo,
  ) {
    this.waveform = simpleSynthWaveform(info.Preset);
  }

  Dispatch(message: MidiMessage): void {
    if (!message) {
      return;
    }
    // 先处理 NoteOn，确保同一消息携带的 NoteOff 能停止刚创建的 voice。
    if (message.NoteOn) {
      if (message.NoteOn.Velocity === 0) {
        this.noteOff(message.NoteOn);
      } else {
        this.noteOn(message.NoteOn);
      }
    }
    if (message.NoteOff) {
      this.noteOff(message.NoteOff);
    }
  }

  GetInfo(): InstrumentInfo {
    return new InstrumentInfo(this.info);
  }

  private noteOn(note: MidiNote): void {
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

  private createVoice(note: MidiNote, now: Seconds): SimpleSynthVoice {
    const gain = this.audio.CreateGain();
    let source: AudioScheduledSourceNode;
    if (this.waveform === "noise") {
      const bufferSource = this.audio.CreateBufferSource();
      this.prepareNoise(bufferSource);
      source = bufferSource;
    } else {
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

  private noteOff(note: MidiNote): void {
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

  private voiceEnded(key: MidiNoteNumber, voice: SimpleSynthVoice): void {
    if (this.voices.get(key) === voice) {
      this.voices.delete(key);
    }
    voice.Source.Disconnect();
    voice.Gain.Disconnect();
  }

  private noteTime(note: MidiNote): Seconds {
    // 绝对位置接近 0 时按即时消息处理；否则直接作为 Web Audio 调度时间。
    return note.Position < 0.001 ? this.audio.GetCurrentTime() : note.Position;
  }

  private prepareNoise(source: AudioBufferSourceNode): void {
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

function simpleSynthWaveform(preset: string): string {
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

function simpleSynthOscillatorType(waveform: string): OscillatorType {
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

////////////////////////////////////////////////////////////////////////////////
// Sequencer cube

export type BankIndex = number;
export type BankName = string;
export type StepIndex = number;
export type StepLocation = number;
export type TrackIndex = number;
export type TrackName = string;

export interface StepTickHandler {
  // Every bank/step/cell exposes the same callback, so the sequencer can walk
  // one logical step without owning the concrete storage layout.
  OnStep(context: StepTickContext): void;
}

export interface BankHolder {
  GetBank(): Bank | null;
}

export class Cell {
  BankIndex: BankIndex = 0;
  StepIndex: StepIndex = 0;
  TrackIndex: TrackIndex = 0;

  constructor(init: Partial<Cell> = {}) {
    Object.assign(this, init);
  }
}

export class CellCache {
  readonly List: SimpleCell[] = [];

  Add(cell: SimpleCell | null): void {
    if (cell) {
      this.List.push(cell);
    }
  }

  Reset(): void {
    this.List.length = 0;
  }

  Sort(compare: (a: SimpleCell, b: SimpleCell) => number): void {
    this.List.sort(compare);
  }
}

export class SimpleCell extends Cell implements StepTickHandler {
  // 普通音符格。ComplexCell.First 指向其首个 SimpleCell，
  // 从而让一个多格音符只在第一个格子上触发一次。
  Context: Cube | null = null;
  OwnerStep: Step | null = null;
  OwnerBank: Bank | null = null;
  OwnerTrack: Track | null = null;
  Enabled = false;
  Note: Note | null = null;
  Complex: ComplexCell | null = null;

  OnStep(context: StepTickContext): void {
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

  DispatchNote(context: StepTickContext, note: Note): void {
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

export class ComplexCell extends Cell implements StepTickHandler {
  // ComplexCell 占用多个 SimpleCell 槽位。这里把 Notes 同时交给首发格，
  // 保持“多音”语义；更复杂时值拆分可由上层 Note 数据表达。
  First: SimpleCell | null = null;
  SimpleCells: SimpleCell[] = [];
  Notes: Note[] = [];

  OnStep(context: StepTickContext): void {
    if (!this.First) {
      return;
    }
    const notes = this.Notes.length > 0
      ? this.Notes
      : this.SimpleCells.map((cell) => cell.Note).filter(
          (note): note is Note => note !== null,
        );
    for (const note of notes) {
      this.First.DispatchNote(context, note);
    }
  }
}

export class CellManager {
  // 立方体中的每个 {bank, step, track} 组合只创建一个 SimpleCell；
  // 不同 Bank/Step/Track 通过 owner 引用共享这个对象。
  private readonly table = new Map<string, SimpleCell>();

  OpenCell(want: Cell | null, autoMake = false): SimpleCell {
    if (!want) {
      throw new Error("cell is nil");
    }
    const key = this.keyForCell(want);
    let cell = this.table.get(key);
    if (!cell) {
      if (!autoMake) {
        throw new Error(
          `no item:[cell bank:${want.BankIndex} step:${want.StepIndex} track:${want.TrackIndex}] in manager`,
        );
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

  KeyForCell(cell: Cell | null): string {
    return this.keyForCell(cell);
  }

  private keyForCell(cell: Cell | null): string {
    return cell
      ? `b${cell.BankIndex}_s${cell.StepIndex}_t${cell.TrackIndex}`
      : "";
  }
}

export class BankArray extends Array<Bank> {
  Capacity(): number {
    return this.length;
  }

  Init(capacity: number): BankArray {
    const count = Math.max(1, Math.trunc(capacity));
    const output = new BankArray();
    for (let index = 0; index < count; index++) {
      const bank = this[index] ?? new Bank();
      bank.Index = index;
      output.push(bank);
    }
    return output;
  }

  GetSlice(length: number): BankSlice {
    return new BankSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
  }

  GetItem(index: BankIndex): Bank | null {
    return this[index] ?? null;
  }
}

export class BankSlice extends Array<Bank> {
  Length(): number {
    return this.length;
  }

  Init(array: BankArray, length: number): BankSlice {
    return new BankSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
  }

  Get(index: number): Bank | null {
    return this[index] ?? null;
  }
}

export class StepArray extends Array<Step> {
  Capacity(): number {
    return this.length;
  }

  Init(capacity: number): StepArray {
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

  GetSlice(length: number): StepSlice {
    return new StepSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
  }

  GetItem(index: StepIndex): Step | null {
    return this[index] ?? null;
  }
}

export class StepSlice extends Array<Step> {
  Length(): number {
    return this.length;
  }

  Init(array: StepArray, length: number): StepSlice {
    return new StepSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
  }

  Get(index: StepIndex): Step | null {
    return this[index] ?? null;
  }
}

export class TrackArray extends Array<Track> {
  Capacity(): number {
    return this.length;
  }

  Init(capacity: number): TrackArray {
    const count = Math.max(1, Math.trunc(capacity));
    const output = new TrackArray();
    for (let index = 0; index < count; index++) {
      const track = this[index] ?? new Track();
      track.Index = index;
      output.push(track);
    }
    return output;
  }

  GetSlice(length: number): TrackSlice {
    return new TrackSlice(...this.slice(0, Math.max(1, Math.trunc(length))));
  }

  GetItem(index: TrackIndex): Track | null {
    return this[index] ?? null;
  }
}

export class TrackSlice extends Array<Track> {
  Length(): number {
    return this.length;
  }

  Init(array: TrackArray, length: number): TrackSlice {
    return new TrackSlice(...array.slice(0, Math.max(1, Math.trunc(length))));
  }

  Get(index: number): Track | null {
    return this[index] ?? null;
  }

  Filter(predicate: (track: Track) => boolean): TrackSlice {
    return new TrackSlice(...Array.prototype.filter.call(this, predicate));
  }
}

export class StepPosition extends RelativePosition {
  Index: StepIndex = 0;
  Location: StepLocation = 0;
}

export class Step {
  Context: Cube | null = null;
  Index: StepIndex = 0;
  Weight = 60;
  readonly Cells = new CellCache();

  OnStep(context: StepTickContext): void {
    if (!context) {
      throw new Error("StepTickContext is nil");
    }
    for (const cell of this.Cells.List) {
      cell.OnStep(context);
    }
  }
}

export class TrackMeta {
  Index: TrackIndex = 0;
  Name: TrackName = "";
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
  Instrument: Instrument | null = null;
}

export class Track {
  Context: Cube | null = null;
  Meta: TrackMeta | null = null;
  Index: TrackIndex = 0;
  readonly Cells = new CellCache();
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
  Context: SequencerContext | null = null;
  readonly CellManager = new CellManager();
  readonly Buffers = new CubeBuffers();
  Banks = new BankSlice();
  Tracks = new TrackSlice();
  Steps = new StepSlice();

  Init(
    context: SequencerContext | null,
    configuration?: Configuration | null,
  ): void {
    const cfg =
      configuration ??
      context?.EngineContext?.Configuration ??
      null;
    if (!context && !cfg) {
      throw new Error("parameter(s) are nil");
    }
    if (!cfg) {
      throw new Error("engine.configuration is nil");
    }

    this.Buffers.BankBuffer = this.Buffers.BankBuffer.Init(
      cfg.BankListCapacity,
    );
    this.Buffers.StepBuffer = this.Buffers.StepBuffer.Init(
      cfg.StepListCapacity,
    );
    this.Buffers.TrackBuffer = this.Buffers.TrackBuffer.Init(
      cfg.TrackListCapacity,
    );

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

  Reload(bankIndex: BankIndex, all: boolean): void {
    if (all) {
      for (const bank of this.Buffers.BankBuffer) {
        bank?.Reload();
      }
      return;
    }
    this.Buffers.BankBuffer.GetItem(bankIndex)?.Reload();
  }

  private initBank(bank: Bank): void {
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
        const cell = this.CellManager.OpenCell(
          new Cell({
            BankIndex: bank.Index,
            StepIndex: stepIndex,
            TrackIndex: trackIndex,
          }),
          true,
        );
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

export class Bank implements StepTickHandler {
  Context: Cube | null = null;
  Index: BankIndex = 0;
  Name: BankName = "";
  Label = "";
  Description = "";
  StepBuffer = new StepArray();
  TrackBuffer = new TrackArray();
  Steps = new StepSlice();
  Tracks = new TrackSlice();

  OnStep(context: StepTickContext): void {
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

  Init(cube: Cube | null): void {
    if (!cube) {
      throw new Error("bank.init() : param is nil");
    }
    this.Context = cube;
    this.StepBuffer = this.StepBuffer.Init(cube.Buffers.StepBuffer.Capacity());
    this.TrackBuffer = this.TrackBuffer.Init(
      cube.Buffers.TrackBuffer.Capacity(),
    );
    this.Steps = this.StepBuffer.GetSlice(cube.Steps.Length());
    this.Tracks = this.TrackBuffer.GetSlice(cube.Tracks.Length());
  }

  Reload(): void {
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
  Profile: MusicProfile | null = null;
  SecondsPerStep: Seconds = 0;
  SecondsPerBeat: Seconds = 0;
  StepsPerBeat = 1;
  StepsPerCycle = 16;
  PreloadSeconds: Seconds = 1;

  Normalize(rewrite = false): SequencerInfo {
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
  Index: StepIndex = 0;
  Location: StepLocation = 0;

  ComputeIndex(info: SequencerInfo | null, rewrite = false): StepIndex {
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
  Context: SequencerContext | null = null;
  WorkTimeline: Timeline | null = null;
  BreakCount = 0;
  Broken = false;
  Info = new SequencerInfo();
  RTC: RTC | null = null;
  Front = new StepTickPos();
  Back = new StepTickPos();
  This = new StepTickPos();
}

export interface Sequencer extends StepTickHandler {
  Update(context: SequencerContext): void;
  GetContext(): SequencerContext;
  GetBanks(): Bank[];
  GetSteps(): Step[];
  GetTracks(): Track[];
}

export interface SequencerFactory {
  CreateSequencer(context: EngineContext): Sequencer;
}

export class SequencerContext {
  // 音序器运行时状态。CurrentBankHolder 负责在 CurrentBank 为空时
  // 自动选择并缓存当前 bank。
  Engine: Engine | null = null;
  EngineContext: EngineContext | null = null;
  Sequencer: Sequencer | null = null;
  CurrentBank: Bank | null = null;
  CurrentBankIndex: BankIndex = 0;
  CurrentBankHolder: BankHolder | null = null;
  Cube = new Cube();
  Info = new SequencerInfo();
  STC: StepTickContext | null = null;
  Updating = false;

  NormalizeSTP(position: StepTickPos, align: boolean): void {
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

  ComputeStepLocation(position: RelativePosition | null): StepLocation {
    // 先把任意时间线的位置转成绝对时间，再换算到 WorkTimeline 的 step 序号。
    if (!position) {
      return 0;
    }
    const workTimeline = this.EngineContext?.WorkTileline;
    const relative = AbsolutePositionRelative(
      position.Absolute(),
      workTimeline,
    );
    const secondsPerStep = Math.max(0.001, this.Info.SecondsPerStep);
    return Math.max(0, Math.trunc(relative.Pos / secondsPerStep));
  }

  ComputeStepIndex(position: RelativePosition | null): StepIndex {
    const location = this.ComputeStepLocation(position);
    const stepsPerCycle = this.Info.StepsPerCycle;
    return stepsPerCycle < 1 ? 0 : location % stepsPerCycle;
  }

  NewBank(index: BankIndex): Bank {
    const bank = new Bank();
    bank.Init(this.Cube);
    bank.Index = index;
    return bank;
  }

  NewStep(index: StepIndex): Step {
    const step = new Step();
    step.Index = index;
    step.Context = this.Cube;
    return step;
  }

  NewTrack(index: TrackIndex): Track {
    const track = new Track();
    track.Index = index;
    track.Context = this.Cube;
    return track;
  }
}

////////////////////////////////////////////////////////////////////////////////
// Clock and real-time positions

export class ClockEvent {
  Clock: Clock | null = null;
  Time: AbsolutePosition = 0;
  Context: EngineContext | null = null;
}

export interface ClockListener {
  OnClock(event: ClockEvent): void;
}

export interface Clock {
  AddListener(listener: ClockListener | null): void;
  RemoveListener(listener: ClockListener | null): void;
  Dispatch(event: ClockEvent | null): void;
}

export class RTC {
  Precision = new StepTickPos();
  Normal = new StepTickPos();
  Raw = new StepTickPos();
}

export class SimpleClock implements Clock {
  // Clock 只负责读取 AudioContext 时间、刷新 RTC，然后通知监听器。
  // 音序器和监控组件因此看到同一个“当前时间”快照。
  private listeners: ClockListener[] = [];

  AddListener(listener: ClockListener | null): void {
    if (listener && !this.listeners.includes(listener)) {
      this.listeners.push(listener);
    }
  }

  RemoveListener(listener: ClockListener | null): void {
    if (!listener) {
      return;
    }
    this.listeners = this.listeners.filter((item) => item !== listener);
  }

  Dispatch(event: ClockEvent | null): void {
    if (!event) {
      throw new Error("lite: nil clock event");
    }
    this.updateRTC(event);
    event.Clock = this;
    const errors: unknown[] = [];
    for (const listener of [...this.listeners]) {
      try {
        listener.OnClock(event);
      } catch (error) {
        errors.push(error);
      }
    }
    if (errors.length === 1) {
      throw errors[0];
    }
    if (errors.length > 1) {
      throw new Error(
        errors
          .map((error) =>
            error instanceof Error ? error.message : String(error),
          )
          .join("; "),
        { cause: errors },
      );
    }
  }

  private updateRTC(event: ClockEvent): void {
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
    normal.Position = new RelativePosition(
      workPosition.Ref,
      secondsPerStep * stepCount,
    );

    event.Time = now;
    context.RTC.Raw = raw;
    context.RTC.Precision = precision;
    context.RTC.Normal = normal;
  }
}

////////////////////////////////////////////////////////////////////////////////
// Sequencer implementation and trigger

class CurrentBankHolder implements BankHolder {
  constructor(private readonly context: SequencerContext) {}

  GetBank(): Bank | null {
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
  // profile、速度或 bank 配置变化后，由 Updating 标记触发一次重算。
  constructor(private readonly context: SequencerContext) {}

  Update(): void {
    this.context.Updating = false;
    Log().Info("innerSequencer.Update()");
    this.computeSpeed();
    this.loadCurrentBank();
  }

  private computeSpeed(): void {
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
      throw new Error(
        `bad value of steps/beat (value=${stepsPerBeat})`,
      );
    }
    const secondsPerBeat = 60 / bpm;
    context.Info.SecondsPerBeat = secondsPerBeat;
    context.Info.SecondsPerStep = secondsPerBeat / stepsPerBeat;
  }

  private loadCurrentBank(): void {
    const list = this.context.Cube.Banks;
    const index = this.context.CurrentBankIndex;
    if (index >= list.Length()) {
      throw new Error(
        `innerSequencerUpdater: bank index is out of array [index:${index} size:${list.Length()}]`,
      );
    }
    const bank = list[index];
    if (!bank) {
      throw new Error(`innerSequencerUpdater: bank[${index}] is nil`);
    }
    this.context.CurrentBank = bank;
  }
}

export class InnerSequencer implements Sequencer {
  constructor(private readonly context: SequencerContext) {}

  GetBanks(): Bank[] {
    return [...this.context.Cube.Banks];
  }

  GetContext(): SequencerContext {
    return this.context;
  }

  GetSteps(): Step[] {
    return [...this.context.Cube.Steps];
  }

  GetTracks(): Track[] {
    return [...this.context.Cube.Tracks];
  }

  OnStep(context: StepTickContext): void {
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

  Update(context: SequencerContext): void {
    if (context.Updating) {
      new SequencerUpdater(this.context).Update();
    }
  }
}

export class SequencerTrigger implements ClockListener {
  // 每次时钟 tick 都从已处理的 Back 推进到预加载前沿 Front。
  // 仅当 Transport.Walking() 为 true 时实际派发 step。
  constructor(private readonly context: SequencerContext) {}

  OnClock(_event: ClockEvent): void {
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

  private isWalking(): boolean {
    return this.context.EngineContext?.TransportContext.Walking() ?? false;
  }

  private updateCurrentPosition(): void {
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

  private prepare(source: StepTickContext | null): StepTickContext {
    const stc = source ?? new StepTickContext();
    stc.Context = this.context;
    stc.WorkTimeline = this.context.EngineContext?.WorkTileline ?? null;
    return stc;
  }

  private checkParameters(stc: StepTickContext): void {
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
      throw new Error(
        `bad stc.Info.StepsPerCycle value: ${stc.Info.StepsPerCycle}`,
      );
    }
    if (stc.Info.PreloadSeconds < 1) {
      throw new Error(
        `bad stc.Info.PreloadSeconds value: ${stc.Info.PreloadSeconds}`,
      );
    }
    if (stc.Info.SecondsPerStep < 0.001) {
      throw new Error(
        `bad stc.Info.SecondsPerStep value: ${stc.Info.SecondsPerStep}`,
      );
    }
    stc.Front.Position.Ref = stc.WorkTimeline;
    stc.This.Position.Ref = stc.WorkTimeline;
    stc.Back.Position.Ref = stc.WorkTimeline;
  }

  private computeAheadPosition(stc: StepTickContext): void {
    const context = stc.Context!;
    const engine = context.EngineContext!;
    const secondsPerStep = Math.max(0.001, stc.Info.SecondsPerStep);
    const preloadSteps = Math.max(
      1,
      Math.trunc(stc.Info.PreloadSeconds / secondsPerStep),
    );
    const front = cloneTickPos(engine.RTC.Normal);
    front.Location += preloadSteps;
    front.Position.Pos += secondsPerStep * preloadSteps;
    front.ComputeIndex(stc.Info, true);
    stc.Front = front;
  }

  private fastForward(stc: StepTickContext): void {
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
      } catch (error) {
        handleError(error);
      }
      current.Position.Pos += secondsPerStep;
      current.Location++;
      current.ComputeIndex(stc.Info, true);
    }
    stc.Back = cloneTickPos(target);
  }
}

function contextSequencer(stc: StepTickContext): Sequencer {
  const sequencer = stc.Context?.Sequencer;
  if (!sequencer) {
    throw new Error("stc.context.Sequencer is nil");
  }
  return sequencer;
}

function cloneTickPos(source: StepTickPos): StepTickPos {
  const destination = new StepTickPos();
  destination.Position = source.Position.Clone();
  destination.Index = source.Index;
  destination.Location = source.Location;
  return destination;
}

export class InnerSequencerFactory implements SequencerFactory {
  CreateSequencer(engineContext: EngineContext): Sequencer {
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

  constructor(init: Partial<MonitorConfig> = {}) {
    Object.assign(this, init);
  }

  Prepare(): this {
    if (this.URL) {
      try {
        const url = new URL(this.URL);
        this.Scheme = url.protocol.replace(/:$/, "");
        this.Host = url.hostname;
        this.Port = Number(url.port) || 0;
      } catch {
        // Keep the raw URL when the runtime URL parser rejects it.
      }
    } else {
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

export type PartSource = Part | PartFactory | PartRegistration;

export class Configuration {
  // Driver 提供底层音频上下文，InstrumentFactory 提供默认乐器。
  // Parts 可以混合 Part、PartFactory 和 PartRegistration。
  Driver: Driver | null = null;
  InstrumentFactory: InstrumentFactory | null = null;
  SampleRate = 0;
  Profile: MusicProfile | null = null;
  BankListSize = 0;
  StepListSize = 0;
  TrackListSize = 0;
  BankListCapacity = 0;
  StepListCapacity = 0;
  TrackListCapacity = 0;
  Parts: PartSource[] = [];
  Monitor: MonitorConfig | null = null;
}

export function DefaultConfig(): Configuration {
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

  String(): string {
    return `engine.[state starting:${this.Starting} started:${this.Started} stopping:${this.Stopping} stopped:${this.Stopped}]`;
  }

  toString(): string {
    return this.String();
  }
}

export interface Transport {
  Play(): void;
  Record(): void;
  Stop(): void;
  Pause(): void;
  Resume(): void;
  Paused(): boolean;
  Recording(): boolean;
  Playing(): boolean;
  Walking(): boolean;
}

export class TransportContext {
  // StopAt 表示暂停/停止后的目标位置；Resume/Play 会据此重定位工作时间线。
  private paused = false;
  private playing = false;
  private recording = false;
  Current = new StepTickPos();
  StopAt: RelativePosition | null = null;

  Walking(): boolean {
    return this.playing && !this.paused;
  }

  IsPaused(): boolean {
    return this.paused;
  }

  IsPlaying(): boolean {
    return this.playing;
  }

  IsRecording(): boolean {
    return this.recording;
  }

  internalSetPaused(value: boolean): void {
    this.paused = value;
  }

  internalSetPlaying(value: boolean): void {
    this.playing = value;
  }

  internalSetRecording(value: boolean): void {
    this.recording = value;
  }
}

export class EngineContext {
  // 所有子系统共享的依赖容器；Context 是对应的 Go 风格类型别名。
  Transport!: Transport;
  LifeManager!: LifeManager;
  PartManager!: PartManager;
  PartRegistry!: PartRegistry;
  Logger: Logger = Log();
  Engine!: Engine;
  Clock!: Clock;
  EventDispatcher!: EventDispatcher;
  EventRunner!: EventRunner;
  TransportContext = new TransportContext();
  SequencerContext!: SequencerContext;
  Configuration!: Configuration;
  AudioContext!: AudioContext;
  AudioContextFactory!: AudioContextFactory;
  CurrentInstrument: Instrument | null = null;
  CurrentInstrumentHolder!: InstrumentHolder;
  MainInstrumentFactory!: InstrumentFactory;
  InstrumentManager!: InstrumentManager;
  InputInstrument: Instrument | null = null;
  InputInstrumentHolder!: InstrumentHolder;
  Timer!: Timer;
  TimerHandler: ((timer: Timer) => void) | null = null;
  RootTileline!: Timeline;
  WorkTileline!: Timeline;
  RTC = new RTC();
  State = new EngineState();

  GetPosAtWork(source: RelativePosition | null): RelativePosition {
    // 已经是 WorkTimeline 相对位置时直接返回，避免二次扣减 offset。
    const position = source ?? new RelativePosition();
    if (!position.Ref || !this.WorkTileline) {
      return position;
    }
    if (position.Ref.IsTypeOf(TimelineType.Work)) {
      return position;
    }
    return AbsolutePositionRelative(
      position.Absolute(),
      this.WorkTileline,
    );
  }

  GetCurrentPosition(destination?: RelativePosition): RelativePosition {
    const dst = destination ?? new RelativePosition();
    const position =
      this.TransportContext.StopAt ?? this.RTC.Precision.Position;
    dst.Ref = position.Ref;
    dst.Pos = position.Pos;
    return dst;
  }

  RebaseToPosition(to: RelativePosition | null): void {
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

export type Context = EngineContext;

export interface Engine {
  GetContext(): EngineContext;
  Create(): void;
  Start(): void;
  Join(): Promise<void>;
  Stop(): void;
  Destroy(): void;
}

export interface EngineFactory {
  CreateEngine(config: Configuration): Engine;
}

function isPart(value: unknown): value is Part {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as Part).GetInfo === "function"
  );
}

function isPartFactory(value: unknown): value is PartFactory {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as PartFactory).CreatePart === "function" &&
    typeof (value as PartFactory).GetRegistration === "function"
  );
}

function isLife(value: unknown): value is Life {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as Life).GetLifeCycle === "function"
  );
}

export class InnerEngineFactory implements EngineFactory {
  CreateEngine(config: Configuration): Engine {
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
    context.EventDispatcher = eventContext.GetDispatcher()!;
    context.EventRunner = eventContext.GetRunner()!;

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

  private initParts(context: EngineContext): void {
    // 通过运行时能力判断 Part 角色，同一对象可以同时实现多种角色。
    for (const source of context.Configuration.Parts) {
      if (source instanceof PartRegistration) {
        context.PartRegistry.Register(source);
      } else if (isPartFactory(source)) {
        context.PartRegistry.RegisterFactory(source);
      } else if (isPart(source)) {
        context.PartRegistry.RegisterPart(source);
      }
    }
  }

  private initLifeCycles(context: EngineContext): void {
    for (const name of context.PartManager.ListNames()) {
      const part = context.PartManager.FindPartByName(name);
      if (isLife(part)) {
        context.LifeManager.Add(part);
      }
    }
  }

  private initInstruments(context: EngineContext): void {
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

export function DefaultEngineFactory(): EngineFactory {
  return new InnerEngineFactory();
}

export class InnerEngine implements Engine {
  // Join 不阻塞 JS 线程，而是返回 Promise，由生命周期运行器等待。
  Context!: EngineContext;

  GetContext(): EngineContext {
    return this.Context;
  }

  Join(): Promise<void> {
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

  Create(): void {
    // 主 Timer 约每 10ms 产生一次时钟事件，实际精度取决于宿主调度器。
    const context = this.Context;
    const timer = context.Timer ?? new Timer();
    context.Timer = timer;
    timer.Delay = 0.01;
    timer.Interval = 0.01;
    timer.Handler = context.TimerHandler;
  }

  Destroy(): void {
    this.Context.Timer = new Timer();
    this.Context.State.Stopped = true;
  }

  Start(): void {
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

  Stop(): void {
    this.Context.State.Stopping = true;
    this.Context.Timer.Stop();
  }
}

export class InnerTransport implements Transport {
  // Transport 只维护播放状态和目标位置，实际时间推进由 Clock 完成。
  Context!: EngineContext;
  TransportContext!: TransportContext;

  Walking(): boolean {
    return this.TransportContext.Walking();
  }

  Pause(): void {
    // 仅在真正 Walking 时记录位置；重复 Pause 不会覆盖 StopAt。
    if (this.TransportContext.Walking()) {
      this.TransportContext.StopAt =
        this.Context.RTC.Precision.Position.Clone();
      this.TransportContext.internalSetPaused(true);
    }
  }

  Paused(): boolean {
    return this.TransportContext.IsPaused();
  }

  Play(): void {
    this.TransportContext.internalSetPlaying(true);
    this.tryRebaseToStopAt();
  }

  Playing(): boolean {
    return this.TransportContext.IsPlaying();
  }

  Record(): void {
    this.TransportContext.internalSetRecording(true);
    this.TransportContext.internalSetPlaying(true);
    this.tryRebaseToStopAt();
  }

  Recording(): boolean {
    return this.TransportContext.IsRecording();
  }

  Resume(): void {
    this.TransportContext.internalSetPaused(false);
    this.tryRebaseToStopAt();
  }

  Stop(): void {
    const context = this.Context;
    this.TransportContext.internalSetRecording(false);
    this.TransportContext.internalSetPlaying(false);
    this.TransportContext.internalSetPaused(false);
    this.TransportContext.StopAt = new RelativePosition(
      context.WorkTileline,
      0,
    );
  }

  ontimer(_timer: Timer): void {
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

  private tryRebaseToStopAt(): void {
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

export class AutoEngineController implements Part, Life {
  // 把 Engine 的 Create/Start/Join/Stop/Destroy 接入通用生命周期。
  // OnRun 等待 Join 完成，确保后续停止阶段不会提前执行。
  private context: EngineContext | null = null;

  GetInfo(): PartInfo {
    return new PartInfo({ Name: "AutoEngineController", Part: this });
  }

  GetLifeCycle(): LifeCycle {
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
  async Run(config: Configuration | null = null): Promise<void> {
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

export function Run(config: Configuration | null = null): Promise<void> {
  return new Boot().Run(config);
}

////////////////////////////////////////////////////////////////////////////////
// Controllers and monitoring

export interface Controller {
  Start(): void;
}

export interface ActionController extends Controller {
  Execute(event: ActionEvent): void;
}

type ActionHandler = (
  context: EngineContext,
  event: ActionEvent,
) => void;

export class InnerActionController implements ActionController {
  // 将控制动作映射为 Transport 或 MIDI 操作。Pause/Resume 共用一个
  // 切换逻辑：Walking 时暂停，否则恢复。
  private readonly handlers = new Map<Action, ActionHandler>();

  constructor(private readonly context: EngineContext) {
    this.handlers.set(Action.Min, () => undefined);
    this.handlers.set(Action.Play, (context) => context.Transport.Play());
    this.handlers.set(Action.Stop, (context) => context.Transport.Stop());
    this.handlers.set(Action.Record, (context) => context.Transport.Record());
    this.handlers.set(Action.Pause, (context) => {
      if (context.Transport.Walking()) {
        context.Transport.Pause();
      } else {
        context.Transport.Resume();
      }
    });
    this.handlers.set(Action.Resume, (context) => {
      if (context.Transport.Walking()) {
        context.Transport.Pause();
      } else {
        context.Transport.Resume();
      }
    });
    this.handlers.set(Action.NoteOn, (context, event) => {
      const instrument =
        context.InputInstrumentHolder.GetInstrument(context);
      instrument?.Dispatch(
        new MidiMessage({
          NoteOn: new MidiNote({
            Key: event.NoteNumber ?? 0,
            Velocity: event.NoteVelocity ?? 0,
            EventType: MidiEventType.NoteOn,
          }),
        }),
      );
    });
    this.handlers.set(Action.NoteOff, (context, event) => {
      const instrument =
        context.InputInstrumentHolder.GetInstrument(context);
      instrument?.Dispatch(
        new MidiMessage({
          NoteOff: new MidiNote({
            Key: event.NoteNumber ?? 0,
            Velocity: event.NoteVelocity ?? 0,
            EventType: MidiEventType.NoteOff,
          }),
        }),
      );
    });
  }

  Start(): void {
    // The action controller is driven by its caller.
  }

  Execute(event: ActionEvent): void {
    const handler = this.handlers.get(event.Action);
    if (!handler) {
      throw new Error(`no handler with action [${event.Action}]`);
    }
    handler(this.context, event);
  }
}

export function NewActionController(context: EngineContext): ActionController {
  return new InnerActionController(context);
}

export class Monitor {
  // 低频状态监控器，只负责日志输出，不参与音频调度。
  private context: EngineContext | null = null;
  private timer: Timer | null = null;
  InputKeyCount = 0;
  InputKeyCode = 0;

  Init(context: EngineContext): void {
    this.context = context;
  }

  Start(): void {
    if (!this.context) {
      throw new Error("monitor is not initialized");
    }
    this.timer = new Timer();
    this.timer.Delay = 1;
    this.timer.Interval = 1 / 24;
    this.timer.Handler = () => this.ontimer();
    this.timer.Start();
  }

  Stop(): void {
    this.timer?.Stop();
    this.timer = null;
  }

  private ontimer(): void {
    const context = this.context;
    if (!context) {
      return;
    }
    const logger = context.Logger;
    logger.Info(
      "--------------------------------------------------------------",
    );
    logger.Info("%s", context.State.String());
    const transport = context.TransportContext;
    logger.Info(
      "transport.[state playing:%s recording:%s paused:%s]",
      transport.IsPlaying(),
      transport.IsRecording(),
      transport.IsPaused(),
    );
    logger.Info(
      "transport.[position raw:%f root:%f work:%f]",
      context.AudioContext.GetCurrentTime(),
      context.RTC.Raw.Position.Pos,
      context.RTC.Precision.Position.Pos,
    );

    const stepContext = context.SequencerContext.STC;
    if (stepContext) {
      logger.Info(
        "[STC(StepTickContext) location:%d index:%d pos:%f ]",
        stepContext.Front.Location,
        stepContext.Front.Index,
        stepContext.Front.Position.Pos,
      );
    } else {
      logger.Info("STC is nil");
    }
    logger.Info(
      "input.[key count:%d code:%d char:'%s']",
      this.InputKeyCount,
      this.InputKeyCode,
      String.fromCharCode(this.InputKeyCode),
    );

    if (context.State.Stopped || context.State.Stopping) {
      this.Stop();
    }
  }
}
