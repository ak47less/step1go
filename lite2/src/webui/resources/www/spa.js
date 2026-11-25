    (() => {
      "use strict";

      const API_ROOT = "/api/v1";
      const STATUS_KEYS = [
        "engine.started",
        "transport.playing",
        "transport.recording",
        "transport.paused",
        "audio.state"
      ];
      const COMMANDS = [
        ["engine-start", "启动引擎"],
        ["engine-restart", "重启引擎"],
        ["engine-stop", "停止引擎"],
        ["play", "播放"],
        ["record", "录制"],
        ["stop", "停止走带"]
      ];
      const MIDI_NOTE_NAMES = [
        "C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"
      ];
      const MIDI_BLACK_PITCH_CLASSES = new Set([1, 3, 6, 8, 10]);
      const MIDI_PC_KEYS = Object.freeze({
        KeyA: 0,
        KeyW: 1,
        KeyS: 2,
        KeyE: 3,
        KeyD: 4,
        KeyF: 5,
        KeyT: 6,
        KeyG: 7,
        KeyY: 8,
        KeyH: 9,
        KeyU: 10,
        KeyJ: 11,
        KeyK: 12
      });

      const state = {
        health: null,
        status: null,
        engine: null,
        transport: null,
        cube: null,
        banks: [],
        steps: [],
        tracks: [],
        cells: [],
        selectedBank: 0,
        selectedTrack: 0,
        selectedCell: null,
        cellSaving: false,
        busy: false,
        loading: false,
        midiOctave: 3,
        midiChannel: 0,
        midiVelocity: 100
      };

      const activeMidiSources = new Map();
      const activeMidiNoteCounts = new Map();
      let midiRequestChain = Promise.resolve();
      let midiLastErrorAt = 0;

      const $ = (selector, root = document) => root.querySelector(selector);
      const $$ = (selector, root = document) => Array.from(root.querySelectorAll(selector));

      const els = {
        connection: $("#connection"),
        connectionLabel: $("#connection-label"),
        refreshButton: $("#refresh-button"),
        errorBanner: $("#error-banner"),
        errorMessage: $("#error-message"),
        dismissError: $("#dismiss-error"),
        transportState: $("#transport-state"),
        positionDisplay: $("#position-display"),
        timelineLabel: $("#timeline-label"),
        positionSeconds: $("#position-seconds"),
        seekForm: $("#seek-form"),
        seekPosition: $("#seek-position"),
        timelineSelect: $("#timeline-select"),
        pauseButton: $("#pause-button"),
        engineState: $("#engine-state"),
        audioState: $("#audio-state"),
        bankTabs: $("#bank-tabs"),
        sequenceGrid: $("#sequence-grid"),
        sequenceMeta: $("#sequence-meta"),
        trackNumber: $("#track-number"),
        trackHeading: $("#track-heading"),
        trackSubheading: $("#track-subheading"),
        trackForm: $("#track-form"),
        saveTrack: $("#save-track"),
        bankForm: $("#bank-form"),
        bankLabel: $("#bank-label"),
        bankDescription: $("#bank-description"),
        saveBank: $("#save-bank"),
        cellReadout: $("#cell-readout"),
        cellForm: $("#cell-form"),
        noteNumber: $("#note-number"),
        noteVelocity: $("#note-velocity"),
        saveCellNote: $("#save-cell-note"),
        clearCellNote: $("#clear-cell-note"),
        midiKeyboard: $("#midi-keyboard"),
        midiOctave: $("#midi-octave"),
        midiChannel: $("#midi-channel"),
        midiVelocity: $("#midi-velocity"),
        midiVelocityValue: $("#midi-velocity-value"),
        midiActivity: $("#midi-activity"),
        actionEditor: $("#action-editor"),
        actionResponse: $("#action-response"),
        executeActions: $("#execute-actions"),
        toastStack: $("#toast-stack"),
        lastUpdated: $("#last-updated")
      };

      function escapeHTML(value) {
        return String(value ?? "")
          .replaceAll("&", "&amp;")
          .replaceAll("<", "&lt;")
          .replaceAll(">", "&gt;")
          .replaceAll('"', "&quot;")
          .replaceAll("'", "&#039;");
      }

      function pad(value, width = 2) {
        return String(value).padStart(width, "0");
      }

      function formatTime(seconds) {
        const total = Math.max(0, Number(seconds) || 0);
        const whole = Math.floor(total * 1000);
        const hours = Math.floor(whole / 3600000);
        const minutes = Math.floor((whole % 3600000) / 60000);
        const secs = Math.floor((whole % 60000) / 1000);
        const millis = whole % 1000;
        if (hours > 0) {
          return `${hours}:${pad(minutes)}:${pad(secs)}.${pad(millis, 3)}`;
        }
        return `${pad(minutes)}:${pad(secs)}.${pad(millis, 3)}`;
      }

      function formatNumber(value, digits = 3) {
        const number = Number(value);
        return Number.isFinite(number) ? number.toFixed(digits) : "0.000";
      }

      const MIDI_PC_SHORTCUTS = Object.freeze(Object.fromEntries(
        Object.entries(MIDI_PC_KEYS).map(([code, semitone]) => [semitone, code.slice(3)])
      ));

      function midiNoteName(note) {
        const pitch = ((note % 12) + 12) % 12;
        return `${MIDI_NOTE_NAMES[pitch]}${Math.floor(note / 12) - 1}`;
      }

      function setMidiActivity(message, mode = "idle") {
        els.midiActivity.textContent = message;
        els.midiActivity.dataset.state = mode;
      }

      function sendMidiEvent(type, note, channel) {
        const velocity = type === "note_on" ? state.midiVelocity : 0;
        const label = `${type === "note_on" ? "Note On" : "Note Off"} · ${midiNoteName(note)}`;
        const payload = {
          events: [{
            type,
            time: -1,
            channel,
            note: {
              number: note,
              velocity
            }
          }]
        };

        setMidiActivity(`${label} · 排队`, "pending");
        midiRequestChain = midiRequestChain
          .then(async () => {
            await api("/midi/events", { method: "POST", body: payload });
            setMidiActivity(`${label} · CH ${channel}`, "active");
          })
          .catch((error) => {
            setMidiActivity(`${label} 失败`, "error");
            const now = Date.now();
            if (now - midiLastErrorAt > 3000) {
              midiLastErrorAt = now;
              showError(`MIDI 发送失败：${describeError(error)}`);
              toast(describeError(error), "error");
            }
          });
      }

      function renderMidiChannels() {
        els.midiChannel.innerHTML = Array.from({ length: 16 }, (_, channel) => (
          `<option value="${channel}">CH ${channel + 1}</option>`
        )).join("");
        els.midiChannel.value = String(state.midiChannel);
      }

      function renderMidiKeyboard() {
        releaseAllMidiNotes();
        const baseNote = (state.midiOctave + 1) * 12;
        const offsets = Array.from({ length: 25 }, (_, offset) => offset);
        const whiteTotal = offsets.filter((offset) => (
          !MIDI_BLACK_PITCH_CLASSES.has((baseNote + offset) % 12)
        )).length;
        const bars = [];
        let whiteSeen = 0;

        for (const offset of offsets) {
          const note = baseNote + offset;
          const pitch = note % 12;
          const black = MIDI_BLACK_PITCH_CLASSES.has(pitch);
          const shortcut = offset <= 12 ? MIDI_PC_SHORTCUTS[offset] || "" : "";
          const shortcutMarkup = shortcut
            ? `<span class="piano-key-shortcut">${escapeHTML(shortcut)}</span>`
            : "";
          const caption = `<span class="piano-key-caption">${escapeHTML(midiNoteName(note))}</span>`;
          const classes = [
            "piano-key",
            black ? "piano-black-key" : "piano-white-key"
          ];

          if (black) {
            const unit = 100 / whiteTotal;
            const width = unit * 0.62;
            const left = whiteSeen * unit - width / 2;
            bars.push(`
              <button class="${classes.join(" ")}" type="button" tabindex="-1"
                data-midi-note="${note}"
                aria-label="${escapeHTML(midiNoteName(note))}，MIDI ${note}"
                style="left:${left.toFixed(3)}%;width:${width.toFixed(3)}%">
                ${shortcutMarkup}${caption}
              </button>
            `);
            continue;
          }

          whiteSeen += 1;
          bars.push(`
            <button class="${classes.join(" ")}" type="button" tabindex="-1"
              data-midi-note="${note}"
              aria-label="${escapeHTML(midiNoteName(note))}，MIDI ${note}">
              ${shortcutMarkup}${caption}
            </button>
          `);
        }
        els.midiKeyboard.innerHTML = bars.join("");
        renderMidiActiveNotes();
      }

      function renderMidiActiveNotes() {
        const activeNotes = new Set(
          Array.from(activeMidiSources.values(), (source) => source.note)
        );
        $$("[data-midi-note]", els.midiKeyboard).forEach((key) => {
          key.classList.toggle("is-active", activeNotes.has(Number(key.dataset.midiNote)));
        });
      }

      function startMidiNote(note, source) {
        if (!Number.isInteger(note) || note < 0 || note > 127 || activeMidiSources.has(source)) {
          return;
        }
        const channel = state.midiChannel;
        const countKey = `${channel}:${note}`;
        const count = activeMidiNoteCounts.get(countKey) || 0;
        activeMidiSources.set(source, { note, channel });
        activeMidiNoteCounts.set(countKey, count + 1);
        if (count === 0) sendMidiEvent("note_on", note, channel);
        renderMidiActiveNotes();
      }

      function stopMidiNote(source) {
        const active = activeMidiSources.get(source);
        if (!active) return;
        activeMidiSources.delete(source);
        const countKey = `${active.channel}:${active.note}`;
        const count = Math.max(0, (activeMidiNoteCounts.get(countKey) || 0) - 1);
        if (count === 0) {
          activeMidiNoteCounts.delete(countKey);
          sendMidiEvent("note_off", active.note, active.channel);
        } else {
          activeMidiNoteCounts.set(countKey, count);
        }
        renderMidiActiveNotes();
      }

      function releaseAllMidiNotes() {
        for (const source of Array.from(activeMidiSources.keys())) {
          stopMidiNote(source);
        }
      }

      function midiPointerSource(pointerId) {
        return `pointer:${pointerId}`;
      }

      function startMidiPointer(event) {
        if (event.pointerType === "mouse" && event.button !== 0) return;
        const key = event.target.closest?.("[data-midi-note]");
        if (!key || !els.midiKeyboard.contains(key)) return;
        event.preventDefault();
        try {
          els.midiKeyboard.setPointerCapture(event.pointerId);
        } catch {
          // Pointer capture is best-effort; the pointerup handler still releases the note.
        }
        stopMidiNote(midiPointerSource(event.pointerId));
        startMidiNote(Number(key.dataset.midiNote), midiPointerSource(event.pointerId));
      }

      function moveMidiPointer(event) {
        const source = midiPointerSource(event.pointerId);
        const current = activeMidiSources.get(source);
        if (!current) return;

        const hit = document.elementFromPoint(event.clientX, event.clientY);
        const key = hit?.closest?.("[data-midi-note]");
        if (!key || !els.midiKeyboard.contains(key)) {
          stopMidiNote(source);
          return;
        }
        const note = Number(key.dataset.midiNote);
        if (note === current.note) return;
        stopMidiNote(source);
        startMidiNote(note, source);
      }

      function stopMidiPointer(event) {
        stopMidiNote(midiPointerSource(event.pointerId));
      }

      function setConnection(mode, label) {
        els.connection.dataset.state = mode;
        els.connectionLabel.textContent = label;
      }

      function showError(message) {
        els.errorMessage.textContent = message;
        els.errorBanner.hidden = false;
      }

      function clearError() {
        els.errorBanner.hidden = true;
      }

      function toast(message, type = "success") {
        const node = document.createElement("div");
        node.className = `toast${type === "error" ? " is-error" : ""}`;
        node.textContent = message;
        els.toastStack.append(node);
        window.setTimeout(() => node.remove(), 3600);
      }

      function describeError(error) {
        if (!error) return "未知错误";
        if (error.code && error.status) {
          return `${error.code} · HTTP ${error.status} · ${error.message}`;
        }
        return error.message || String(error);
      }

      async function api(path, options = {}) {
        const request = {
          ...options,
          headers: {
            Accept: "application/json",
            ...(options.headers || {})
          }
        };
        if (request.body && typeof request.body !== "string") {
          request.body = JSON.stringify(request.body);
          request.headers["Content-Type"] = "application/json";
        }

        let response;
        try {
          response = await fetch(API_ROOT + path, request);
        } catch (cause) {
          setConnection("offline", "连接失败");
          const error = new Error("无法连接后端服务");
          error.cause = cause;
          throw error;
        }

        setConnection("online", "已连接");
        const text = await response.text();
        let data = null;
        if (text) {
          try {
            data = JSON.parse(text);
          } catch {
            data = { raw: text };
          }
        }

        if (!response.ok) {
          const payload = data?.error || {};
          const error = new Error(payload.message || `请求失败（HTTP ${response.status}）`);
          error.status = response.status;
          error.code = payload.code || "HTTP_ERROR";
          error.details = payload.details;
          error.payload = data;
          throw error;
        }
        return data;
      }

      function settle(name, result, failures) {
        if (result.status === "fulfilled") {
          state[name] = result.value;
          return;
        }
        failures.push(`${name}: ${describeError(result.reason)}`);
      }

      async function loadDashboard(options = {}) {
        if (state.loading) return;
        state.loading = true;
        document.body.classList.add("is-loading");
        els.refreshButton.disabled = true;

        const requests = [
          ["health", api("/health")],
          ["status", api("/status")],
          ["engine", api("/engine")],
          ["transport", api("/transport")],
          ["cube", api("/cube")],
          ["banksResponse", api("/banks")],
          ["stepsResponse", api("/steps")],
          ["tracksResponse", api("/tracks")]
        ];
        const settled = await Promise.allSettled(requests.map(([, promise]) => promise));
        const failures = [];

        settled.forEach((result, index) => {
          const name = requests[index][0];
          if (name === "banksResponse") {
            if (result.status === "fulfilled") {
              state.banks = Array.isArray(result.value?.items) ? result.value.items : [];
            } else {
              failures.push(`banks: ${describeError(result.reason)}`);
            }
            return;
          }
          if (name === "stepsResponse") {
            if (result.status === "fulfilled") {
              state.steps = Array.isArray(result.value?.items) ? result.value.items : [];
            } else {
              failures.push(`steps: ${describeError(result.reason)}`);
            }
            return;
          }
          if (name === "tracksResponse") {
            if (result.status === "fulfilled") {
              state.tracks = Array.isArray(result.value?.items) ? result.value.items : [];
            } else {
              failures.push(`tracks: ${describeError(result.reason)}`);
            }
            return;
          }
          settle(name, result, failures);
        });

        if (state.banks.length > 0) {
          const exists = state.banks.some((bank) => bank.index === state.selectedBank);
          if (!exists) state.selectedBank = state.banks[0].index;
        }
        if (state.tracks.length > 0) {
          const exists = state.tracks.some((track) => track.index === state.selectedTrack);
          if (!exists) state.selectedTrack = state.tracks[0].index;
        }

        await loadCells(failures);
        renderAll();

        if (failures.length > 0) {
          showError(`部分数据读取失败：${failures.slice(0, 3).join("；")}${failures.length > 3 ? `；另有 ${failures.length - 3} 项` : ""}`);
        } else if (!options.quiet) {
          clearError();
        }

        state.loading = false;
        document.body.classList.remove("is-loading");
        els.refreshButton.disabled = false;
        els.lastUpdated.textContent = `同步于 ${new Date().toLocaleTimeString("zh-CN", { hour12: false })}`;
      }

      async function loadCells(failures = []) {
        if (!state.banks.length) {
          state.cells = [];
          return;
        }
        try {
          const data = await api(`/cells?bankIndex=${encodeURIComponent(state.selectedBank)}`);
          state.cells = Array.isArray(data?.items) ? data.items : [];
          if (state.selectedCell) {
            const selectedName = state.selectedCell.name;
            state.selectedCell = state.cells.find((cell) => cell.name === selectedName) || null;
          }
        } catch (error) {
          state.cells = [];
          failures.push(`cells: ${describeError(error)}`);
        }
      }

      function renderAll() {
        renderEngine();
        renderTransport();
        renderStatus();
        renderTimelines();
        renderBanks();
        renderSequence();
        renderTrackEditor();
        renderBankEditor();
        renderCellReadout();
      }

      function renderEngine() {
        const engine = state.engine;
        const engineName = engine?.state ? String(engine.state).toUpperCase() : "UNKNOWN";
        els.engineState.textContent = engineName;
        els.audioState.textContent = engine?.audio?.state
          ? `Audio · ${engine.audio.state}`
          : "音频状态未知";
      }

      function renderTransport() {
        const transport = state.transport;
        const transportState = transport?.state || "unknown";
        const labels = {
          stopped: "已停止",
          playing: "播放中",
          recording: "录制中",
          paused: "已暂停",
          unknown: "未知状态"
        };
        els.transportState.textContent = labels[transportState] || transportState;
        els.transportState.classList.toggle("is-active", ["playing", "recording", "paused"].includes(transportState));
        els.transportState.classList.toggle("is-recording", transportState === "recording");

        const position = Number(transport?.position) || 0;
        els.positionDisplay.textContent = formatTime(position);
        els.positionSeconds.textContent = `${formatNumber(position)} s`;
        const timeline = transport?.timeline || "@document";
        els.timelineLabel.textContent = timeline;
        if (document.activeElement !== els.seekPosition) {
          els.seekPosition.value = String(position);
        }
        if (document.activeElement !== els.timelineSelect) {
          els.timelineSelect.value = timeline;
        }

        const paused = Boolean(transport?.paused) || transportState === "paused";
        els.pauseButton.querySelector("span").textContent = paused ? "继续" : "暂停";
        els.pauseButton.dataset.mode = paused ? "resume" : "pause";
      }

      function renderStatus() {
        const status = state.status?.status || {};
        STATUS_KEYS.forEach((key) => {
          const card = $(`[data-status-card="${key}"]`);
          const valueNode = $("[data-status-value]", card);
          if (!card || !valueNode) return;
          const value = status[key];
          card.classList.remove("is-on", "is-recording");
          if (value === undefined || value === null) {
            valueNode.textContent = "—";
            return;
          }
          if (typeof value === "number" || value === "0" || value === "1") {
            const active = Number(value) === 1;
            valueNode.textContent = active ? "ON" : "OFF";
            card.classList.toggle("is-on", active && key !== "transport.recording");
            card.classList.toggle("is-recording", active && key === "transport.recording");
            return;
          }
          valueNode.textContent = String(value).toUpperCase();
          card.classList.toggle("is-on", key === "audio.state" && value === "running");
        });
      }

      function renderTimelines() {
        const timelines = state.cube?.timelines || {
          document: "@document",
          transport: "@transport",
          step: "@step",
          frame: "@frame"
        };
        const currentValue = els.timelineSelect.value;
        els.timelineSelect.innerHTML = Object.entries(timelines)
          .map(([label, value]) => `<option value="${escapeHTML(value)}">${escapeHTML(value || label)}</option>`)
          .join("");
        if (timelines.document) els.timelineSelect.value = timelines.document;
        const preferred = state.transport?.timeline || currentValue;
        if (preferred && $$("option", els.timelineSelect).some((option) => option.value === preferred)) {
          els.timelineSelect.value = preferred;
        }
      }

      function renderBanks() {
        els.bankTabs.innerHTML = state.banks.map((bank) => `
          <button class="bank-tab${bank.index === state.selectedBank ? " is-active" : ""}"
            type="button" data-bank-index="${bank.index}">
            ${escapeHTML(bank.label || `Bank ${bank.index}`)}
          </button>
        `).join("");

        const bank = selectedBank();
        if (bank) {
          els.bankLabel.value = bank.label || "";
          els.bankDescription.value = bank.description || "";
          els.saveBank.disabled = false;
        } else {
          els.bankLabel.value = "";
          els.bankDescription.value = "";
          els.saveBank.disabled = true;
        }
      }

      function selectedBank() {
        return state.banks.find((bank) => bank.index === state.selectedBank) || null;
      }

      function selectedTrack() {
        return state.tracks.find((track) => track.index === state.selectedTrack) || null;
      }

      function renderSequence() {
        const stepCount = state.steps.length || state.cube?.steps?.length || 0;
        els.sequenceGrid.style.setProperty("--step-count", Math.max(stepCount, 1));
        const bank = selectedBank();
        els.sequenceMeta.textContent = bank
          ? `${bank.label || `Bank ${bank.index}`} · ${state.tracks.length} Tracks · ${stepCount} Steps`
          : "等待 Cube 数据";

        if (!state.tracks.length || !stepCount) {
          els.sequenceGrid.innerHTML = '<div class="sequence-empty">暂无可用轨道或 Step。</div>';
          return;
        }

        const cellMap = new Map(state.cells.map((cell) => [cell.name, cell]));
        const blocks = ['<div class="sequence-corner">Tracks / Steps</div>'];
        for (const step of state.steps) {
          blocks.push(`<div class="step-head">S${pad(step.index + 1)}</div>`);
        }

        state.tracks.forEach((track) => {
          const hue = 72 + (track.index * 47) % 180;
          blocks.push(`
            <button class="track-label${track.index === state.selectedTrack ? " is-active" : ""}"
              type="button" data-track-index="${track.index}"
              style="--track-hue:${hue}">
              <span class="track-index">${pad(track.index + 1)}</span>
              <span class="track-copy">
                <strong>${escapeHTML(track.label || track.name || `Track ${track.index}`)}</strong>
                <small>CH ${track.channel} · ${Number(track.gainDb).toFixed(1)} dB</small>
              </span>
            </button>
          `);
          for (const step of state.steps) {
            const name = `b${state.selectedBank}_s${step.index}_t${track.index}`;
            const cell = cellMap.get(name);
            const selected = state.selectedCell?.name === name;
            const note = cell?.note || null;
            const classes = ["seq-cell"];
            if (note) classes.push("has-note");
            if (selected) classes.push("is-selected");
            const noteLabel = note ? String(note.number) : cell ? "·" : "–";
            const title = note
              ? `${name} · Note ${note.number} · Velocity ${note.velocity}`
              : name;
            blocks.push(`
              <button class="${classes.join(" ")}" type="button"
                data-cell-name="${escapeHTML(name)}"
                aria-label="Cell ${escapeHTML(name)}"
                title="${escapeHTML(title)}">
                <span>${escapeHTML(noteLabel)}</span>
              </button>
            `);
          }
        });
        els.sequenceGrid.innerHTML = blocks.join("");
      }

      function renderTrackEditor() {
        const track = selectedTrack();
        const controls = $$("input, textarea, select", els.trackForm);
        if (!track) {
          els.trackNumber.textContent = "—";
          els.trackHeading.textContent = "未选择轨道";
          els.trackSubheading.textContent = "选择左侧轨道开始编辑";
          els.saveTrack.disabled = true;
          controls.forEach((control) => {
            control.disabled = true;
          });
          return;
        }

        controls.forEach((control) => {
          control.disabled = false;
        });
        els.saveTrack.disabled = false;
        els.trackNumber.textContent = pad(track.index + 1);
        els.trackHeading.textContent = track.label || track.name || `Track ${track.index}`;
        els.trackSubheading.textContent = `INDEX ${track.index} · MIDI CH ${track.channel}`;
        $("#track-name").value = track.name || "";
        $("#track-label").value = track.label || "";
        $("#track-description").value = track.description || "";
        $("#track-order").value = Number(track.order) || 0;
        $("#track-channel").value = Number(track.channel);
        $("#track-gain").value = Number(track.gainDb) || 0;
        $('[name="enabled"]', els.trackForm).checked = Boolean(track.enabled);
        $('[name="solo"]', els.trackForm).checked = Boolean(track.solo);
        $('[name="mute"]', els.trackForm).checked = Boolean(track.mute);
      }

      function renderBankEditor() {
        const bank = selectedBank();
        if (!bank) {
          els.bankLabel.value = "";
          els.bankDescription.value = "";
          els.saveBank.disabled = true;
          return;
        }
        els.bankLabel.value = bank.label || "";
        els.bankDescription.value = bank.description || "";
        els.saveBank.disabled = false;
      }

      function renderCellReadout() {
        const cell = state.selectedCell;
        if (!cell) {
          els.cellReadout.innerHTML = "<strong>未选择 Cell</strong><p>点击序列网格中的单元格编辑 MIDI Note。</p>";
          els.noteNumber.value = 60;
          els.noteVelocity.value = 100;
          updateCellEditorState();
          return;
        }
        const note = cell.note || null;
        const noteState = note
          ? `Note <strong>${escapeHTML(note.number)}</strong> · Velocity <strong>${escapeHTML(note.velocity)}</strong>`
          : "尚未设置 Note";
        els.cellReadout.innerHTML = `
          <strong>${escapeHTML(cell.name)}</strong>
          <p>Bank ${cell.bankIndex} · Step ${cell.stepIndex} · Track ${cell.trackIndex}<br>
          ${noteState}</p>
        `;
        els.noteNumber.value = note ? note.number : 60;
        els.noteVelocity.value = note ? note.velocity : 100;
        updateCellEditorState();
      }

      function updateCellEditorState() {
        const cell = state.selectedCell;
        const disabled = !cell || state.cellSaving || state.busy;
        els.noteNumber.disabled = disabled;
        els.noteVelocity.disabled = disabled;
        els.saveCellNote.disabled = disabled;
        els.clearCellNote.disabled = disabled || !cell?.note;
      }

      function setBusy(busy) {
        state.busy = busy;
        $$("[data-command]").forEach((button) => {
          button.disabled = busy;
        });
        $("button[type='submit']", els.seekForm).disabled = busy;
        updateCellEditorState();
      }

      async function runCommand(name) {
        if (state.busy) return;
        const paths = {
          "engine-start": ["/engine/start", "引擎启动请求已发送"],
          "engine-restart": ["/engine/restart", "引擎已重启"],
          "engine-stop": ["/engine/stop", "引擎已停止"],
          play: ["/transport/play", "已开始播放"],
          record: ["/transport/record", "已开始录制"],
          stop: ["/transport/stop", "走带已停止"]
        };
        let command = paths[name];
        if (name === "pause-resume") {
          command = els.pauseButton.dataset.mode === "resume"
            ? ["/transport/resume", "已从暂停恢复"]
            : ["/transport/pause", "已暂停"];
        }
        if (!command) return;

        setBusy(true);
        try {
          await api(command[0], { method: "POST", body: {} });
          toast(command[1]);
          await loadDashboard({ quiet: true });
          clearError();
        } catch (error) {
          showError(describeError(error));
          toast(describeError(error), "error");
        } finally {
          setBusy(false);
        }
      }

      async function seek(event) {
        event.preventDefault();
        const position = Number(els.seekPosition.value);
        if (!Number.isFinite(position) || position < 0) {
          showError("定位位置必须是大于等于 0 的秒数。");
          return;
        }
        setBusy(true);
        try {
          await api("/transport/position", {
            method: "PUT",
            body: {
              position,
              timeline: els.timelineSelect.value || "@document"
            }
          });
          toast(`已定位到 ${formatTime(position)}`);
          await loadDashboard({ quiet: true });
          clearError();
        } catch (error) {
          showError(describeError(error));
        } finally {
          setBusy(false);
        }
      }

      async function selectBank(index) {
        if (index === state.selectedBank) return;
        state.selectedBank = index;
        state.selectedCell = null;
        renderBanks();
        renderSequence();
        renderBankEditor();
        renderCellReadout();
        const failures = [];
        await loadCells(failures);
        renderSequence();
        if (failures.length) showError(failures.join("；"));
      }

      function selectTrack(index) {
        state.selectedTrack = index;
        renderSequence();
        renderTrackEditor();
      }

      function selectCell(name) {
        const cell = state.cells.find((item) => item.name === name);
        state.selectedCell = cell || {
          name,
          bankIndex: state.selectedBank,
          stepIndex: null,
          trackIndex: null
        };
        renderSequence();
        renderCellReadout();
      }

      function setCellSaving(saving) {
        state.cellSaving = saving;
        updateCellEditorState();
      }

      function replaceCell(updated) {
        state.cells = state.cells.map((cell) => cell.name === updated.name ? updated : cell);
        if (state.selectedCell?.name === updated.name) {
          state.selectedCell = updated;
        }
        renderSequence();
        renderCellReadout();
      }

      function readNoteValue(input, label) {
        const raw = input.value.trim();
        const value = Number(raw);
        if (raw === "" || !Number.isInteger(value) || value < 0 || value > 127) {
          showError(`${label}必须是 0 到 127 之间的整数。`);
          input.focus();
          return null;
        }
        return value;
      }

      async function saveCellNote(event) {
        event.preventDefault();
        const cell = state.selectedCell;
        if (!cell || state.cellSaving || state.busy) return;

        const number = readNoteValue(els.noteNumber, "MIDI 音高");
        if (number === null) return;
        const velocity = readNoteValue(els.noteVelocity, "力度");
        if (velocity === null) return;

        const cellName = cell.name;
        setCellSaving(true);
        try {
          const updated = await api(`/cells/${encodeURIComponent(cellName)}`, {
            method: "PATCH",
            body: { note: { number, velocity } }
          });
          replaceCell(updated);
          toast(`${cellName} 的 Note 已保存`);
          clearError();
        } catch (error) {
          showError(describeError(error));
          toast(describeError(error), "error");
        } finally {
          setCellSaving(false);
        }
      }

      async function clearCellNote() {
        const cell = state.selectedCell;
        if (!cell?.note || state.cellSaving || state.busy) return;

        const cellName = cell.name;
        setCellSaving(true);
        try {
          const updated = await api(`/cells/${encodeURIComponent(cellName)}`, {
            method: "PATCH",
            body: { note: null }
          });
          replaceCell(updated);
          toast(`${cellName} 的 Note 已清除`);
          clearError();
        } catch (error) {
          showError(describeError(error));
          toast(describeError(error), "error");
        } finally {
          setCellSaving(false);
        }
      }

      async function saveTrack(event) {
        event.preventDefault();
        const track = selectedTrack();
        if (!track) return;
        const payload = {
          name: $("#track-name").value,
          label: $("#track-label").value,
          description: $("#track-description").value,
          order: Number($("#track-order").value),
          channel: Number($("#track-channel").value),
          gainDb: Number($("#track-gain").value),
          enabled: $('[name="enabled"]', els.trackForm).checked,
          solo: $('[name="solo"]', els.trackForm).checked,
          mute: $('[name="mute"]', els.trackForm).checked
        };
        els.saveTrack.disabled = true;
        try {
          const updated = await api(`/tracks/${track.index}`, { method: "PATCH", body: payload });
          state.tracks = state.tracks.map((item) => item.index === track.index ? updated : item);
          renderSequence();
          renderTrackEditor();
          toast("Track 已保存");
          clearError();
        } catch (error) {
          showError(describeError(error));
          toast(describeError(error), "error");
        } finally {
          els.saveTrack.disabled = false;
        }
      }

      async function saveBank(event) {
        event.preventDefault();
        const bank = selectedBank();
        if (!bank) return;
        els.saveBank.disabled = true;
        try {
          const updated = await api(`/banks/${bank.index}`, {
            method: "PATCH",
            body: {
              label: els.bankLabel.value,
              description: els.bankDescription.value
            }
          });
          state.banks = state.banks.map((item) => item.index === bank.index ? updated : item);
          renderBanks();
          renderSequence();
          toast("Bank 已保存");
          clearError();
        } catch (error) {
          showError(describeError(error));
          toast(describeError(error), "error");
        } finally {
          els.saveBank.disabled = false;
        }
      }

      async function refreshLive() {
        if (document.hidden || state.loading || state.busy) return;
        const requests = [
          ["status", "/status"],
          ["engine", "/engine"],
          ["transport", "/transport"]
        ];
        const results = await Promise.allSettled(requests.map(([, path]) => api(path)));
        results.forEach((result, index) => {
          if (result.status === "fulfilled") state[requests[index][0]] = result.value;
        });
        renderEngine();
        renderTransport();
        renderStatus();
      }

      async function executeActions() {
        let payload;
        try {
          payload = JSON.parse(els.actionEditor.value);
        } catch (error) {
          showError(`Action JSON 无效：${error.message}`);
          els.actionResponse.textContent = error.message;
          return;
        }

        els.executeActions.disabled = true;
        els.actionResponse.textContent = "执行中…";
        try {
          const data = await api("/actions", { method: "POST", body: payload });
          els.actionResponse.textContent = JSON.stringify(data, null, 2);
          toast("Action 执行完成");
          await loadDashboard({ quiet: true });
        } catch (error) {
          els.actionResponse.textContent = JSON.stringify(error.payload || {
            error: {
              code: error.code || "HTTP_ERROR",
              message: error.message
            }
          }, null, 2);
          showError(describeError(error));
        } finally {
          els.executeActions.disabled = false;
        }
      }

      function fillStatusAction() {
        els.actionEditor.value = JSON.stringify({
          actions: [{
            name: "status",
            params: { all: true }
          }]
        }, null, 2);
      }

      function fillStartAction() {
        els.actionEditor.value = JSON.stringify({
          actions: [
            { name: "start_engine", params: {} },
            { name: "play", params: {} },
            {
              name: "status",
              params: {
                all: false,
                keys: ["engine.started", "transport.playing"]
              }
            }
          ]
        }, null, 2);
      }

      function wireEvents() {
        els.refreshButton.addEventListener("click", () => loadDashboard());
        els.dismissError.addEventListener("click", clearError);
        els.seekForm.addEventListener("submit", seek);
        els.trackForm.addEventListener("submit", saveTrack);
        els.cellForm.addEventListener("submit", saveCellNote);
        els.clearCellNote.addEventListener("click", clearCellNote);
        els.bankForm.addEventListener("submit", saveBank);
        els.executeActions.addEventListener("click", executeActions);
        $("#fill-status-action").addEventListener("click", fillStatusAction);
        $("#fill-start-action").addEventListener("click", fillStartAction);

        els.midiOctave.addEventListener("change", () => {
          state.midiOctave = Number(els.midiOctave.value) || 3;
          renderMidiKeyboard();
        });
        els.midiChannel.addEventListener("change", () => {
          state.midiChannel = Number(els.midiChannel.value) || 0;
          setMidiActivity(`输入通道切换为 CH ${state.midiChannel + 1}`, "idle");
        });
        els.midiVelocity.addEventListener("input", () => {
          state.midiVelocity = Math.max(1, Math.min(127, Number(els.midiVelocity.value) || 1));
          els.midiVelocityValue.textContent = String(state.midiVelocity);
        });
        els.midiKeyboard.addEventListener("pointerdown", startMidiPointer);
        els.midiKeyboard.addEventListener("pointermove", moveMidiPointer);
        els.midiKeyboard.addEventListener("pointerup", stopMidiPointer);
        els.midiKeyboard.addEventListener("pointercancel", stopMidiPointer);
        els.midiKeyboard.addEventListener("lostpointercapture", stopMidiPointer);

        document.addEventListener("click", (event) => {
          const commandButton = event.target.closest("[data-command]");
          if (commandButton) {
            runCommand(commandButton.dataset.command);
            return;
          }
          const bankButton = event.target.closest("[data-bank-index]");
          if (bankButton) {
            selectBank(Number(bankButton.dataset.bankIndex));
            return;
          }
          const trackButton = event.target.closest("[data-track-index]");
          if (trackButton) {
            selectTrack(Number(trackButton.dataset.trackIndex));
            return;
          }
          const cellButton = event.target.closest("[data-cell-name]");
          if (cellButton) {
            selectCell(cellButton.dataset.cellName);
          }
        });

        document.addEventListener("keydown", (event) => {
          const target = event.target;
          const editing = target instanceof HTMLInputElement ||
            target instanceof HTMLTextAreaElement ||
            target instanceof HTMLSelectElement;
          if (editing || event.ctrlKey || event.metaKey || event.altKey) return;

          const semitone = MIDI_PC_KEYS[event.code];
          if (semitone !== undefined) {
            event.preventDefault();
            if (!event.repeat) {
              const note = (state.midiOctave + 1) * 12 + semitone;
              startMidiNote(note, `keyboard:${event.code}`);
            }
            return;
          }

          if (event.repeat || event.code !== "Space") return;
          event.preventDefault();
          runCommand("pause-resume");
        });

        document.addEventListener("keyup", (event) => {
          const source = `keyboard:${event.code}`;
          if (!activeMidiSources.has(source)) return;
          event.preventDefault();
          stopMidiNote(source);
        });

        window.addEventListener("blur", releaseAllMidiNotes);

        document.addEventListener("visibilitychange", () => {
          if (document.hidden) {
            releaseAllMidiNotes();
          } else {
            refreshLive();
          }
        });
      }

      function init() {
        wireEvents();
        renderMidiChannels();
        els.midiVelocityValue.textContent = String(state.midiVelocity);
        renderMidiKeyboard();
        loadDashboard();
        window.setInterval(refreshLive, 4000);
      }

      init();
    })();
  
