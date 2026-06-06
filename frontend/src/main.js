import {
  StartListening,
  StopListening,
  Send,
  StartBatch,
  StopBatch,
} from '../wailsjs/go/main/App';

import { EventsOn } from '../wailsjs/runtime/runtime';

// --- DOM refs ---
const portInput        = document.getElementById('port-input');
const btnStart         = document.getElementById('btn-start');
const btnStop          = document.getElementById('btn-stop');
const statusDot        = document.getElementById('status-dot');
const statusText       = document.getElementById('status-text');

const barcodeInput     = document.getElementById('barcode-input');
const btnSend          = document.getElementById('btn-send');

const delayInput       = document.getElementById('delay-input');
const btnBatchStart    = document.getElementById('btn-batch-start');
const btnBatchStop     = document.getElementById('btn-batch-stop');
const batchTextarea    = document.getElementById('batch-textarea');
const batchProgressRow = document.getElementById('batch-progress-row');
const batchCounter     = document.getElementById('batch-counter');
const batchBar         = document.getElementById('batch-bar');

const loopCheckbox    = document.getElementById('loop-checkbox');
const loopTimesLabel  = document.getElementById('loop-times-label');
const loopTimesInput  = document.getElementById('loop-times-input');

const logList          = document.getElementById('log-list');
const btnClear         = document.getElementById('btn-clear');

// --- State: idle | listening | connected ---
let state = 'idle';

function applyState(next) {
  state = next;

  const idle      = state === 'idle';
  const connected = state === 'connected';

  btnStart.disabled      = !idle;
  portInput.disabled     = !idle;
  btnStop.disabled       = idle;

  barcodeInput.disabled  = !connected;
  btnSend.disabled       = !connected;

  batchTextarea.disabled = !connected;
  btnBatchStart.disabled = !connected;
  loopCheckbox.disabled  = !connected;
  loopTimesInput.disabled = !connected;

  if (idle || state === 'listening') {
    btnBatchStop.disabled = true;
    batchProgressRow.classList.add('hidden');
  }

  const dotClass = { idle: 'dot-idle', listening: 'dot-listening', connected: 'dot-connected' }[state];
  statusDot.className = `dot ${dotClass}`;
}

// --- Connection buttons ---
btnStart.addEventListener('click', async () => {
  const port = parseInt(portInput.value, 10);
  if (!port || port < 1 || port > 65535) {
    appendLog('ERR', 'Invalid port number');
    return;
  }
  try {
    await StartListening(port);
  } catch (err) {
    appendLog('ERR', String(err));
  }
});

btnStop.addEventListener('click', () => {
  StopListening();
});

// --- Manual send ---
btnSend.addEventListener('click', sendBarcode);
barcodeInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') sendBarcode();
});

async function sendBarcode() {
  const value = barcodeInput.value.trim();
  if (!value) return;
  try {
    await Send(value);
    barcodeInput.value = '';
    barcodeInput.focus();
  } catch (err) {
    appendLog('ERR', String(err));
  }
}

// --- Batch mode ---
btnBatchStart.addEventListener('click', async () => {
  const items = batchTextarea.value
    .split('\n')
    .map(s => s.trim())
    .filter(s => s.length > 0);

  if (items.length === 0) {
    appendLog('ERR', 'Batch list is empty');
    return;
  }

  const delay = parseInt(delayInput.value, 10);
  if (isNaN(delay) || delay < 0) {
    appendLog('ERR', 'Invalid delay value');
    return;
  }

  let loops = 1;
  if (loopCheckbox.checked) {
    const raw = loopTimesInput.value.trim();
    if (raw === '') {
      loops = 0;
    } else {
      const n = parseInt(raw, 10);
      if (isNaN(n) || n < 0) {
        appendLog('ERR', 'Invalid loop count');
        return;
      }
      loops = n; // 0 means infinite
    }
  }

  try {
    await StartBatch(items, delay, loops);
    btnBatchStart.disabled = true;
    btnBatchStop.disabled = false;
    const cycleTotal = loops === 0 ? '∞' : String(loops);
    batchCounter.textContent = `Cycle 1 / ${cycleTotal} — 0 / ${items.length}`;
    batchBar.value = 0;
    batchBar.max = items.length;
    batchProgressRow.classList.remove('hidden');
  } catch (err) {
    appendLog('ERR', String(err));
  }
});

btnBatchStop.addEventListener('click', () => {
  StopBatch();
});

loopCheckbox.addEventListener('change', () => {
  const show = loopCheckbox.checked;
  loopTimesLabel.classList.toggle('hidden', !show);
  loopTimesInput.classList.toggle('hidden', !show);
});

// --- Activity log ---
function appendLog(level, message) {
  const now = new Date();
  const time = now.toTimeString().slice(0, 8);
  const li = document.createElement('li');
  li.innerHTML = `
    <span class="log-time">${time}</span>
    <span class="log-level ${level}">${level}</span>
    <span class="log-msg">${escapeHtml(message)}</span>
  `;
  logList.prepend(li);
}

function escapeHtml(str) {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

btnClear.addEventListener('click', () => {
  logList.innerHTML = '';
});

// --- Wails event listeners ---
EventsOn('status:listening', ({ port }) => {
  applyState('listening');
  statusText.textContent = `Listening on :${port}`;
});

EventsOn('status:connected', ({ remoteAddr }) => {
  applyState('connected');
  statusText.textContent = `Connected — ${remoteAddr}`;
});

EventsOn('status:disconnected', () => {
  applyState('listening');
  statusText.textContent = 'Listening — client disconnected';
});

EventsOn('status:idle', () => {
  applyState('idle');
  statusText.textContent = 'Idle';
});

EventsOn('log:entry', ({ time, level, message }) => {
  const li = document.createElement('li');
  li.innerHTML = `
    <span class="log-time">${escapeHtml(time)}</span>
    <span class="log-level ${escapeHtml(level)}">${escapeHtml(level)}</span>
    <span class="log-msg">${escapeHtml(message)}</span>
  `;
  logList.prepend(li);
});

EventsOn('batch:progress', ({ index, total, cycle, totalCycles }) => {
  const cycleTotal = totalCycles === 0 ? '∞' : String(totalCycles);
  batchCounter.textContent = `Cycle ${cycle} / ${cycleTotal} — ${index} / ${total}`;
  batchBar.value = index;
  batchBar.max = total;
});

EventsOn('batch:done', () => {
  btnBatchStart.disabled = false;
  btnBatchStop.disabled = true;
  batchProgressRow.classList.add('hidden');
});

// Initial state
applyState('idle');
