#!/usr/bin/env node
// Launcher for the AUXO MCP server.
//
// The server is a Go binary shipped in one npm package per platform, for
// example @on2it/auxo-mcp-darwin-arm64. npm installs only the package that
// matches the current os and cpu (see "optionalDependencies" in
// package.json). This script finds that binary and runs it with the same
// arguments and environment.
//
// Stdout is the MCP stdio channel, so this script never writes to it. All
// diagnostics go to stderr.

"use strict";

const { spawn } = require("child_process");
const fs = require("fs");
const path = require("path");

// The npm scope is derived from this package's own name, so renaming the
// scope only requires changing scripts/npm-prepare.sh.
const SCOPE = require("../package.json").name.split("/")[0];
const platform = `${process.platform}-${process.arch}`;
const platformPackage = `${SCOPE}/auxo-mcp-${platform}`;
const binaryName = process.platform === "win32" ? "auxo-mcp-server.exe" : "auxo-mcp-server";

function fail(message) {
  process.stderr.write(`auxo-mcp: ${message}\n`);
  process.exit(1);
}

function findBinary() {
  let packageJson;
  try {
    packageJson = require.resolve(`${platformPackage}/package.json`);
  } catch (err) {
    fail(
      `no AUXO MCP server binary is installed for ${platform}.\n` +
        `The package ${platformPackage} is missing. It is an optional dependency, so it is skipped when\n` +
        `npm runs with --no-optional, --omit=optional or --ignore-optional, and it does not exist for\n` +
        `unsupported platforms. Supported: darwin-arm64, darwin-x64, linux-arm64, linux-x64, win32-x64.\n` +
        `Reinstall without those flags, for example: npm install @on2it/auxo-mcp`
    );
  }
  const binary = path.join(path.dirname(packageJson), "bin", binaryName);
  if (!fs.existsSync(binary)) {
    fail(`${platformPackage} is installed but ${binary} is missing. Reinstall @on2it/auxo-mcp.`);
  }
  return binary;
}

// npm tarballs do not always keep the executable bit, so restore it.
function ensureExecutable(binary) {
  if (process.platform === "win32") return;
  try {
    fs.accessSync(binary, fs.constants.X_OK);
  } catch (err) {
    try {
      fs.chmodSync(binary, 0o755);
    } catch (chmodErr) {
      process.stderr.write(`auxo-mcp: could not make ${binary} executable: ${chmodErr.message}\n`);
    }
  }
}

function run() {
  const binary = findBinary();
  ensureExecutable(binary);

  const child = spawn(binary, process.argv.slice(2), {
    stdio: "inherit",
    env: process.env,
  });

  const forward = (signal) => {
    if (child.exitCode === null && child.signalCode === null) child.kill(signal);
  };
  const signals = ["SIGINT", "SIGTERM"];
  for (const signal of signals) process.on(signal, forward);

  child.on("error", (err) => fail(`failed to start ${binary}: ${err.message}`));

  child.on("exit", (code, signal) => {
    for (const s of signals) process.removeListener(s, forward);
    if (signal) {
      // Exit the same way the server did so the caller sees the signal.
      process.kill(process.pid, signal);
      return;
    }
    process.exit(code === null ? 1 : code);
  });
}

run();
