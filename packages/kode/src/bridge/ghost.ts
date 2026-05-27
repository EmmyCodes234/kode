// Ghost Branch bridge — spawns `kode loop --branches N` as a subprocess
// and streams progress events back to the TUI via structured JSON on stderr.

import { spawn } from "child_process"
import { resolve } from "path"
import { resolveKodeBinary } from "./gatekeeper"

export interface GhostBranchConfig {
  task: string
  branches: number
  projectDir: string
  model?: string
  testCommand?: string
}

export interface GhostBranchResult {
  id: string
  strategy: string
  status: "PASS" | "FAIL"
  score: number
  duration: number
  blast_radius: number
  gates_passed: number
  token_cost: number
  error?: string
}

export interface GhostSummary {
  status: string
  task: string
  branches: number
  winner: GhostBranchResult | null
  total_time: number
  total_cost: number
  all_branches: GhostBranchResult[]
}

export type GhostProgressEvent =
  | { phase: "started"; task: string; branchCount: number }
  | { phase: "branch_update"; branchId: string; status: string; detail: string }
  | { phase: "scoring"; message: string }
  | { phase: "complete"; summary: GhostSummary }
  | { phase: "error"; message: string }

export type GhostProgressCallback = (event: GhostProgressEvent) => void

/**
 * Run Ghost Branches via the Go engine.
 *
 * Spawns `kode loop <task> --branches N` and parses the JSON output.
 * Progress events are emitted to the callback for TUI rendering.
 */
export async function runGhostBranches(
  config: GhostBranchConfig,
  onProgress?: GhostProgressCallback,
): Promise<GhostSummary> {
  const binary = resolveKodeBinary()

  const args = [
    "loop",
    config.task,
    "--branches", String(config.branches),
    "--project-dir", config.projectDir,
  ]

  if (config.model) {
    args.push("--model", config.model)
  }
  if (config.testCommand) {
    args.push("--test-command", config.testCommand)
  }

  onProgress?.({
    phase: "started",
    task: config.task,
    branchCount: config.branches,
  })

  return new Promise<GhostSummary>((resolvePromise, reject) => {
    const proc = spawn(binary, args, {
      timeout: 300_000, // 5 minute timeout for ghost branches
      stdio: ["ignore", "pipe", "pipe"],
      cwd: config.projectDir,
    })

    let stdoutBuf = ""

    proc.stdout!.on("data", (d: Buffer) => {
      stdoutBuf += d.toString()
    })

    proc.stderr!.on("data", (d: Buffer) => {
      const lines = d.toString().split("\n").filter((l) => l.trim())
      for (const line of lines) {
        // Parse structured progress from stderr
        const winnerMatch = line.match(/Winner: (\w+) \((\w+)\) — score ([\d.]+)/)
        if (winnerMatch && onProgress) {
          onProgress({
            phase: "scoring",
            message: `Winner: ${winnerMatch[1]} (${winnerMatch[2]}) — score ${winnerMatch[3]}`,
          })
          continue
        }

        const branchMatch = line.match(/\[([+x!])\] (\w+) \((\w+)\) — (\w+) — Score: ([\d.]+)/)
        if (branchMatch && onProgress) {
          onProgress({
            phase: "branch_update",
            branchId: branchMatch[2],
            status: branchMatch[4],
            detail: `${branchMatch[3]} — Score: ${branchMatch[5]}${branchMatch[1] === "+" ? " — WINNER" : ""}`,
          })
          continue
        }

        const ghostHeader = line.match(/Ghost Branch Mode/)
        if (ghostHeader && onProgress) {
          onProgress({
            phase: "started",
            task: config.task,
            branchCount: config.branches,
          })
        }
      }
    })

    proc.on("error", (err: NodeJS.ErrnoException) => {
      if (err.code === "ENOENT") {
        reject(new Error(`kode binary not found at ${binary}`))
      } else {
        reject(err)
      }
    })

    proc.on("close", (code) => {
      try {
        const parsed = JSON.parse(stdoutBuf)

        const summary: GhostSummary = {
          status: parsed.status ?? "FAIL",
          task: parsed.task ?? config.task,
          branches: parsed.branches ?? config.branches,
          winner: parsed.winner ?? null,
          total_time: parsed.total_time ?? 0,
          total_cost: parsed.total_cost ?? 0,
          all_branches: parsed.all_branches ?? [],
        }

        onProgress?.({ phase: "complete", summary })
        resolvePromise(summary)
      } catch (parseErr) {
        const error = code !== 0
          ? `Ghost branch process exited with code ${code}`
          : `Failed to parse ghost branch output: ${parseErr}`

        onProgress?.({ phase: "error", message: error })
        reject(new Error(error))
      }
    })
  })
}
