import { query, escapeHTML } from "./ui.js";
import { lookupName } from "./presentation.js";

export function bindCatalogDialog(workspace) {
  query("#catalog-button").addEventListener("click", () => {
    const state = workspace.getState();
    if (!state) return;
    const catalog = state.catalog;
    const inProfile = (entries) =>
      entries.filter((entry) => entry.profile_id === state.profile_id);
    const entry = (name, detail) =>
      `<div class="catalog-entry">${escapeHTML(name)}<small>${escapeHTML(detail)}</small></div>`;
    const section = (title, entries) =>
      `<section class="catalog-section"><h3>${escapeHTML(title)}</h3>${entries.join("")}</section>`;
    query("#catalog-body").innerHTML =
      section(
        "Profile members",
        inProfile(catalog.people).map((member) =>
          entry(member.name, member.id),
        ),
      ) +
      section(
        "Wallet groups",
        inProfile(catalog.wallet_groups).map((group) =>
          entry(
            group.name,
            group.members
              .map((id) => lookupName(catalog.wallets, id))
              .join(", "),
          ),
        ),
      ) +
      section(
        "Recipient groups",
        inProfile(catalog.recipient_groups).map((group) =>
          entry(
            group.name,
            group.members
              .map((id) => lookupName(catalog.recipients, id))
              .join(", "),
          ),
        ),
      ) +
      section(
        "Recipients",
        inProfile(catalog.recipients).map((recipient) =>
          entry(recipient.name, recipient.id),
        ),
      );
    query("#catalog-dialog").showModal();
  });
}
