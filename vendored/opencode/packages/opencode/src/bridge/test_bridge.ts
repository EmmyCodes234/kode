import { spawn } from "child_process"
import { resolve } from "path"

const KODE_BIN = process.env.KODE_BIN ? resolve(process.env.KODE_BIN) : resolve("bin/kode.exe")
const PROJECT_DIR = resolve(".")

interface Verdict {
  task_id: string
  status: "PASS" | "FAIL"
  rounds_used: number
  applied_hunks: string[]
  failed_hunks: Record<string, string>
}

function runVerify(label: string, hunksFile: string): Promise<{ verdict: Verdict; exitCode: number }> {
  return new Promise((resolvePromise) => {
    const start = Date.now()
    const proc = spawn(KODE_BIN, ["verify", "--input", hunksFile, "--project-dir", PROJECT_DIR])
    let stdout = ""
    let stderr = ""

    proc.stdout.on("data", (d: Buffer) => { stdout += d.toString() })
    proc.stderr.on("data", (d: Buffer) => { stderr += d.toString() })
    proc.on("close", (code: number | null) => {
      const elapsed = Date.now() - start
      const verdict: Verdict = stdout ? JSON.parse(stdout) : { task_id: "error", status: "FAIL", rounds_used: 0, applied_hunks: [], failed_hunks: { error: stderr || "no output" } }
      console.log(`  └─ Response: ${elapsed}ms, exit code ${code}`)
      resolvePromise({ verdict, exitCode: code ?? 1 })
    })
  })
}

async function main() {
  console.log("┌─ Verification Gate Bridge Test")
  console.log("│")
  console.log(`│  Binary: ${KODE_BIN}`)
  console.log(`│  Project: ${PROJECT_DIR}`)
  console.log("│")

  // Test 1: Valid patch should PASS
  console.log("├─ [Test 1] Valid patch (should PASS)")
  const valid = await runVerify("valid", resolve("testdata/valid_patch.json"))
  const validPass = valid.verdict.status === "PASS"
  console.log(`│  ${validPass ? "✓ PASS" : "✗ FAIL"} — verdict: ${valid.verdict.status}`)

  // Test 2: Invalid patch should FAIL
  console.log("├─ [Test 2] Invalid patch (should FAIL)")
  const invalid = await runVerify("invalid", resolve("testdata/invalid_patch.json"))
  const invalidFail = invalid.verdict.status === "FAIL"
  console.log(`│  ${invalidFail ? "✓ FAIL" : "✗ PASS"} — verdict: ${invalid.verdict.status}`)
  if (invalid.verdict.failed_hunks) {
    for (const [hunkId, reason] of Object.entries(invalid.verdict.failed_hunks)) {
      console.log(`│    └─ ${hunkId}: ${reason}`)
    }
  }

  // Summary
  console.log("│")
  if (validPass && invalidFail) {
    console.log("└─ ✓ Bridge validated — gate correctly passes/fails as expected")
    process.exit(0)
  } else {
    console.log("└─ ✗ Bridge validation FAILED — unexpected results")
    process.exit(1)
  }
}

main().catch((err) => {
  console.error("Bridge test error:", err)
  process.exit(1)
})
