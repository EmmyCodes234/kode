import { Config } from "effect"

function truthy(key: string) {
  const value = process.env[key]?.toLowerCase()
  return value === "true" || value === "1"
}

const KODE_EXPERIMENTAL = truthy("KODE_EXPERIMENTAL")
const copy = process.env["KODE_EXPERIMENTAL_DISABLE_COPY_ON_SELECT"]

export const Flag = {
  OTEL_EXPORTER_OTLP_ENDPOINT: process.env["OTEL_EXPORTER_OTLP_ENDPOINT"],
  OTEL_EXPORTER_OTLP_HEADERS: process.env["OTEL_EXPORTER_OTLP_HEADERS"],

  KODE_AUTO_HEAP_SNAPSHOT: truthy("KODE_AUTO_HEAP_SNAPSHOT"),
  KODE_GIT_BASH_PATH: process.env["KODE_GIT_BASH_PATH"],
  KODE_CONFIG: process.env["KODE_CONFIG"],
  KODE_CONFIG_CONTENT: process.env["KODE_CONFIG_CONTENT"],
  KODE_DISABLE_AUTOUPDATE: truthy("KODE_DISABLE_AUTOUPDATE"),
  KODE_ALWAYS_NOTIFY_UPDATE: truthy("KODE_ALWAYS_NOTIFY_UPDATE"),
  KODE_DISABLE_PRUNE: truthy("KODE_DISABLE_PRUNE"),
  KODE_DISABLE_TERMINAL_TITLE: truthy("KODE_DISABLE_TERMINAL_TITLE"),
  KODE_SHOW_TTFD: truthy("KODE_SHOW_TTFD"),
  KODE_DISABLE_AUTOCOMPACT: truthy("KODE_DISABLE_AUTOCOMPACT"),
  KODE_DISABLE_MODELS_FETCH: truthy("KODE_DISABLE_MODELS_FETCH"),
  KODE_DISABLE_MOUSE: truthy("KODE_DISABLE_MOUSE"),
  KODE_FAKE_VCS: process.env["KODE_FAKE_VCS"],
  KODE_SERVER_PASSWORD: process.env["KODE_SERVER_PASSWORD"],
  KODE_SERVER_USERNAME: process.env["KODE_SERVER_USERNAME"],

  // Experimental
  KODE_EXPERIMENTAL_FILEWATCHER: Config.boolean("KODE_EXPERIMENTAL_FILEWATCHER").pipe(
    Config.withDefault(false),
  ),
  KODE_EXPERIMENTAL_DISABLE_FILEWATCHER: Config.boolean("KODE_EXPERIMENTAL_DISABLE_FILEWATCHER").pipe(
    Config.withDefault(false),
  ),
  KODE_EXPERIMENTAL_DISABLE_COPY_ON_SELECT:
    copy === undefined ? process.platform === "win32" : truthy("KODE_EXPERIMENTAL_DISABLE_COPY_ON_SELECT"),
  KODE_MODELS_URL: process.env["KODE_MODELS_URL"],
  KODE_MODELS_PATH: process.env["KODE_MODELS_PATH"],
  KODE_DB: process.env["KODE_DB"],

  KODE_WORKSPACE_ID: process.env["KODE_WORKSPACE_ID"],
  KODE_EXPERIMENTAL_WORKSPACES: KODE_EXPERIMENTAL || truthy("KODE_EXPERIMENTAL_WORKSPACES"),

  // Evaluated at access time (not module load) because tests, the CLI, and
  // external tooling set these env vars at runtime.
  get KODE_DISABLE_PROJECT_CONFIG() {
    return truthy("KODE_DISABLE_PROJECT_CONFIG")
  },
  get KODE_TUI_CONFIG() {
    return process.env["KODE_TUI_CONFIG"]
  },
  get KODE_CONFIG_DIR() {
    return process.env["KODE_CONFIG_DIR"]
  },
  get KODE_PURE() {
    return truthy("KODE_PURE")
  },
  get KODE_PERMISSION() {
    return process.env["KODE_PERMISSION"]
  },
  get KODE_PLUGIN_META_FILE() {
    return process.env["KODE_PLUGIN_META_FILE"]
  },
  get KODE_CLIENT() {
    return process.env["KODE_CLIENT"] ?? "cli"
  },
}
