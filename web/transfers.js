import { api } from "./api.js";
import { query, escapeHTML, populateSelect, showError } from "./ui.js";
import {
  lookupName,
  formatAmount,
  formatTimestamp,
  renderStatusBadge,
  renderPolicyConditions,
} from "./presentation.js";
export function createTransfersView(workspace) {
  let state;
  let profile;
  const refresh = workspace.refresh;
  const profileEntries = (collection) =>
    collection.filter((item) => item.profile_id === profile);
  function synchronize() {
    state = workspace.getState();
    profile = workspace.getProfileId();
  }
  let selectedTransferId = null;
  query("#detail-dialog").addEventListener("close", () => {
    selectedTransferId = null;
  });
  function renderTransferHistory() {
    query("#transfers-list").innerHTML = state.transfers.length
      ? `<div class="table-scroll"><table><thead><tr><th>TRANSFER / CREATED UTC</th><th>AMOUNT</th><th>ROUTE</th><th>POLICY</th><th>STATUS</th><th></th></tr></thead><tbody>${state.transfers
          .map(
            (transfer) => `
    <tr data-transfer="${escapeHTML(transfer.id)}"><td>${escapeHTML(transfer.id.toUpperCase())}<span class="cell-sub">${escapeHTML(formatTimestamp(transfer.created_at))}</span></td><td class="amount-cell">${escapeHTML(formatAmount(transfer.amount))} ${escapeHTML(transfer.currency)}</td><td>${escapeHTML(lookupName(state.catalog.wallets, transfer.wallet_id))}<span class="cell-sub">→ ${escapeHTML(lookupName(state.catalog.recipients, transfer.recipient_id))}</span></td><td>${escapeHTML(transfer.decision.policy_name || "No match")}<span class="cell-sub">${transfer.decision.outcome === "ALLOW" ? "Allowed at creation" : "Rejected at creation"}</span></td><td>${renderStatusBadge(transfer.status)}</td><td><button class="table-button" data-transfer="${escapeHTML(transfer.id)}" aria-label="View ${escapeHTML(transfer.id)}">View ↗</button></td></tr>`,
          )
          .join("")}</tbody></table></div>`
      : '<div class="empty">No transfers yet.<br>Create a transfer to see how the policies apply.</div>';
  }
  function populateTransferForm() {
    const form = query("#transfer-form");
    populateSelect(
      form.elements.initiator_id,
      profileEntries(state.catalog.people).map((p) => [p.id, p.name]),
    );
    populateSelect(
      form.elements.wallet_id,
      profileEntries(state.catalog.wallets).map((wallet) => [
        wallet.id,
        `${wallet.name} · ${wallet.currency}`,
      ]),
    );
    populateSelect(
      form.elements.recipient_id,
      profileEntries(state.catalog.recipients).map((recipient) => [
        recipient.id,
        recipient.name,
      ]),
    );
    updateWalletContext();
  }
  function updateWalletContext() {
    const wallet = state.wallets.find(
      (wallet) =>
        wallet.id === query("#transfer-form").elements.wallet_id.value,
    );
    query("#transfer-currency").textContent = wallet?.currency || "";
    query("#transfer-available").textContent = wallet
      ? `${formatAmount(wallet.available)} ${wallet.currency} available`
      : "";
  }
  function renderTransferActions(transfer) {
    let result = "";
    if (transfer.status === "PENDING_SIGNATURE")
      result +=
        '<button class="button primary small" data-action="sign">Sign & send</button>';
    if (transfer.status === "SENT")
      result +=
        '<button class="button primary small" data-action="confirm">Confirm</button>';
    if (transfer.status === "PENDING_SIGNATURE")
      result +=
        '<button class="button secondary small" data-action="reject">Reject</button>';
    if (transfer.status === "PENDING_SIGNATURE")
      result +=
        '<button class="button secondary small" data-action="cancel">Cancel transfer</button>';
    if (["PENDING_SIGNATURE", "SENT"].includes(transfer.status))
      result +=
        '<button class="button secondary small" data-action="fail">Simulate failure</button>';
    return (
      result ||
      '<span class="field-hint">This transfer is in a final state.</span>'
    );
  }
  function openTransferDetails(id, open = true) {
    const transfer = state.transfers.find((item) => item.id === id);
    if (!transfer) {
      query("#detail-dialog").close();
      return;
    }
    selectedTransferId = id;
    query("#detail-title").textContent = transfer.id.toUpperCase();
    const decision = transfer.decision;
    const snapshot = decision.matched_policy
      ? `<div class="condition-chips">${renderPolicyConditions(decision.matched_policy, state.catalog)}</div>`
      : '<p class="field-hint">No policy matched at creation.</p>';
    query("#detail-body").innerHTML =
      `<div class="detail-amount">${escapeHTML(formatAmount(transfer.amount))} ${escapeHTML(transfer.currency)}</div>${renderStatusBadge(transfer.status)}
    <dl class="detail-grid"><div><dt>From wallet</dt><dd>${escapeHTML(lookupName(state.catalog.wallets, transfer.wallet_id))}</dd></div><div><dt>To recipient</dt><dd>${escapeHTML(lookupName(state.catalog.recipients, transfer.recipient_id))}</dd></div><div><dt>Initiator</dt><dd>${escapeHTML(lookupName(state.catalog.people, transfer.initiator_id))}</dd></div><div><dt>Created UTC</dt><dd>${escapeHTML(formatTimestamp(transfer.created_at))}</dd></div><div><dt>Updated UTC</dt><dd>${escapeHTML(formatTimestamp(transfer.updated_at))}</dd></div><div><dt>Policy at creation</dt><dd>${escapeHTML(decision.policy_name || "No match")}</dd></div>${transfer.note ? `<div><dt>Note</dt><dd>${escapeHTML(transfer.note)}</dd></div>` : ""}</dl>
    <div class="detail-actions">${renderTransferActions(transfer)}</div><p id="detail-error" class="form-error" role="alert" hidden></p>
    <div class="detail-section"><h3>Policy conditions at creation</h3>${snapshot}</div>
    <div class="detail-section"><h3>Evaluation trace</h3>${decision.trace.map((rule, index) => `<div class="trace-item ${rule.matched ? "matched" : ""}"><div class="trace-title"><strong>${index + 1}. ${escapeHTML(rule.policy_name)}</strong><span>${rule.matched ? "Matched" : "Continue"}</span></div>${rule.checks.map((condition) => `<div class="trace-check"><span class="${condition.matched ? "pass" : "miss"}">${condition.matched ? "✓" : "×"}</span><strong>${escapeHTML(condition.name)}</strong><small>${escapeHTML(condition.detail)}</small></div>`).join("")}</div>`).join("") || '<p class="field-hint">There were no policies in this profile.</p>'}</div>`;
    if (open) query("#detail-dialog").showModal();
  }
  query("#transfer-form").addEventListener("change", (event) => {
    if (event.target.name === "wallet_id") updateWalletContext();
  });
  query("#transfer-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const button = event.submitter;
    button.disabled = true;
    query("#transfer-error").hidden = true;
    const input = {
      profile_id: profile,
      ...Object.fromEntries(new FormData(event.target)),
    };
    input.amount = input.amount.trim();
    try {
      const transfer = await api.createTransfer(input);
      await refresh();
      if (transfer.profile_id === profile) {
        query("#latest-result").hidden = false;
        query("#latest-result").classList.toggle(
          "rejected",
          transfer.decision.outcome !== "ALLOW",
        );
        query("#latest-result").innerHTML =
          `<div>${escapeHTML(transfer.id.toUpperCase())} created · ${transfer.decision.outcome === "ALLOW" ? "Allowed" : "Rejected"}<small>${escapeHTML(transfer.decision.policy_name || "No matching policy")}</small></div><button data-transfer="${escapeHTML(transfer.id)}">View decision ↗</button>`;
      }
    } catch (error) {
      showError("#transfer-error", error);
    } finally {
      button.disabled = false;
    }
  });
  document.addEventListener("click", (event) => {
    const row = event.target.closest("[data-transfer]");
    if (row) openTransferDetails(row.dataset.transfer);
  });
  query("#detail-body").addEventListener("click", async (event) => {
    const button = event.target.closest("[data-action]");
    if (!button || button.disabled) return;
    const id = selectedTransferId,
      originalProfile = profile;
    const command = {
      profile_id: originalProfile,
      action: button.dataset.action,
    };
    button.disabled = true;
    try {
      await api.applyTransferAction(id, command);
      await refresh();
      if (
        profile === originalProfile &&
        selectedTransferId === id &&
        query("#detail-dialog").open
      )
        openTransferDetails(id, false);
    } catch (error) {
      if (profile === originalProfile && selectedTransferId === id)
        showError("#detail-error", error);
    } finally {
      if (button.isConnected) button.disabled = false;
    }
  });
  return {
    render() {
      synchronize();
      renderTransferHistory();
      updateWalletContext();
    },
    populateForm() {
      synchronize();
      populateTransferForm();
    },
    resetForm() {
      synchronize();
      selectedTransferId = null;
      query("#latest-result").hidden = true;
      query("#transfer-error").hidden = true;
      query("#transfer-form").reset();
      populateTransferForm();
    },
  };
}
