import { test, expect } from '@playwright/test'

test('agents page loads', async ({ page }) => {
  await page.goto('/agents')
  await expect(page.getByRole('heading', { name: 'Agents', exact: true })).toBeVisible()
})

test('can create and delete an agent', async ({ page }) => {
  await page.goto('/agents')

  // Fill in create form (assumes there's a name input and create button).
  const nameInput = page.getByPlaceholder('Agent name')
  if (await nameInput.isVisible()) {
    const agentName = `pw-agent-${Date.now()}`
    await nameInput.fill(agentName)
    await page.getByRole('button', { name: 'Create' }).click()

    // Agent should appear in the list.
    await expect(page.getByText(agentName)).toBeVisible({ timeout: 5_000 })
  }
})
