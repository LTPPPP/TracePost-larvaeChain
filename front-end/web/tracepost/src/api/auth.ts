const endpoint = process.env.NEXT_PUBLIC_API_URL;

interface RegisterData {
  company_id: string;
  email: string;
  password: string;
  role: string;
  username: string;
}

interface LoginData {
  username: string;
  password: string;
}

// Register
export async function register(data: RegisterData) {
  return await fetch(`${endpoint}/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });
}

// Login
export async function login(username: string, password: string) {
  return await fetch(`${endpoint}/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      username,
      password
    })
  });
}

export async function refreshToken() {
  return await fetch(`${endpoint}/auth/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
      // Add token if needed
      // 'Authorization': `Bearer ${currentToken}`
    }
  });
}

export type { RegisterData, LoginData };
