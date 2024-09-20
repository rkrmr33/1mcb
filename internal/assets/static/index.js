let togglerId = '';

window.onload = async () => {
    const cbxs = document.getElementById('checkboxes');

    try {
        await setupEventSource(handleUpdates.bind(null, cbxs));
        console.log('[+] Event-Source opened');
    } catch {
        console.log('Could not open event-source');
    }
};

function handleUpdates(cbxs, event) {
    switch (event.type) {
        case 0: break; // Keepalive event
        case 1: // Hello event
            togglerId = event.payload.togglerId;
            console.log(togglerId)
            break;
        case 2: // Update event
            const { index, state } = event.payload;

            cbxs.children[index].children[0].checked = state;
            break;
        default:
            console.log('unsupported event type:', event.type)
    }
}

function cbClicked(cb) {
    const { id, checked } = cb;

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