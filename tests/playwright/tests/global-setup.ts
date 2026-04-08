import { test as setup, expect } from '@playwright/test'
import * as fs from 'fs'
import * as path from 'path'

setup('authenticate', async ({ page }) => {
  const suffix = Math.random().toString(36).substring(2, 10)
  const email = `pw-e2e-${suffix}@test.local`
  const password = `testpass-${suffix}`

  // Navigate to the app — should show login page.
  await page.goto('/')
  await expect(page.getByRole('heading', { name: /sign in|create account/i })).toBeVisible({ timeout: 10_000 })

  // Switch to register mode if needed.
  const registerLink = page.getByRole('button', { name: 'Register' })
  if (await registerLink.isVisible()) {
    await registerLink.click()
  }

  // Fill registration form.
  const nameField = page.getByLabel('Name')
  if (await nameField.isVisible()) {
    await nameField.fill(`PW User ${suffix}`)
  }
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)

  // Submit (button text depends on mode).
  const createBtn = page.getByRole('button', { name: 'Create account' })
  const signInBtn = page.getByRole('button', { name: 'Sign in' })
  if (await createBtn.isVisible()) {
    await createBtn.click()
  } else {
    await signInBtn.click()
  }

  // Wait for redirect to dashboard.
  await expect(page.getByText('Dashboard')).toBeVisible({ timeout: 10_000 })

  // Save auth state.
  const authDir = path.join(__dirname, '.auth')
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true })
  }
  await page.context().storageState({ path: path.join(authDir, 'user.json') })
})
