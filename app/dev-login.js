// Development login script for browser console
// Copy and paste this into your browser's console while on the Night Owls app

async function devLogin() {
	try {
		// First register/get a user
		const registerResponse = await fetch('/api/auth/register', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				phone: '+27821234567',
				name: 'Admin User'
			})
		});

		if (!registerResponse.ok) {
			console.log('User might already exist, trying dev login...');
		}

		// Use dev login to get a real JWT token
		const devLoginResponse = await fetch('/api/auth/dev-login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				phone: '+27821234567'
			})
		});

		if (!devLoginResponse.ok) {
			throw new Error(`Dev login failed: ${devLoginResponse.status}`);
		}

		const loginData = await devLoginResponse.json();

		// Set the session in localStorage (this is what the app uses)
		const userSession = {
			isAuthenticated: true,
			id: loginData.user.id.toString(),
			name: loginData.user.name,
			phone: loginData.user.phone,
			role: loginData.user.role,
			token: loginData.token
		};

		localStorage.setItem('user-session', JSON.stringify(userSession));

		console.log(
			'✅ Successfully logged in as:',
			loginData.user.name,
			'(' + loginData.user.role + ')'
		);
		console.log('🔄 Refreshing page...');

		// Refresh the page to apply the authentication
		window.location.reload();
	} catch (error) {
		console.error('❌ Login failed:', error);
	}
}

console.log('🚀 Development login function loaded!');
console.log('�� Run: devLogin()');
