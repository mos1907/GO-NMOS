import process from "node:process";

const required = [
  "vite",
  "svelte",
  "postcss",
  "tailwindcss",
  "autoprefixer",
  "@sveltejs/vite-plugin-svelte"
];

function ok(msg) {
  process.stdout.write(`[ok] ${msg}\n`);
}

function warn(msg) {
  process.stderr.write(`[warn] ${msg}\n`);
}

function fail(msg) {
  process.stderr.write(`[fail] ${msg}\n`);
}

ok(`node: ${process.version}`);

let missing = [];
for (const name of required) {
  try {
    const resolved = await import.meta.resolve?.(name);
    // Node's resolve may return null in older versions; fall back to createRequire.
    if (resolved) ok(`${name}: ${resolved}`);
    else ok(`${name}: resolved`);
  } catch {
    missing.push(name);
    fail(`${name}: NOT FOUND`);
  }
}

if (missing.length) {
  warn("");
  warn("Eksik frontend dependency bulundu. Cozum:");
  warn("  cd frontend && npm install");
  warn("");
  warn(`Eksikler: ${missing.join(", ")}`);
  process.exit(1);
}

ok("Frontend dependencies OK.");

