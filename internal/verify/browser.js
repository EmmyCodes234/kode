import { chromium } from 'playwright'
import fs from 'fs'
import path from 'path'

// Parse CLI args
const args = process.argv.slice(2)
let url = 'http://localhost:3000'
let task = ''
let videoPath = ''
let screenshotPath = ''

for (let i = 0; i < args.length; i++) {
  if (args[i] === '--url' && args[i + 1]) {
    url = args[i + 1]
  } else if (args[i] === '--task' && args[i + 1]) {
    task = args[i + 1]
  } else if (args[i] === '--video' && args[i + 1]) {
    videoPath = args[i + 1]
  } else if (args[i] === '--screenshot' && args[i + 1]) {
    screenshotPath = args[i + 1]
  }
}

// Log status for Go engine parser
function streamStep(name, status, message) {
  console.log(JSON.stringify({ step: name, status: status, message: message }))
}

async function run() {
  streamStep('Browser Launch', 'START', 'Launching headless Chromium...')
  const browser = await chromium.launch({ headless: true })
  
  const videoDir = videoPath ? path.dirname(videoPath) : null
  const contextOpts = {}
  if (videoDir) {
    contextOpts.recordVideo = {
      dir: videoDir,
      size: { width: 1280, height: 720 }
    }
  }

  const context = await browser.newContext(contextOpts)
  const page = await context.newPage()

  // Listen to console errors
  const consoleErrors = []
  page.on('pageerror', (err) => {
    consoleErrors.push(err)
    streamStep('Console Error', 'ERROR', err.message)
  })

  try {
    streamStep('Navigation', 'RUNNING', `Navigating to ${url}...`)
    await page.goto(url, { waitUntil: 'networkidle' })
    streamStep('Navigation', 'PASS', `Successfully loaded ${url}`)

    // If task instructions exist
    if (task) {
      // Check if it's dynamic JavaScript code
      if (task.startsWith('javascript:')) {
        const code = task.slice('javascript:'.length)
        streamStep('Custom Verification Script', 'RUNNING', 'Running raw JS script...')
        // We run it by passing the page context wrapped in an async function
        const fn = new Function('page', `return (async () => { ${code} })()`)
        await fn(page)
        streamStep('Custom Verification Script', 'PASS', 'Finished running verification script')
      } else {
        // High-level task instructions: parse dynamic actions split by ';'
        const steps = task.split(';').map(s => s.trim()).filter(Boolean)
        for (const step of steps) {
          const colonIdx = step.indexOf(':')
          if (colonIdx === -1) {
            // Check text existence or click containing text
            streamStep(step, 'RUNNING', `Locating element containing "${step}"...`)
            const loc = page.locator(`text=${step}`).first()
            if (await loc.count() > 0) {
              await loc.click()
              streamStep(step, 'PASS', `Clicked element containing "${step}"`)
            } else {
              const content = await page.textContent('body')
              if (content.includes(step)) {
                streamStep(step, 'PASS', `Verified page contains text "${step}"`)
              } else {
                throw new Error(`Step failed: text "${step}" not found`)
              }
            }
          } else {
            const action = step.slice(0, colonIdx).trim().toLowerCase()
            const value = step.slice(colonIdx + 1).trim()
            if (action === 'click') {
              streamStep(`Click ${value}`, 'RUNNING', `Clicking ${value}...`)
              await page.click(value)
              streamStep(`Click ${value}`, 'PASS', `Successfully clicked ${value}`)
            } else if (action === 'fill' || action === 'type') {
              const pipeIdx = value.indexOf('|')
              if (pipeIdx === -1) {
                throw new Error(`Invalid fill action: ${value} (expected selector|text)`)
              }
              const selector = value.slice(0, pipeIdx).trim()
              const text = value.slice(pipeIdx + 1).trim()
              streamStep(`Type ${selector}`, 'RUNNING', `Filling ${selector} with "${text}"...`)
              await page.fill(selector, text)
              streamStep(`Type ${selector}`, 'PASS', `Successfully filled ${selector}`)
            } else if (action === 'verifytext' || action === 'asserttext') {
              streamStep(`Verify Text: ${value}`, 'RUNNING', `Asserting page contains "${value}"...`)
              const content = await page.textContent('body')
              if (content.includes(value)) {
                streamStep(`Verify Text: ${value}`, 'PASS', `Confirmed text "${value}" exists`)
              } else {
                throw new Error(`Assertion failed: "${value}" not found on page`)
              }
            } else if (action === 'assert' || action === 'assertexists') {
              streamStep(`Assert Exists: ${value}`, 'RUNNING', `Asserting selector "${value}" exists...`)
              await page.waitForSelector(value, { timeout: 5000 })
              streamStep(`Assert Exists: ${value}`, 'PASS', `Selector "${value}" is present`)
            } else if (action === 'wait') {
              streamStep(`Wait: ${value}`, 'RUNNING', `Waiting...`)
              if (isNaN(value)) {
                await page.waitForSelector(value, { timeout: 5000 })
              } else {
                await page.waitForTimeout(parseInt(value, 10))
              }
              streamStep(`Wait: ${value}`, 'PASS', `Wait finished`)
            }
          }
        }
      }
    }

    if (screenshotPath) {
      await page.screenshot({ path: screenshotPath, fullPage: true })
      streamStep('Screenshot', 'PASS', `Saved screenshot to ${screenshotPath}`)
    }

    await context.close()
    
    // Rename video if needed
    if (videoPath) {
      const video = page.video()
      if (video) {
        const tempPath = await video.path()
        fs.mkdirSync(path.dirname(videoPath), { recursive: true })
        fs.renameSync(tempPath, videoPath)
        streamStep('Video Recording', 'PASS', `Saved walkthrough video to ${videoPath}`)
      }
    }

    await browser.close()
    streamStep('Browser Verification', 'PASS', 'UI Verification completed successfully.')
    process.exit(0)

  } catch (err) {
    if (screenshotPath) {
      try {
        await page.screenshot({ path: screenshotPath, fullPage: true })
        streamStep('Screenshot on Failure', 'PASS', `Saved crash screenshot to ${screenshotPath}`)
      } catch (e) {}
    }
    streamStep('Browser Verification', 'FAIL', err.message)
    await browser.close()
    process.exit(1)
  }
}

run()
