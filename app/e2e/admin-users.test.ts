import { test, expect, type Page } from '@playwright/test';
import { loginAsAdmin } from './test-utils';

// Test data constants
const TEST_USERS = {
	USER_1: { name: 'Test User Alpha', phone: '+15551111111', role: 'owl' },
	USER_2: { name: 'Test User Beta', phone: '+15552222222', role: 'supervisor' },
	USER_3: { name: 'Test User Gamma', phone: '+15553333333', role: 'owl' },
	USER_4: { name: 'Test User Delta', phone: '+15554444444', role: 'guest' },
	ADMIN_USER: { name: 'Test Admin', phone: '+15555555555', role: 'admin' }
};

const USER_ROLES = ['guest', 'owl', 'supervisor', 'admin'];

async function navigateToUsers(page: Page) {
	await page.goto('/admin/users');
	await expect(page.locator('h1, h2')).toContainText('Users');
}

async function createUser(page: Page, name: string, phone: string, role: string = 'owl') {
	await page.click('button:has-text("Create User")');

	// Wait for form to be visible
	await expect(page.locator('form')).toBeVisible();

	// Fill form
	await page.fill('input[name="name"]', name);
	await page.fill('input[name="phone"]', phone);
	await page.selectOption('select[name="role"]', role);

	// Submit form
	await page.click('button[type="submit"]');

	// Verify success
	await expect(page.locator('.toast')).toContainText('successfully');
}

async function deleteUser(page: Page, userName: string) {
	// Find user in list and click delete
	const userRow = page.locator(`[data-testid="user-item"]:has-text("${userName}")`);
	await userRow.locator('button:has-text("Delete")').click();

	// Confirm deletion
	await page.click('button:has-text("Yes")');

	// Verify success
	await expect(page.locator('.toast')).toContainText('deleted successfully');
}

test.describe('Admin Users Management - CRUD Operations', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToUsers(page);
	});

	test('Admin can create a new user with all fields', async ({ page }) => {
		await createUser(page, TEST_USERS.USER_1.name, TEST_USERS.USER_1.phone, TEST_USERS.USER_1.role);

		// Verify user appears in list
		await expect(page.locator(`text=${TEST_USERS.USER_1.name}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_USERS.USER_1.phone}`)).toBeVisible();
	});

	test('Admin can create users with different roles', async ({ page }) => {
		for (const role of USER_ROLES) {
			const testUser = {
				name: `Test ${role} User`,
				phone: `+1555${Math.floor(Math.random() * 9000) + 1000}000`,
				role: role
			};

			await createUser(page, testUser.name, testUser.phone, testUser.role);

			// Verify user appears with correct role indicator
			await expect(page.locator(`text=${testUser.name}`)).toBeVisible();

			// Check role-specific icon (admin has shield, others have user icon)
			if (role === 'admin') {
				await expect(
					page.locator(
						`[data-testid="user-item"]:has-text("${testUser.name}") svg[class*="shield"]`
					)
				).toBeVisible();
			} else {
				await expect(
					page.locator(`[data-testid="user-item"]:has-text("${testUser.name}") svg[class*="user"]`)
				).toBeVisible();
			}
		}
	});

	test('Admin can view user details', async ({ page }) => {
		// Create a user first
		await createUser(page, TEST_USERS.USER_2.name, TEST_USERS.USER_2.phone, TEST_USERS.USER_2.role);

		// Click on user to view details
		await page.click(`[data-testid="user-item"]:has-text("${TEST_USERS.USER_2.name}")`);

		// Verify user details are displayed
		await expect(page.locator(`text=${TEST_USERS.USER_2.name}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_USERS.USER_2.phone}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_USERS.USER_2.role}`)).toBeVisible();
	});

	test('Admin can edit user information', async ({ page }) => {
		// Create a user first
		await createUser(page, TEST_USERS.USER_3.name, TEST_USERS.USER_3.phone, TEST_USERS.USER_3.role);

		// Click on user to edit
		await page.click(`[data-testid="user-item"]:has-text("${TEST_USERS.USER_3.name}")`);

		// Wait for edit form
		await expect(page.locator('form')).toBeVisible();

		// Modify user information
		const updatedName = TEST_USERS.USER_3.name + ' Updated';
		await page.fill('input[name="name"]', updatedName);
		await page.selectOption('select[name="role"]', 'supervisor');

		// Save changes
		await page.click('button:has-text("Save")');

		// Verify success
		await expect(page.locator('.toast')).toContainText('successfully');

		// Verify changes are reflected
		await expect(page.locator(`text=${updatedName}`)).toBeVisible();
		await expect(
			page.locator(`[data-testid="user-item"]:has-text("${updatedName}") svg[class*="shield"]`)
		).toBeVisible();
	});

	test('Admin can delete a user', async ({ page }) => {
		// Create a user first
		await createUser(page, TEST_USERS.USER_4.name, TEST_USERS.USER_4.phone, TEST_USERS.USER_4.role);

		// Verify user exists
		await expect(page.locator(`text=${TEST_USERS.USER_4.name}`)).toBeVisible();

		// Delete user
		await deleteUser(page, TEST_USERS.USER_4.name);

		// Verify user no longer appears in list
		await expect(page.locator(`text=${TEST_USERS.USER_4.name}`)).not.toBeVisible();
	});
});

test.describe('Admin Users Management - Form Validation', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToUsers(page);
	});

	test('Form validation prevents creating user without required fields', async ({ page }) => {
		await page.click('button:has-text("Create User")');

		// Try to submit empty form
		await page.click('button[type="submit"]');

		// Should show validation errors
		await expect(page.locator('text=Name is required')).toBeVisible();
		await expect(page.locator('text=Phone is required')).toBeVisible();
	});

	test('Form validation prevents invalid phone numbers', async ({ page }) => {
		await page.click('button:has-text("Create User")');

		// Fill form with invalid phone
		await page.fill('input[name="name"]', 'Test User');
		await page.fill('input[name="phone"]', 'invalid-phone');
		await page.selectOption('select[name="role"]', 'owl');

		await page.click('button[type="submit"]');

		// Should show phone validation error
		await expect(page.locator('text=Invalid phone number')).toBeVisible();
	});

	test('Form validation prevents duplicate phone numbers', async ({ page }) => {
		// Create first user
		await createUser(page, 'First User', '+15556666666', 'owl');

		// Try to create second user with same phone
		await page.click('button:has-text("Create User")');
		await page.fill('input[name="name"]', 'Second User');
		await page.fill('input[name="phone"]', '+15556666666');
		await page.selectOption('select[name="role"]', 'owl');

		await page.click('button[type="submit"]');

		// Should show duplicate phone error
		await expect(page.locator('text=Phone number already exists')).toBeVisible();
	});

	test('Form validation handles long names gracefully', async ({ page }) => {
		await page.click('button:has-text("Create User")');

		// Try very long name
		const longName = 'A'.repeat(256);
		await page.fill('input[name="name"]', longName);
		await page.fill('input[name="phone"]', '+15557777777');
		await page.selectOption('select[name="role"]', 'owl');

		await page.click('button[type="submit"]');

		// Should either truncate or show length validation error
		await expect(page.locator('.toast, text=too long, text=exceeds')).toBeVisible();
	});
});

test.describe('Admin Users Management - Search and Filtering', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToUsers(page);

		// Create test users for filtering
		await createUser(page, 'Alice Johnson', '+15551111111', 'owl');
		await createUser(page, 'Bob Smith', '+15552222222', 'supervisor');
		await createUser(page, 'Carol Williams', '+15553333333', 'admin');
		await createUser(page, 'David Brown', '+15554444444', 'guest');
	});

	test('Can search users by name', async ({ page }) => {
		// Search for specific user
		await page.fill('input[placeholder*="Search"]', 'Alice');

		// Should show only matching user
		await expect(page.locator('text=Alice Johnson')).toBeVisible();
		await expect(page.locator('text=Bob Smith')).not.toBeVisible();
		await expect(page.locator('text=Carol Williams')).not.toBeVisible();
		await expect(page.locator('text=David Brown')).not.toBeVisible();
	});

	test('Can search users by phone number', async ({ page }) => {
		// Search by phone
		await page.fill('input[placeholder*="Search"]', '+15552222222');

		// Should show only matching user
		await expect(page.locator('text=Bob Smith')).toBeVisible();
		await expect(page.locator('text=Alice Johnson')).not.toBeVisible();
	});

	test('Search is case insensitive', async ({ page }) => {
		// Search with different case
		await page.fill('input[placeholder*="Search"]', 'alice');

		// Should still find the user
		await expect(page.locator('text=Alice Johnson')).toBeVisible();
	});

	test('Can clear search filter', async ({ page }) => {
		// Search for user
		await page.fill('input[placeholder*="Search"]', 'Alice');
		await expect(page.locator('text=Alice Johnson')).toBeVisible();
		await expect(page.locator('text=Bob Smith')).not.toBeVisible();

		// Clear search
		await page.fill('input[placeholder*="Search"]', '');

		// Should show all users again
		await expect(page.locator('text=Alice Johnson')).toBeVisible();
		await expect(page.locator('text=Bob Smith')).toBeVisible();
		await expect(page.locator('text=Carol Williams')).toBeVisible();
		await expect(page.locator('text=David Brown')).toBeVisible();
	});

	test('Search with no results shows appropriate message', async ({ page }) => {
		// Search for non-existent user
		await page.fill('input[placeholder*="Search"]', 'NonExistentUser');

		// Should show no results message
		await expect(page.locator('text=No users found')).toBeVisible();
	});
});

test.describe('Admin Users Management - Bulk Operations', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToUsers(page);

		// Create multiple test users for bulk operations
		await createUser(page, 'Bulk User 1', '+15551111111', 'owl');
		await createUser(page, 'Bulk User 2', '+15552222222', 'owl');
		await createUser(page, 'Bulk User 3', '+15553333333', 'supervisor');
		await createUser(page, 'Bulk User 4', '+15554444444', 'guest');
	});

	test('Can enable bulk selection mode', async ({ page }) => {
		// Enable bulk mode
		await page.click('input[id="bulk-mode"]');

		// Should show checkboxes for all users
		const checkboxes = page.locator('input[type="checkbox"]');
		const checkboxCount = await checkboxes.count();
		expect(checkboxCount).toBeGreaterThan(4); // At least 4 users + select all
	});

	test('Can select individual users in bulk mode', async ({ page }) => {
		// Enable bulk mode
		await page.click('input[id="bulk-mode"]');

		// Select specific users
		await page.click(`[data-testid="user-item"]:has-text("Bulk User 1") input[type="checkbox"]`);
		await page.click(`[data-testid="user-item"]:has-text("Bulk User 2") input[type="checkbox"]`);

		// Should show bulk actions toolbar
		await expect(page.locator('.bulk-actions, [data-testid="bulk-actions"]')).toBeVisible();

		// Should show selection count
		await expect(page.locator('text=2 selected, text=2 of')).toBeVisible();
	});

	test('Can select all users', async ({ page }) => {
		// Enable bulk mode
		await page.click('input[id="bulk-mode"]');

		// Click select all
		await page.click('input[type="checkbox"]:has-text("Select All"), label:has-text("Select All")');

		// Should show all users selected
		await expect(page.locator('text=4 selected, text=4 of, text=All')).toBeVisible();
	});

	test('Can deselect all users', async ({ page }) => {
		// Enable bulk mode and select all
		await page.click('input[id="bulk-mode"]');
		await page.click('input[type="checkbox"]:has-text("Select All"), label:has-text("Select All")');

		// Deselect all
		await page.click(
			'input[type="checkbox"]:has-text("Deselect All"), label:has-text("Deselect All")'
		);

		// Should show no users selected
		await expect(page.locator('[data-testid="bulk-actions"]')).not.toBeVisible();
	});

	test('Can bulk delete multiple users', async ({ page }) => {
		// Enable bulk mode
		await page.click('input[id="bulk-mode"]');

		// Select users to delete
		await page.click(`[data-testid="user-item"]:has-text("Bulk User 3") input[type="checkbox"]`);
		await page.click(`[data-testid="user-item"]:has-text("Bulk User 4") input[type="checkbox"]`);

		// Click bulk delete
		await page.click('button:has-text("Delete Selected")');

		// Confirm deletion
		await page.click('button:has-text("Yes")');

		// Verify success
		await expect(page.locator('.toast')).toContainText('deleted successfully');

		// Verify users are removed
		await expect(page.locator('text=Bulk User 3')).not.toBeVisible();
		await expect(page.locator('text=Bulk User 4')).not.toBeVisible();

		// Verify other users remain
		await expect(page.locator('text=Bulk User 1')).toBeVisible();
		await expect(page.locator('text=Bulk User 2')).toBeVisible();
	});

	test('Can exit bulk mode', async ({ page }) => {
		// Enable bulk mode
		await page.click('input[id="bulk-mode"]');

		// Select some users
		await page.click(`[data-testid="user-item"]:has-text("Bulk User 1") input[type="checkbox"]`);

		// Exit bulk mode
		await page.click('input[id="bulk-mode"]');

		// Should hide checkboxes and bulk actions
		await expect(page.locator('[data-testid="bulk-actions"]')).not.toBeVisible();
		const checkboxes = page.locator(`[data-testid="user-item"] input[type="checkbox"]`);
		const checkboxCount = await checkboxes.count();
		expect(checkboxCount).toBe(0);
	});
});

test.describe('Admin Users Management - Role Management', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToUsers(page);
	});

	test('Can change user role from guest to owl', async ({ page }) => {
		// Create guest user
		await createUser(page, 'Guest User', '+15551111111', 'guest');

		// Edit user and change role
		await page.click(`[data-testid="user-item"]:has-text("Guest User")`);
		await page.selectOption('select[name="role"]', 'owl');
		await page.click('button:has-text("Save")');

		// Verify role change
		await expect(page.locator('.toast')).toContainText('successfully');

		// Verify role icon changed (guest to owl should change icon)
		await page.goto('/admin/users'); // Refresh to see updated list
		await expect(
			page.locator(`[data-testid="user-item"]:has-text("Guest User") svg[class*="user"]`)
		).toBeVisible();
	});

	test('Can promote owl to supervisor', async ({ page }) => {
		// Create owl user
		await createUser(page, 'Owl User', '+15552222222', 'owl');

		// Edit user and promote to supervisor
		await page.click(`[data-testid="user-item"]:has-text("Owl User")`);
		await page.selectOption('select[name="role"]', 'supervisor');
		await page.click('button:has-text("Save")');

		// Verify promotion
		await expect(page.locator('.toast')).toContainText('successfully');
	});

	test('Can promote user to admin', async ({ page }) => {
		// Create supervisor user
		await createUser(page, 'Supervisor User', '+15553333333', 'supervisor');

		// Edit user and promote to admin
		await page.click(`[data-testid="user-item"]:has-text("Supervisor User")`);
		await page.selectOption('select[name="role"]', 'admin');
		await page.click('button:has-text("Save")');

		// Verify promotion
		await expect(page.locator('.toast')).toContainText('successfully');

		// Verify admin icon appears
		await page.goto('/admin/users'); // Refresh to see updated list
		await expect(
			page.locator(`[data-testid="user-item"]:has-text("Supervisor User") svg[class*="shield"]`)
		).toBeVisible();
	});

	test('Can demote admin to regular user', async ({ page }) => {
		// Create admin user
		await createUser(page, 'Admin User', '+15554444444', 'admin');

		// Edit user and demote
		await page.click(`[data-testid="user-item"]:has-text("Admin User")`);
		await page.selectOption('select[name="role"]', 'owl');
		await page.click('button:has-text("Save")');

		// Verify demotion
		await expect(page.locator('.toast')).toContainText('successfully');

		// Verify icon changed from shield to user
		await page.goto('/admin/users'); // Refresh to see updated list
		await expect(
			page.locator(`[data-testid="user-item"]:has-text("Admin User") svg[class*="user"]`)
		).toBeVisible();
	});
});

test.describe('Admin Users Management - Error Handling', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToUsers(page);
	});

	test('Handles network errors gracefully during user creation', async ({ page }) => {
		// Simulate network failure
		await page.route('**/api/admin/users', (route) => route.abort());

		await page.click('button:has-text("Create User")');
		await page.fill('input[name="name"]', 'Network Test User');
		await page.fill('input[name="phone"]', '+15555555555');
		await page.selectOption('select[name="role"]', 'owl');

		await page.click('button[type="submit"]');

		// Should show network error
		await expect(page.locator('text=Network error, text=Failed to')).toBeVisible();
	});

	test('Handles server errors gracefully', async ({ page }) => {
		// Simulate server error
		await page.route('**/api/admin/users', (route) =>
			route.fulfill({ status: 500, body: JSON.stringify({ error: 'Internal server error' }) })
		);

		await page.click('button:has-text("Create User")');
		await page.fill('input[name="name"]', 'Server Error Test');
		await page.fill('input[name="phone"]', '+15555555555');
		await page.selectOption('select[name="role"]', 'owl');

		await page.click('button[type="submit"]');

		// Should show user-friendly error message
		await expect(page.locator('text=Something went wrong')).toBeVisible();
	});

	test('Handles unauthorized access', async ({ page }) => {
		// Simulate unauthorized response
		await page.route('**/api/admin/users', (route) =>
			route.fulfill({ status: 403, body: JSON.stringify({ error: 'Unauthorized' }) })
		);

		await page.goto('/admin/users');

		// Should show unauthorized message or redirect to login
		await expect(page.locator('text=Unauthorized, text=Access denied')).toBeVisible();
	});
});

test.describe('Admin Users Management - Accessibility', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToUsers(page);
	});

	test('Keyboard navigation works throughout user management', async ({ page }) => {
		// Tab through main interface elements
		await page.keyboard.press('Tab');
		await expect(page.locator(':focus')).toBeVisible();

		// Should be able to reach create user button
		await page.keyboard.press('Tab');
		await page.keyboard.press('Tab');
		const focusedElement = await page.locator(':focus').textContent();
		expect(focusedElement).toContain('Create');

		// Open form with keyboard
		await page.keyboard.press('Enter');
		await expect(page.locator('form')).toBeVisible();

		// Navigate form with keyboard
		await page.keyboard.press('Tab');
		await expect(page.locator(':focus')).toHaveAttribute('name', 'name');
	});

	test('Screen reader labels are present', async ({ page }) => {
		await page.click('button:has-text("Create User")');

		// Check for proper labels
		await expect(page.locator('label[for*="name"]')).toBeVisible();
		await expect(page.locator('label[for*="phone"]')).toBeVisible();
		await expect(page.locator('label[for*="role"]')).toBeVisible();

		// Check form has proper heading
		await expect(page.locator('h1, h2')).toContainText('Create, New User');
	});

	test('Error messages are announced properly', async ({ page }) => {
		await page.click('button:has-text("Create User")');

		// Submit empty form
		await page.click('button[type="submit"]');

		// Error messages should have proper ARIA attributes
		await expect(page.locator('[role="alert"], [aria-live="polite"]')).toBeVisible();
	});
});

test.describe('Admin Users Management - Performance', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
	});

	test('Users page loads quickly with many users', async ({ page }) => {
		// Navigate and measure load time
		const startTime = Date.now();
		await page.goto('/admin/users');
		await page.waitForSelector('[data-testid="user-item"], text=No users found');
		const loadTime = Date.now() - startTime;

		// Should load within reasonable time
		expect(loadTime).toBeLessThan(5000); // 5 seconds
	});

	test('Search performs well with large user list', async ({ page }) => {
		await page.goto('/admin/users');

		const searchStartTime = Date.now();
		await page.fill('input[placeholder*="Search"]', 'Test');

		// Wait for search results
		await page.waitForTimeout(500); // Allow for debouncing

		const searchTime = Date.now() - searchStartTime;
		expect(searchTime).toBeLessThan(2000); // 2 seconds
	});
});
