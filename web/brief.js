const content = document.querySelector("#brief-content");
const contents = document.querySelector("#brief-contents");

function linkTarget(target) {
  if (target === "README.md")
    return "https://github.com/anfractuous-echo/policy-workbench#readme";
  if (/^https:\/\//.test(target) || /^#[a-z0-9-]+$/.test(target)) return target;
  return null;
}

function inline(element, text) {
  const tokens = /\*\*([^*]+)\*\*|`([^`]+)`|\[([^\]]+)\]\(([^)\s]+)\)/g;
  let start = 0;
  for (const match of text.matchAll(tokens)) {
    element.append(document.createTextNode(text.slice(start, match.index)));
    const target = match[4] && linkTarget(match[4]);
    if (match[3] && !target) element.append(document.createTextNode(match[0]));
    else {
      const node = document.createElement(
        match[1] ? "strong" : match[2] ? "code" : "a",
      );
      node.textContent = match[1] || match[2] || match[3];
      if (target) {
        node.href = target;
        if (target.startsWith("https://")) {
          node.target = "_blank";
          node.rel = "noopener noreferrer";
        }
      }
      element.append(node);
    }
    start = match.index + match[0].length;
  }
  element.append(document.createTextNode(text.slice(start)));
}

function block(tag, text, parent) {
  const node = document.createElement(tag);
  inline(node, text);
  parent.append(node);
  return node;
}

// Render the Markdown features used by TASK.md, with text nodes rather than raw HTML.
function renderBrief(markdown) {
  const lines = markdown.split(/\r?\n/);
  let section = content;
  for (let index = 0; index < lines.length; ) {
    const line = lines[index].trim();
    if (!line) {
      index++;
      continue;
    }
    if (line.startsWith("# ")) {
      document.querySelector("#brief-title").textContent = line.slice(2);
      index++;
      continue;
    }
    if (line.startsWith("## ")) {
      const title = line.slice(3);
      section = document.createElement("section");
      section.className = "brief-section";
      section.id = title
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/-$/, "");
      block("h2", title, section);
      content.append(section);
      const link = block("a", title, contents);
      link.href = `#${section.id}`;
      const template = title.startsWith("Example 1:")
        ? "#allow-flow"
        : title.startsWith("Example 2:")
          ? "#reject-flow"
          : null;
      if (template)
        section.append(
          document.querySelector(template).content.cloneNode(true),
        );
      index++;
      continue;
    }
    if (line.startsWith("|")) {
      const rows = [];
      while (index < lines.length && lines[index].trim().startsWith("|")) {
        const cells = lines[index++]
          .trim()
          .replace(/^\||\|$/g, "")
          .split("|")
          .map((cell) => cell.trim());
        if (!cells.every((cell) => /^:?-+:?$/.test(cell))) rows.push(cells);
      }
      const wrapper = document.createElement("div");
      wrapper.className = "brief-table";
      wrapper.tabIndex = 0;
      wrapper.setAttribute("role", "region");
      wrapper.setAttribute(
        "aria-label",
        `${section.querySelector("h2")?.textContent || "Brief"} table`,
      );
      const table = document.createElement("table");
      for (let rowIndex = 0; rowIndex < rows.length; rowIndex++) {
        const group =
          table.querySelector(rowIndex ? "tbody" : "thead") ||
          table.appendChild(
            document.createElement(rowIndex ? "tbody" : "thead"),
          );
        const row = group.appendChild(document.createElement("tr"));
        for (const value of rows[rowIndex]) {
          const cell = block(rowIndex ? "td" : "th", value, row);
          if (!rowIndex) cell.scope = "col";
        }
      }
      wrapper.append(table);
      section.append(wrapper);
      continue;
    }
    if (/^(?:- |\d+\. )/.test(line)) {
      const ordered = /^\d+\. /.test(line);
      const list = section.appendChild(
        document.createElement(ordered ? "ol" : "ul"),
      );
      const pattern = ordered ? /^\d+\. / : /^- /;
      while (index < lines.length && pattern.test(lines[index].trim()))
        block("li", lines[index++].trim().replace(pattern, ""), list);
      continue;
    }
    const paragraph = [];
    while (
      index < lines.length &&
      lines[index].trim() &&
      !/^(?:#|\||- |\d+\. )/.test(lines[index].trim())
    )
      paragraph.push(lines[index++].trim());
    if (paragraph.length) block("p", paragraph.join(" "), section);
    else {
      block("p", line, section);
      index++;
    }
  }
}

try {
  const response = await fetch("/brief.md", { cache: "no-store" });
  if (!response.ok) throw new Error("Brief unavailable");
  renderBrief(await response.text());
} catch {
  document.querySelector("#brief-error").hidden = false;
} finally {
  document.querySelector("#brief-loading").hidden = true;
}
