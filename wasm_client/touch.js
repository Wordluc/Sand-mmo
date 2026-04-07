// multitouch.js
document.addEventListener("DOMContentLoaded", () => {
	const buttons = document.querySelectorAll(".touchButton");
	buttons.forEach(btn => {
		btn.addEventListener("touchstart", (e) => {
			btn.classList.add("onTouch")
			btn.click();
		}, { passive: false });

		// For held buttons
		let interval = null;
		btn.addEventListener("touchstart", (e) => {
			if (interval) return;
			interval = setInterval(() => btn.click(), 100);
		}, { passive: false });

		btn.addEventListener("touchend", () => {
			btn.classList.remove("onTouch")
			clearInterval(interval);
			interval = null;
		});

		btn.addEventListener("touchcancel", () => {
			btn.classList.remove("onTouch")
			clearInterval(interval);
			interval = null;
		});
	});
});
