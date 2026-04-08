import { test, expect } from '@playwright/test'

// These tests run WITHOUT pre-authenticated state.
test.use({ storageState: { cookies: [], origins: [] } })

test('shows login page when unauthenticated', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Sign in' })).toBeVisible()
})

test('can register and log in', async ({ page }) => {
  const suffix = Math.random().toString(36).substring(2, 10)
  const email = `pw-login-${suffix}@test.local`
  const password = `testpass-${suffix}`

  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Sign in' })).toBeVisible()

  // Switch to register mode.
  await page.getByRole('button', { name: 'Register' }).click()
  await expect(page.getByRole('heading', { name: 'Create account' })).toBeVisible()

  // Fill and submit registration.
  await page.getByLabel('Name').fill(`Login Test ${suffix}`)
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)
  await page.getByRole('button', { name: 'Create account' }).click()

  // Should land on dashboard.
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible({ timeout: 10_000 })
})

test('shows error for invalid credentials', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Sign in' })).toBeVisible()

  await page.getByLabel('Email').fill('wrong@test.local')
  await page.getByLabel('Password').fill('wrongpassword')
  await page.getByRole('button', { name: 'Sign in' }).click()

  await expect(page.getByText('invalid credentials')).toBeVisible({ timeout: 5_000 })
})
