#!/usr/bin/env node
// Smoke test for an MCP stdio server.
//
// Usage: node scripts/mcp-smoke-test.js [--expect-tools] -- <command> [args...]
//
// Starts <command>, sends an MCP initialize request, the initialized
// notification and a tools/list request over stdin, and checks that stdout
// carries valid JSON-RPC responses and nothing else. With --expect-tools, it
// also requires at least one tool and checks that every tool has a title and
// sets readOnlyHint or destructiveHint. Exits non-zero on any failure.

"use strict";

const { spawn } = require("child_process");

const argv = process.argv.slice(2);
const sep = argv.indexOf("--");
if (sep === -1 || sep === argv.length - 1) {
  console.error("Usage: mcp-smoke-test.js [--expect-tools] -- <command> [args...]");
  process.exit(2);
}
const expectTools = argv.slice(0, sep).includes("--expect-tools");
const [command, ...args] = argv.slice(sep + 1);
const TIMEOUT_MS = 120000;

const child = spawn(command, args, {
  stdio: ["pipe", "pipe", "inherit"],
  shell: process.platform === "win32", // npx is npx.cmd on Windows
});

const send = (msg) => child.stdin.write(JSON.stringify(msg) + "\n");
const pending = new Map();
let buffer = "";
let failed = false;

function finish(ok, message) {
  if (failed) return;
  if (!ok) failed = true;
  console.log(ok ? `PASS ${message}` : `FAIL ${message}`);
  if (!ok || pending.size === 0) {
    child.stdin.end();
    setTimeout(() => {
      child.kill("SIGTERM");
      process.exit(ok ? 0 : 1);
    }, 500).unref();
  }
}

const timer = setTimeout(() => finish(false, `no response within ${TIMEOUT_MS / 1000}s`), TIMEOUT_MS);

child.on("error", (err) => finish(false, `could not start ${command}: ${err.message}`));
child.on("exit", (code, signal) => {
  if (pending.size > 0) finish(false, `server exited early (code ${code}, signal ${signal})`);
});

child.stdout.on("data", (chunk) => {
  buffer += chunk.toString("utf8");
  let nl;
  while ((nl = buffer.indexOf("\n")) !== -1) {
    const line = buffer.slice(0, nl).trim();
    buffer = buffer.slice(nl + 1);
    if (line === "") continue;
    let msg;
    try {
      msg = JSON.parse(line);
    } catch (err) {
      return finish(false, `non-JSON output on stdout: ${line.slice(0, 200)}`);
    }
    if (msg.jsonrpc !== "2.0") return finish(false, `not a JSON-RPC 2.0 message: ${line.slice(0, 200)}`);
    if (msg.id === undefined || !pending.has(msg.id)) continue; // notification or unrelated
    const handler = pending.get(msg.id);
    pending.delete(msg.id);
    if (msg.error) return finish(false, `${handler.name} returned error: ${JSON.stringify(msg.error)}`);
    handler.onResult(msg.result);
    if (pending.size === 0) {
      clearTimeout(timer);
      finish(true, "server answered initialize and tools/list");
    }
  }
});

function request(id, method, params, onResult) {
  pending.set(id, { name: method, onResult });
  send({ jsonrpc: "2.0", id, method, params });
}

request(
  1,
  "initialize",
  {
    protocolVersion: "2025-06-18",
    capabilities: {},
    clientInfo: { name: "auxo-mcp-smoke-test", version: "0" },
  },
  (result) => {
    const info = result.serverInfo || {};
    console.log(`initialize: ${info.name} ${info.version} (protocol ${result.protocolVersion})`);
    send({ jsonrpc: "2.0", method: "notifications/initialized" });
    request(2, "tools/list", {}, (listResult) => {
      const tools = listResult.tools || [];
      console.log(`tools/list: ${tools.length} tools`);
      const problems = [];
      for (const tool of tools) {
        const a = tool.annotations || {};
        const title = tool.title || a.title;
        const hints = [];
        if (a.readOnlyHint === true) hints.push("readOnly");
        if (typeof a.destructiveHint === "boolean") hints.push(`destructive=${a.destructiveHint}`);
        console.log(`  ${tool.name.padEnd(34)} ${String(title).padEnd(38)} ${hints.join(", ")}`);
        if (!title) problems.push(`${tool.name}: no title`);
        if (hints.length === 0) problems.push(`${tool.name}: no readOnlyHint or destructiveHint`);
      }
      if (expectTools && tools.length === 0) problems.push("no tools listed");
      if (problems.length > 0) finish(false, problems.join("; "));
    });
  }
);
