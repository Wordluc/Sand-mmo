var queue_chunks = new Map()


function setup_websocket(ws) {
	ws.binaryType = "arraybuffer"
	ws.onmessage = (e) => {
		const view = new DataView(e.data);
		const chunkId = view.getUint16(0, false);
		const payload = new Uint8Array(e.data, 0);
		queue_chunks.set(chunkId, payload)
	}
}

function get_all_chunks_binary() {
	const chunks = Array.from(queue_chunks.values());

	let totalSize = 0;
	for (const payload of chunks) {
		totalSize += payload.byteLength;
	}

	const buffer = new Uint8Array(totalSize);
	let offset = 0;
	for (const payload of chunks) {
		buffer.set(payload, offset);
		offset += payload.byteLength;
	}

	return buffer;
}

function clear_all_queued_chunks() {
	queue_chunks = new Map()
}
