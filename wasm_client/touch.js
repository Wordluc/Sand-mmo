// multitouch.js
document.addEventListener("DOMContentLoaded", () => {
	const buttons = document.querySelectorAll("button, .btn");

	buttons.forEach(btn => {
		btn.addEventListener("touchstart", (e) => {
			e.preventDefault(); // prevents delay and double-firing with onclick
			btn.click();
		}, { passive: false });

		// For held buttons (like dpad), fire repeatedly while held
		let interval = null;
		btn.addEventListener("touchstart", (e) => {
			if (interval) return;
			interval = setInterval(() => btn.click(), 100);
		}, { passive: false });

		btn.addEventListener("touchend", () => {
			clearInterval(interval);
			interval = null;
		});

		btn.addEventListener("touchcancel", () => {
			clearInterval(interval);
			interval = null;
		});
	});
});
