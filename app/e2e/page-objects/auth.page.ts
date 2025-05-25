import { type Page, type Locator, expect } from '@playwright/test';

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

  constructor(page: Page) {
    this.page = page;
    
    // Registration page
    this.registerHeading = page.getByRole('heading', { name: 'Join the Community Watch' });
    this.nameInput = page.getByLabel('Full Name');
    this.phoneInput = page.getByLabel('Phone Number');
    this.createAccountButton = page.getByRole('button', { name: 'Create account' });
    
    // Login page
    this.loginHeading = page.getByRole('heading', { name: 'Enter verification code' });
    this.otpInput = page.locator('[data-input-otp-root]');
    this.sendCodeButton = page.getByRole('button', { name: 'Send verification code' });
    this.verifyButton = page.getByRole('button', { name: 'Verify & Continue' });
    this.goBackButton = page.getByRole('button', { name: 'Wrong phone number? Go back' });
    
    // Homepage
    this.joinUsButton = page.getByRole('button', { name: 'Join Us' });
    
    // Messages
    this.successMessage = page.getByText('Registration successful!').or(page.getByText('Login successful!'));
    this.errorMessage = page.getByText(/verification failed/i).or(page.getByText(/error/i));
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
    await this.page.getByLabel('Phone Number').fill(phone);
  }

  async submitRegistration() {
    await this.page.getByRole('button', { name: 'Create account' }).click();
  }

  async register(name: string, phone: string) {
    await this.fillRegistrationForm(name, phone);
    await this.submitRegistration();
  }

  async fillPhoneForLogin(phone: string) {
    await this.phoneInput.fill(phone);
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
    await this.page.goto('/');
    
    // Set the correct localStorage structure matching UserSessionData interface
    await this.page.evaluate(() => {
      const userSessionData = {
        isAuthenticated: true,
        id: '1', // Must be string, not number
        name: 'Alice Admin',
        phone: '+27821234567',
        role: 'admin', // Must match UserRole type exactly
        token: 'mock-jwt-token'
      };
      localStorage.setItem('user-session', JSON.stringify(userSessionData));
    });
    
    // Wait a moment for the store to update
    await this.page.waitForTimeout(500);
    
    // Navigate to admin and handle potential redirect
    await this.page.goto('/admin');
    await this.page.waitForLoadState('networkidle');
    
    // Verify we stayed on admin (not redirected to login)
    await expect(this.page).toHaveURL('/admin');
  }

  async loginAsVolunteer() {
    await this.page.goto('/');
    
    // Set the correct localStorage structure for volunteer (owl role)
    await this.page.evaluate(() => {
      const userSessionData = {
        isAuthenticated: true,
        id: '2', // Must be string, not number
        name: 'Bob Volunteer',
        phone: '+27821234568',
        role: 'owl', // Must match UserRole type exactly
        token: 'mock-jwt-token'
      };
      localStorage.setItem('user-session', JSON.stringify(userSessionData));
    });
    
    // Wait a moment for the store to update
    await this.page.waitForTimeout(500);
    
    // Navigate to shifts page
    await this.page.goto('/shifts');
    await this.page.waitForLoadState('networkidle');
    
    // Verify we're on shifts page
    await expect(this.page).toHaveURL('/shifts');
  }

  async logout() {
    await this.page.evaluate(() => {
      localStorage.removeItem('user-session');
    });
  }

  async fillLoginForm(phone: string, name?: string) {
    if (name) {
      await this.page.getByLabel('Name (optional)').fill(name);
    }
    await this.page.getByLabel('Phone Number').fill(phone);
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
} 