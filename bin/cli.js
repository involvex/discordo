#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

// Get the directory where this script is located
const binDir = path.dirname(__filename);
const parentDir = path.dirname(binDir);

// Determine the binary name based on platform
const isWindows = process.platform === 'win32';
const binaryName = isWindows ? 'disgo-cli.exe' : 'disgo-cli';
const binaryPath = path.join(parentDir, binaryName);

// All arguments after the script name
const args = process.argv.slice(2);

const child = spawn(binaryPath, args, {
  stdio: 'inherit',
  cwd: process.cwd()
});

child.on('close', (code) => {
  process.exit(code);
});

child.on('error', (err) => {
  console.error('Failed to start disgo-cli:', err);
  process.exit(1);
});
