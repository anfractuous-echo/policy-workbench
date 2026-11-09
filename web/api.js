async function requestJSON(path, method = "GET", body) {
  const response = await fetch(path, {
    method,
    headers: body === undefined ? {} : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await response.text();
  let data;
  try {
    data = JSON.parse(text);
  } catch {
    throw new Error(`Request failed (${response.status}). Refresh and retry.`);
  }
  if (!response.ok)
    throw new Error(data.error || `Request failed (${response.status}).`);
  return data;
}
export const api = {
  getWorkspace: (profileId) =>
    requestJSON(`/api/state?profile_id=${encodeURIComponent(profileId)}`),
  createPolicy: (policy) => requestJSON("/api/policies", "POST", policy),
  deletePolicy: (profileId, policyId) =>
    requestJSON(
      `/api/policies/${encodeURIComponent(policyId)}?profile_id=${encodeURIComponent(profileId)}`,
      "DELETE",
    ),
  reorderPolicies: (profileId, ids) =>
    requestJSON("/api/policies/order", "PUT", { profile_id: profileId, ids }),
  createTransfer: (transfer) => requestJSON("/api/transfers", "POST", transfer),
  applyTransferAction: (transferId, command) =>
    requestJSON(
      `/api/transfers/${encodeURIComponent(transferId)}/actions`,
      "POST",
      command,
    ),
  advanceClock: (hours) => requestJSON("/api/clock", "POST", { hours }),
  resetWorkspace: () => requestJSON("/api/reset", "POST", {}),
};
