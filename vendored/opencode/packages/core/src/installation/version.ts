declare global {
  const KODE_VERSION: string
  const KODE_CHANNEL: string
}

export const InstallationVersion = typeof KODE_VERSION === "string" ? KODE_VERSION : "local"
export const InstallationChannel = typeof KODE_CHANNEL === "string" ? KODE_CHANNEL : "local"
export const InstallationLocal = InstallationChannel === "local"
