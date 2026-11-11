import { api } from "./api.js";
import { query, toast, confirmAction, bindDialogControls } from "./ui.js";
import { formatTimestamp } from "./presentation.js";
import { createPoliciesView } from "./policies.js";
import { createTransfersView } from "./transfers.js";
import { createWalletsView } from "./wallets.js";
import { bindCatalogDialog } from "./catalog.js";
let state;
const viewNames = ["wallets", "policies", "transfers"];
const initialLocation = new URLSearchParams(window.location.search);
const initialProfile = initialLocation.get("profile");
const initialView = initialLocation.get("view");
let profileId = Array.from(query("#profile").options).some(
  (option) => option.value === initialProfile,
)
  ? initialProfile
  : "treasury";
let currentView = viewNames.includes(initialView) ? initialView : "wallets";
let refreshSequence = 0;
const workspace = {
  getState: () => (workspaceReady() ? state : undefined),
  getProfileId: () => profileId,
  refresh: refreshWorkspace,
};
bindDialogControls();
const wallets = createWalletsView(workspace);
const policies = createPoliciesView(workspace);
const transfers = createTransfersView(workspace);
bindCatalogDialog(workspace);
function workspaceReady() {
  return state?.profile_id === profileId;
}
function showView(view) {
  currentView = view;
  for (const name of viewNames) {
    query(`#${name}-view`).hidden = !workspaceReady() || view !== name;
  }
  query("#breadcrumb-current").textContent = {
    wallets: "Wallets",
    policies: "Policies",
    transfers: "Transfers",
  }[view];
  document.querySelectorAll("[data-view]").forEach((button) => {
    button.classList.toggle("selected", button.dataset.view === view);
    button.setAttribute(
      "aria-current",
      button.dataset.view === view ? "page" : "false",
    );
  });
  if (workspaceReady()) {
    const url = new URL(window.location.href);
    url.searchParams.set("profile", profileId);
    url.searchParams.set("view", currentView);
    if (url.href !== window.location.href)
      window.history.replaceState(null, "", url);
  }
}
async function refreshWorkspace() {
  const sequence = ++refreshSequence,
    requestedProfile = profileId;
  let next;
  try {
    next = await api.getWorkspace(requestedProfile);
  } catch (error) {
    if (sequence !== refreshSequence || requestedProfile !== profileId)
      return false;
    throw error;
  }
  if (sequence !== refreshSequence || requestedProfile !== profileId)
    return false;
  const profileChanged = state?.profile_id !== next.profile_id;
  state = next;
  query("#loading").hidden = true;
  query("#clock").textContent = formatTimestamp(state.now);
  query("#clock").dateTime = state.now;
  query("#policy-count").textContent = state.policies.length;
  query("#transfer-count").textContent = state.transfers.length;
  query("#active-count").textContent = state.policies.length;
  query("#history-count").textContent = state.transfers.length;
  query("#wallet-count").textContent = state.wallets.length;
  wallets.render();
  policies.render();
  transfers.render();
  if (profileChanged) transfers.resetForm();
  showView(currentView);
  return true;
}
document
  .querySelectorAll("[data-view]")
  .forEach((button) =>
    button.addEventListener("click", () => showView(button.dataset.view)),
  );
query("#profile").addEventListener("change", async (event) => {
  profileId = event.target.value;
  const requestedProfile = profileId;
  query("#wallets-view").hidden = true;
  query("#policies-view").hidden = true;
  query("#transfers-view").hidden = true;
  query("#loading").textContent = "Loading workspace…";
  query("#loading").hidden = false;
  try {
    await refreshWorkspace();
  } catch (error) {
    if (requestedProfile !== profileId) return;
    profileId = state?.profile_id || "treasury";
    query("#profile").value = profileId;
    query("#loading").hidden = true;
    showView(currentView);
    toast(error.message);
  }
});
document.querySelectorAll("[data-hours]").forEach((button) =>
  button.addEventListener("click", async () => {
    button.disabled = true;
    try {
      await api.advanceClock(Number(button.dataset.hours));
      await refreshWorkspace();
    } catch (error) {
      toast(error.message);
    } finally {
      button.disabled = false;
    }
  }),
);
query("#reset-button").addEventListener("click", () =>
  confirmAction(
    "Reset demo data?",
    "Replace policies and transfers in both profiles with the original demo data and reset the simulation clock?",
    async () => {
      await api.resetWorkspace();
      await refreshWorkspace();
      transfers.resetForm();
      toast("Demo data restored");
    },
  ),
);
query("#profile").value = profileId;
refreshWorkspace().catch((error) => {
  query("#loading").textContent = error.message;
});
