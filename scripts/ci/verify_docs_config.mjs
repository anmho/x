#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";

const repoRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), "../..");
const docsBase = path.join(repoRoot, "docs", "mintlify");

const configCandidates = ["docs.json", "mint.json"];
const configName = configCandidates.find((name) => fs.existsSync(path.join(docsBase, name)));

if (!configName) {
  console.error("error: missing docs/mintlify/docs.json or docs/mintlify/mint.json");
  process.exit(1);
}

const configPath = path.join(docsBase, configName);
const raw = fs.readFileSync(configPath, "utf8");
let config;
try {
  config = JSON.parse(raw);
} catch (error) {
  console.error(`error: invalid JSON in ${configName}: ${error.message}`);
  process.exit(1);
}

const pages = [];
const collectPages = (node) => {
  if (!node) return;

  if (Array.isArray(node)) {
    for (const item of node) {
      collectPages(item);
    }
    return;
  }

  if (typeof node === "string") {
    pages.push(node);
    return;
  }

  if (typeof node === "object") {
    collectPages(node.pages);
    collectPages(node.groups);
    collectPages(node.tabs);
    collectPages(node.navigation);
  }
};

collectPages(config.navigation ?? config);

const missing = pages.filter((page) => {
  const mdx = path.join(docsBase, `${page}.mdx`);
  const md = path.join(docsBase, `${page}.md`);
  return !fs.existsSync(mdx) && !fs.existsSync(md);
});

if (missing.length > 0) {
  console.error(`error: missing Mintlify pages: ${missing.join(", ")}`);
  process.exit(1);
}

if (configName === "docs.json" && typeof config.theme !== "string") {
  console.error("error: docs.json must define a string `theme`.");
  process.exit(1);
}

console.log(`docs config verified: ${configName} (${pages.length} nav page entries)`);
