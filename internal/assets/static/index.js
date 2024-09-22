'use strict';

const ONE_MILLION = 1_000_000;

let togglerId = '';
let virtualizedList
let state;
const bitsInBucket = Uint8Array.BYTES_PER_ELEMENT * 8;

window.onload = async () => {
    const cbxs = document.getElementById('checkboxes');
    const inRow = 28;

    const res = await fetch('/api/state');
    state = new Uint8Array(await res.arrayBuffer());

    virtualizedList = new VirtualizedList.default(cbxs, {
        height: cbxs.style.height, // The height of the container
        rowCount: state.length * bitsInBucket / inRow,
        renderRow: rowIndex => {
            const row = document.createElement('div');
            row.className = 'checkbox-row';
            console.log('rendered', rowIndex);

            for (let i = 0; i < inRow; i++) {
                const cbId = rowIndex * inRow + i;
                const cbIdStr = cbId.toString();
                const cb = document.createElement('div');
                cb.className = 'checkbox-wrapper';

                const input = document.createElement('input');
                input.id = cbIdStr;
                input.type = 'checkbox';
                input.onclick = cbClicked;
                cb.appendChild(input);

                if (getStateIdx(cbId)) {
                    input.setAttribute('checked', 'checked');
                }

                const label = document.createElement('label');
                label.className = 'cbx';
                label.setAttribute('for', cbIdStr);
                cb.appendChild(label);

                row.appendChild(cb);
            }

            return row;
        },
        rowHeight: 28,
        overscanCount: 100,
    });

    try {
        await setupEventSource(handleUpdates);
        console.log('[+] Event-Source opened');
    } catch {
        console.log('Could not open event-source');
    }
};

function getStateIdx(idx) {
    const bucketIdx = Math.floor(idx / bitsInBucket);
    const idxInBucket = idx % bitsInBucket;
    return (state[bucketIdx] & (1 << idxInBucket)) !== 0
}

function setStateIdx(idx, val) {
    const bucketIdx = Math.floor(idx / bitsInBucket);
    const idxInBucket = idx % bitsInBucket;

    const cbx = document.getElementById(idx);
    if (cbx) {
        cbx.checked = val;
    }

    if (val) {
        state[bucketIdx] |= 1 << idxInBucket;
    } else {
        state[bucketIdx] &= ~(1 << idxInBucket);
    }
}

function handleUpdates(event) {
    switch (event.type) {
        case 0: break; // Keepalive event
        case 1: // Hello event
            togglerId = event.payload.togglerId;
            break;
        case 2: // Update event
            const { index, state } = event.payload;
            setStateIdx(index, state);
            break;
        default:
            console.log('unsupported event type:', event.type)
    }
}

function cbClicked() {
    const { id, checked } = this;

    setStateIdx(id, checked);

    fetch('/api/toggle', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            index: parseInt(id),
            state: checked,
            togglerId,
        })
    })
}

function setupEventSource(handler) {
    const url = `${window.location.toString().replace(/\/$/, '')}/api/events`;
    console.log(`[+] Starting event-source: ${url}`);

    return new Promise((res, rej) => {
        const es = new EventSource(url);

        es.onopen = res;
        es.onerror = rej;
        es.onmessage = (msg) => {
            const data = JSON.parse(msg.data);
            handler(data);
        };
    });
}