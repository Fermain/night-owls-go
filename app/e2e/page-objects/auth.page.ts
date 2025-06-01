import { type Page, type Locator, expect } from '@playwright/test';
import { fillPhoneInput } from '../utils/form-helpers';
import { setAuthState, mockUsers, clearAuthState } from '../utils/auth-helpers';

export class AuthPage {
	readonly page: Page;

	// Locators
	readonly registerHeading: Locator;
	readonly loginHeading: Locator;
	readonly nameInput: Locator;
	readonly phoneInput: Locator;
	readonly otpInput: Locator;
	readonly createAccountButton: Locator;
	readonly sendCodeButton: Locator;
	readonly verifyButton: Locator;
	readonly joinUsButton: Locator;
	readonly goBackButton: Locator;
	readonly successMessage: Locator;
	readonly errorMessage: Locator;

	// Form elements (consolidated)
	readonly nameField: Locator;
	readonly otpField: Locator;

	// Buttons - Updated to current text
	readonly loginButton: Locator;
	readonly registerButton: Locator;

	// Links
	readonly loginLink: Locator;
	readonly registerLink: Locator;

	constructor(page: Page) {
		this.page = page;

		// Registration page
		this.registerHeading = page.getByRole('heading', { name: 'Join the Night Owls Control' });
		this.nameInput = page.getByLabel('Full Name');
		this.phoneInput = page.locator('input[type="tel"]');
		this.createAccountButton = page.getByRole('button', { name: 'Create account' });

		// Login page
		this.loginHeading = page.getByRole('heading', { name: 'Enter verification code' });
		this.otpInput = page.locator('[data-input-otp-root]');
		this.sendCodeButton = page.getByRole('button', { name: 'Send verification code' });
		this.verifyButton = page.getByRole('button', { name: 'Verify & Continue' });
		this.goBackButton = page.getByRole('button', { name: 'Wrong phone number? Go back' });

		// Homepage
		this.joinUsButton = page.getByRole('link', { name: /become an owl/i });

		// Messages
		this.successMessage = page
			.getByText('Registration successful!')
			.or(page.getByText('Login successful!'));
		this.errorMessage = page.getByText(/verification failed/i).or(page.getByText(/error/i));

		// Consolidated form elements (no duplicates)
		this.nameField = page.getByLabel('Full Name');
		this.otpField = page.getByPlaceholder(/enter.*code|otp/i);

		// Buttons - Updated to current text
		this.loginButton = page.getByRole('button', { name: /send verification code|sign in/i });
		this.registerButton = page.getByRole('button', { name: /create account/i });

		// Links
		this.loginLink = page.getByRole('link', { name: /sign in/i });
		this.registerLink = page.getByRole('link', { name: /create.*account|register/i });
	}

	async goto() {
		await this.page.goto('/');
	}

	async gotoRegister() {
		await this.page.goto('/register');
	}

	async gotoLogin() {
		await this.page.goto('/login');
	}

	async clickJoinUs() {
		await this.joinUsButton.click();
		await expect(this.page).toHaveURL('/register');
	}

	async navigateToRegister() {
		await this.page.goto('/register');
	}

	async navigateToLogin() {
		await this.page.goto('/login');
	}

	async fillRegistrationForm(name: string, phone: string) {
		await this.page.getByLabel('Full Name').fill(name);
		await fillPhoneInput(this.page, phone);
	}

	async submitRegistration() {
		await this.page.getByRole('button', { name: 'Create account' }).click();
	}

	async register(name: string, phone: string) {
		await this.fillRegistrationForm(name, phone);
		await this.submitRegistration();
	}

	async fillPhoneForLogin(phone: string) {
		await fillPhoneInput(this.page, phone);
		await this.sendCodeButton.click();
		await expect(this.loginHeading).toBeVisible();
	}

	async enterOTP(otp: string) {
		await expect(this.otpInput).toBeVisible();
		for (let i = 0; i < otp.length; i++) {
			await this.page.keyboard.type(otp[i]);
		}
	}

	async submitOTP() {
		await this.verifyButton.click();
	}

	async verifyOTP(otp: string) {
		await this.enterOTP(otp);
		await this.submitOTP();
	}

	async expectSuccessfulLogin() {
		await expect(this.successMessage).toBeVisible();
		await expect(this.page).toHaveURL('/admin');
	}

	async expectLoginError() {
		await expect(this.errorMessage).toBeVisible();
	}

	async expectOTPCleared() {
		const otpInput = this.page.locator('[data-input-otp-root] input').first();
		await expect(otpInput).toHaveValue('');
	}

	async goBack() {
		await this.goBackButton.click();
	}

	async loginAsAdmin() {
		await setAuthState(this.page, mockUsers.admin);
		await this.page.goto('/admin');
		console.log('✅ Logged in as admin via page object');
	}

	async loginAsVolunteer() {
		await setAuthState(this.page, mockUsers.volunteer);
		await this.page.goto('/');
		console.log('✅ Logged in as volunteer via page object');
	}

	async logout() {
		await clearAuthState(this.page);
		await this.page.goto('/');
		console.log('✅ Logged out via page object');
	}

	async fillLoginForm(phone: string, name?: string) {
		if (name) {
			await this.page.getByLabel('Name (optional)').fill(name);
		}
		await fillPhoneInput(this.page, phone);
	}

	async submitLoginForm() {
		await this.page.getByRole('button', { name: 'Send verification code' }).click();
	}

	async fillOTPCode(code: string) {
		const otpInputs = this.page.locator('[data-input-otp] input');

		for (let i = 0; i < code.length && i < 6; i++) {
			await otpInputs.nth(i).fill(code[i]);
		}
	}

	async submitOTPCode() {
		await this.page.getByRole('button', { name: 'Verify & Continue' }).click();
	}

	async waitForSuccessMessage() {
		await this.page.waitForSelector('text=/Registration successful|OTP sent|Login successful/i');
	}

	async waitForRedirectTo(path: string) {
		await this.page.waitForURL(`**${path}`);
	}

	async registerUser(name: string, phone: string) {
		await this.navigateToRegister();
		await this.fillRegistrationForm(name, phone);
		await this.submitRegistration();
		await this.waitForSuccessMessage();
	}

	async loginUser(phone: string, otpCode: string, name?: string) {
		await this.navigateToLogin();
		await this.fillLoginForm(phone, name);
		await this.submitLoginForm();

		await this.page.waitForSelector('text=/Enter verification code/i');

		await this.fillOTPCode(otpCode);
		await this.submitOTPCode();
		await this.waitForRedirectTo('/admin');
	}

	// OTP input helper methods
	async fillOtpInput(otp: string) {
		const otpInputs = this.page.locator('[data-input-otp] input');
		
		for (let i = 0; i < otp.length; i++) {
			await otpInputs.nth(i).fill(otp[i]);
		}
	}

	// Login flow methods
	async login(phone: string, otp: string = '123456') {
		await this.goto();
		await this.joinUsButton.click();
		await fillPhoneInput(this.page, phone);
		await this.createAccountButton.click();
		await this.fillOtpInput(otp);
		await this.verifyButton.click();
	}
}
