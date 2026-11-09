export const query = (selector) => document.querySelector(selector);
export const escapeHTML = (value) =>
  String(value ?? "").replace(
    /[&<>"']/g,
    (character) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        character
      ],
  );
export const renderOption = (value, label) =>
  `<option value="${escapeHTML(value)}">${escapeHTML(label)}</option>`;
let toastTimer;
let confirmation = null;

export function toast(message) {
  clearTimeout(toastTimer);
  query("#toast").textContent = message;
  query("#toast").hidden = false;
  toastTimer = setTimeout(() => {
    query("#toast").hidden = true;
  }, 4500);
}
export function showError(selector, error) {
  query(selector).textContent = error.message;
  query(selector).hidden = false;
}
export function populateSelect(select, entries, fallback = "") {
  const old = select.value;
  select.innerHTML = entries
    .map(([id, label]) => renderOption(id, label))
    .join("");
  if (entries.some(([id]) => id === old)) select.value = old;
  else if (entries.some(([id]) => id === fallback)) select.value = fallback;
}
export function confirmAction(title, text, action) {
  query("#confirm-title").textContent = title;
  query("#confirm-text").textContent = text;
  query("#confirm-error").hidden = true;
  confirmation = action;
  query("#confirm-dialog").showModal();
}

export function bindDialogControls() {
  document
    .querySelectorAll(".close-dialog")
    .forEach((button) =>
      button.addEventListener("click", () => button.closest("dialog").close()),
    );
  query("#confirm-dialog").addEventListener("close", () => {
    confirmation = null;
  });
  query("#confirm-action").addEventListener("click", async (event) => {
    if (!confirmation) return;
    const button = event.currentTarget;
    button.disabled = true;
    try {
      await confirmation();
      query("#confirm-dialog").close();
    } catch (error) {
      showError("#confirm-error", error);
    } finally {
      button.disabled = false;
    }
  });
}
