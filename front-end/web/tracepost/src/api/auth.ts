const endpoint = process.env.NEXT_PUBLIC_API_URL;

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
