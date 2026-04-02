var is_modal_open = false;

function toggle_modal() {
	const modal = document.getElementById("info-modal");
	if (is_modal_open) {
		modal.style.display = "none";
		is_modal_open = false;
	} else {
		modal.style.display = "block";
		is_modal_open = true; // was false
	}
}
