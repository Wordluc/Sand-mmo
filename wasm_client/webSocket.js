var queue_chunks = new Map()


function setup_websocket(ws, chunk_size) {
	ws.binaryType = "arraybuffer"
	ws.onmessage = (e) => {
		const buffer = new Uint8Array(e.data);
		let offset = 0;
		while (offset < buffer.length) {
			const chunkId =
				(buffer[offset] << 8) | buffer[offset + 1];

			const chunk = buffer.slice(offset, offset + chunk_size);

			queue_chunks.set(chunkId, chunk);

			offset += chunk_size;
		}
	};
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
