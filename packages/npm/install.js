const { existsSync, chmodSync, mkdirSync, createWriteStream, readFileSync, writeFileSync } = require("fs");
const { join } = require("path");
const os = require("os");
const https = require("https");

const pkg = require("./package.json");
const version = pkg.version;
const repo = "sicario-labs/kode";

const platformMap = {
  win32: { os: "windows", ext: ".exe" },
  darwin: { os: "darwin", ext: "" },
  linux: { os: "linux", ext: "" },
};

const archMap = {
  x64: "amd64",
  arm64: "arm64",
};

const plat = platformMap[process.platform];
const arch = archMap[process.arch];

if (!plat || !arch) {
  console.error(`Unsupported platform: ${process.platform} ${process.arch}`);
  process.exit(1);
}

const binaryName = `kode-${plat.os}-${arch}${plat.ext}`;
const installDir = join(__dirname, "bin");
const binaryPath = join(installDir, `kode${plat.ext}`);
const tuiDir = join(os.homedir(), ".kode", "tui");

const versionedURL = (name) => `https://github.com/${repo}/releases/download/v${version}/${name}`;

let exitCode = 0;

main().catch((err) => {
  console.error(err.message);
  exitCode = 1;
}).finally(() => process.exit(exitCode));

async function main() {
  // Step 1: Download Go binary
  if (!existsSync(binaryPath)) {
    mkdirSync(installDir, { recursive: true });
    console.log(`Downloading Kode v${version} (${binaryName})...`);
    try {
      await download(versionedURL(binaryName), binaryPath);
      try { chmodSync(binaryPath, 0o755); } catch {}
      console.log(`Kode binary installed to ${binaryPath}`);
    } catch (err) {
      console.error(`Binary download failed: ${err.message}`);
      exitCode = 1;
    }
  }

  // Step 2: Download compiled TUI binary to ~/.kode/tui/bin/
  const versionFile = join(tuiDir, ".kode-tui-version");
  let installedVersion = "";
  try { installedVersion = readFileSync(versionFile, "utf8").trim(); } catch {}
  if (installedVersion !== version) {
    const binDir = join(tuiDir, "bin");
    mkdirSync(binDir, { recursive: true });
    const tuiBinaryName = `kode-tui-${plat.os}-${process.arch === "x64" ? "x64" : "arm64"}${plat.ext}`;
    const tuiPath = join(binDir, `kode-tui${plat.ext}`);
    console.log(`Downloading Kode TUI binary (${tuiBinaryName})...`);
    try {
      await download(versionedURL(tuiBinaryName), tuiPath);
      try { chmodSync(tuiPath, 0o755); } catch {}
      writeFileSync(versionFile, version, "utf8");
      console.log(`TUI binary installed to ${tuiPath}`);
    } catch (err) {
      console.error(`TUI binary download failed: ${err.message}`);
    }
  }
}

// --- helpers ---

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = createWriteStream(dest);
    const req = https.get(url, { headers: { "User-Agent": "kode-installer" } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        file.close();
        download(res.headers.location, dest).then(resolve, reject);
        return;
      }
      if (res.statusCode !== 200) {
        file.close();
        reject(new Error(`HTTP ${res.statusCode}`));
        return;
      }
      res.pipe(file);
      file.on("finish", () => file.close((err) => err ? reject(err) : resolve()));
    });
    req.on("error", (err) => { file.close(); reject(err); });
  });
}


