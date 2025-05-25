// Browser-side MSW setup for intercepting requests in the browser context
export const setupMSWInBrowser = `
import { setupWorker } from 'msw/browser';

// Import our handlers (we'll need to inline them since we can't import from Node.js context)
const handlers = [
  // Authentication endpoints
  {
    method: 'POST',
    url: '/api/auth/register',
    handler: async (request) => {
      const body = await request.json();
      return Response.json({
        success: true,
        message: 'Registration successful!',
        user: {
          id: Date.now(),
          name: body.name,
          phone: body.phone,
          role: 'guest'
        }
      });
    }
  },
  
  {
    method: 'POST', 
    url: '/api/auth/verify',
    handler: async (request) => {
      const body = await request.json();
      
      if (!/^\\d{6}$/.test(body.otp)) {
        return Response.json(
          { error: 'Invalid OTP format' },
          { status: 400 }
        );
      }

      return Response.json({
        success: true,
        message: 'Login successful!',
        token: 'mock-jwt-token',
        user: {
          id: Date.now(),
          name: 'Test User',
          phone: body.phone,
          role: 'guest'
        }
      });
    }
  }
];

// This will be injected into the browser context
if (typeof window !== 'undefined') {
  console.log('Setting up MSW in browser...');
  // Browser MSW setup would go here
}
`; 